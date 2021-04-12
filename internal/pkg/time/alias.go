package time

import "time"

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

func Now() time.Time { return time.Now() }

func Parse(layout, value string) (time.Time, error) { return time.Parse(layout, value) }

type Time = time.Time

type Weekday = time.Weekday
