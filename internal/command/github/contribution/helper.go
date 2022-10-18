package contribution

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func Datetime(t time.Time) string {
	now := time.Now().In(t.Location())
	sign := "-"
	if t.After(now) {
		sign = "+"
	}

	days := t.Sub(now) / xtime.Day
	if days < 0 {
		days = -days
	}
	tail := t.Sub(now) % xtime.Day
	if tail < 0 {
		tail = -tail
	}
	normalized := strings.ToUpper(tail.Truncate(time.Second).String())

	if days > 0 {
		return fmt.Sprintf("%s%dd%s", sign, days, normalized)
	}
	return fmt.Sprintf("%s%s", sign, normalized)
}

func FallbackDate(args []string) time.Time {
	fallback := time.Now()
	if len(args) > 0 {
		raw := strings.Split(args[0], "/")
		rawDate := raw[0]
		if rawDate != "" && rawDate != "git" {
			return fallback
		}
	}

	repo, err := git.PlainOpenWithOptions("", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return fallback
	}
	head, err := repo.Head()
	if err != nil {
		return fallback
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fallback
	}
	return commit.Author.When
}

func ParseDate(
	args []string,
	defaultDate time.Time,
	defaultWeeks int,
) (contribution.DateOptions, error) {
	// trick to skip length check
	args = append(args, "")

	var (
		opts contribution.DateOptions
		err  error
	)
	var rawDate, rawWeeks string
	raw := strings.Split(args[0], "/")
	switch len(raw) {
	case 2:
		rawDate, rawWeeks = raw[0], raw[1]
	case 1:
		rawDate, rawWeeks = raw[0], ""
	default:
		return opts, fmt.Errorf("too many parts")
	}

	var date time.Time
	switch l := len(rawDate); {
	case rawDate == "" || rawDate == "git":
		date = defaultDate
	case rawDate == "now":
		date = time.Now()
	case l == len(xtime.YearOnly):
		date, err = time.Parse(xtime.YearOnly, rawDate)
	case l == len(xtime.YearAndMonth):
		date, err = time.Parse(xtime.YearAndMonth, rawDate)
	case l == len(xtime.DateOnly):
		date, err = time.Parse(xtime.DateOnly, rawDate)
	case l == 20 || l == len(time.RFC3339):
		date, err = time.Parse(time.RFC3339, rawDate)
	default:
		err = fmt.Errorf("unsupported format")
	}
	if err != nil {
		return opts, fmt.Errorf("parse date %q: %w", rawDate, err)
	}
	opts.Value = date

	var weeks = defaultWeeks
	if rawWeeks != "" {
		weeks, err = strconv.Atoi(rawWeeks)
		if err != nil {
			return opts, fmt.Errorf("parse weeks %q: %w", rawWeeks, err)
		}
		// +%d and positive %d have the same value, but different semantic
		// invariant: len(rawWeeks) > 0, because weeks > 0
		if weeks > 0 && rawWeeks[0] != '+' {
			opts.Half = true
		}
	} else {
		opts.Half = true
	}
	opts.Weeks = weeks

	return opts, nil
}

func TableView(
	cmd *cobra.Command,
	heats contribution.HeatMap,
	scope xtime.Range,
	opts ...func(time.Time, string) string,
) {
	assert.True(func() bool { return scope.From().Weekday() == time.Sunday })

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Day / Week"},
		},
	}
	var weeks int
	for i := scope.From(); i.Before(scope.To()); i = i.Add(xtime.Week) {
		_, week := i.ISOWeek()
		table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
			Align: simpletable.AlignCenter,
			Text:  fmt.Sprintf("#%02d", week+1), // Gregorian correction, see LookupRange
		})
		weeks++
	}
	table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
		Align: simpletable.AlignCenter,
		Text:  "Date",
	})

	for i, cursor := time.Sunday, scope.From(); i <= time.Saturday; i++ {
		row := append(make([]*simpletable.Cell, 0, weeks+1), &simpletable.Cell{Text: i.String()})
		for j := 0; j < weeks; j++ {
			cell := cursor.Add(time.Duration(j) * xtime.Week)

			count := heats.Count(cell)
			text := "-"
			if count > 0 {
				text = strconv.FormatUint(uint64(count), 10)
			} else if cell.After(scope.To()) {
				text = "?"
			}
			for _, opt := range opts {
				text = opt(cell, text)
			}

			row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: text})
			if j+1 == weeks {
				row = append(row, &simpletable.Cell{
					Align: simpletable.AlignCenter,
					Text:  cell.Format(xtime.DayStamp),
				})
			}
		}
		cursor = cursor.Add(xtime.Day)
		table.Body.Cells = append(table.Body.Cells, row)
	}

	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{
				Align: simpletable.AlignRight,
				Span:  len(table.Header.Cells),
				Text:  "Stats: coming soon",
			},
		},
	}
	table.SetStyle(simpletable.StyleCompactLite)
	cmd.PrintErrln("\n" + table.String() + "\n")
}
