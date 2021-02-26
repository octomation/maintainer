package time

import "time"

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
//  - [min, max]: Between(u, min, max)
//  - (min, max): min.Before(u) && max.After(u)
//  - [min, max): BeforeOrEqual(min, u) && max.After(u)
//  - (min, max]: min.Before(u) && AfterOrEqual(max, u)
//
func Between(u, min, max time.Time) bool {
	return BeforeOrEqual(min, u) && AfterOrEqual(max, u)
}

type Range struct {
	// invariants:
	//  - from < to, from and to have the same time.Location
	//  - from and to are zero only both, Range is zero in this case
	//  - from and to could be equal, Range is zero in this case
	from, to time.Time
}

func (r Range) From() time.Time {
	return r.from
}

func (r Range) To() time.Time {
	return r.to
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

func RangeByWeeks(t time.Time, weeks int) Range {
	min := TruncateToDay(t)
	max := min.AddDate(0, 0, 1).Add(-time.Nanosecond)

	if weeks > 0 {
		day, days := t.Weekday(), 7*(weeks/2)
		min = min.AddDate(0, 0, int(time.Monday-day)-days)
		max = max.AddDate(0, 0, int(time.Saturday+1-day)+days)
	}

	return Range{min, max}
}
