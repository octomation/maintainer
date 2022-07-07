package contribution

import (
	"time"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// Suggest finds a week with gaps in the contribution heatmap
// and returns an appropriate day to contribute.
func Suggest(
	chm HeatMap,
	start time.Time,
	end time.Time,
	basis int,
) HistogramByWeekdayRow {
	defaults := HistogramByWeekdayRow{
		Day: start,
		Sum: basis,
	}

	for t := start; t.Before(end); t = t.Add(xtime.Week) {
		week := xtime.RangeByWeeks(t, 0, false).Shift(-xtime.Day) // shift Sunday
		data := HistogramByCount(chm.Subset(week), OrderByCount)
		sunday := week.From()

		// good week: no gaps and enough contributions
		if len(data) == 1 && data[0].Count >= defaults.Sum {
			continue
		}

		// bad week: no contributions
		if len(data) == 0 {
			return HistogramByWeekdayRow{Day: sunday, Sum: defaults.Sum}
		}

		// otherwise, we choose the maximum amount of contributions
		// it's the last element in the histogram because it's sorted by count ASC
		count := data[len(data)-1].Count
		if count < defaults.Sum {
			count = defaults.Sum
		}
		// and try to find an appropriate day to contribute
		day := sunday
		for i := time.Sunday; i <= time.Saturday; i++ {
			if chm[day] != count {
				break
			}
			day = day.Add(xtime.Day)
		}
		return HistogramByWeekdayRow{Day: day, Sum: count}
	}

	return defaults
}
