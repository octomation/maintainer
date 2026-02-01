package hub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/hub/mirror"
)

// Errors of a triage write, mapped to HTTP statuses by the API.
var (
	ErrNotFound        = errors.New("no such task")
	ErrDuplicateMirror = errors.New("the task has duplicate mirrors")
	ErrNotIndexed      = errors.New("the index is not ready")
)

// Index holds the current snapshot and the rules of its publication: a scan
// reads the mirrors without the lock and publishes under it only when the
// generation it started from is still current, otherwise it takes the
// entries of the writes it missed; a write holds the lock from the file to
// the new snapshot.
type Index struct {
	discover   func(ctx context.Context) ([]Root, error)
	weights    config.Weights
	repoWeight func(Root) float64
	now        func() time.Time

	mu       sync.Mutex // publication and writes
	current  atomic.Pointer[Snapshot]
	writes   []write // under mu: the paths written, with the generation each published
	scanMu   sync.Mutex
	scanning atomic.Int32
	pending  atomic.Bool
	wake     chan struct{}
}

type write struct {
	generation uint64
	path       string
}

// New returns an index over the roots discover finds.
func New(discover func(ctx context.Context) ([]Root, error), weights config.Weights, repoWeight func(Root) float64) *Index {
	if repoWeight == nil {
		repoWeight = func(Root) float64 { return 1 }
	}
	return &Index{discover: discover, weights: weights, repoWeight: repoWeight, now: time.Now, wake: make(chan struct{}, 1)}
}

// Snapshot returns the current snapshot, nil before the first scan.
func (ix *Index) Snapshot() *Snapshot { return ix.current.Load() }

// SetNow replaces the clock of the snapshots: tests and fixtures pin it.
func (ix *Index) SetNow(now func() time.Time) { ix.now = now }

// Refreshing reports a scan in progress.
func (ix *Index) Refreshing() bool { return ix.scanning.Load() > 0 || ix.pending.Load() }

// Start runs the initial scan in the background and serves the refresh
// requests until ctx ends. Errors are passed to report.
func (ix *Index) Start(ctx context.Context, report func(error)) {
	go func() {
		if err := ix.Scan(ctx); err != nil && report != nil {
			report(err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ix.wake:
				ix.pending.Store(false)
				if err := ix.Scan(ctx); err != nil && report != nil {
					report(err)
				}
			}
		}
	}()
}

// Refresh asks for a full scan; requests during a scan coalesce into one
// more scan after it.
func (ix *Index) Refresh() {
	ix.pending.Store(true)
	select {
	case ix.wake <- struct{}{}:
	default:
	}
}

// Scan discovers the roots, reads every mirror file and publishes a new
// generation. A malformed file keeps its previous entry, flagged stale.
func (ix *Index) Scan(ctx context.Context) error {
	ix.scanMu.Lock()
	defer ix.scanMu.Unlock()
	ix.scanning.Add(1)
	defer ix.scanning.Add(-1)
	var from uint64
	if cur := ix.current.Load(); cur != nil {
		from = cur.Generation
	}
	discovered, err := ix.discover(ctx)
	if err != nil {
		return fmt.Errorf("discover mirror roots: %w", err)
	}
	// The real path of a root is remembered: a write compares it again.
	roots := make([]Root, 0, len(discovered))
	var unresolved []string
	for _, root := range discovered {
		if root.Real == "" {
			real, err := filepath.EvalSymlinks(root.Path)
			if err != nil {
				unresolved = append(unresolved, fmt.Sprintf("%s: %v", root.Path, err))
				continue
			}
			root.Real = real
		}
		if root.Info == nil {
			info, err := os.Stat(root.Real)
			if err != nil {
				unresolved = append(unresolved, fmt.Sprintf("%s: %v", root.Path, err))
				continue
			}
			root.Info = info
		}
		roots = append(roots, root)
	}
	entries := map[string]*entry{}
	for _, root := range roots {
		if err := ctx.Err(); err != nil {
			return err
		}
		for _, e := range readRoot(root) {
			entries[e.path] = e
		}
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()
	cur := ix.current.Load()
	if cur != nil {
		if cur.Generation != from {
			// Writes published since the scan started are fresher than its
			// parse of the same files.
			for _, w := range ix.writes {
				if w.generation > from {
					if e, ok := cur.entries[w.path]; ok {
						if _, listed := entries[w.path]; listed {
							entries[w.path] = e
						}
					}
				}
			}
		}
		for path, e := range entries {
			if e.kind == kindBroken {
				if prev, ok := cur.entries[path]; ok && prev.kind != kindBroken {
					kept := *prev
					kept.stale = true
					entries[path] = &kept
					// The warning of the malformed file is reported below.
					kept.err = e.err
				}
			}
		}
	}
	snap := build(entries, roots, ix.now(), ix.weights, ix.repoWeight)
	for _, path := range sortedPaths(entries) {
		if e := entries[path]; e.stale && e.err != "" {
			snap.Warnings = append(snap.Warnings, fmt.Sprintf("%s: %s (the previous entry is kept)", path, e.err))
		}
	}
	for _, u := range unresolved {
		snap.Warnings = append(snap.Warnings, u+" (the root is left out)")
	}
	snap.Generation = from + 1
	if cur != nil {
		snap.Generation = cur.Generation + 1
	}
	ix.current.Store(snap)
	ix.writes = nil
	return nil
}

// Patch applies a triage to the task of key, whose file must still be the
// given version, and publishes the new generation.
func (ix *Index) Patch(key, version string, triage mirror.Triage) (*Task, *Snapshot, error) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	cur := ix.current.Load()
	if cur == nil {
		return nil, nil, ErrNotIndexed
	}
	task := cur.byKey[key]
	if task == nil {
		return nil, nil, ErrNotFound
	}
	if task.Flags.DuplicateMirror {
		return nil, nil, ErrDuplicateMirror
	}
	if task.Version != version {
		return nil, nil, mirror.ErrVersionMismatch
	}
	// The root, or an ancestor of it, may have been replaced by a symlink
	// since discovery: the real path must still be the one seen then.
	if real, err := filepath.EvalSymlinks(task.root.Path); err != nil || real != task.root.Real {
		return nil, nil, fmt.Errorf("%w: the mirror root %s no longer resolves to %s", mirror.ErrNotEditable, task.root.Path, task.root.Real)
	}
	_, err := mirror.Patch(task.root.Path, task.path, version, triage, mirror.Options{Identity: task.root.Info})
	switch {
	case err == nil:
	case errors.Is(err, mirror.ErrVersionMismatch):
		// The file changed on disk behind the index: publish its current
		// parse, so a reload gets the current version instead of the stale
		// one and the next write can succeed.
		ix.publishFile(cur, task)
		return nil, nil, err
	case errors.Is(err, mirror.ErrDirectorySync):
		// The file is replaced: the index follows it; the caller reports the
		// failed sync.
	default:
		return nil, nil, err
	}
	snap := ix.publishFile(cur, task)
	return snap.byKey[key], snap, err
}

// publishFile reparses the file of task into a new generation over cur.
func (ix *Index) publishFile(cur *Snapshot, task *Task) *Snapshot {
	entries := make(map[string]*entry, len(cur.entries))
	for path, e := range cur.entries {
		entries[path] = e
	}
	entries[task.path] = readEntry(task.root, task.path)
	snap := build(entries, cur.Roots, ix.now(), ix.weights, ix.repoWeight)
	snap.Generation = cur.Generation + 1
	ix.writes = append(ix.writes, write{generation: snap.Generation, path: task.path})
	if len(ix.writes) > 4096 {
		ix.writes = ix.writes[len(ix.writes)-4096:]
	}
	ix.current.Store(snap)
	return snap
}

func sortedPaths(entries map[string]*entry) []string {
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
