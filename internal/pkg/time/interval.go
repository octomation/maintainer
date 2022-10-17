package time

import (
	"sort"
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

func Everyday(intervals ...Interval) Schedule {
	assert.True(func() bool { return len(intervals) > 0 })
	sort.Slice(intervals, func(i, j int) bool {
		assert.True(func() bool { return intervals[i].isValid() })
		assert.True(func() bool { return intervals[j].isValid() })
		assert.True(func() bool { return !intervals[i].isIntersected(intervals[j]) })

		return intervals[i].from.volume() < intervals[j].from.volume()
	})

	return Schedule{
		time.Sunday:    intervals,
		time.Monday:    intervals,
		time.Tuesday:   intervals,
		time.Wednesday: intervals,
		time.Thursday:  intervals,
		time.Friday:    intervals,
		time.Saturday:  intervals,
	}
}

func Hours(from, to int, d time.Duration) Interval {
	assert.True(func() bool { return 0 <= from && from < to && to <= 24 })
	assert.True(func() bool { return d >= 0 })

	return Interval{from: Clock{from, 0, 0}, to: Clock{to, 0, 0}, duration: d}
}

type Clock struct{ hour, min, sec int }

func (c Clock) From(t time.Time) Clock {
	h, m, s := t.Clock()
	return Clock{h, m, s}
}

func (c Clock) CopyTo(t time.Time) time.Time {
	var delta time.Duration
	delta += time.Hour * time.Duration(c.hour)
	delta += time.Minute * time.Duration(c.min)
	delta += time.Second * time.Duration(c.sec)
	return TruncateToDay(t).Add(delta)
}

func (c Clock) volume() int { return 60*60*c.hour + 60*c.min + c.sec }

type Interval struct {
	from, to Clock
	duration time.Duration
}

func (i Interval) Contains(t time.Time) bool {
	clock := Clock{}.From(t)
	return i.from.volume() <= clock.volume() && clock.volume() <= i.to.volume()
}

func (i Interval) isValid() bool { return i.from.volume() < i.to.volume() }
func (i Interval) isIntersected(j Interval) bool {
	return i.from.volume() < j.to.volume() && j.from.volume() < i.to.volume()
}

type Schedule map[time.Weekday][]Interval

func (s Schedule) Suggest(t time.Time) time.Time {
	if len(s) == 0 || len(s[t.Weekday()]) == 0 {
		return time.Time{}
	}

	intervals := s[t.Weekday()]
	assert.True(func() bool {
		return sort.SliceIsSorted(intervals, func(i, j int) bool {
			return intervals[i].to.volume() < intervals[j].from.volume()
		})
	})

	clock := Clock{}.From(t)
	for _, interval := range intervals {
		if interval.Contains(t) {
			return t
		}
		if interval.to.volume() < clock.volume() {
			continue
		}
		return interval.from.CopyTo(t)
	}

	return time.Time{}
}
