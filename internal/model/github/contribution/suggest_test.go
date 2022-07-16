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
		chm    HeatMap
		start  time.Time
		end    time.Time
		target uint

		expected HistogramByWeekdayRow
	}{
		"empty heatmap": {
			chm:    make(HeatMap),
			start:  xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			end:    time.Now().UTC(),
			target: 5,
			expected: HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
				Sum: 5,
			},
		},
		"empty week": {
			chm:    BuildHeatMap(load(t, "testdata/kamilsk.2019.html")),
			start:  xtime.UTC().Year(2019).Month(time.October).Day(7).Time(),
			end:    time.Now().UTC(),
			target: 5,
			expected: HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2019).Month(time.October).Day(7).Time(),
				Sum: 5,
			},
		},
		"full week": {
			chm:    BuildHeatMap(load(t, "testdata/kamilsk.2021.html")),
			start:  xtime.UTC().Year(2021).Month(time.April).Day(28).Time(),
			end:    time.Now().UTC(),
			target: 5,
			expected: HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.April).Day(28).Time(),
				Sum: 10,
			},
		},
		"week with gaps": {
			chm:    BuildHeatMap(load(t, "testdata/kamilsk.2019.html")),
			start:  xtime.UTC().Year(2019).Month(time.December).Day(17).Time(),
			end:    time.Now().UTC(),
			target: 5,
			expected: HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2019).Month(time.December).Day(17).Time(),
				Sum: 5,
			},
		},
		"issue#68: missed Saturday": {
			chm: func() HeatMap {
				chm := BuildHeatMap(load(t, "testdata/kamilsk.2021.html"))
				delete(chm, xtime.UTC().Year(2021).Month(time.December).Day(18).Time())
				return chm
			}(),
			start:  xtime.UTC().Year(2021).Month(time.December).Day(12).Time(),
			end:    time.Now().UTC(),
			target: 5,
			expected: HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.December).Day(18).Time(),
				Sum: 10,
			},
		},
		"issue#119: max Saturday": {
			chm:    BuildHeatMap(load(t, "testdata/kamilsk.2021.html")),
			start:  xtime.UTC().Year(2021).Month(time.April).Day(3).Time(),
			end:    time.Now().UTC(),
			target: 5,
			expected: HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.April).Day(4).Time(),
				Sum: 8,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Suggest(test.chm, test.start, test.end, test.target))
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
