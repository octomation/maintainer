package fetch

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/service/workspace"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// scanOccupancy snapshots every rendered target immediately before planning.
// Existing paths are fail-closed: the Planner may proceed only when the
// Adopter has independently verified the path as the expected repository.
// A target below a symlink or a file inside root is unsafe, never clear.
func (s *Service) scanOccupancy(snapshots []github.RepoSnapshot, st *state.State) map[string]Occupancy {
	result := make(map[string]Occupancy, len(snapshots))
	for _, snap := range snapshots {
		rec, _ := st.ByID(snap.ID)
		// Pins are verified from disk identity and never materialised. A path
		// that cannot be rendered is the planner's conflict for that repository.
		if pin, err := s.planner.paths.PinnedPath(snap, rec); err != nil || pin != "" {
			continue
		}
		path, err := s.planner.paths.Resolve(snap)
		if err != nil {
			continue
		}
		switch err := workspace.Destination(s.root, path); {
		case err == nil:
			result[path] = OccupancyClear
		case errors.Is(err, fs.ErrExist):
			result[path] = OccupancyForeign
		case errors.Is(err, workspace.ErrSymlinkParent), errors.Is(err, workspace.ErrFileParent):
			result[path] = OccupancyUnsafe
		default:
			result[path] = OccupancyUnknown // e.g. an unreadable parent, reported by the scan
		}
	}
	return result
}

// boundedSync enforces the workspace boundary at the mutation point: the
// destination is checked again right before Git creates or renames into it,
// so a symlink parent or a path that appeared after planning is refused.
type boundedSync struct {
	gitsvc.GitSync
	root string
}

func (b boundedSync) Clone(ctx context.Context, opt gitsvc.CloneOptions) error {
	if err := workspace.Destination(b.root, opt.Path); err != nil {
		return fmt.Errorf("refusing to clone into %s: %w", opt.Path, err)
	}
	return b.GitSync.Clone(ctx, opt)
}

func (b boundedSync) Move(from, to string) error {
	if err := workspace.Destination(b.root, to); err != nil {
		return fmt.Errorf("refusing to move %s to %s: %w", from, to, err)
	}
	return b.GitSync.Move(from, to)
}

// caseAliases finds rendered targets that differ from their recorded path only
// in letter case and denote the same directory: a case-insensitive filesystem
// already shows the checkout under the new spelling, so a rename that changes
// only the case needs no move (§7.5). Both paths are compared without
// following a final symbolic link.
func (s *Service) caseAliases(snapshots []github.RepoSnapshot, st *state.State) map[string]string {
	aliases := make(map[string]string)
	for _, snap := range snapshots {
		rec, ok := st.ByID(snap.ID)
		if !ok {
			continue
		}
		if pin, err := s.planner.paths.PinnedPath(snap, rec); err != nil || pin != "" {
			continue // pins never move
		}
		target, err := s.planner.paths.Resolve(snap)
		if err != nil || target == rec.Path || !strings.EqualFold(target, rec.Path) {
			continue
		}
		recorded, err := os.Lstat(rec.Path)
		if err != nil {
			continue
		}
		if shown, err := os.Lstat(target); err == nil && os.SameFile(recorded, shown) {
			aliases[target] = rec.Path
		}
	}
	return aliases
}
