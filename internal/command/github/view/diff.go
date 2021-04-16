package view

import (
	"fmt"
	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func Diff(
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
		// TODO:unclear explain
		if i == 0 {
			row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: "-"})
			continue
		}
		txt := "-"
		if count := data[i-1].Report[time.Sunday]; count != 0 {
			txt = fmt.Sprintf("%+d", count)
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
