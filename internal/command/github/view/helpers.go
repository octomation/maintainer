package view

import (
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

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
