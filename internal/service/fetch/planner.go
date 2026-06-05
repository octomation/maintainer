package fetch

import (
	"fmt"
	"sort"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// DiskClone is a clone discovered on disk by the Adopter, resolved to a stable
// GitHub id (§4.4 pass 3). ID is 0 when the remote could not be resolved.
type DiskClone struct {
	Path      string
	RemoteURL string // canonical, credential-free
	Transport string // ssh|https observed
	Owner     string
	Name      string
	ID        int64
	Origins   int // number of origin URLs (>1 ⇒ conflict)
}

// Occupancy classifies what occupies a would-be target path (§7.5).
type Occupancy int

// Occupancy values.
const (
	OccupancyClear   Occupancy = iota // absent or empty directory
	OccupancyForeign                  // file / non-empty non-git dir / bare / .git file
)

// ConfirmStatus is the outcome of a GET /repositories/{id} re-verification (§10).
type ConfirmStatus int

// Confirmation outcomes.
const (
	ConfirmGone         ConfirmStatus = iota // 404 → orphan
	ConfirmFound                             // 200 → still exists (noop)
	ConfirmInaccessible                      // 401/403 → kept, flagged
	ConfirmLegalHold                         // 451 → kept, flagged
	ConfirmTransient                         // 5xx / network → leave unchanged
)

// Confirmation re-verifies a state id that vanished from discovery.
type Confirmation struct {
	Status   ConfirmStatus
	Snapshot *github.RepoSnapshot // populated when Status == ConfirmFound
}

// PlanInput aggregates every fact the pure Planner needs. Disk and network
// facts are gathered by the service before planning, so the Planner does no
// I/O (§9).
type PlanInput struct {
	Snapshots     []github.RepoSnapshot
	State         *state.State
	Clones        []DiskClone
	Confirmations map[int64]Confirmation
	Occupancy     map[string]Occupancy // keyed by cleaned target path; optional
}

// Planner turns {API, State, Disk} facts into an ordered, conflict-checked
// action list. It is pure and deterministic (§9).
type Planner struct {
	cnf   *config.Fetch
	paths *PathResolver
}

// NewPlanner wires a Planner from config and a path resolver.
func NewPlanner(cnf *config.Fetch, paths *PathResolver) *Planner {
	return &Planner{cnf: cnf, paths: paths}
}

// Plan computes the action list. The result is sorted by apply order then id,
// with target-path collisions resolved up front (§7.4).
func (p *Planner) Plan(in PlanInput) ([]Action, error) {
	snapByID := make(map[int64]github.RepoSnapshot, len(in.Snapshots))
	for _, s := range in.Snapshots {
		snapByID[s.ID] = s
	}
	clonesByID := make(map[int64][]DiskClone)
	clonesByPath := make(map[string]DiskClone)
	for _, c := range in.Clones {
		if c.ID != 0 {
			clonesByID[c.ID] = append(clonesByID[c.ID], c)
		}
		clonesByPath[c.Path] = c
	}

	// The universe of repository ids across API, state and disk.
	ids := unionIDs(in)

	actions := make([]Action, 0, len(ids))
	for _, id := range ids {
		snap, hasSnap := snapByID[id]
		rec, hasRec := in.State.ByID(id)

		owner, name := bestName(snap, hasSnap, rec, hasRec)
		if p.cnf.Ignored(id, owner, name) {
			continue // ignore = true suppresses the whole pipeline (§4.2)
		}

		switch {
		case hasSnap:
			act, err := p.planPresent(snap, rec, hasRec, clonesByID[id], clonesByPath, in.Occupancy)
			if err != nil {
				return nil, err
			}
			if act != nil {
				actions = append(actions, *act)
			}
		case hasRec:
			if act := p.planMissing(*rec, in.Confirmations[id]); act != nil {
				actions = append(actions, *act)
			}
		}
	}

	resolveTargetCollisions(actions)
	sort.SliceStable(actions, func(i, j int) bool {
		if actions[i].order() != actions[j].order() {
			return actions[i].order() < actions[j].order()
		}
		return actions[i].ID < actions[j].ID
	})
	return actions, nil
}

// planPresent handles the API-present rows of the drift table (§10).
func (p *Planner) planPresent(
	snap github.RepoSnapshot,
	rec *state.Record,
	hasRec bool,
	clones []DiskClone,
	clonesByPath map[string]DiskClone,
	occupancy map[string]Occupancy,
) (*Action, error) {
	target, err := p.paths.Resolve(snap)
	if err != nil {
		return nil, fmt.Errorf("resolve path for %s (id=%d): %w", snap.FullName(), snap.ID, err)
	}
	transport := p.cnf.CloneURLFor(snap.SourceProfile, snap.Owner, snap.ID, snap.Name)
	filtered := p.filtered(snap)

	base := Action{
		ID: snap.ID, NodeID: snap.NodeID, Owner: snap.Owner, Name: snap.Name,
		Transport: transport, Profile: snap.SourceProfile, Snapshot: &snap,
	}
	snapCopy := snap

	if !hasRec {
		// API Y, State N — adopt an existing clone, clone fresh, or conflict.
		switch {
		case len(uniquePaths(clones)) > 1:
			a := base
			a.Kind = KindConflict
			a.Reason = "same repository id found at multiple locations on disk"
			a.Path = target
			return &a, nil
		case len(clones) == 1:
			c := clones[0]
			a := base
			a.Kind = KindAdopt
			a.Path = c.Path
			a.RemoteURL = c.RemoteURL
			a.Transport = c.Transport
			a.Record = recordFromSnapshot(snapCopy, c.Path, c.RemoteURL, c.Transport)
			return &a, nil
		default:
			if other, ok := clonesByPath[target]; ok && other.ID != snap.ID {
				a := base
				a.Kind = KindConflict
				a.Path = target
				a.Reason = "rendered path holds a different repository (remote mismatch)"
				return &a, nil
			}
			if occupancy[target] == OccupancyForeign {
				a := base
				a.Kind = KindConflict
				a.Path = target
				a.Reason = "rendered path is occupied by a file or non-Git directory"
				return &a, nil
			}
			if filtered {
				return nil, nil // a new repo excluded by a filter is simply skipped (§4.2)
			}
			a := base
			a.Kind = KindClone
			a.Path = target
			a.RemoteURL = canonicalURL(transport, snap.Owner, snap.Name)
			a.Record = recordFromSnapshot(snapCopy, target, a.RemoteURL, transport)
			return &a, nil
		}
	}

	// API Y, State Y — fetch / move / update_remote / relocate / clone.
	a := base
	a.Record = rec
	a.Transport = rec.CloneURL // observed transport governs update_remote (§7.1)
	if a.Transport == "" {
		a.Transport = transport
	}

	recPresent := cloneAt(clones, rec.Path)
	if !recPresent {
		elsewhere := uniquePathsExcept(clones, rec.Path)
		switch len(elsewhere) {
		case 1:
			a.Kind = KindRelocate
			a.FromPath = rec.Path
			a.ToPath = elsewhere[0]
			a.Path = elsewhere[0]
			return &a, nil
		case 0:
			if filtered {
				a.Kind = KindNoop
				a.Filtered = true
				a.Path = rec.Path
				return &a, nil
			}
			a.Kind = KindClone
			a.Path = target
			a.RemoteURL = canonicalURL(transport, snap.Owner, snap.Name)
			a.Transport = transport
			fresh := recordFromSnapshot(snap, target, a.RemoteURL, transport)
			fresh.FirstSeen = rec.FirstSeen // preserve original first_seen across a re-clone
			a.Record = fresh
			return &a, nil
		default:
			a.Kind = KindConflict
			a.Path = rec.Path
			a.Reason = "same repository id found at multiple locations on disk"
			return &a, nil
		}
	}

	nameChanged := snap.Owner != rec.OwnerLogin || snap.Name != rec.Name
	external := p.paths.External(snap)
	pathDiffers := target != rec.Path
	canonical := canonicalURL(a.Transport, snap.Owner, snap.Name)
	remoteDrift := canonical != rec.RemoteURL

	switch {
	case pathDiffers && !external:
		// Move; the target must be clear (§7.5).
		if other, ok := clonesByPath[target]; ok && other.ID != snap.ID {
			a.Kind = KindConflict
			a.Path = target
			a.Reason = "move target holds a different repository"
			return &a, nil
		}
		if occupancy[target] == OccupancyForeign {
			a.Kind = KindConflict
			a.Path = target
			a.Reason = "move target is occupied by a file or non-Git directory"
			return &a, nil
		}
		a.Kind = KindMove
		a.FromPath = rec.Path
		a.ToPath = target
		a.Path = target
		a.FromName = rec.OwnerLogin + "/" + rec.Name
		a.UpdateRemote = nameChanged || remoteDrift
		if a.UpdateRemote {
			a.RemoteURL = canonical
		}
		return &a, nil
	case remoteDrift:
		a.Kind = KindUpdateRemote
		a.Path = rec.Path
		a.RemoteURL = canonical
		a.UpdateRemote = true
		return &a, nil
	default:
		a.Kind = KindFetch
		a.Path = rec.Path
		a.Filtered = filtered
		return &a, nil
	}
}

// planMissing handles the API-absent rows of the drift table (§10): a tracked
// repository no longer returned by discovery, re-verified by id.
func (p *Planner) planMissing(rec state.Record, c Confirmation) *Action {
	base := Action{
		ID: rec.ID, NodeID: rec.NodeID, Owner: rec.OwnerLogin, Name: rec.Name,
		Path: rec.Path, Record: &rec,
	}
	switch c.Status {
	case ConfirmFound:
		base.Kind = KindNoop
		base.Reason = "still present on GitHub (visibility narrowed)"
		return &base
	case ConfirmInaccessible:
		base.Kind = KindNoop
		base.Flag = FlagInaccessible
		base.Reason = "access lost (401/403); record kept"
		return &base
	case ConfirmLegalHold:
		base.Kind = KindNoop
		base.Flag = FlagLegalHold
		base.Reason = "legal hold (451); record kept"
		return &base
	case ConfirmTransient:
		return nil // leave unchanged; the service reports the transient error
	default: // ConfirmGone (404)
		base.Kind = KindOrphan
		base.Reason = "gone on GitHub (404); local clone retained"
		return &base
	}
}

func (p *Planner) filtered(snap github.RepoSnapshot) bool {
	f := p.cnf.Filters
	return (f.ExcludeArchived && snap.IsArchived) ||
		(f.ExcludeForks && snap.IsFork) ||
		(f.ExcludeTemplates && snap.IsTemplate)
}

// resolveTargetCollisions flags any two materialising actions (clone/move)
// that resolve to the same canonical path as conflicts for both (§7.4).
func resolveTargetCollisions(actions []Action) {
	claims := make(map[string][]int)
	for i, a := range actions {
		var path string
		switch a.Kind {
		case KindClone:
			path = a.Path
		case KindMove:
			path = a.ToPath
		default:
			continue
		}
		claims[path] = append(claims[path], i)
	}
	for path, idx := range claims {
		if len(idx) < 2 {
			continue
		}
		for _, i := range idx {
			actions[i].Kind = KindConflict
			actions[i].Reason = fmt.Sprintf("target path %q claimed by %d repositories", path, len(idx))
		}
	}
}

func unionIDs(in PlanInput) []int64 {
	seen := make(map[int64]bool)
	var ids []int64
	add := func(id int64) {
		if id != 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, s := range in.Snapshots {
		add(s.ID)
	}
	for i := range in.State.Repos {
		add(in.State.Repos[i].ID)
	}
	for _, c := range in.Clones {
		add(c.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func bestName(snap github.RepoSnapshot, hasSnap bool, rec *state.Record, hasRec bool) (owner, name string) {
	if hasSnap {
		return snap.Owner, snap.Name
	}
	if hasRec {
		return rec.OwnerLogin, rec.Name
	}
	return "", ""
}

func cloneAt(clones []DiskClone, path string) bool {
	for _, c := range clones {
		if c.Path == path {
			return true
		}
	}
	return false
}

func uniquePaths(clones []DiskClone) []string {
	seen := make(map[string]bool)
	var out []string
	for _, c := range clones {
		if !seen[c.Path] {
			seen[c.Path] = true
			out = append(out, c.Path)
		}
	}
	return out
}

func uniquePathsExcept(clones []DiskClone, except string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, c := range clones {
		if c.Path == except || seen[c.Path] {
			continue
		}
		seen[c.Path] = true
		out = append(out, c.Path)
	}
	return out
}

func canonicalURL(transport, owner, name string) string {
	if transport == config.TransportSSH {
		return fmt.Sprintf("git@github.com:%s/%s.git", owner, name)
	}
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, name)
}

func recordFromSnapshot(snap github.RepoSnapshot, path, remote, transport string) *state.Record {
	return &state.Record{
		ID:               snap.ID,
		NodeID:           snap.NodeID,
		OwnerLogin:       snap.Owner,
		Name:             snap.Name,
		Visibility:       string(snap.Visibility),
		Path:             path,
		RemoteURL:        remote,
		CloneURL:         transport,
		SourceProfile:    snap.SourceProfile,
		DefaultBranch:    snap.DefaultBranch,
		IsFork:           snap.IsFork,
		IsTemplate:       snap.IsTemplate,
		ArchivedOnGitHub: snap.IsArchived,
	}
}
