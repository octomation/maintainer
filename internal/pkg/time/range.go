package time

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

func NewRange(base, from, to time.Time) Range {
	assert.True(func() bool { return Between(from, to, base) })
	assert.True(func() bool { return from.Location() == base.Location() })
	assert.True(func() bool { return to.Location() == base.Location() })

	return Range{base, from, to}
}

func RangeByWeeks(date time.Time, weeks int, half bool) Range {
	assert.True(func() bool { return !half || (half && weeks > 0) })

	min := TruncateToDay(date)
	max := min.Add(Day - time.Nanosecond)

	const week = 7 // days in week
	day := date.Weekday()
	if day == time.Sunday {
		day = time.Saturday + 1
	}
	monday := int(time.Monday - day)
	sunday := int(time.Saturday - day + 1)

	days := week * weeks
	if weeks < 0 {
		days *= -1 // semantic
		min = min.AddDate(0, 0, monday-days)
		max = max.AddDate(0, 0, sunday)
		return NewRange(date, min, max)
	}

	if half {
		days = week * (weeks / 2)
		min = min.AddDate(0, 0, monday-days)
		max = max.AddDate(0, 0, sunday+days)
	} else {
		min = min.AddDate(0, 0, monday)
		max = max.AddDate(0, 0, sunday+days)
	}
	return NewRange(date, min, max)
}

func RangeByMonths(date time.Time, months int, half bool) Range {
	assert.True(func() bool { return !half || (half && months > 0) })

	min := TruncateToMonth(date)
	max := min.AddDate(0, 1, 0).Add(-time.Nanosecond)

	if months < 0 {
		min = min.AddDate(0, months, 0)
		return NewRange(date, min, max)
	}

	if half {
		min = min.AddDate(0, -months/2, 0)
		max = max.AddDate(0, months/2, 0)
	} else {
		max = max.AddDate(0, months, 0)
	}
	return NewRange(date, min, max)
}

func RangeByYears(date time.Time, years int, half bool) Range {
	assert.True(func() bool { return !half || (half && years > 0) })

	min := TruncateToYear(date)
	max := min.AddDate(1, 0, 0).Add(-time.Nanosecond)

	if years < 0 {
		min = min.AddDate(years, 0, 0)
		return NewRange(date, min, max)
	}

	if half {
		min = min.AddDate(-years/2, 0, 0)
		max = max.AddDate(years/2, 0, 0)
	} else {
		max = max.AddDate(years, 0, 0)
	}
	return NewRange(date, min, max)
}

type Range struct{ base, from, to time.Time }

func (r Range) Base() time.Time { return r.base }
func (r Range) From() time.Time { return r.from }
func (r Range) To() time.Time   { return r.to }

func (r Range) Contains(t time.Time) bool {
	return Between(r.from, r.to, t)
}

func (r Range) ExcludeFuture() Range {
	if now := time.Now(); now.Before(r.to) {
		r.to = now
	}
	return r
}

func (r Range) ExpandLeft(t time.Time) Range {
	assert.True(func() bool { return t.Before(r.from) })

	r.from = t
	return r
}

func (r Range) ExpandRight(t time.Time) Range {
	assert.True(func() bool { return t.After(r.to) })

	r.to = t
	return r
}

func (r Range) Shift(shift time.Duration) Range {
	assert.True(func() bool { return Between(r.from.Add(shift), r.to.Add(shift), r.base) })

	r.from = r.from.Add(shift)
	r.to = r.to.Add(shift)
	return r
}
