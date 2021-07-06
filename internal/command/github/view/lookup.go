package view

import (
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

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
