package time

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

func NewRange(from, to time.Time) Range {
	assert.True(func() bool { return !from.IsZero() })
	assert.True(func() bool { return from.Before(to) })

	return Range{from, to}
}

func RangeByWeeks(date time.Time, weeks int, half bool) Range {
	assert.True(func() bool { return !half || (half && weeks > 0) })

	min := TruncateToDay(date)
	max := min.Add(Day - time.Nanosecond)

	const week = 7
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
		return NewRange(min, max)
	}

	if half {
		days = week * (weeks / 2)
		min = min.AddDate(0, 0, monday-days)
		max = max.AddDate(0, 0, sunday+days)
	} else {
		min = min.AddDate(0, 0, monday)
		max = max.AddDate(0, 0, sunday+days)
	}
	return NewRange(min, max)
}

func GregorianWeeks(date time.Time, weeks int, half bool) Range {
	r := RangeByWeeks(date, weeks, half)
	if date.Weekday() == time.Sunday {
		return r.Shift(6 * Day)
	}
	return r.Shift(-Day)
}

func RangeByMonths(date time.Time, months int, half bool) Range {
	assert.True(func() bool { return !half || (half && months > 0) })

	min := TruncateToMonth(date)
	max := min.AddDate(0, 1, 0).Add(-time.Nanosecond)

	if months < 0 {
		min = min.AddDate(0, months, 0)
		return NewRange(min, max)
	}

	if half {
		min = min.AddDate(0, -months/2, 0)
		max = max.AddDate(0, months/2, 0)
	} else {
		max = max.AddDate(0, months, 0)
	}
	return NewRange(min, max)
}

func RangeByYears(date time.Time, years int, half bool) Range {
	assert.True(func() bool { return !half || (half && years > 0) })

	min := TruncateToYear(date)
	max := min.AddDate(1, 0, 0).Add(-time.Nanosecond)

	if years < 0 {
		min = min.AddDate(years, 0, 0)
		return NewRange(min, max)
	}

	if half {
		min = min.AddDate(-years/2, 0, 0)
		max = max.AddDate(years/2, 0, 0)
	} else {
		max = max.AddDate(years, 0, 0)
	}
	return NewRange(min, max)
}

type Range struct{ from, to time.Time }

func (r Range) From() time.Time { return r.from }
func (r Range) To() time.Time   { return r.to }

func (r Range) IsZero() bool              { return r.from.IsZero() }
func (r Range) Contains(t time.Time) bool { return Between(r.from, r.to, t) }

func (r Range) Since(t time.Time) Range { return NewRange(t, r.to) }
func (r Range) Until(t time.Time) Range { return NewRange(r.from, t) }

func (r Range) Shift(shift time.Duration) Range {
	return NewRange(r.from.Add(shift), r.to.Add(shift))
}

func (r Range) ExcludeFuture() Range {
	if now := time.Now(); now.Before(r.to) {
		return NewRange(r.from, now)
	}
	return r
}
