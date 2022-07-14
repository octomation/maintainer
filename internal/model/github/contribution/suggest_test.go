package contribution_test

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

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

func TestSuggest(t *testing.T) {
	tests := map[string]struct {
		// input
		chm    HeatMap
		start  time.Time
		end    time.Time
		target uint

		// output
		expected HistogramByWeekdayRow
	}{
		"empty heatmap": {
			make(HeatMap),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
				Sum: 5,
			},
		},
		"issue#68: missed zero": {
			func() HeatMap {
				Dec2020 := xtime.UTC().Year(2020).Month(time.December)
				Jan2021 := xtime.UTC().Year(2021).Month(time.January)

				// Sunday        6     7     9
				// Monday        6     7     5
				// Tuesday       6     7    12
				// Wednesday     6     7    10
				// Thursday      6     7     6
				// Friday        6     7     4
				// Saturday      6     -     6

				chm := make(HeatMap)
				chm.SetCount(Dec2020.Day(27).Time(), 6)
				chm.SetCount(Dec2020.Day(28).Time(), 6)
				chm.SetCount(Dec2020.Day(29).Time(), 6)
				chm.SetCount(Dec2020.Day(30).Time(), 6)
				chm.SetCount(Dec2020.Day(31).Time(), 6)
				chm.SetCount(Jan2021.Day(1).Time(), 6)
				chm.SetCount(Jan2021.Day(2).Time(), 6)

				chm.SetCount(Jan2021.Day(3).Time(), 7)
				chm.SetCount(Jan2021.Day(4).Time(), 7)
				chm.SetCount(Jan2021.Day(5).Time(), 7)
				chm.SetCount(Jan2021.Day(6).Time(), 7)
				chm.SetCount(Jan2021.Day(7).Time(), 7)
				chm.SetCount(Jan2021.Day(8).Time(), 7)

				chm.SetCount(Jan2021.Day(10).Time(), 9)
				chm.SetCount(Jan2021.Day(11).Time(), 5)
				chm.SetCount(Jan2021.Day(12).Time(), 12)
				chm.SetCount(Jan2021.Day(13).Time(), 10)
				chm.SetCount(Jan2021.Day(14).Time(), 6)
				chm.SetCount(Jan2021.Day(15).Time(), 4)
				chm.SetCount(Jan2021.Day(16).Time(), 6)
				return chm
			}(),
			xtime.UTC().Year(2021).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.January).Day(9).Time(),
				Sum: 7,
			},
		},
		"full week with some distribution": {
			load(t, "testdata/kamilsk.2021.html"),
			xtime.UTC().Year(2021).Month(time.September).Day(15).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.September).Day(15).Time(),
				Sum: 10,
			},
		},
		"week without contributions": {
			load(t, "testdata/kamilsk.2021.html"),
			xtime.UTC().Year(2021).Month(time.October).Day(7).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(7).Time(),
				Sum: 6,
			},
		},
		"week with gaps": {
			load(t, "testdata/kamilsk.2021.html"),
			xtime.UTC().Year(2021).Month(time.October).Day(16).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(16).Time(),
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

func load(t testing.TB, name string) HeatMap {
	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	return BuildHeatMap(doc)
}
