package contribution

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	"go.octolab.org/toolset/maintainer/internal/pkg/assert/checks"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// HeatMap contains how many contributions have been made in a time.
type HeatMap map[time.Time]uint

// Count returns how many contributions have been made in the specified time.
func (chm HeatMap) Count(ts time.Time) uint {
	assert.True(func() bool { return ts.Location() == time.UTC })
	assert.True(func() bool { return checks.ZeroClock(ts.Clock()) })

	return chm[ts]
}

// SetCount sets how many contributions have been made to the specified time.
func (chm HeatMap) SetCount(ts time.Time, count uint) {
	assert.True(func() bool { return ts.Location() == time.UTC })
	assert.True(func() bool { return checks.ZeroClock(ts.Clock()) })

	chm[ts] = count
}

// Subset returns a subset of contribution heatmap in the provided time range.
func (chm HeatMap) Subset(scope xtime.Range) HeatMap {
	subset := make(HeatMap)

	for ts, count := range chm {
		if scope.Contains(ts) {
			subset.SetCount(ts, count)
		}
	}

	return subset
}

// Diff calculates the difference between two heatmaps.
func (chm HeatMap) Diff(src HeatMap) HeatMap {
	diff := make(HeatMap)

	keys := make(map[time.Time]struct{}, len(chm)+len(src))
	for ts := range chm {
		keys[ts] = struct{}{}
	}
	for ts := range src {
		keys[ts] = struct{}{}
	}
	for ts := range keys {
		if delta := src.Count(ts) - chm.Count(ts); delta != 0 {
			diff.SetCount(ts, delta)
		}
	}

	return diff
}

// From returns minimum time of the heatmap, otherwise the zero time instant.
func (chm HeatMap) From() time.Time {
	var min time.Time
	for ts := range chm {
		if ts.Before(min) || min.IsZero() {
			min = ts
		}
	}
	return min
}

// To returns maximum time of the heatmap, otherwise the zero time instant.
func (chm HeatMap) To() time.Time {
	var max time.Time
	for ts := range chm {
		if ts.After(max) {
			max = ts
		}
	}
	return max
}
