package time

import "time"

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

func Year(year int) Builder {
	return Builder{yyyy: year, mm: 1, dd: 1}
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
