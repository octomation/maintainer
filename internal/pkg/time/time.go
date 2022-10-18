package time

import "time"

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

const (
	DayStamp     = "Jan _2"
	DateOnly     = "2006-01-02"
	YearAndMonth = "2006-01"
	YearOnly     = "2006"
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
//   - [from, to]: Between(u, from, to)
//   - (from, to): from.Before(u) && to.After(u)
//   - [from, to): BeforeOrEqual(from, u) && to.After(u)
//   - (from, to]: from.Before(u) && AfterOrEqual(to, u)
func Between(from, to, u time.Time) bool {
	return BeforeOrEqual(from, u) && AfterOrEqual(to, u)
}

type Transformation func(time.Time) time.Time

func (fn Transformation) Apply(t time.Time) time.Time { return fn(t) }

func TruncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func TruncateToWeek(t time.Time) time.Time {
	day := t.Weekday()
	if day == time.Sunday {
		day = 7
	}
	return TruncateToDay(t).Add(Day * time.Duration(time.Monday-day))
}

func TruncateToMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func TruncateToYear(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, t.Location())
}
