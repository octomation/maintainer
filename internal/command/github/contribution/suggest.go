package contribution

import (
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
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

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))

		// data provisioning
		opts, err := ParseDate(args, FallbackDate(args), 5)
		if err != nil {
			return err
		}

		since := opts.Value
		until := time.Now()
		scope := xtime.RangeByWeeks(since, opts.Weeks, opts.Half).ExpandRight(until)
		chm, err := service.ContributionHeatMap(cmd.Context(), scope)
		if err != nil {
			return err
		}

		suggestion := contribution.Suggest(chm, since, until, target)
		opts.Value = suggestion.Day // reuse options
		area := contribution.LookupRange(opts)
		data := contribution.HistogramByWeekday(chm.Subset(area), false)

		// data presentation
		return view.Suggest(cmd, area, data, view.SuggestOption{
			Suggest: contribution.HistogramByWeekdayRow{
				Day: suggestion.Day,
				Sum: suggestion.Target,
			},
			Current: chm.Count(suggestion.Day),
			Delta:   delta,
			Short:   short,
		})
	}

	return cmd
}
