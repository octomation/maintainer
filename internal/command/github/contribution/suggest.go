package contribution

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/pkg/time/jitter"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Suggest(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	var (
		delta  bool
		short  bool
		target uint
	)
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Flags().BoolVar(&delta, "delta", false, "shows relatively")
	cmd.Flags().BoolVar(&short, "short", false, "shows only date")
	cmd.Flags().UintVar(&target, "target", 5, "minimum contributions")
	// TODO:configure setup from flags
	// TODO:extend support Location
	schedule := xtime.Everyday(xtime.Hours(5, 19, 0)) // TODO:extend UTC correction

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))

		// data provisioning
		opts, err := ParseDate(args, FallbackDate(args), 5)
		if err != nil {
			return err
		}

		scope := contribution.LookupRange(opts).Until(time.Now())
		chm, err := service.ContributionHeatMap(cmd.Context(), scope)
		if err != nil {
			return err
		}

		suggestion := contribution.Suggest(chm, scope.Since(opts.Value), schedule, target)
		if suggestion.Time.IsZero() {
			return fmt.Errorf("nothing to suggest")
		}
		suggestion.Time = suggestion.Time.Add(jitter.FullRandom().Apply(time.Hour))
		opts.Value = suggestion.Time // reuse options
		area := contribution.LookupRange(opts).ExcludeFuture()

		// data presentation
		if !short {
			accent := xtime.TruncateToDay(suggestion.Time)
			TableView(cmd, chm, area, func(day time.Time, txt string) string {
				if !day.Equal(accent) {
					return txt
				}
				if txt == "-" || txt == "?" {
					return "*"
				}
				return txt + "*"
			})
		}
		cmd.PrintErr("Suggestion is ")
		if delta {
			cmd.PrintErr(suggestion.Time.Local().Format(time.RFC3339), ": ")
			cmd.Print(Datetime(suggestion.Time.Local()))
		} else {
			cmd.Print(suggestion.Time.Local().Format(time.RFC3339))
		}
		cmd.PrintErrf(", %d → %d\n", suggestion.Actual, suggestion.Target)
		return nil
	}

	return cmd
}
