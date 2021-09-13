package contribution_test

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestHeatMap_Subset(t *testing.T) {
	chm := make(HeatMap)
	chm.SetCount(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3)
	chm.SetCount(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2)
	chm.SetCount(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8)
	chm.SetCount(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7)
	chm.SetCount(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1)

	t.Run("one week", func(t *testing.T) {
		ts := time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC)
		subset := chm.Subset(xtime.RangeByWeeks(ts, 0, false))
		assert.Len(t, subset, 3)
	})

	t.Run("one week behind", func(t *testing.T) {
		ts := time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC)
		subset := chm.Subset(xtime.RangeByWeeks(ts, -1, false))
		assert.Len(t, subset, 4)
	})

	t.Run("one week ahead", func(t *testing.T) {
		ts := time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC)
		subset := chm.Subset(xtime.RangeByWeeks(ts, 1, false))
		assert.Len(t, subset, 7)
	})

	t.Run("one week around", func(t *testing.T) {
		ts := time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC)
		subset := chm.Subset(xtime.RangeByWeeks(ts, 1, true))
		assert.Len(t, subset, 3)
	})

	t.Run("three weeks around", func(t *testing.T) {
		ts := time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC)
		subset := chm.Subset(xtime.RangeByWeeks(ts, 3, true))
		assert.Len(t, subset, 8)
	})
}

func TestHistogramByCount(t *testing.T) {
	chm := make(HeatMap)
	chm.SetCount(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3)
	chm.SetCount(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2)
	chm.SetCount(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8)
	chm.SetCount(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7)
	chm.SetCount(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1)

	expected := map[int]int{
		1: 4,
		2: 1,
		3: 1,
		7: 1,
		8: 1,
	}

	histogram := HistogramByCount(chm)
	require.Len(t, histogram, len(expected))
	for i, row := range histogram {
		assert.Equal(t, expected[row.Count], row.Frequency, i)
	}
}

func TestHistogramByDate(t *testing.T) {
	chm := make(HeatMap)
	chm.SetCount(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3)
	chm.SetCount(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2)
	chm.SetCount(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8)
	chm.SetCount(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7)
	chm.SetCount(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1)

	t.Run("grouped by day", func(t *testing.T) {
		expected := map[string]int{
			"2013-11-13": 1,
			"2013-11-20": 1,
			"2013-11-21": 3,
			"2013-11-24": 1,
			"2013-11-25": 2,
			"2013-11-26": 8,
			"2013-11-28": 7,
			"2013-11-29": 1,
		}

		histogram := HistogramByDate(chm, xtime.RFC3339Day)
		require.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})

	t.Run("grouped by month", func(t *testing.T) {
		expected := map[string]int{
			"2013-11": 24,
		}

		histogram := HistogramByDate(chm, xtime.RFC3339Month)
		require.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})
}

func TestHistogramByWeekday(t *testing.T) {
	chm := make(HeatMap)
	chm.SetCount(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1) // Wednesday
	chm.SetCount(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1) // Wednesday
	chm.SetCount(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3) // Thursday
	chm.SetCount(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1) // Sunday
	chm.SetCount(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2) // Monday
	chm.SetCount(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8) // Tuesday
	chm.SetCount(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7) // Thursday
	chm.SetCount(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1) // Friday

	t.Run("grouped", func(t *testing.T) {
		expected := map[time.Weekday]int{
			time.Sunday:    1,
			time.Monday:    2,
			time.Tuesday:   8,
			time.Wednesday: 2,
			time.Thursday:  10,
			time.Friday:    1,
		}

		histogram := HistogramByWeekday(chm, true)
		require.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Day.Weekday()], row.Sum, i)
		}
	})

	t.Run("ungrouped", func(t *testing.T) {
		expected := map[time.Weekday][]int{
			time.Sunday:    {1},
			time.Monday:    {2},
			time.Tuesday:   {8},
			time.Wednesday: {1, 1},
			time.Thursday:  {3, 7},
			time.Friday:    {1},
		}

		total := 0
		for _, s := range expected {
			total += len(s)
		}

		histogram := HistogramByWeekday(chm, false)
		require.Len(t, histogram, total)
		for i, row := range histogram {
			sum, weekday := 0, row.Day.Weekday()
			sum, expected[weekday] = expected[weekday][0], expected[weekday][1:]
			assert.Equal(t, sum, row.Sum, i)
		}
	})
}

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
		"issue#68: missed zero": {
			golden(t, "issue-68.golden.json"),
			time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: time.Date(2021, time.September, 11, 0, 0, 0, 0, time.UTC),
				Sum: 7,
			},
		},
		"full week with some distribution": {
			golden(t, "issue-68.golden.json"),
			time.Date(2021, time.September, 15, 0, 0, 0, 0, time.UTC),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: time.Date(2021, time.September, 12, 0, 0, 0, 0, time.UTC),
				Sum: 12,
			},
		},
		"week without contributions": {
			golden(t, "issue-68.golden.json"),
			time.Date(2021, time.October, 7, 0, 0, 0, 0, time.UTC),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: time.Date(2021, time.October, 3, 0, 0, 0, 0, time.UTC),
				Sum: 5,
			},
		},
		"week with gaps": {
			golden(t, "issue-68.golden.json"),
			time.Date(2021, time.October, 16, 0, 0, 0, 0, time.UTC),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: time.Date(2021, time.October, 11, 0, 0, 0, 0, time.UTC),
				Sum: 8,
			},
		},
		"empty contribution heatmap": {
			make(HeatMap),
			time.Date(2021, time.October, 5, 0, 0, 0, 0, time.UTC),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: time.Date(2021, time.October, 3, 0, 0, 0, 0, time.UTC),
				Sum: 5,
			},
		},
		"no range": {
			make(HeatMap),
			time.Date(2021, time.October, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2021, time.October, 5, 0, 0, 0, 0, time.UTC),
			5,

			HistogramByWeekdayRow{
				Day: time.Date(2021, time.October, 5, 0, 0, 0, 0, time.UTC),
				Sum: 5,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Suggest(test.chm, test.start, test.end, test.basis))
		})
	}
}

func golden(t testing.TB, name string) HeatMap {
	src := FileSource{
		Provider: afero.NewBasePathFs(afero.NewOsFs(), "testdata"),
		Path:     name,
	}
	chm, err := src.Fetch(context.TODO())
	require.NoError(t, err)
	return chm
}
