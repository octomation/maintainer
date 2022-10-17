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
// and returns an appropriate day to contribute.
//
// Will normalize dates to UTC.
func Suggest(
	heats HeatMap,
	scope xtime.Range,
	hours xtime.Schedule,
	basis uint,
) Suggestion {
	assert.True(func() bool { return heats != nil })
	assert.True(func() bool { return !scope.IsZero() })
	assert.True(func() bool { return hours != nil })
	assert.True(func() bool { return basis > 0 })

	// normalize dates to UTC
	since, until := scope.From().UTC(), scope.To().UTC()
	suggestion := since

	var dist WeekDistribution
	day := xtime.TruncateToDay(since)
	week := xtime.GregorianWeeks(since, 0, false)
STEP:
	for cursor := week.From(); cursor.Before(until); {
		for i := time.Sunday; i <= time.Saturday; i++ {
			dist[i] = heats.Count(cursor)
			cursor = cursor.Add(xtime.Day)
		}

		for i := day.Weekday(); i <= time.Saturday; i++ {
			suggested, value := dist.Suggest(i, basis)
			if suggested == -1 {
				day = cursor
				continue STEP
			}
			if delta := suggested - day.Weekday(); delta > 0 {
				day = day.Add(time.Duration(delta) * xtime.Day)
			}
			if day.After(suggestion) {
				suggestion = day
			}
			suggestion = hours.Suggest(suggestion)
			if suggestion.IsZero() {
				i = suggested
				continue
			}
			return Suggestion{Time: suggestion, Actual: heats.Count(day), Target: value}
		}
		day = cursor
	}
	return Suggestion{Time: suggestion, Actual: 0, Target: basis}
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
