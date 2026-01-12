package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alexeyco/simpletable"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// ContributionDiff prints a row per changed day: the count in the base,
// the count in the head, and the signed difference between them.
// A day missing from a source is shown as "-" on its side.
//
//	 Day          before   after   diff
//	------------ -------- ------- ------
//	 2013-11-13     1        5      +4
//	 2013-11-14     5        2      -3
//	The diff between base{"file:before.json"} and head{"file:after.json"}
//	head also has 2 days that base lacks, 2013-11-17…2013-11-18, 1 with contributions listed above
//
// The baseOnly and headOnly days are the days only one source covers,
// a line under the table names them, because a day without contributions
// there has no row.
func ContributionDiff(
	printer interface{ Println(...interface{}) },
	diff []contribution.DayDiff,
	baseOnly, headOnly []time.Time,
	base, head string,
) error {
	if len(diff) == 0 {
		printer.Println(fmt.Sprintf("There is no diff between base{%q} and head{%q}", base, head))
		coverage(printer, diff, baseOnly, headOnly)
		return nil
	}

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Day"},
			{Align: simpletable.AlignCenter, Text: "before"},
			{Align: simpletable.AlignCenter, Text: "after"},
			{Align: simpletable.AlignCenter, Text: "diff"},
		},
	}
	for _, day := range diff {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: day.Day.Format(xtime.DateOnly)},
			{Align: simpletable.AlignCenter, Text: count(day.Before, day.InBase)},
			{Align: simpletable.AlignCenter, Text: count(day.After, day.InHead)},
			{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%+d", day.Delta())},
		})
	}

	table.SetStyle(simpletable.StyleCompactLite)
	printer.Println(table.String())
	printer.Println(fmt.Sprintf("The diff between base{%q} and head{%q}", base, head))
	coverage(printer, diff, baseOnly, headOnly)
	return nil
}

// coverage prints a line per source that has days the other one lacks.
// Such a day with contributions is a row of the diff, so they are counted
// by the rows missing from the other side.
func coverage(
	printer interface{ Println(...interface{}) },
	diff []contribution.DayDiff,
	baseOnly, headOnly []time.Time,
) {
	var inBaseOnly, inHeadOnly int
	for _, day := range diff {
		if !day.InHead {
			inBaseOnly++
		}
		if !day.InBase {
			inHeadOnly++
		}
	}
	if line := extra("base", "head", baseOnly, inBaseOnly); line != "" {
		printer.Println(line)
	}
	if line := extra("head", "base", headOnly, inHeadOnly); line != "" {
		printer.Println(line)
	}
}

// extra describes the days one source has and the other one lacks,
// or returns an empty string if there are no such days.
func extra(side, other string, days []time.Time, contributed int) string {
	if len(days) == 0 {
		return ""
	}

	noun, span := "days", days[0].Format(xtime.DateOnly)+"…"+days[len(days)-1].Format(xtime.DateOnly)
	if len(days) == 1 {
		noun, span = "day", days[0].Format(xtime.DateOnly)
	}
	tail := "without contributions"
	if contributed > 0 {
		tail = fmt.Sprintf("%d with contributions listed above", contributed)
	}
	return fmt.Sprintf("%s also has %d %s that %s lacks, %s, %s", side, len(days), noun, other, span, tail)
}

// count returns the count of a day, or "-" if the source has no such day.
func count(value uint, present bool) string {
	if !present {
		return "-"
	}
	return strconv.FormatUint(uint64(value), 10)
}
