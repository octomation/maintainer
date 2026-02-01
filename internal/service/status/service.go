package status

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/config"
	fetchsvc "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/service/workspace"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// Discovery is what Collect knows before any Git inspection: the candidate
// rows with their paths, pins and diagnostics, and the placeholders. Rows
// finishes it without Git, Inspect with it.
type Discovery struct {
	cnf           *config.Fetch
	home          string
	root          string
	owners        []string
	rows          []Row
	extra         []Row
	listed        map[string]bool
	byName        map[string]state.Record
	pins          map[string]string
	explicitPaths map[string]bool
	obstructed    map[string]string
}

// Collect combines disk discovery, remembered checkouts and explicit pins,
// and inspects every checkout with Git. State is only a snapshot loaded by
// the caller and is never mutated.
func Collect(ctx context.Context, cnf *config.Fetch, st *state.State, home, cwd string, owners []string, concurrency int) ([]Row, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("concurrency must be positive")
	}
	d, err := Discover(ctx, cnf, st, home, cwd, owners)
	if err != nil {
		return nil, err
	}
	return d.Inspect(ctx, concurrency)
}

// Discover combines disk discovery, remembered checkouts and explicit pins
// without running Git. State is only a snapshot loaded by the caller and is
// never mutated.
func Discover(ctx context.Context, cnf *config.Fetch, st *state.State, home, cwd string, owners []string) (*Discovery, error) {
	renderer, err := fetchsvc.NewPathRenderer(cnf.WorkspaceConfig().Root, home, cwd)
	if err != nil {
		return nil, err
	}
	paths := fetchsvc.NewPathResolver(cnf, renderer)
	scope, err := paths.Scope()
	if err != nil {
		return nil, err
	}
	byPath := map[string]Row{}
	byName := map[string]state.Record{}
	pins := map[string]string{}
	pinnedPaths := map[string]bool{}
	explicitPaths := map[string]bool{}
	orphans := map[string]string{}
	activePaths := map[string]string{}
	add := func(path string, rec state.Record) {
		path = filepath.Clean(path)
		if _, exists := byPath[path]; exists && rec.ID == 0 {
			return
		}
		name := filepath.Base(path)
		if rec.OwnerLogin != "" {
			name = rec.OwnerLogin + "/" + rec.Name
		}
		byPath[path] = Row{ID: rec.ID, Repository: name, Path: path, DefaultBranch: rec.DefaultBranch, Visibility: rec.Visibility}
		if rec.RemoteStatus == "gone" {
			orphans[path] = workspace.OrphanRemoteGone
			row := byPath[path]
			row.RemoteCheckedAt = rec.RemoteCheckedAt
			byPath[path] = row
		}
	}
	for _, rec := range st.Repos {
		byName[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)] = rec
		if cnf.Ignored(rec.ID, rec.OwnerLogin, rec.Name) {
			continue
		}
		snap := snapshot(rec)
		pin, err := paths.PinnedPath(snap, &rec)
		if err != nil {
			continue // the rule below reports it
		}
		active := rec.Path
		if pin != "" {
			active = pin
		}
		for _, old := range append(append([]string(nil), rec.PreviousPaths...), rec.Path) {
			if old == "" || old == active {
				continue
			}
			if _, err := os.Stat(filepath.Join(old, ".git")); !os.IsNotExist(err) {
				add(old, rec)
				orphans[old], activePaths[old] = workspace.OrphanDuplicatePin, active
			}
		}
		if !paths.Authorised(rec, snap) {
			add(rec.Path, rec)
			orphans[filepath.Clean(rec.Path)] = workspace.OrphanOutOfScope
			continue
		}
		if pin != "" {
			pinnedPaths[pin] = true
			if !paths.WorkspacePin(snap, &rec) {
				pins[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)] = pin
				explicitPaths[pin] = true
			}
			add(pin, rec)
		} else if rec.Path != "" {
			add(rec.Path, rec)
		}
	}
	// Rules without state still support literal paths and owner/name templates.
	// A rule that references other unknown fields is an unresolved pin, and a
	// rule that cannot be rendered is an error of its repository: both are
	// shown, but nothing is inspected in their place.
	var extra []Row
	for _, rule := range cnf.Repos {
		if rule.Ignore || rule.Path == "" {
			continue
		}
		rec := state.Record{ID: rule.Match.ID, OwnerLogin: rule.Match.Owner, Name: rule.Match.Name}
		if known, ok := st.ByID(rec.ID); rec.ID != 0 && ok {
			rec = *known
		} else if known, ok := byName[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)]; rec.ID == 0 && ok {
			rec = known
		}
		pin, err := paths.PinnedPath(snapshot(rec), &rec)
		if err != nil {
			if applicable, _ := cnf.RepoOverride(rec.ID, rec.OwnerLogin, rec.Name); applicable != rule {
				continue // the applicable rule reports it
			}
			name := fmt.Sprintf("id=%d", rec.ID)
			if rec.OwnerLogin != "" {
				name = rec.OwnerLogin + "/" + rec.Name
			}
			row := Row{ID: rec.ID, Repository: name, Path: rule.Path, Status: StatusUnresolvedPin, Pinned: true, displayPath: rule.Path, placeholder: true}
			if !errors.Is(err, fetchsvc.ErrUnresolvedPin) {
				row.Status, row.Error = "error", err.Error()
			}
			extra = append(extra, row)
			continue
		}
		add(pin, rec)
		pinnedPaths[pin] = true
		explicitPaths[pin] = true
		if rec.OwnerLogin != "" {
			pins[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)] = pin
		}
	}
	// Obstructions are shown, never inspected: the scan is incomplete there.
	// A remembered checkout at such a path keeps its row, marked as the error.
	obstructed := map[string]string{}
	err = scope.Walk(ctx, func(path string, pinned bool) error {
		add(path, state.Record{})
		if pinned {
			pinnedPaths[path] = true
		}
		return nil
	}, func(path string, reason error) error {
		obstructed[path] = "scan: " + reason.Error()
		extra = append(extra, Row{
			Repository: filepath.Base(path), Path: path, Status: "error", Error: obstructed[path],
			displayPath: relativePath(path, renderer.Root(), home), placeholder: true,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", renderer.Root(), err)
	}
	rows := make([]Row, 0, len(byPath))
	for _, row := range byPath {
		row.displayPath = relativePath(row.Path, renderer.Root(), home)
		row.Pinned = pinnedPaths[row.Path]
		row.OrphanReason, row.ActivePath = orphans[row.Path], activePaths[row.Path]
		if row.OrphanReason == workspace.OrphanDuplicatePin || row.OrphanReason == workspace.OrphanOutOfScope {
			row.Pinned = false
		}
		rows = append(rows, row)
	}
	// The map order is random: a stable order keeps Rows and Inspect alike.
	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	listed := make(map[string]bool, len(byPath))
	for path := range byPath {
		listed[path] = true
	}
	return &Discovery{cnf: cnf, home: home, root: renderer.Root(), owners: owners, rows: rows, extra: extra,
		listed: listed, byName: byName, pins: pins, explicitPaths: explicitPaths, obstructed: obstructed}, nil
}

// Root is the workspace root the rows were discovered under.
func (d *Discovery) Root() string { return d.root }

// Rows returns the rows without inspecting Git: identity comes from state
// only, and the Git columns stay empty. The order and the diagnostics follow
// the same rules as Inspect.
func (d *Discovery) Rows() []Row {
	return d.finish(slices.Clone(d.rows))
}

// Inspect reads every checkout with Git, at most concurrency at a time, and
// returns the finished rows.
func (d *Discovery) Inspect(ctx context.Context, concurrency int) ([]Row, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("concurrency must be positive")
	}
	rows := slices.Clone(d.rows)
	var group errgroup.Group
	group.SetLimit(concurrency)
	for i := range rows {
		group.Go(func() error {
			expected := rows[i].Repository
			rows[i] = Inspect(ctx, rows[i])
			if rows[i].ID != 0 && strings.Contains(expected, "/") && !strings.EqualFold(expected, rows[i].Repository) {
				rows[i].Error = fmt.Sprintf("origin is %s; expected %s from state (run fetch to verify identity)", rows[i].Repository, expected)
				rows[i].Repository, rows[i].Status = expected, "error"
			}
			return nil
		})
	}
	_ = group.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return d.finish(rows), nil
}

// finish applies the rules that need every row: learned pins, duplicates,
// ignores, obstructions, the owner filter, the placeholders and the order.
func (d *Discovery) finish(rows []Row) []Row {
	cnf, owners, byName, explicitPaths, obstructed := d.cnf, d.owners, d.byName, d.explicitPaths, d.obstructed
	pins := maps.Clone(d.pins)
	// A literal ID-only pin can learn its display identity from its origin.
	for _, row := range rows {
		if explicitPaths[row.Path] && row.identified() {
			pins[strings.ToLower(row.Repository)] = row.Path
		}
	}
	// Broad workspace pins do not silently pick one of several checkouts. Only
	// owner/name identities collide: a checkout named after its directory is
	// keyed by its path, and an orphan row is a copy that is already reported.
	duplicates := map[string]int{}
	for _, row := range rows {
		if row.identified() && row.OrphanReason == "" {
			duplicates[strings.ToLower(row.Repository)]++
		}
	}
	selected := make([]Row, 0, len(rows))
	for _, row := range rows {
		key := strings.ToLower(row.Repository)
		if rec, ok := byName[key]; ok && row.ID == 0 {
			row.ID = rec.ID
			if row.DefaultBranch == "" {
				row.DefaultBranch = rec.DefaultBranch
			}
		}
		owner, name, _ := strings.Cut(row.Repository, "/")
		if cnf.Ignored(row.ID, owner, name) {
			continue
		}
		if pin, ok := pins[key]; ok {
			if row.Path != pin {
				row.OrphanReason, row.ActivePath, row.Pinned = workspace.OrphanDuplicatePin, pin, false
			} else {
				row.Pinned = true
			}
		}
		if duplicates[key] > 1 && pins[key] == "" && row.OrphanReason == "" {
			row.Status, row.Error = "error", "multiple checkouts for this repository; select one with a per-repo path"
		}
		if reason, ok := obstructed[row.Path]; ok {
			row.Status, row.Error = "error", reason // the scan reason outranks the others
		}
		if !selectedOwner(owners, owner) {
			continue
		}
		selected = append(selected, row)
	}
	for _, row := range d.extra {
		if d.listed[row.Path] {
			continue
		}
		// A path without a known owner is kept under every owner filter.
		if owner, _, known := strings.Cut(row.Repository, "/"); known && !selectedOwner(owners, owner) {
			continue
		}
		selected = append(selected, row)
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Repository != selected[j].Repository {
			return selected[i].Repository < selected[j].Repository
		}
		return selected[i].Path < selected[j].Path
	})
	return selected
}

// identified reports whether origin or state names the row owner/name; the
// fallback directory name never contains a separator.
func (row Row) identified() bool {
	owner, name, ok := strings.Cut(row.Repository, "/")
	return ok && owner != "" && name != ""
}

// StatusUnresolvedPin marks a per-repo path that cannot be rendered offline:
// it references a field that only a state record supplies.
const StatusUnresolvedPin = "unresolved pin"

func selectedOwner(owners []string, owner string) bool {
	if len(owners) == 0 {
		return true
	}
	for _, filter := range owners {
		if strings.EqualFold(owner, filter) {
			return true
		}
	}
	return false
}

func snapshot(rec state.Record) github.RepoSnapshot {
	return github.RepoSnapshot{ID: rec.ID, Owner: rec.OwnerLogin, Name: rec.Name,
		Visibility: github.Visibility(rec.Visibility), DefaultBranch: rec.DefaultBranch,
		IsFork: rec.IsFork, IsTemplate: rec.IsTemplate, IsArchived: rec.ArchivedOnGitHub}
}
