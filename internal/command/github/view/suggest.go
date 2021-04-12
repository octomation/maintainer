package view

import (
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func Suggest(
	printer interface{ Println(...interface{}) },
	scope time.Range,
	histogram []contribution.HistogramByWeekdayRow,
	suggest contribution.HistogramByWeekdayRow,
	current int,
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
				Text: fmt.Sprintf("Contributions for %s: %dd, %[4]d -> %[3]d",
					suggest.Day.Format(time.RFC3339Day),
					suggest.Day.Sub(time.Now().UTC())/time.Day,
					suggest.Sum,
					current,
				),
			},
		},
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	return nil
}
