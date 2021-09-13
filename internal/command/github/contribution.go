package github

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/exec"
	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/config/flag"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/pkg/unsafe"
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
		RunE: exec.ContributionDiff(cnf),
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
			zero := unsafe.ReturnBool(cmd.Flags().GetBool("with-zero"))

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
				if !zero && row.Count == 0 {
					continue
				}
				fmt.Printf("%3d %s\n", row.Count, strings.Repeat("#", row.Frequency))
			}
			return nil
		},
	}
	histogram.Flags().Bool("with-zero", false, "shows zero-counted rows")
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

			// input validation: date[/{+-}weeks]
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
	// $ maintainer github contribution suggest 2013-11/10
	// $ maintainer github contribution suggest --target=5 2013/+10
	// $ maintainer github contribution suggest --short 2013/-10
	//
	suggest := cobra.Command{
		Use:  "suggest",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// dependencies and defaults
			service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
			date, weeks, half := time.TruncateToYear(time.Now().UTC()), 5, true
			delta := unsafe.ReturnBool(cmd.Flags().GetBool("delta"))
			short := unsafe.ReturnBool(cmd.Flags().GetBool("short"))
			target := unsafe.ReturnInt(cmd.Flags().GetInt("target"))

			// input validation: date(year,+month,+week{day})[/{+-}weeks]
			if len(args) == 1 {
				var err error
				wrap := func(err error) error {
					return fmt.Errorf(
						"please provide argument in format YYYY[-mm[-dd]][/[+|-]weeks], e.g., 2006-01: %w",
						fmt.Errorf("invalid argument %q: %w", args[0], err),
					)
				}

				input := args[0]
				raw := strings.Split(input, "/")
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
					input = raw[0]
				default:
					return wrap(fmt.Errorf("too many parts"))
				}

				switch len(input) {
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

			value := contribution.Suggest(chm, start, scope.To(), target)
			area := time.RangeByWeeks(value.Day, weeks, half).Shift(-time.Day) // Sunday
			data := contribution.HistogramByWeekday(chm.Subset(area), false)

			// data presentation
			return view.Suggest(cmd, area, data, view.SuggestOption{
				Suggest: value,
				Current: chm[value.Day],
				Delta:   delta,
				Short:   short,
			})
		},
	}
	suggest.Flags().Bool("delta", false, "shows relatively")
	suggest.Flags().Bool("short", false, "shows only date")
	suggest.Flags().Int("target", 5, "minimum contributions")
	cmd.AddCommand(&suggest)

	return &cmd
}
