package contribution

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type Suggestion struct {
	Time   time.Time
	Actual uint
	Target uint
}

// Suggest finds a week with gaps in the contribution heatmap
// and returns an appropriate moment to contribute: the first one
// between since and until, both inclusive, that is within the schedule
// on a day with fewer contributions than its week requires.
// It returns the zero Suggestion if there is no such moment.
//
// Will normalize dates to UTC.
func Suggest(
	heats HeatMap,
	since, until time.Time,
	hours xtime.Schedule,
	basis uint,
) Suggestion {
	assert.True(func() bool { return heats != nil })
	assert.True(func() bool { return !since.IsZero() })
	assert.True(func() bool { return hours != nil })
	assert.True(func() bool { return basis > 0 })

	// normalize dates to UTC
	since, until = since.UTC(), until.UTC()

	var (
		dist WeekDistribution
		week time.Time
	)
	for day := xtime.TruncateToDay(since); !day.After(until); day = day.Add(xtime.Day) {
		// a Gregorian week starts on Sunday
		if sunday := day.AddDate(0, 0, -int(day.Weekday())); !sunday.Equal(week) {
			week = sunday
			for i := time.Sunday; i <= time.Saturday; i++ {
				dist[i] = heats.Count(week.AddDate(0, 0, int(i)))
			}
		}

		// the day is a gap if the first gap since the day is the day itself
		suggested, value := dist.Suggest(day.Weekday(), basis)
		if suggested != day.Weekday() {
			continue
		}
		suggestion := day
		if suggestion.Before(since) {
			suggestion = since
		}
		suggestion = hours.Suggest(suggestion)
		if suggestion.IsZero() {
			continue // the schedule of the day is over
		}
		if suggestion.After(until) {
			break
		}
		return Suggestion{Time: suggestion, Actual: dist[day.Weekday()], Target: value}
	}
	return Suggestion{}
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
