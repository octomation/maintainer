package github

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Contribution(cnf *config.Tool) *cobra.Command {
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
	// $ maintainer github contribution histogram 2013-11    # month
	// $ maintainer github contribution histogram 2013-11-20 # week
	//
	histogram := cobra.Command{
		Use:  "histogram",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			construct, date := xtime.RangeByWeeks, time.Now().UTC()

			// input validation: date(year,+month,+week{day})
			if len(args) == 1 {
				var err error
				wrap := func(err error) error {
					return fmt.Errorf(
						"please provide argument in format YYYY[-mm[-dd]], e.g., 2006-01: %w",
						fmt.Errorf("invalid argument %q: %w", args[0], err),
					)
				}

				switch input := args[0]; len(input) {
				case len(xtime.RFC3339Year):
					date, err = time.Parse(xtime.RFC3339Year, input)
					construct = xtime.RangeByYears
				case len(xtime.RFC3339Month):
					date, err = time.Parse(xtime.RFC3339Month, input)
					construct = xtime.RangeByMonths
				case len(xtime.RFC3339Day):
					date, err = time.Parse(xtime.RFC3339Day, input)
				default:
					err = fmt.Errorf("unsupported format")
				}
				if err != nil {
					return wrap(err)
				}
			}

			// data provisioning
			scope := construct(date, 0, false).Shift(-xtime.Day).ExcludeFuture()
			chm, err := service.ContributionHeatMap(cmd.Context(), scope)
			if err != nil {
				return err
			}

			// data presentation
			data := contribution.HistogramByCount(chm, contribution.OrderByCount)
			for _, row := range data {
				fmt.Printf("%3d %s\n", row.Count, strings.Repeat("#", row.Frequency))
			}
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
	// $ maintainer github contribution lookup            # -> now()/-1
	// $ maintainer github contribution lookup 2013-12-03 # -> 2013-12-03/-1
	// $ maintainer github contribution lookup now/3      # -> now()/3 == now()/-1
	// $ maintainer github contribution lookup /3         # -> now()/3 == now()/-1
	//
	lookup := cobra.Command{
		Use:  "lookup",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			date, weeks, half := time.Now().UTC(), -1, false

			// input validation: date/{+-}weeks
			if len(args) == 1 {
				var err error
				wrap := func(err error) error {
					return fmt.Errorf(
						"please provide argument in format YYYY-mm-dd[/[+|-]weeks], e.g., 2006-01-02/3: %w",
						fmt.Errorf("invalid argument %q: %w", args[0], err),
					)
				}

				raw := strings.Split(args[0], "/")
				switch len(raw) {
				case 2:
					weeks, err = strconv.Atoi(raw[1])
					if err != nil {
						return wrap(err)
					}
					// +%d and positive %d have the same value, but different semantic
					// invariant: len(raw[1]) > 0, because weeks > 0 and invariant(time.RangeByWeeks)
					if weeks > 0 && raw[1][0] != '+' {
						half = true
					}
					fallthrough
				case 1:
					if raw[0] != "now" && raw[0] != "" {
						date, err = time.Parse(xtime.RFC3339Day, raw[0])
					}
					if err != nil {
						return wrap(err)
					}
				default:
					return wrap(fmt.Errorf("too many parts"))
				}
			}

			// data provisioning
			scope := xtime.RangeByWeeks(date, weeks, half).Shift(-xtime.Day).ExcludeFuture()
			chm, err := service.ContributionHeatMap(cmd.Context(), scope)
			if err != nil {
				return err
			}

			// data presentation
			data := contribution.HistogramByWeekday(chm, false)
			return view.Lookup(cmd, scope, data)
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
	// $ maintainer github contribution suggest 2013-11-20
	//
	suggest := cobra.Command{
		Use:  "suggest",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			date := xtime.TruncateToYear(time.Now().UTC())
			weeks, target := 5, 5 // TODO:magic replace by params

			// input validation: date(year,+month,+week{day})
			if len(args) == 1 {
				var err error
				wrap := func(err error) error {
					return fmt.Errorf(
						"please provide argument in format YYYY[-mm[-dd]], e.g., 2006-01: %w",
						fmt.Errorf("invalid argument %q: %w", args[0], err),
					)
				}

				switch input := args[0]; len(input) {
				case len(xtime.RFC3339Year):
					date, err = time.Parse(xtime.RFC3339Year, input)
				case len(xtime.RFC3339Month):
					date, err = time.Parse(xtime.RFC3339Month, input)
				case len(xtime.RFC3339Day):
					date, err = time.Parse(xtime.RFC3339Day, input)
					date = xtime.TruncateToWeek(date)
				default:
					err = fmt.Errorf("unsupported format")
				}
				if err != nil {
					return wrap(err)
				}
			}

			// data provisioning
			start := xtime.TruncateToWeek(date) // Monday
			scope := xtime.NewRange(
				start.Add(-2*xtime.Week-xtime.Day), // buffer from left side with Sunday
				time.Now().UTC(),
			)
			chm, err := service.ContributionHeatMap(cmd.Context(), scope)
			if err != nil {
				return err
			}

			var suggest contribution.HistogramByWeekdayRow
			standard := contribution.HistogramByWeekdayRow{
				Day: start,
				Sum: target,
			}
			for week, end := start, scope.To(); week.Before(end); week = week.Add(xtime.Week) {
				data := contribution.HistogramByCount(
					chm.Subset(xtime.RangeByWeeks(week, 0, false).Shift(-xtime.Day)), // Sunday
					contribution.OrderByCount,
				)

				// good week
				if len(data) == 1 && data[0].Count >= standard.Sum {
					continue
				}

				// Sunday
				day := week.Add(-xtime.Day)

				// bad week
				if len(data) == 0 {
					suggest.Day = day
					suggest.Sum = standard.Sum
					break
				}

				// otherwise
				target := data[len(data)-1].Count // because it's sorted by frequency
				if target < standard.Sum {
					target = standard.Sum
				}
				suggest.Sum = target
				for i := time.Sunday; i <= time.Saturday; i++ {
					if chm[day] != target {
						suggest.Day = day
						break
					}
					day = day.Add(xtime.Day)
				}
				break
			}

			// data presentation
			area := xtime.RangeByWeeks(suggest.Day, weeks, true).Shift(-xtime.Day) // Sunday
			data := contribution.HistogramByWeekday(chm.Subset(area), false)
			return view.Suggest(cmd, area, data, suggest, chm[suggest.Day])
		},
	}
	cmd.AddCommand(&suggest)

	return &cmd
}
