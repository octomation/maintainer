package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type SuggestOption struct {
	Suggest contribution.HistogramByWeekdayRow
	Current uint
	Delta   bool
	Short   bool
}

// TODO:refactoring combine with Lookup, use HeatMap as input
// TODO:refactoring extract "table builder", compare with others views

func Suggest(
	printer interface{ Println(...interface{}) },

	scope xtime.Range,
	histogram []contribution.HistogramByWeekdayRow,

	option SuggestOption,
) error {
	now := time.Now().UTC()

	var suggestion string
	if option.Delta {
		suggestion = fmt.Sprintf("%dd", option.Suggest.Day.Sub(now)/xtime.Day)
	} else {
		day := xtime.CopyClock(now, option.Suggest.Day).In(time.Local)
		suggestion = day.Format(time.RFC3339)
	}
	if option.Short {
		printer.Println(suggestion)
		return nil
	} else if option.Delta {
		suggestion = fmt.Sprintf("%s: %s",
			option.Suggest.Day.Format(xtime.RFC3339Day),
			suggestion,
		)
	}

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
				txt = strconv.FormatUint(uint64(count), 10)
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
				Text: fmt.Sprintf("Suggestion is %s, %d → %d",
					suggestion,
					option.Current,
					option.Suggest.Sum,
				),
			},
		},
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	return nil
}
