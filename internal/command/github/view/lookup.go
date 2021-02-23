package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alexeyco/simpletable"
)

type WeekReport struct {
	Number int
	Report map[time.Weekday]int
}

func Lookup(printer interface{ Println(...interface{}) }, data []WeekReport) error {
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
			row = append(row, &simpletable.Cell{Text: txt})
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	return nil
}
