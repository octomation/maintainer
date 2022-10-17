package view

import (
	"fmt"
	"time"

	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// TODO:refactor simplify and remove the implementation

func ContributionDiff(
	printer interface{ Println(...interface{}) },
	heatmap contribution.HeatMap,
	base, head string,
) error {
	data := prepare(heatmap)
	table := simpletable.New()

	if len(data) == 0 {
		printer.Println(fmt.Sprintf("There is no diff between head{%q} → base{%q}", head, base))
		return nil
	}

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Day / Week"},
		},
	}
	for _, week := range data {
		table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
			Align: simpletable.AlignCenter,
			Text:  fmt.Sprintf("#%d", week.Number),
		})
	}

	for i := time.Sunday; i <= time.Saturday; i++ {
		row := make([]*simpletable.Cell, 0, len(data)+1)
		row = append(row, &simpletable.Cell{Text: i.String()})
		for _, week := range data {
			txt := "-"
			if count := week.Report[i]; count != 0 {
				txt = fmt.Sprintf("%+d", count)
			}
			row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: txt})
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{
				Span: len(table.Header.Cells),
				Text: fmt.Sprintf("The diff between head{%q} → base{%q}", head, base),
			},
		},
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	return nil
}

// TODO:refactor simplify and remove the implementation

func prepare(heatmap contribution.HeatMap) []WeekReport {
	report := make([]WeekReport, 0, 8)

	start := xtime.TruncateToWeek(heatmap.From())
	for week, end := start, heatmap.To(); week.Before(end); week = week.Add(xtime.Week) {
		subset := heatmap.Subset(xtime.GregorianWeeks(week, 0, false))
		if len(subset) == 0 {
			continue
		}

		_, num := week.ISOWeek()
		row := WeekReport{
			Number: num,
			Report: make(map[time.Weekday]uint, len(subset)),
		}
		for ts, count := range subset {
			row.Report[ts.Weekday()] = count
		}
		report = append(report, row)
	}

	// shift Sunday to the right and cleanup empty weeks
	last := len(report) - 1
	if last > -1 {
		_, week := heatmap.To().Add(xtime.Week).ISOWeek()
		report = append(report, WeekReport{
			Number: week,
			Report: make(map[time.Weekday]uint),
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

type WeekReport struct {
	Number int
	Report map[time.Weekday]uint
}
