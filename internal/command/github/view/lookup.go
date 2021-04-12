package view

import (
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type WeekReport struct {
	Number int
	Report map[time.Weekday]int
}

func Lookup(
	printer interface{ Println(...interface{}) },
	scope time.Range,
	histogram []contribution.HistogramByWeekdayRow,
) error {
	data := convert(scope, histogram)
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Day / Week"},
		},
	}
	for i, week := range data {
		if shiftIsNeeded(i, week.Report) {
			continue
		}
		table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
			Align: simpletable.AlignCenter,
			Text:  fmt.Sprintf("#%d", week.Number),
		})
	}

	// shift Sunday to the right
	row := make([]*simpletable.Cell, 0, 4)
	row = append(row, &simpletable.Cell{Text: time.Sunday.String()})
	for i := range data {
		if shiftIsNeeded(i, data[i].Report) {
			continue
		}
		if i == 0 {
			row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: "?"})
			continue
		}
		txt := "-"
		count, present := data[i-1].Report[time.Sunday]
		if count > 0 {
			txt = strconv.Itoa(count)
		} else if !present {
			txt = "?"
		}
		row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: txt})
	}
	table.Body.Cells = append(table.Body.Cells, row)

	for i := time.Monday; i <= time.Saturday; i++ {
		row = make([]*simpletable.Cell, 0, 4)
		row = append(row, &simpletable.Cell{Text: i.String()})
		for j, week := range data {
			if shiftIsNeeded(j, week.Report) {
				continue
			}
			txt := "-"
			count, present := week.Report[i]
			if count > 0 {
				txt = strconv.Itoa(count)
			} else if !present {
				txt = "?"
			}
			row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: txt})
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{
				Span: len(table.Header.Cells),
				Text: fmt.Sprintf("Contributions are on the range from %s to %s",
					scope.From().Format(time.RFC3339Day),
					scope.To().Format(time.RFC3339Day),
				),
			},
		},
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	return nil
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
