package time

import "time"

func CopyClock(from, to time.Time) time.Time {
	to = TruncateToDay(to)

	h, m, s := from.Clock()
	var delta time.Duration
	delta += time.Hour * time.Duration(h)
	delta += time.Minute * time.Duration(m)
	delta += time.Second * time.Duration(s)

	return to.Add(delta)
}
