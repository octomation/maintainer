package view

import (
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type WeekReport struct {
	Number int
	Report map[time.Weekday]int
}

func convert(
	scope time.Range,
	histogram []contribution.HistogramByWeekdayRow,
) []WeekReport {
	report := make([]WeekReport, 0, 4)
	prev, idx := 0, -1
	for day := scope.From(); day.Before(scope.To()); day = day.Add(time.Day) {
		_, week := day.ISOWeek()
		if week != prev {
			prev = week
			idx++
		}

		if len(report) < idx+1 {
			report = append(report, WeekReport{
				Number: week,
				Report: make(map[time.Weekday]int),
			})
		}

		var count int
		if len(histogram) > 0 {
			row := histogram[0]
			if row.Day.Equal(day) {
				histogram = histogram[1:]
				count = row.Sum
			}
		}
		report[idx].Report[day.Weekday()] = count
	}
	return report
}

func prepare(heatmap contribution.HeatMap) []WeekReport {
	report := make([]WeekReport, 0, 8)

	start := time.TruncateToWeek(heatmap.From())
	for week, end := start, heatmap.To(); week.Before(end); week = week.Add(time.Week) {
		subset := heatmap.Subset(time.RangeByWeeks(week, 0, false).Shift(-time.Day))
		if len(subset) == 0 {
			continue
		}

		_, num := week.ISOWeek()
		row := WeekReport{
			Number: num,
			Report: make(map[time.Weekday]int, len(subset)),
		}
		for ts, count := range subset {
			row.Report[ts.Weekday()] = count
		}
		report = append(report, row)
	}

	return report
}

// If it's a first-week report with a single entry for Sunday,
// we skip it completely.
//
// It's because GitHub shows the contribution chart started on Sunday
// of the previous week. For that reason we have to shift it to the right
// and compensate `.Shift(-time.Day)` call for the scope.
func shiftIsNeeded(idx int, report map[time.Weekday]int) bool {
	_, is := report[time.Sunday]
	return idx == 0 && len(report) == 1 && is
}
