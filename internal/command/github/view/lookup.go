package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alexeyco/simpletable"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type WeekReport struct {
	Number int
	Report map[time.Weekday]int
}

func Lookup(r xtime.Range, data []WeekReport, printer interface{ Println(...interface{}) }) error {
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Day / Week"},
		},
	}
	for _, report := range data {
		table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
			Align: simpletable.AlignCenter,
			Text:  fmt.Sprintf("#%d", report.Number),
		})
	}

	for i := time.Sunday; i <= time.Saturday; i++ {
		row := make([]*simpletable.Cell, 0, 4)
		row = append(row, &simpletable.Cell{Text: i.String()})
		for _, week := range data {
			txt := "-"
			count, present := week.Report[i]
			if count > 0 {
				txt = strconv.Itoa(week.Report[i])
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
					r.From().Format(xtime.RFC3339Day),
					r.To().Format(xtime.RFC3339Day),
				),
			},
		},
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	return nil
}
