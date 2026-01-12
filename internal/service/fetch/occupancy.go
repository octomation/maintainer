package fetch

import (
	"fmt"
	"os"

	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// scanOccupancy snapshots every rendered target immediately before planning.
// Existing paths are fail-closed: the Planner may proceed only when the
// Adopter has independently verified the path as the expected repository.
func (s *Service) scanOccupancy(snapshots []github.RepoSnapshot) (map[string]Occupancy, error) {
	result := make(map[string]Occupancy, len(snapshots))
	for _, snap := range snapshots {
		path, err := s.planner.paths.Resolve(snap)
		if err != nil {
			return nil, fmt.Errorf("resolve path for %s (id=%d): %w", snap.FullName(), snap.ID, err)
		}
		if _, err := os.Lstat(path); err == nil {
			result[path] = OccupancyForeign
		} else if os.IsNotExist(err) {
			result[path] = OccupancyClear
		} else {
			return nil, fmt.Errorf("inspect target path %s: %w", path, err)
		}
	}
	return result, nil
}
