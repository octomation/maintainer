package contribution

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Snapshot(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	cmd.Args = cobra.MaximumNArgs(1)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
		date := xtime.TruncateToYear(time.Now().UTC())

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
			case len(xtime.YearOnly):
				date, err = time.Parse(xtime.YearOnly, input)
			default:
				err = fmt.Errorf("unsupported format")
			}
			if err != nil {
				return wrap(err)
			}
		}

		// data provisioning
		scope := xtime.RangeByYears(date, 0, false).ExcludeFuture()
		chm, err := service.ContributionHeatMap(cmd.Context(), scope)
		if err != nil {
			return err
		}

		// data presentation
		return json.NewEncoder(cmd.OutOrStdout()).Encode(chm)
	}

	return cmd
}
