package contribution

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Histogram(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	var (
		zero bool
	)
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Short = "Show how many days have each contribution count"
	cmd.Long = `Shows a histogram of the contributions of the token owner: a row per
count of contributions, with a "#" per day that has it.

The argument is YYYY for the year, YYYY-MM for the month, or YYYY-MM-DD for
the week of the day, from Sunday to Saturday; without it, the current week.
The period is cut at now, a period in the future is an error, and the year
is not before 1970. The row of days without contributions is shown only
with --with-zero.
` + tokenNote
	cmd.Example = `  maintainer github contribution histogram
  maintainer github contribution histogram 2013
  maintainer github contribution histogram 2013-11
  maintainer github contribution histogram --with-zero 2013-11-20`
	cmd.Flags().BoolVar(&zero, "with-zero", false, "shows zero-counted rows")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		now := time.Now()

		// data provisioning
		scope, err := HistogramScope(args, now)
		if err != nil {
			return err
		}
		if err := RequireToken(cnf); err != nil {
			return err
		}
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
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
			fmt.Printf("%3d %s\n", row.Count, strings.Repeat("#", int(row.Frequency)))
		}
		return nil
	}

	return cmd
}

// HistogramScope returns the days the histogram counts: the year, the month,
// or the week of the argument in format YYYY[-mm[-dd]], or the current week
// without it, cut at now.
func HistogramScope(args []string, now time.Time) (xtime.Range, error) {
	construct, date := xtime.GregorianWeeks, now.UTC()

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
		case len(xtime.YearOnly):
			date, err = time.Parse(xtime.YearOnly, input)
			construct = xtime.RangeByYears
		case len(xtime.YearAndMonth):
			date, err = time.Parse(xtime.YearAndMonth, input)
			construct = xtime.RangeByMonths
		case len(xtime.DateOnly):
			date, err = time.Parse(xtime.DateOnly, input)
		default:
			err = fmt.Errorf("unsupported format")
		}
		if err == nil && date.Year() < contribution.MinYear {
			err = fmt.Errorf("the year is before %d", contribution.MinYear)
		}
		if err != nil {
			return xtime.Range{}, wrap(err)
		}
	}

	return construct(date, 0, false).ExcludeFuture(now)
}
