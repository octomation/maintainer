package contribution_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestHistogramByCount(t *testing.T) {
	Nov2013 := xtime.UTC().Year(2013).Month(time.November)

	chm := make(HeatMap)
	chm.SetCount(Nov2013.Day(13).Time(), 1)
	chm.SetCount(Nov2013.Day(20).Time(), 1)
	chm.SetCount(Nov2013.Day(21).Time(), 3)
	chm.SetCount(Nov2013.Day(24).Time(), 1)
	chm.SetCount(Nov2013.Day(25).Time(), 2)
	chm.SetCount(Nov2013.Day(26).Time(), 8)
	chm.SetCount(Nov2013.Day(28).Time(), 7)
	chm.SetCount(Nov2013.Day(29).Time(), 1)

	expected := map[uint]uint{
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
	Nov2013 := xtime.UTC().Year(2013).Month(time.November)

	chm := make(HeatMap)
	chm.SetCount(Nov2013.Day(13).Time(), 1)
	chm.SetCount(Nov2013.Day(20).Time(), 1)
	chm.SetCount(Nov2013.Day(21).Time(), 3)
	chm.SetCount(Nov2013.Day(24).Time(), 1)
	chm.SetCount(Nov2013.Day(25).Time(), 2)
	chm.SetCount(Nov2013.Day(26).Time(), 8)
	chm.SetCount(Nov2013.Day(28).Time(), 7)
	chm.SetCount(Nov2013.Day(29).Time(), 1)

	t.Run("grouped by day", func(t *testing.T) {
		expected := map[string]uint{
			"2013-11-13": 1,
			"2013-11-20": 1,
			"2013-11-21": 3,
			"2013-11-24": 1,
			"2013-11-25": 2,
			"2013-11-26": 8,
			"2013-11-28": 7,
			"2013-11-29": 1,
		}

		histogram := HistogramByDate(chm, xtime.DateOnly)
		require.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})

	t.Run("grouped by month", func(t *testing.T) {
		expected := map[string]uint{
			"2013-11": 24,
		}

		histogram := HistogramByDate(chm, xtime.YearAndMonth)
		require.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})
}

func TestHistogramByWeekday(t *testing.T) {
	Nov2013 := xtime.UTC().Year(2013).Month(time.November)

	chm := make(HeatMap)
	chm.SetCount(Nov2013.Day(13).Time(), 1) // Wednesday
	chm.SetCount(Nov2013.Day(20).Time(), 1) // Wednesday
	chm.SetCount(Nov2013.Day(21).Time(), 3) // Thursday
	chm.SetCount(Nov2013.Day(24).Time(), 1) // Sunday
	chm.SetCount(Nov2013.Day(25).Time(), 2) // Monday
	chm.SetCount(Nov2013.Day(26).Time(), 8) // Tuesday
	chm.SetCount(Nov2013.Day(28).Time(), 7) // Thursday
	chm.SetCount(Nov2013.Day(29).Time(), 1) // Friday

	t.Run("grouped", func(t *testing.T) {
		expected := map[time.Weekday]uint{
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
		expected := map[time.Weekday][]uint{
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
			sum, weekday := uint(0), row.Day.Weekday()
			sum, expected[weekday] = expected[weekday][0], expected[weekday][1:]
			assert.Equal(t, sum, row.Sum, i)
		}
	})
}
