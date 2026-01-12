package contribution

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Snapshot(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Short = "Save the contributions of a year as JSON"
	cmd.Long = `Prints the contributions of the token owner for a year as a JSON object:
the count of contributions by the day, as RFC 3339 midnight UTC, including
the days without contributions. The result is a snapshot for diff.

The argument is YYYY, the current year without it. The year is cut at now,
a year in the future is an error, and it is not before 1970.
` + tokenNote
	cmd.Example = `  maintainer github contribution snapshot > snap.json
  maintainer github contribution snapshot 2013 | tee /tmp/snap.2013.json | jq`

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		now := time.Now()

		// data provisioning
		scope, err := SnapshotScope(args, now)
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
		return json.NewEncoder(cmd.OutOrStdout()).Encode(chm)
	}

	return cmd
}

// SnapshotScope returns the days the snapshot takes: the year of
// the argument in format YYYY, or the current one without it, cut at now.
func SnapshotScope(args []string, now time.Time) (xtime.Range, error) {
	date := xtime.TruncateToYear(now.UTC())

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
		if err == nil && date.Year() < contribution.MinYear {
			err = fmt.Errorf("the year is before %d", contribution.MinYear)
		}
		if err != nil {
			return xtime.Range{}, wrap(err)
		}
	}

	return xtime.RangeByYears(date, 0, false).ExcludeFuture(now)
}
