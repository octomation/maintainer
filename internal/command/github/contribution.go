package github

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func Contribution(github GitHub) *cobra.Command {
	cmd := cobra.Command{
		Use: "contribution",
	}

	//
	// $ maintainer github contribution suggest --since=2021-01-01
	//
	// https://github.com/kamilsk?tab=overview&from=2021-12-01&to=2021-12-31
	//
	// $('.js-calendar-graph-svg rect.ContributionCalendar-day')
	//   data-date
	//   data-level
	//
	suggest := cobra.Command{
		Use: "suggest",
		RunE: func(cmd *cobra.Command, args []string) error {
			chm, err := github.ContributionHeatMap(cmd.Context(), time.Now())
			if err != nil {
				return err
			}

			cmd.Println("-123d", chm)
			return nil
		},
	}
	cmd.AddCommand(&suggest)

	//
	// $ maintainer github contribution lookup 2013-12-03/9
	//
	//  Day / Week   #45   #46   #47   #48   #49   #50   #51   #52   #1
	// ------------ ----- ----- ----- ----- ----- ----- ----- ----- ----
	//  Sunday        -     -     -     1     -     -     -     -    -
	//  Monday        -     -     -     2     1     2     -     -    -
	//  Tuesday       -     -     -     8     1     -     -     2    -
	//  Wednesday     -     1     1     -     3     -     -     2    ?
	//  Thursday      -     -     3     7     1     7     4     -    ?
	//  Friday        -     -     -     1     2     -     3     2    ?
	//  Saturday      -     -     -     -     -     -     -     -    ?
	// ------------ ----- ----- ----- ----- ----- ----- ----- ----- ----
	//  Contributions are on the range from 2013-11-03 to 2013-12-31
	//
	lookup := cobra.Command{
		Use:  "lookup",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// defaults
			date, weeks := time.Now().Add(-xtime.Week), 3

			if len(args) == 1 {
				var err error
				raw := strings.Split(args[0], "/")
				switch len(raw) {
				case 2:
					weeks, err = strconv.Atoi(raw[1])
					if err != nil {
						return err
					}
					fallthrough
				case 1:
					date, err = time.Parse(xtime.RFC3339Day, raw[0])
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf(
						"please provide in format YYYY-mm-dd[/weeks], e.g., 2006-01-02/3: %w",
						fmt.Errorf("invalid argument %q", args[0]),
					)
				}
			}

			chm, err := github.ContributionHeatMap(cmd.Context(), date)
			if err != nil {
				return err
			}

			scope := xtime.
				RangeByWeeks(date, weeks).
				Shift(-xtime.Day).
				ExcludeFuture().
				TrimByYear(date.Year())
			histogram := contribution.HistogramByWeekday(chm.Subset(scope.From(), scope.To()), false)
			report := make([]view.WeekReport, 0, 4)

			prev, idx := 0, -1
			for i := scope.From(); i.Before(scope.To()); i = i.Add(xtime.Day) {
				_, week := i.ISOWeek()
				if week != prev {
					prev = week
					idx++
				}

				if len(report) < idx+1 {
					report = append(report, view.WeekReport{
						Number: week,
						Report: make(map[time.Weekday]int),
					})
				}

				var count int
				if len(histogram) > 0 {
					row := histogram[0]
					if row.Day.Equal(i) {
						histogram = histogram[1:]
						count = row.Sum
					}
				}
				report[idx].Report[i.Weekday()] = count
			}

			return view.Lookup(scope, report, cmd)
		},
	}
	cmd.AddCommand(&lookup)

	return &cmd
}
