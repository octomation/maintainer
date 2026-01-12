package contribution

import (
	"sort"
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

// DayDiff describes how the contribution count of a day has changed
// from a base heatmap to a head one. A day missing from a heatmap counts
// as zero there, and InBase or InHead tells it apart from a present zero.
type DayDiff struct {
	Day            time.Time
	Before, After  uint
	InBase, InHead bool
}

// Delta returns the signed change of the count, e.g., -3 if 5 became 2.
func (diff DayDiff) Delta() int64 {
	return int64(diff.After) - int64(diff.Before)
}

// Diff compares the heatmap as a base with the head one day by day.
// It returns only the days where the count has changed, sorted by day.
func (chm HeatMap) Diff(head HeatMap) []DayDiff {
	keys := make(map[time.Time]struct{}, len(chm)+len(head))
	for ts := range chm {
		keys[ts] = struct{}{}
	}
	for ts := range head {
		keys[ts] = struct{}{}
	}

	diff := make([]DayDiff, 0, 8)
	for ts := range keys {
		before, inBase := chm[ts]
		after, inHead := head[ts]
		if before != after {
			diff = append(diff, DayDiff{
				Day:    ts,
				Before: before,
				After:  after,
				InBase: inBase,
				InHead: inHead,
			})
		}
	}
	sort.Slice(diff, func(i, j int) bool { return diff[i].Day.Before(diff[j].Day) })

	return diff
}

// Only returns the days the heatmap has and the other one lacks,
// sorted by day, e.g., the days a longer period covers beyond a shorter one.
func (chm HeatMap) Only(other HeatMap) []time.Time {
	days := make([]time.Time, 0, 8)
	for ts := range chm {
		if _, present := other[ts]; !present {
			days = append(days, ts)
		}
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	return days
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
