package contribution

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type Suggestion struct {
	Day    time.Time
	Actual uint
	Target uint
}

// Suggest finds a week with gaps in the contribution heatmap
// and returns an appropriate day to contribute.
//
// Will normalize dates to UTC.
func Suggest(heats HeatMap, since time.Time, until time.Time, basis uint) Suggestion {
	assert.True(func() bool { return heats != nil })
	assert.True(func() bool { return since.Before(until) })
	assert.True(func() bool { return basis > 0 })

	// normalize dates to UTC
	since, until = since.UTC(), until.UTC()

	var dist WeekDistribution
	day, weekday := xtime.TruncateToDay(since), since.Weekday()
	week := ShiftRange(xtime.RangeByWeeks(since, 0, false))
	for cursor := week.From(); cursor.Before(until); {
		for i := time.Sunday; i <= time.Saturday; i++ {
			dist[i] = heats.Count(cursor)
			cursor = cursor.Add(xtime.Day)
		}
		suggestion, value := dist.Suggest(weekday, basis)
		if suggestion == -1 {
			day, weekday = cursor, time.Sunday
			continue
		}
		day = day.Add(xtime.Day * time.Duration(suggestion-weekday))
		return Suggestion{Day: day, Actual: heats.Count(day), Target: value}
	}
	return Suggestion{Day: day, Actual: 0, Target: basis}
}

type WeekDistribution [7]uint

func (week WeekDistribution) Suggest(day time.Weekday, basis uint) (time.Weekday, uint) {
	assert.True(func() bool { return basis > 0 })

	value := week.max()
	if value < basis {
		value = basis
	}
	for i := day; i <= time.Saturday; i++ {
		if week[i] < value {
			return i, value
		}
	}
	return -1, value
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
