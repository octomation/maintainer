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
	Current int
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
	now := xtime.Now().UTC()

	var suggestion string
	if option.Delta {
		suggestion = fmt.Sprintf("%dd", option.Suggest.Day.Sub(now)/xtime.Day)
	} else {
		day := xtime.CopyClock(now, option.Suggest.Day).In(time.Local)
		suggestion = fmt.Sprintf("%s", day.Format(time.RFC3339))
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
