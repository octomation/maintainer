package contribution

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func Histogram(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	var (
		zero bool
	)
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Short = "Show how many days have each contribution count"
	cmd.Long = `Shows a histogram of the contributions of the token owner: a row per
count of contributions, with the number of days that have it and a bar of
"#", one per day. Bars longer than 40 characters are scaled down so that
the longest is 40, and a row with days keeps at least one "#".

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
			return exit.WithUser(err)
		}
		if err := RequireToken(cnf); err != nil {
			return exit.WithUser(err)
		}
		service := NewService(cmd, cnf)
		chm, err := service.ContributionHeatMap(cmd.Context(), scope)
		if err != nil {
			return err
		}

		// data presentation
		data := contribution.HistogramByCount(chm, contribution.OrderByCount)
		HistogramView(cmd.OutOrStdout(), data, zero)
		return nil
	}

	return cmd
}

// HistogramBar is the longest bar of the histogram, in characters.
const HistogramBar = 40

// HistogramView writes the rows in their order as a table of the count,
// the exact number of days and a bar of "#", one per day, scaled down when
// the most frequent count has more than HistogramBar days. A row with days
// keeps at least one "#". The row of zero contributions is written only
// with zero.
//
//	Contributions  Days  Distribution
//	1                 4  ####
//	3                 1  #
func HistogramView(w io.Writer, rows []contribution.HistogramByCountRow, zero bool) {
	shown := make([]contribution.HistogramByCountRow, 0, len(rows))
	var most uint
	for _, row := range rows {
		if !zero && row.Count == 0 {
			continue
		}
		shown = append(shown, row)
		most = max(most, row.Frequency)
	}

	const count, days = "Contributions", "Days"
	cw, dw := len(count), len(days)
	for _, row := range shown {
		cw = max(cw, len(fmt.Sprint(row.Count)))
		dw = max(dw, len(fmt.Sprint(row.Frequency)))
	}

	_, _ = fmt.Fprintf(w, "%-*s  %*s  %s\n", cw, count, dw, days, "Distribution")
	for _, row := range shown {
		bar := row.Frequency
		if most > HistogramBar {
			bar = (row.Frequency*HistogramBar + most/2) / most // rounded to the nearest
		}
		if row.Frequency > 0 {
			bar = max(bar, 1)
		}
		_, _ = fmt.Fprintf(w, "%-*d  %*d  %s\n", cw, row.Count, dw, row.Frequency, strings.Repeat("#", int(bar)))
	}
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
