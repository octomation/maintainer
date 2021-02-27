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
	// $ maintainer github contribution histogram 2013
	//
	//  1 #######
	//  2 ######
	//  3 ###
	//  4 #
	//  7 ##
	//  8 #
	//
	// $ maintainer github contribution histogram 2013-11
	//
	histogram := cobra.Command{
		Use:  "histogram",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(&histogram)

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
	// $ maintainer github contribution lookup            # -> now()/3
	// $ maintainer github contribution lookup 2013-12-03 # -> 2013-12-03/3
	// $ maintainer github contribution lookup now/3      # -> now()/3
	// $ maintainer github contribution lookup /3         # -> now()/3
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
					if raw[0] != "now" && raw[0] != "" {
						date, err = time.Parse(xtime.RFC3339Day, raw[0])
					}
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
			data := contribution.HistogramByWeekday(chm.Subset(scope), false)
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
				if len(data) > 0 {
					row := data[0]
					if row.Day.Equal(i) {
						data = data[1:]
						count = row.Sum
					}
				}
				report[idx].Report[i.Weekday()] = count
			}

			return view.Lookup(scope, report, cmd)
		},
	}
	cmd.AddCommand(&lookup)

	//
	// $ maintainer github contribution suggest 2013
	//
	//   Day / Week   #45   #46   #47   #48   #49   #50
	//  ------------ ----- ----- ----- ----- ----- -----
	//   Sunday        -     -     -     1     -     -
	//   Monday        -     -     -     2     1     2
	//   Tuesday       -     -     -     8     1     -
	//   Wednesday     -     1     1     -     3     -
	//   Thursday      -     -     3     7     1     7
	//   Friday        -     -     -     1     2     -
	//   Saturday      -     -     -     -     -     -
	//  ------------ ----- ----- ----- ----- ----- -----
	//   Contributions for 2013-11-10: -154d, 0 -> 5
	//
	// $ maintainer github contribution suggest 2013-11
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

	return &cmd
}
