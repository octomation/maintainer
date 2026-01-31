package fetch

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/service/workspace"
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
	Origins   int  // number of origin fetch URLs (>1 ⇒ conflict)
	Pinned    bool // discovered through workspace.pins or an explicit path
	Error     string
	Worktrees []string // linked worktrees of the checkout

	// former is set on a disputed checkout at a previous path of the recorded
	// id: its conflict leaves the recorded id's active checkout planned.
	former bool
	// rebound is set when the recorded id was kept although the origin's name
	// belongs to another repository now: fetching through it is unsafe.
	rebound bool
}

// Occupancy classifies what occupies a would-be target path (§7.5).
type Occupancy int

// Occupancy values.
const (
	OccupancyClear   Occupancy = iota // path is absent
	OccupancyForeign                  // any existing path not verified as the expected clone
	OccupancyUnsafe                   // a parent inside root is a symlink or not a directory
	OccupancyUnknown                  // the path could not be inspected
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
	ConfirmUnverified                        // 404/401/403 from another profile's token → leave unchanged
)

// Confirmation re-verifies a state id that vanished from discovery.
type Confirmation struct {
	Status   ConfirmStatus
	Snapshot *github.RepoSnapshot // populated when Status == ConfirmFound
	Reason   string               // why an unverified record stays as it is
}

// PlanInput aggregates every fact the pure Planner needs. Disk and network
// facts are gathered by the service before planning, so the Planner does no
// I/O (§9).
type PlanInput struct {
	KnownCheckouts map[string]bool // existing exact historical paths; reporting only
	Snapshots      []github.RepoSnapshot
	State          *state.State
	Clones         []DiskClone
	Confirmations  map[int64]Confirmation
	Occupancy      map[string]Occupancy // keyed by cleaned target path; optional
	Deselected     map[int64]string     // outside --owner/--profile: id → reason; report-only
	// Aliases maps a rendered target to the recorded checkout path it already
	// denotes: the spellings differ only in letter case, and the filesystem
	// confirms one directory (os.SameFile). Such a target needs no move.
	Aliases map[string]string
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
	if _, err := p.paths.Scope(); err != nil {
		return nil, err
	}
	snapByID := make(map[int64]github.RepoSnapshot, len(in.Snapshots))
	for _, s := range in.Snapshots {
		snapByID[s.ID] = s
	}
	// Transfers can leave the selected owners. A successful ID confirmation
	// still supplies authoritative metadata for reconciliation.
	for _, snap := range confirmedRenames(in.State, in.Confirmations) {
		if _, exists := snapByID[snap.ID]; !exists {
			snapByID[snap.ID] = snap
		}
	}
	in.Clones = attribute(in.State, in.Clones)
	clonesByID := make(map[int64][]DiskClone)
	clonesByPath := make(map[string]DiskClone)
	disputed := make(map[string]bool)
	for _, c := range in.Clones {
		if c.ID != 0 && c.Error == "" {
			clonesByID[c.ID] = append(clonesByID[c.ID], c)
		}
		clonesByPath[c.Path] = c
	}

	// The universe of repository ids across API, state and disk.
	ids := unionIDs(in)

	actions := make([]Action, 0, len(ids))
	blocked := map[int64]bool{}
	for _, c := range in.Clones {
		if c.Error != "" {
			if p.cnf.Ignored(c.ID, c.Owner, c.Name) {
				continue
			}
			// A checkout bound to an id outside the selection is left alone,
			// as that repository is: its row is the noop below.
			if _, outside := in.Deselected[c.ID]; c.ID != 0 && outside {
				continue
			}
			if !c.former {
				blocked[c.ID] = true
			}
			disputed[c.Path] = true
			actions = append(actions, Action{Kind: KindConflict, ID: c.ID, Owner: c.Owner, Name: c.Name, Path: c.Path, Reason: c.Error})
		}
	}
	for _, id := range ids {
		if blocked[id] {
			continue
		}
		snap, hasSnap := snapByID[id]
		rec, hasRec := in.State.ByID(id)

		owner, name := bestName(snap, hasSnap, rec, hasRec)
		if p.cnf.Ignored(id, owner, name) {
			continue // ignore = true suppresses the whole pipeline (§4.2)
		}
		if reason, outside := in.Deselected[id]; outside {
			actions = append(actions, p.deselected(id, owner, name, snap, rec, reason))
			continue
		}
		if hasRec && !p.paths.Authorised(*rec, snap) {
			actions = append(actions, orphanAction(id, owner, name, rec.Path, workspace.OrphanOutOfScope, ""))
			continue
		}
		// A specific pin selects mutation authority, not visibility. Other
		// checkouts remain visible even when the selected pin is missing.
		pinSnap := snap
		if !hasSnap && hasRec {
			pinSnap = github.RepoSnapshot{ID: id, Owner: rec.OwnerLogin, Name: rec.Name}
		}
		pin, err := p.paths.PinnedPath(pinSnap, rec)
		if err != nil {
			actions = append(actions, Action{Kind: KindConflict, ID: id, Owner: owner, Name: name, Reason: err.Error()})
			continue
		}
		if pin != "" && !p.paths.WorkspacePin(pinSnap, rec) {
			seen := map[string]bool{pin: true}
			addOrphan := func(path string) {
				if path == "" || seen[path] || disputed[path] {
					return
				}
				seen[path] = true
				actions = append(actions, orphanAction(id, owner, name, path, workspace.OrphanDuplicatePin, pin))
			}
			for _, c := range clonesByID[id] {
				addOrphan(c.Path)
			}
			if hasRec {
				for _, path := range append(append([]string(nil), rec.PreviousPaths...), rec.Path) {
					if in.KnownCheckouts[path] {
						addOrphan(path)
					}
				}
			}
		}

		switch {
		case hasSnap:
			act, err := p.planPresent(snap, rec, hasRec, clonesByID[id], clonesByPath, in.Occupancy, in.Aliases)
			if err != nil {
				return nil, err
			}
			if act != nil {
				actions = append(actions, *act)
			}
		case hasRec:
			c, confirmed := in.Confirmations[id]
			if !confirmed {
				continue
			} // absence of evidence is not a 404
			if act := p.planMissing(*rec, c); act != nil {
				if pin != "" {
					act.Path = pin
				}
				actions = append(actions, *act)
			}
		}
	}

	existing := make([]string, 0, len(in.Clones)+len(in.KnownCheckouts))
	for _, c := range in.Clones {
		existing = append(existing, c.Path)
	}
	for path := range in.KnownCheckouts {
		existing = append(existing, path)
	}
	resolveTargetCollisions(actions, p.paths.scope.Root, existing)
	actions = withRefreshes(actions)
	sort.SliceStable(actions, func(i, j int) bool {
		if actions[i].order() != actions[j].order() {
			return actions[i].order() < actions[j].order()
		}
		if actions[i].ID != actions[j].ID {
			return actions[i].ID < actions[j].ID
		}
		return actions[i].Path < actions[j].Path
	})
	return actions, nil
}

// deselected is the report-only row of a repository this invocation leaves
// alone: no request, no mutation, no new observation.
func (p *Planner) deselected(id int64, owner, name string, snap github.RepoSnapshot, rec *state.Record, reason string) Action {
	a := Action{Kind: KindNoop, ID: id, Owner: owner, Name: name, Reason: reason, Deselected: true}
	if rec != nil {
		a.Path = rec.Path
	} else if pin, err := p.paths.PinnedPath(snap, nil); err == nil {
		a.Path = pin
	}
	return a
}

// attribute keeps a checkout bound to the id the state records for its path.
// The name a checkout's origin resolves to may belong to another repository by
// now: a rename freed it, and a new repository took it. The recorded id wins
// when the recorded remote is still the checkout's origin; otherwise the
// checkout is a conflict that names both ids, never an adoption (§4.4).
func attribute(st *state.State, clones []DiskClone) []DiskClone {
	out := make([]DiskClone, 0, len(clones))
	for _, c := range clones {
		if c.ID != 0 && c.Error == "" {
			if rec, former := recordedAt(st, c.Path); rec != nil && rec.ID != c.ID {
				if sameRepository(rec.RemoteURL, c.RemoteURL) {
					c.ID, c.rebound = rec.ID, true
				} else {
					c.Error = fmt.Sprintf("checkout is recorded as id=%d (%s/%s), but its origin %s/%s belongs to id=%d now",
						rec.ID, rec.OwnerLogin, rec.Name, c.Owner, c.Name, c.ID)
					c.ID, c.former = rec.ID, former
				}
			}
		}
		out = append(out, c)
	}
	return out
}

// recordedAt finds the record that owns path as its active path or, failing
// that, as one of its previous paths (former).
func recordedAt(st *state.State, path string) (rec *state.Record, former bool) {
	if st == nil {
		return nil, false
	}
	for i := range st.Repos {
		if st.Repos[i].Path == path {
			return &st.Repos[i], false
		}
	}
	for i := range st.Repos {
		for _, previous := range st.Repos[i].PreviousPaths {
			if previous == path {
				return &st.Repos[i], true
			}
		}
	}
	return nil, false
}

// sameRepository compares two GitHub remotes by host, owner and name, which
// GitHub treats case-insensitively; the transport does not matter.
func sameRepository(a, b string) bool {
	ah, ao, an, _, aok := parseRemote(a)
	bh, bo, bn, _, bok := parseRemote(b)
	return aok && bok && strings.EqualFold(ah, bh) && strings.EqualFold(ao, bo) && strings.EqualFold(an, bn)
}

func orphanAction(id int64, owner, name, path, reason, active string) Action {
	return Action{Kind: KindOrphan, ID: id, Owner: owner, Name: name, Path: path,
		OrphanReason: reason, ActivePath: active, Reason: workspace.OrphanDescription(reason, active)}
}

// Confirmation of an unchanged repository outside discovery remains a noop.
// Only a rename/transfer needs to re-enter reconciliation from confirmation.
func confirmedRenames(st *state.State, confirmations map[int64]Confirmation) []github.RepoSnapshot {
	var snapshots []github.RepoSnapshot
	for id, c := range confirmations {
		if c.Status != ConfirmFound || c.Snapshot == nil || c.Snapshot.ID != id {
			continue
		}
		if rec, ok := st.ByID(id); ok && (rec.OwnerLogin != c.Snapshot.Owner || rec.Name != c.Snapshot.Name) {
			snapshots = append(snapshots, *c.Snapshot)
		}
	}
	return snapshots
}

// planPresent handles the API-present rows of the drift table (§10).
func (p *Planner) planPresent(
	snap github.RepoSnapshot,
	rec *state.Record,
	hasRec bool,
	clones []DiskClone,
	clonesByPath map[string]DiskClone,
	occupancy map[string]Occupancy,
	aliases map[string]string,
) (*Action, error) {
	transport := p.cnf.CloneURLFor(snap.SourceProfile, snap.Owner, snap.ID, snap.Name)
	filtered := p.filtered(snap)

	base := Action{
		ID: snap.ID, NodeID: snap.NodeID, Owner: snap.Owner, Name: snap.Name,
		Transport: transport, Profile: snap.SourceProfile, Snapshot: &snap,
	}
	pinned, err := p.paths.PinnedPath(snap, rec)
	if err != nil {
		base.Kind, base.Reason = KindConflict, err.Error()
		return &base, nil
	}
	if pinned != "" {
		if filtered && !hasRec {
			return notAdopted(base, pinned), nil
		}
		a := planPinned(base, pinned, rec, clones)
		if p.paths.WorkspacePin(snap, rec) {
			if len(uniquePaths(clones)) > 1 {
				a.Kind, a.Reason = KindConflict, "workspace pin does not select between duplicate checkouts"
			}
			if a.Record != nil {
				a.Record.PinSource = "workspace"
			}
		}
		return a, nil
	}
	for _, c := range clones {
		if p.paths.scope.Pinned(c.Path) {
			if filtered && !hasRec {
				return notAdopted(base, c.Path), nil
			}
			if len(uniquePaths(clones)) > 1 {
				base.Kind, base.Path, base.Reason = KindConflict, c.Path, "workspace pin does not select between duplicate checkouts"
				return &base, nil
			}
			a := planPinned(base, c.Path, rec, clones)
			if a.Record != nil {
				a.Record.PinSource = "workspace"
			}
			return a, nil
		}
	}
	// Explicit pins have already been handled. Ordinary disk facts must not
	// turn an arbitrary path into an adopted checkout, even in a pure plan.
	var managed []DiskClone
	for _, c := range clones {
		if p.paths.scope.Managed(c.Path) {
			managed = append(managed, c)
		}
	}
	clones = managed
	if len(uniquePaths(clones)) > 1 {
		base.Kind, base.Path, base.Reason = KindConflict, clones[0].Path, "same repository id found at multiple locations on disk"
		return &base, nil
	}
	// One repository whose path cannot be rendered never blocks the others.
	target, err := p.paths.Resolve(snap)
	if err != nil {
		base.Kind, base.Reason = KindConflict, "cannot resolve the target path: "+err.Error()
		return &base, nil
	}
	if p.paths.scope.Pinned(target) {
		base.Kind, base.Path, base.Reason = KindConflict, target, "target is inside workspace.pins; only existing pinned checkouts are allowed"
		return &base, nil
	}
	snapCopy := snap

	if !hasRec {
		// API Y, State N — adopt an existing clone, clone fresh, or conflict.
		switch {
		case len(clones) == 1 && filtered:
			return notAdopted(base, clones[0].Path), nil
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
			if a := conflictAtTarget(base, target, snap.ID, clonesByPath, occupancy); a != nil {
				return a, nil
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

	diskClone, recPresent := cloneAt(clones, rec.Path)
	if !recPresent {
		elsewhere := uniquePathsExcept(clones, rec.Path)
		switch len(elsewhere) {
		case 1:
			a.Kind = KindRelocate
			a.FromPath = rec.Path
			a.ToPath = elsewhere[0]
			a.Path = elsewhere[0]
			found, _ := cloneAt(clones, elsewhere[0])
			a.staleOrigin = found.rebound
			return &a, nil
		case 0:
			if filtered {
				a.Kind = KindNoop
				a.Filtered = true
				a.Path = rec.Path
				return &a, nil
			}
			if conflict := conflictAtTarget(base, target, snap.ID, clonesByPath, occupancy); conflict != nil {
				return conflict, nil
			}
			a.Kind = KindClone
			a.Path = target
			a.RemoteURL = canonicalURL(transport, snap.Owner, snap.Name)
			a.Transport = transport
			fresh := recordFromSnapshot(snap, target, a.RemoteURL, transport)
			fresh.FirstSeen = rec.FirstSeen // preserve original first_seen across a re-clone
			fresh.PreviousPaths = append([]string(nil), rec.PreviousPaths...)
			a.Record = fresh
			return &a, nil
		default:
			a.Kind = KindConflict
			a.Path = rec.Path
			a.Reason = "same repository id found at multiple locations on disk"
			return &a, nil
		}
	}
	if diskClone.Transport != "" {
		a.Transport = diskClone.Transport
	}

	nameChanged := snap.Owner != rec.OwnerLogin || snap.Name != rec.Name
	external := p.paths.External(snap)
	// A case-only rename on a case-insensitive filesystem is settled when the
	// target already is the recorded checkout: only its remote follows.
	pathDiffers := target != rec.Path && aliases[target] != rec.Path
	canonical := canonicalURL(a.Transport, snap.Owner, snap.Name)
	actualRemote := rec.RemoteURL
	if diskClone.RemoteURL != "" {
		actualRemote = diskClone.RemoteURL
	}
	remoteDrift := canonical != actualRemote

	switch {
	case pathDiffers && !external:
		// Move; the target must be clear (§7.5).
		if conflict := conflictAtTarget(base, target, snap.ID, clonesByPath, occupancy); conflict != nil {
			return conflict, nil
		}
		a.Kind = KindMove
		a.FromPath = rec.Path
		a.ToPath = target
		a.Path = target
		a.FromName = rec.OwnerLogin + "/" + rec.Name
		a.Worktrees = diskClone.Worktrees
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

// notAdopted reports an untracked checkout of a filtered repository: filters
// skip adoptions as they skip clones (§4.2), on every path.
func notAdopted(base Action, path string) *Action {
	base.Kind = KindNoop
	base.Path = path
	base.Filtered = true
	base.Reason = "excluded by [filters]; the checkout is not adopted"
	return &base
}

func conflictAtTarget(
	base Action,
	target string,
	expectedID int64,
	clonesByPath map[string]DiskClone,
	occupancy map[string]Occupancy,
) *Action {
	if other, ok := clonesByPath[target]; ok && other.ID != expectedID {
		a := base
		a.Kind = KindConflict
		a.Path, a.target = target, target
		switch {
		case other.Error != "":
			a.Reason = other.Error
		case other.ID == 0:
			a.Reason = "target contains a Git repository whose fetch origin cannot be matched unambiguously"
		default:
			a.Reason = "target contains a different Git repository"
		}
		return &a
	}
	reason := map[Occupancy]string{
		OccupancyForeign: "target path already exists and is not the expected repository; refusing to overwrite it",
		OccupancyUnsafe:  "a parent of the target path inside the workspace root is a symbolic link or not a directory; refusing to write through it",
		OccupancyUnknown: "the target path cannot be inspected; refusing to write into it",
	}[occupancy[target]]
	if reason == "" {
		return nil
	}
	a := base
	a.Kind = KindConflict
	a.Path, a.target = target, target
	a.Reason = reason
	return &a
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
	case ConfirmUnverified:
		base.Kind = KindNoop
		base.Flag = FlagUnverified
		base.Reason = c.Reason
		return &base
	default: // ConfirmGone (404)
		base.Kind = KindOrphan
		base.OrphanReason = workspace.OrphanRemoteGone
		base.Reason = workspace.OrphanDescription(base.OrphanReason, "")
		return &base
	}
}

func (p *Planner) filtered(snap github.RepoSnapshot) bool {
	f := p.cnf.Filters
	return (f.ExcludeArchived && snap.IsArchived) ||
		(f.ExcludeForks && snap.IsFork) ||
		(f.ExcludeTemplates && snap.IsTemplate)
}

// resolveTargetCollisions flags materialising actions (clone/move) whose
// targets coincide or nest, with each other or with an existing checkout
// inside root, as conflicts for every participant, before any side effect
// (§7.4). A target already refused, such as an occupied one, still takes
// part, so the checkout it overlaps becomes a conflict too. A checkout at root
// or above it contains the whole workspace and is not a collision.
func resolveTargetCollisions(actions []Action, root string, existing []string) {
	type claim struct {
		index int
		path  string
	}
	var claims []claim
	for i, a := range actions {
		switch {
		case a.Kind == KindClone:
			claims = append(claims, claim{i, a.Path})
		case a.Kind == KindMove:
			claims = append(claims, claim{i, a.ToPath})
		case a.Kind == KindConflict && a.target != "":
			claims = append(claims, claim{i, a.target})
		}
	}
	reasons := make(map[int]string)
	conflict := func(i int, reason string) {
		if _, set := reasons[i]; !set {
			reasons[i] = reason
		}
	}
	byPath := make(map[string][]int)
	for _, c := range claims {
		byPath[c.path] = append(byPath[c.path], c.index)
	}
	for path, idx := range byPath {
		if len(idx) < 2 {
			continue
		}
		for _, i := range idx {
			conflict(i, fmt.Sprintf("target path %q claimed by %d repositories", path, len(idx)))
		}
	}
	nested := func(a, b string) bool { return a != b && (workspace.Within(a, b) || workspace.Within(b, a)) }
	name := func(a Action) string { return a.Owner + "/" + a.Name }
	for x := range claims {
		for _, y := range claims[x+1:] {
			if nested(claims[x].path, y.path) {
				conflict(claims[x].index, fmt.Sprintf("target path overlaps the target path %q of %s", y.path, name(actions[y.index])))
				conflict(y.index, fmt.Sprintf("target path overlaps the target path %q of %s", claims[x].path, name(actions[claims[x].index])))
			}
		}
	}
	existing = slices.Clone(existing)
	slices.Sort(existing)
	for _, path := range slices.Compact(existing) {
		if path == root || !workspace.Within(root, path) {
			continue
		}
		for _, c := range claims {
			if !nested(c.path, path) {
				continue
			}
			conflict(c.index, fmt.Sprintf("target path overlaps the existing checkout %q", path))
			for i, a := range actions {
				location := a.Path
				if a.Kind == KindMove {
					location = a.FromPath
				}
				if i != c.index && a.Executable() && location == path {
					conflict(i, fmt.Sprintf("checkout overlaps the target path %q of %s", c.path, name(actions[c.index])))
				}
			}
		}
	}
	for i, reason := range reasons {
		if actions[i].Kind != KindConflict { // an earlier reason stays
			actions[i].Kind, actions[i].Reason = KindConflict, reason
		}
	}
}

// withRefreshes follows every reconciliation of an existing checkout with one
// fetch of its refs in the same apply. The fetch runs with the routine ones,
// last; the applier skips it when the reconciliation fails (§7.4). A checkout
// whose origin names another repository now is fetched only after its
// update_remote, never through that origin.
func withRefreshes(actions []Action) []Action {
	for _, a := range actions {
		if !a.reconciles() || a.staleOrigin {
			continue
		}
		actions = append(actions, Action{
			Kind: KindFetch, ID: a.ID, NodeID: a.NodeID, Owner: a.Owner, Name: a.Name,
			Path: a.Path, Transport: a.Transport, Profile: a.Profile,
			After: a.Kind, Snapshot: a.Snapshot,
		})
	}
	return actions
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

func cloneAt(clones []DiskClone, path string) (DiskClone, bool) {
	for _, c := range clones {
		if c.Path == path {
			return c, true
		}
	}
	return DiskClone{}, false
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
