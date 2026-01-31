package fetch

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

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
