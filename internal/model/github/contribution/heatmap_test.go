package contribution_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
)

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

		histogram := HistogramByDate(chm, "2006-01-02")
		require.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})

	t.Run("grouped by month", func(t *testing.T) {
		expected := map[string]int{
			"2013-11": 24,
		}

		histogram := HistogramByDate(chm, "2006-01")
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
			assert.Equal(t, expected[row.Day], row.Sum, i)
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
			shift := expected[row.Day][0]
			expected[row.Day] = expected[row.Day][1:]

			assert.Equal(t, shift, row.Sum, i)
		}
	})
}
