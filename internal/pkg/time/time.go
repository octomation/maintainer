package time

import "time"

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

const (
	RFC3339Day   = "2006-01-02"
	RFC3339Month = "2006-01"
	RFC3339Year  = "2006"
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

func CopyClock(from, to time.Time) time.Time {
	to = TruncateToDay(to)

	h, m, s := from.Clock()
	var delta time.Duration
	delta += time.Hour * time.Duration(h)
	delta += time.Minute * time.Duration(m)
	delta += time.Second * time.Duration(s)

	return to.Add(delta)
}

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
	return Year(y).Location(t.Location()).Time()
}

func UTC() Builder {
	return Builder{mm: 1, dd: 1, l: time.UTC}
}

func Year(year int) Builder {
	return Builder{yyyy: year, mm: 1, dd: 1}
}

type Builder struct {
	yyyy int
	mm   time.Month
	dd   int
	h    int
	m    int
	s    int
	ns   int
	l    *time.Location
	d    time.Duration
}

func (b Builder) Year(year int) Builder {
	b.yyyy = year
	return b
}

func (b Builder) Month(month time.Month) Builder {
	b.mm = month
	return b
}

func (b Builder) Day(day int) Builder {
	b.dd = day
	return b
}

func (b Builder) Hour(hour int) Builder {
	b.h = hour
	return b
}

func (b Builder) Minute(minute int) Builder {
	b.m = minute
	return b
}

func (b Builder) Second(second int) Builder {
	b.s = second
	return b
}

func (b Builder) Nanosecond(ns int) Builder {
	b.ns = ns
	return b
}

func (b Builder) Location(loc *time.Location) Builder {
	b.l = loc
	return b
}

func (b Builder) Add(d time.Duration) Builder {
	b.d += d
	return b
}

func (b Builder) Time() time.Time {
	if b.l == nil {
		b.l = time.Local
	}
	t := time.Date(b.yyyy, b.mm, b.dd, b.h, b.m, b.s, b.ns, b.l)
	if b.d != 0 {
		t = t.Add(b.d)
	}
	return t
}

func (b Builder) Format(layout string) string {
	return b.Time().Format(layout)
}
