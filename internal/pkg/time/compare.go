package time

import "time"

func AfterOrEqual(t, u time.Time) bool {
	return t.After(u) || t.Equal(u)
}

func BeforeOrEqual(t, u time.Time) bool {
	return t.Before(u) || t.Equal(u)
}

func Between(u, min, max time.Time) bool {
	return BeforeOrEqual(min, u) && AfterOrEqual(max, u)
}
