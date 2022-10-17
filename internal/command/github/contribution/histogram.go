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
	cmd.Flags().BoolVar(&zero, "with-zero", false, "shows zero-counted rows")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
		construct, date := xtime.GregorianWeeks, time.Now().UTC()

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
			fmt.Printf("%3d %s\n", row.Count, strings.Repeat("#", int(row.Frequency)))
		}
		return nil
	}

	return cmd
}
