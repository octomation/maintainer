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

func TestSuggest(t *testing.T) {
	tests := map[string]struct {
		// input
		chm   HeatMap
		start time.Time
		end   time.Time
		basis int

		// output
		expected HistogramByWeekdayRow
	}{
		"zero range": {
			make(HeatMap),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
				Sum: 5,
			},
		},

		// ---

		"empty heatmap": {
			make(HeatMap),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(3).Time(), // <- 5
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
				chm.SetCount(Jan2021.Day(9).Time(), 0) // <- skip

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

		// ---

		"full week with some distribution": {
			load(t, "testdata/kamilsk.2021.html"),
			xtime.UTC().Year(2021).Month(time.September).Day(15).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.September).Day(12).Time(),
				Sum: 10,
			},
		},
		"week without contributions": {
			load(t, "testdata/kamilsk.2021.html"),
			xtime.UTC().Year(2021).Month(time.October).Day(7).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(3).Time(),
				Sum: 6,
			},
		},
		"week with gaps": {
			load(t, "testdata/kamilsk.2021.html"),
			xtime.UTC().Year(2021).Month(time.October).Day(16).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(10).Time(),
				Sum: 8,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Suggest(test.chm, test.start, test.end, test.basis))
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
