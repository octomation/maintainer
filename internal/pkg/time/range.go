package time

import (
	"fmt"
	"time"

	"go.octolab.org/errors"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

// ErrFuturePeriod reports a range that starts at or after now:
// nothing has happened in it yet.
const ErrFuturePeriod = errors.Message("the period starts in the future")

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

// GregorianWeek returns the week of the date, whose weeks start on Sunday:
// the week that contains January 1 is the first one, and the year of a week
// is the year of its Saturday. The date is taken in its own location.
func GregorianWeek(date time.Time) (year, week int) {
	y, m, d := date.Date()
	day := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	sunday := day.AddDate(0, 0, -int(day.Weekday()))

	year = sunday.AddDate(0, 0, 6).Year()
	first := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	first = first.AddDate(0, 0, -int(first.Weekday()))
	return year, int(sunday.Sub(first)/Week) + 1
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

func (r Range) Since(from time.Time) Range { return NewRange(from, r.to) }
func (r Range) Until(to time.Time) Range   { return NewRange(r.from, to) }

func (r Range) Shift(shift time.Duration) Range {
	return NewRange(r.from.Add(shift), r.to.Add(shift))
}

// ExcludeFuture cuts the range at now. A range that starts at or after now
// has no past part, so it fails with ErrFuturePeriod instead.
func (r Range) ExcludeFuture(now time.Time) (Range, error) {
	if !r.from.Before(now) {
		return Range{}, fmt.Errorf("%w: %s, now is %s", ErrFuturePeriod,
			r.from.UTC().Format(time.RFC3339), now.UTC().Format(time.RFC3339))
	}
	if now.Before(r.to) {
		return NewRange(r.from, now), nil
	}
	return r, nil
}
