package github

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/config/flag"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Contribution(cnf *config.Tool) *cobra.Command {
	cmd := cobra.Command{
		Use: "contribution",
	}

	//
	// $ maintainer github contribution diff --base=/tmp/snap.01.2013.json --head=/tmp/snap.02.2013.json
	//
	//  Day / Week                  #46             #48             #49           #50
	// ---------------------- --------------- --------------- --------------- -----------
	//  Sunday                       -               -               -             -
	//  Monday                       -               -               -             -
	//  Tuesday                      -               -               -             -
	//  Wednesday                   +4               -              +1             -
	//  Thursday                     -               -               -            +1
	//  Friday                       -              +2               -             -
	//  Saturday                     -               -               -             -
	// ---------------------- --------------- --------------- --------------- -----------
	//  The diff between head{"/tmp/snap.02.2013.json"} → base{"/tmp/snap.01.2013.json"}
	//
	// $ maintainer github contribution diff --base=/tmp/snap.01.2013.json 2013
	//
	diff := cobra.Command{
		Use:  "diff",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			date := time.TruncateToYear(time.Now().UTC())

			// input validation: files{params}, date(year){args}
			var baseSource, headSource string
			dst, err := flag.Adopt(cmd.Flags()).GetFile("base")
			if err != nil {
				return err
			}
			if dst == nil {
				return fmt.Errorf("please provide a base file by `--base` parameter")
			}
			baseSource = dst.Name()

			src, err := flag.Adopt(cmd.Flags()).GetFile("head")
			if err != nil {
				return err
			}
			if src == nil && len(args) == 0 {
				return fmt.Errorf("please provide a compared file by `--head` parameter or year in args")
			}
			if src != nil && len(args) > 0 {
				return fmt.Errorf("please omit `--head` or argument, only one of them is allowed")
			}
			if len(args) == 1 {
				var err error
				wrap := func(err error) error {
					return fmt.Errorf(
						"please provide argument in format YYYY, e.g., 2006: %w",
						fmt.Errorf("invalid argument %q: %w", args[0], err),
					)
				}

				switch input := args[0]; len(input) {
				case len(time.RFC3339Year):
					date, err = time.Parse(time.RFC3339Year, input)
				default:
					err = fmt.Errorf("unsupported format")
				}
				if err != nil {
					return wrap(err)
				}
				headSource = fmt.Sprintf("upstream:year(%s)", date.Format(time.RFC3339Year))
			} else {
				headSource = src.Name()
			}

			// data provisioning
			var (
				base contribution.HeatMap
				head contribution.HeatMap
			)
			if err := json.NewDecoder(dst).Decode(&base); err != nil {
				return err
			}
			if src != nil {
				if err := json.NewDecoder(src).Decode(&head); err != nil {
					return err
				}
			} else {
				scope := time.RangeByYears(date, 0, false).ExcludeFuture()
				head, err = service.ContributionHeatMap(cmd.Context(), scope)
				if err != nil {
					return err
				}
			}

			// data presentation
			return view.Diff(cmd, base.Diff(head), baseSource, headSource)
		},
	}
	flag.Adopt(diff.Flags()).File("base", "", "path to a base file")
	flag.Adopt(diff.Flags()).File("head", "", "path to a head file")
	cmd.AddCommand(&diff)

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
			construct, date := time.RangeByWeeks, time.Now().UTC()

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
				case len(time.RFC3339Year):
					date, err = time.Parse(time.RFC3339Year, input)
					construct = time.RangeByYears
				case len(time.RFC3339Month):
					date, err = time.Parse(time.RFC3339Month, input)
					construct = time.RangeByMonths
				case len(time.RFC3339Day):
					date, err = time.Parse(time.RFC3339Day, input)
				default:
					err = fmt.Errorf("unsupported format")
				}
				if err != nil {
					return wrap(err)
				}
			}

			// data provisioning
			scope := construct(date, 0, false).ExcludeFuture()
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
	// $ maintainer github contribution lookup            # → now()/-1
	// $ maintainer github contribution lookup 2013-12-03 # → 2013-12-03/-1
	// $ maintainer github contribution lookup now/3      # → now()/3 == now()/-1
	// $ maintainer github contribution lookup /3         # → now()/3 == now()/-1
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
						date, err = time.Parse(time.RFC3339Day, raw[0])
					}
					if err != nil {
						return wrap(err)
					}
				default:
					return wrap(fmt.Errorf("too many parts"))
				}
			}

			// data provisioning
			scope := time.RangeByWeeks(date, weeks, half).Shift(-time.Day).ExcludeFuture()
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
	// $ maintainer github contribution snapshot 2013 | tee /tmp/snap.01.2013.json | jq
	//
	// {
	//   "2013-11-13T00:00:00Z": 1,
	//   ...
	//   "2013-12-27T00:00:00Z": 2
	// }
	//
	snapshot := cobra.Command{
		Use:  "snapshot",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			date := time.TruncateToYear(time.Now().UTC())

			// input validation: date(year)
			if len(args) == 1 {
				var err error
				wrap := func(err error) error {
					return fmt.Errorf(
						"please provide argument in format YYYY, e.g., 2006: %w",
						fmt.Errorf("invalid argument %q: %w", args[0], err),
					)
				}

				switch input := args[0]; len(input) {
				case len(time.RFC3339Year):
					date, err = time.Parse(time.RFC3339Year, input)
				default:
					err = fmt.Errorf("unsupported format")
				}
				if err != nil {
					return wrap(err)
				}
			}

			// data provisioning
			scope := time.RangeByYears(date, 0, false).ExcludeFuture()
			chm, err := service.ContributionHeatMap(cmd.Context(), scope)
			if err != nil {
				return err
			}

			// data presentation
			return json.NewEncoder(cmd.OutOrStdout()).Encode(chm)
		},
	}
	cmd.AddCommand(&snapshot)

	//
	// $ maintainer github contribution suggest --delta 2013-11-20
	//
	//  Day / Week    #45    #46    #47    #48   #49
	// ------------- ------ ------ ------ ----- -----
	//  Sunday         -      -      -      1     -
	//  Monday         -      -      -      2     1
	//  Tuesday        -      -      -      8     1
	//  Wednesday      -      1      1      -     3
	//  Thursday       -      -      3      7     1
	//  Friday         -      -      -      1     2
	//  Saturday       -      -      -      -     -
	// ------------- ------ ------ ------ ----- -----
	//  Contributions for 2013-11-17: -3119d, 0 → 5
	//
	// $ maintainer github contribution suggest 2013-11
	// $ maintainer github contribution suggest 2013
	// $ maintainer github contribution suggest --short 2013
	//
	suggest := cobra.Command{
		Use:  "suggest",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			date := time.TruncateToYear(time.Now().UTC())
			delta, _ := cmd.Flags().GetBool("delta")
			short, _ := cmd.Flags().GetBool("short")
			target, _ := cmd.Flags().GetInt("target")
			weeks, _ := cmd.Flags().GetInt("weeks")

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
				case len(time.RFC3339Year):
					date, err = time.Parse(time.RFC3339Year, input)
				case len(time.RFC3339Month):
					date, err = time.Parse(time.RFC3339Month, input)
				case len(time.RFC3339Day):
					date, err = time.Parse(time.RFC3339Day, input)
				default:
					err = fmt.Errorf("unsupported format")
				}
				if err != nil {
					return wrap(err)
				}
			}

			// data provisioning
			start := time.TruncateToWeek(date)
			scope := time.NewRange(
				start.Add(-2*time.Week-time.Day), // buffer from left side with Sunday
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
			for week, end := start, scope.To(); week.Before(end); week = week.Add(time.Week) {
				data := contribution.HistogramByCount(
					chm.Subset(time.RangeByWeeks(week, 0, false).Shift(-time.Day)), // Sunday
					contribution.OrderByCount,
				)

				// good week
				if len(data) == 1 && data[0].Count >= standard.Sum {
					continue
				}

				// Sunday
				day := week.Add(-time.Day)

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
					day = day.Add(time.Day)
				}
				break
			}

			// data presentation
			area := time.RangeByWeeks(suggest.Day, weeks, true).Shift(-time.Day) // Sunday
			data := contribution.HistogramByWeekday(chm.Subset(area), false)
			return view.Suggest(cmd, area, data, view.SuggestOption{
				Suggest: suggest,
				Current: chm[suggest.Day],
				Delta:   delta,
				Short:   short,
			})
		},
	}
	suggest.Flags().Bool("delta", false, "shows relatively")
	suggest.Flags().Bool("short", false, "shows only date")
	suggest.Flags().Int("target", 5, "minimum contributions")
	suggest.Flags().Int("weeks", 5, "how many weeks to show")
	cmd.AddCommand(&suggest)

	return &cmd
}
