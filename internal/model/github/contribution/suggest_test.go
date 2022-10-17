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
		scope xtime.Range
		hours xtime.Schedule
		basis uint

		expected Suggestion
	}{
		"empty heatmap": {
			heats: make(HeatMap),
			scope: xtime.NewRange(
				xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
				time.Now(),
			),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.October).Day(5).Hour(8).Time(),
				Actual: 0,
				Target: 5,
			},
		},
		"empty week": {
			heats: BuildHeatMap(load(t, "testdata/kamilsk.2019.html")),
			scope: xtime.NewRange(
				xtime.UTC().Year(2019).Month(time.October).Day(7).Time(),
				time.Now(),
			),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2019).Month(time.October).Day(7).Hour(8).Time(),
				Actual: 0,
				Target: 5,
			},
		},
		"full week": {
			heats: BuildHeatMap(load(t, "testdata/kamilsk.2021.html")),
			scope: xtime.NewRange(
				xtime.UTC().Year(2021).Month(time.April).Day(28).Time(),
				time.Now(),
			),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.April).Day(28).Hour(8).Time(),
				Actual: 4,
				Target: 10,
			},
		},
		"week with gaps": {
			heats: BuildHeatMap(load(t, "testdata/kamilsk.2019.html")),
			scope: xtime.NewRange(
				xtime.UTC().Year(2019).Month(time.December).Day(17).Time(),
				time.Now(),
			),
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
				chm := BuildHeatMap(load(t, "testdata/kamilsk.2021.html"))
				delete(chm, xtime.UTC().Year(2021).Month(time.December).Day(18).Time())
				return chm
			}(),
			scope: xtime.NewRange(
				xtime.UTC().Year(2021).Month(time.December).Day(12).Time(),
				time.Now(),
			),
			hours: xtime.Everyday(xtime.Hours(8, 22, 0)),
			basis: 5,
			expected: Suggestion{
				Time:   xtime.UTC().Year(2021).Month(time.December).Day(18).Hour(8).Time(),
				Actual: 0,
				Target: 10,
			},
		},
		"issue#119: max Saturday": {
			heats: BuildHeatMap(load(t, "testdata/kamilsk.2021.html")),
			scope: xtime.NewRange(
				xtime.UTC().Year(2021).Month(time.April).Day(3).Time(),
				time.Now(),
			),
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
			assert.Equal(t, test.expected, Suggest(test.heats, test.scope, test.hours, test.basis))
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
