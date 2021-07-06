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

	// shift Sunday to the right and cleanup empty weeks
	last := len(report) - 1
	if last > -1 {
		_, week := scope.To().Add(time.Week).ISOWeek()
		report = append(report, WeekReport{
			Number: week,
			Report: make(map[time.Weekday]int),
		})
		for i := last + 1; i > 0; i-- {
			if count, present := report[i-1].Report[time.Sunday]; present {
				report[i].Report[time.Sunday] = count
				delete(report[i-1].Report, time.Sunday)
			}
		}
	}
	cleaned := make([]WeekReport, 0, len(report))
	for _, row := range report {
		if len(row.Report) > 0 {
			cleaned = append(cleaned, row)
		}
	}
	report = cleaned

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

	// shift Sunday to the right and cleanup empty weeks
	last := len(report) - 1
	if last > -1 {
		_, week := heatmap.To().Add(time.Week).ISOWeek()
		report = append(report, WeekReport{
			Number: week,
			Report: make(map[time.Weekday]int),
		})
		for i := last + 1; i > 0; i-- {
			if count, present := report[i-1].Report[time.Sunday]; present {
				report[i].Report[time.Sunday] = count
				delete(report[i-1].Report, time.Sunday)
			}
		}
	}
	cleaned := make([]WeekReport, 0, len(report))
	for _, row := range report {
		if len(row.Report) > 0 {
			cleaned = append(cleaned, row)
		}
	}
	report = cleaned

	return report
}
