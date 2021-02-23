package time

import "time"

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

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

func RangeByWeeks(t time.Time, weeks int) (time.Time, time.Time) {
	min := TruncateToDay(t)
	max := min.AddDate(0, 0, 1).Add(-time.Nanosecond)

	if weeks > 0 {
		day, days := t.Weekday(), 7*(weeks/2)
		min = min.AddDate(0, 0, -(int(day-time.Sunday) + days))
		max = max.AddDate(0, 0, int(time.Saturday-day)+days)
	}

	return min, max
}
