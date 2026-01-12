package contribution_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestSuggest(t *testing.T) {
	tests := map[string]struct {
		heats HeatMap
		since time.Time
		until time.Time
		hours xtime.Schedule
		basis uint

		expected Suggestion
	}{
		"empty heatmap": {
			heats: make(HeatMap),
			since: xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			until: xtime.UTC().Year(2022).Time(),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.October).Day(5).Hour(8).Time(),
				Actual: 0,
				Target: 5,
			},
		},
		"empty week": {
			heats: load(t, "testdata/kamilsk.2019.json"),
			since: xtime.UTC().Year(2019).Month(time.October).Day(7).Time(),
			until: xtime.UTC().Year(2020).Time(),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2019).Month(time.October).Day(7).Hour(8).Time(),
				Actual: 0,
				Target: 5,
			},
		},
		"full week": {
			heats: load(t, "testdata/kamilsk.2021.json"),
			since: xtime.UTC().Year(2021).Month(time.April).Day(28).Time(),
			until: xtime.UTC().Year(2022).Time(),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.April).Day(28).Hour(8).Time(),
				Actual: 3,
				Target: 10,
			},
		},
		"week with gaps": {
			heats: load(t, "testdata/kamilsk.2019.json"),
			since: xtime.UTC().Year(2019).Month(time.December).Day(17).Time(),
			until: xtime.UTC().Year(2020).Time(),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2019).Month(time.December).Day(17).Hour(8).Time(),
				Actual: 0,
				Target: 5,
			},
		},
		"issue#68: missed Saturday": {
			heats: func() HeatMap {
				chm := load(t, "testdata/kamilsk.2021.json")
				delete(chm, xtime.UTC().Year(2021).Month(time.December).Day(2).Time())
				return chm
			}(),
			since: xtime.UTC().Year(2021).Month(time.November).Day(28).Time(),
			until: xtime.UTC().Year(2022).Time(),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.December).Day(2).Hour(8).Time(),
				Actual: 0,
				Target: 15,
			},
		},
		"issue#119: max Saturday": {
			heats: load(t, "testdata/kamilsk.2021.json"),
			since: xtime.UTC().Year(2021).Month(time.April).Day(3).Time(),
			until: xtime.UTC().Year(2022).Time(),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.April).Day(4).Hour(8).Time(),
				Actual: 7,
				Target: 8,
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Suggest(test.heats, test.since, test.until, test.hours, test.basis))
		})
	}
}

// TestSuggest_Bounds covers the window of a suggestion: it is never
// before the anchor, nor after now, nor outside the schedule, and nothing
// is suggested if there is no such moment.
func TestSuggest_Bounds(t *testing.T) {
	sunday := xtime.UTC().Year(2026).Month(time.September).Day(20) // the week of now
	now := sunday.Day(25).Hour(12).Minute(34).Second(56)           // Friday
	hours, basis := xtime.Everyday(xtime.Hours(5, 19, 0)), uint(5) // the schedule of suggest
	week := func(counts ...uint) HeatMap {
		chm := make(HeatMap)
		for i, count := range counts {
			chm.SetCount(sunday.Day(20+i).Time(), count)
		}
		return chm
	}

	tests := map[string]struct {
		heats HeatMap
		since time.Time
		until time.Time

		expected Suggestion
	}{
		"anchor in the past on a gap": {
			heats:    week(5, 5, 5, 5, 5, 0),
			since:    now.Hour(10).Time(),
			until:    now.Time(),
			expected: Suggestion{Time: now.Hour(10).Time(), Actual: 0, Target: 5},
		},
		"anchor in the past before a gap": {
			heats:    week(5, 5, 5, 5, 2, 0),
			since:    sunday.Day(23).Hour(8).Time(), // Wednesday
			until:    now.Time(),
			expected: Suggestion{Time: sunday.Day(24).Hour(5).Time(), Actual: 2, Target: 5},
		},
		"the target of the week is above the basis": {
			heats:    week(12, 12, 12, 7, 12, 12),
			since:    sunday.Time(),
			until:    now.Time(),
			expected: Suggestion{Time: sunday.Day(23).Hour(5).Time(), Actual: 7, Target: 12},
		},
		"week filled up to now": {
			heats: week(5, 5, 5, 5, 5, 5), // Saturday is a gap, but in the future
			since: sunday.Time(),
			until: now.Time(),
		},
		"anchor after the schedule the same evening": {
			heats: week(5, 5, 5, 5, 5, 0), // Friday is a gap, but its schedule is over
			since: now.Hour(20).Time(),
			until: now.Hour(21).Time(),
		},
		"no gap in the window": {
			heats: week(10, 10, 10, 10, 10, 10, 10),
			since: sunday.Day(23).Hour(10).Time(),
			until: now.Time(),
		},
		"anchor is now on a gap": {
			heats:    week(5, 5, 5, 5, 5, 3),
			since:    now.Time(),
			until:    now.Time(),
			expected: Suggestion{Time: now.Time(), Actual: 3, Target: 5},
		},
		"anchor is now on target": {
			heats: week(5, 5, 5, 5, 5, 5), // Saturday is a gap, but in the future
			since: now.Time(),
			until: now.Time(),
		},
		"anchor is now before the schedule": {
			heats: week(5, 5, 5, 5, 5, 0),
			since: now.Hour(3).Time(),
			until: now.Hour(3).Time(),
		},
		"anchor is now after the schedule": {
			heats: week(5, 5, 5, 5, 5, 0),
			since: now.Hour(20).Time(),
			until: now.Hour(20).Time(),
		},
		"anchor an hour ahead": {
			heats: week(5, 5, 5, 5, 5, 0),
			since: now.Add(time.Hour).Time(),
			until: now.Time(),
		},
		"anchor a day ahead": {
			heats: week(5, 5, 5, 5, 5, 0),
			since: now.Add(xtime.Day).Time(),
			until: now.Time(),
		},
		"anchor a year ahead": {
			heats: week(5, 5, 5, 5, 5, 0),
			since: now.Year(2027).Time(),
			until: now.Time(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			suggestion := Suggest(test.heats, test.since, test.until, hours, basis)
			assert.Equal(t, test.expected, suggestion)
			if suggestion.Time.IsZero() {
				return
			}
			assert.False(t, suggestion.Time.Before(test.since), "before the anchor")
			assert.False(t, suggestion.Time.After(test.until), "after now")
			assert.False(t, hours.End(suggestion.Time).IsZero(), "outside the schedule")
			assert.Equal(t, test.heats.Count(xtime.TruncateToDay(suggestion.Time)), suggestion.Actual)
		})
	}
}

func TestWeekDistribution_Suggest(t *testing.T) {
	tests := map[string]struct {
		week  WeekDistribution
		since time.Weekday
		basis uint

		day time.Weekday
		val uint
	}{
		"empty week, beginning": {
			week:  [7]uint{},
			since: time.Sunday,
			basis: 5,

			day: time.Sunday,
			val: 5,
		},
		"empty week, midweek": {
			week:  [7]uint{},
			since: time.Wednesday,
			basis: 5,

			day: time.Wednesday,
			val: 5,
		},
		"empty week, ending": {
			week:  [7]uint{},
			since: time.Saturday,
			basis: 5,

			day: time.Saturday,
			val: 5,
		},
		"ascending week, beginning": {
			week:  [7]uint{1, 2, 3, 4, 5, 6, 7},
			since: time.Sunday,
			basis: 5,

			day: time.Sunday,
			val: 7,
		},
		"ascending week, midweek": {
			week:  [7]uint{1, 2, 3, 4, 5, 6, 7},
			since: time.Wednesday,
			basis: 5,

			day: time.Wednesday,
			val: 7,
		},
		"ascending week, ending": {
			week:  [7]uint{1, 2, 3, 4, 5, 6, 7},
			since: time.Saturday,
			basis: 5,

			day: -1,
			val: 7,
		},
		"descending week, beginning": {
			week:  [7]uint{7, 6, 5, 4, 3, 2, 1},
			since: time.Sunday,
			basis: 5,

			day: time.Monday,
			val: 7,
		},
		"descending week, midweek": {
			week:  [7]uint{7, 6, 5, 4, 3, 2, 1},
			since: time.Wednesday,
			basis: 5,

			day: time.Wednesday,
			val: 7,
		},
		"descending week, ending": {
			week:  [7]uint{7, 6, 5, 4, 3, 2, 1},
			since: time.Saturday,
			basis: 5,

			day: time.Saturday,
			val: 7,
		},
		"convex week, beginning": {
			week:  [7]uint{1, 2, 3, 4, 3, 2, 1},
			since: time.Sunday,
			basis: 5,

			day: time.Sunday,
			val: 5,
		},
		"convex week, midweek": {
			week:  [7]uint{1, 2, 3, 4, 3, 2, 1},
			since: time.Wednesday,
			basis: 5,

			day: time.Wednesday,
			val: 5,
		},
		"convex week, ending": {
			week:  [7]uint{1, 2, 3, 4, 3, 2, 1},
			since: time.Saturday,
			basis: 5,

			day: time.Saturday,
			val: 5,
		},
		"sunken week, beginning": {
			week:  [7]uint{7, 5, 2, 1, 3, 4, 6},
			since: time.Sunday,
			basis: 5,

			day: time.Monday,
			val: 7,
		},
		"sunken week, midweek": {
			week:  [7]uint{7, 5, 2, 1, 3, 4, 6},
			since: time.Wednesday,
			basis: 5,

			day: time.Wednesday,
			val: 7,
		},
		"sunken week, ending": {
			week:  [7]uint{7, 5, 2, 1, 3, 4, 6},
			since: time.Saturday,
			basis: 5,

			day: time.Saturday,
			val: 7,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			day, val := test.week.Suggest(test.since, test.basis)
			assert.Equal(t, test.day, day)
			assert.Equal(t, test.val, val)
		})
	}
}
