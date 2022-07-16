package contribution

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// Suggest finds a week with gaps in the contribution heatmap
// and returns an appropriate day to contribute.
func Suggest(
	chm HeatMap,
	start time.Time,
	end time.Time,
	target uint,
) HistogramByWeekdayRow {
	assert.True(func() bool { return chm != nil })
	assert.True(func() bool { return start.Before(end) })
	assert.True(func() bool { return target > 0 })

	var dist WeekDistribution
	week := xtime.RangeByWeeks(start, 0, false).Shift(-xtime.Day) // shift Sunday
	day, weekday := start, start.Weekday()
	for cursor := week.From(); cursor.Before(end); {
		for i := time.Sunday; i <= time.Saturday; i++ {
			dist[i] = chm[cursor]
			cursor = cursor.Add(xtime.Day)
		}
		suggestion, value := dist.Suggest(weekday, target)
		if suggestion == -1 {
			day, weekday = cursor, time.Sunday
			continue
		}
		return HistogramByWeekdayRow{
			Day: day.Add(xtime.Day * time.Duration(suggestion-weekday)),
			Sum: value,
		}
	}
	return HistogramByWeekdayRow{}
}

type WeekDistribution [7]uint

func (week WeekDistribution) Suggest(since time.Weekday, basis uint) (time.Weekday, uint) {
	value := week.max()
	if value < basis {
		value = basis
	}
	for i := since; i <= time.Saturday; i++ {
		if week[i] < value {
			return i, value
		}
	}
	return -1, value
}

func (week WeekDistribution) min() uint {
	min := week[time.Sunday]
	for i := time.Monday; i <= time.Saturday; i++ {
		if week[i] < min {
			min = week[i]
		}
	}
	return min
}

func (week WeekDistribution) max() uint {
	max := week[time.Sunday]
	for i := time.Monday; i <= time.Saturday; i++ {
		if week[i] > max {
			max = week[i]
		}
	}
	return max
}
