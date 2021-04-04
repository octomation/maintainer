package time

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

func AfterOrEqual(t, u time.Time) bool {
	return t.After(u) || t.Equal(u)
}

func BeforeOrEqual(t, u time.Time) bool {
	return t.Before(u) || t.Equal(u)
}

// Between returns true if min <= u, u <= max.
// If you want to exclude some border, please use built-in Before or After methods:
//
//  - [from, to]: Between(u, from, to)
//  - (from, to): from.Before(u) && to.After(u)
//  - [from, to): BeforeOrEqual(from, u) && to.After(u)
//  - (from, to]: from.Before(u) && AfterOrEqual(to, u)
//
func Between(from, to, u time.Time) bool {
	return BeforeOrEqual(from, u) && AfterOrEqual(to, u)
}

type Range struct {
	// invariants:
	//  - from < to, from and to have the same time.Location
	//  - from and to are zero only both, Range is zero in this case
	//  - from and to could be equal, Range is zero in this case
	from, to time.Time
}

func (r Range) Contains(t time.Time) bool {
	return Between(r.from, r.to, t)
}

func (r Range) From() time.Time {
	return r.from
}

func (r Range) To() time.Time {
	return r.to
}

func (r Range) ExcludeFuture() Range {
	if now := time.Now(); now.Before(r.to) {
		r.to = now
	}
	return r
}

func (r Range) IsZero() bool {
	return r.from.IsZero() || r.to.IsZero() || r.from.Equal(r.to)
}

func (r Range) Shift(shift time.Duration) Range {
	r.from = r.from.Add(shift)
	r.to = r.to.Add(shift)
	return r
}

func (r Range) TrimByYear(year int) Range {
	if year < r.from.Year() || year > r.to.Year() {
		return Range{}
	}

	if r.from.Year() < year {
		r.from = time.Date(year, 1, 1, 0, 0, 0, 0, r.from.Location())
	}
	if r.to.Year() > year {
		r.to = time.Date(year+1, 1, 1, 0, 0, 0, 0, r.to.Location()).Add(-time.Nanosecond)
	}
	return r
}

func RangeByWeeks(t time.Time, weeks int, half bool) Range {
	assert.True(func() bool { return !half || (half && weeks > 0) })

	min := TruncateToDay(t)
	max := min.Add(Day - time.Nanosecond)

	if weeks == 0 {
		return Range{min, max}
	}

	day, week := t.Weekday(), 7                                      // days in week
	monday, sunday := int(time.Monday-day), int(time.Saturday-day+1) // compensate Sunday

	if weeks < 0 {
		weeks *= -1 // semantic
		min = min.AddDate(0, 0, monday-week*weeks)
		max = max.AddDate(0, 0, sunday)
		return Range{min, max}
	}

	days := week * weeks
	if half {
		days = week * (weeks / 2)
		min = min.AddDate(0, 0, monday-days)
		max = max.AddDate(0, 0, sunday+days)
	} else {
		min = min.AddDate(0, 0, monday)
		max = max.AddDate(0, 0, sunday+days)
	}
	return Range{min, max}
}

func TruncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func TruncateToMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func TruncateToYear(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, 1, 1, 0, 0, 0, 0, t.Location())
}
