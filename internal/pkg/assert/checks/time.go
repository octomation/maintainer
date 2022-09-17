package checks

func ZeroClock(hour, min, sec int) bool {
	return hour == 0 && min == 0 && sec == 0
}
