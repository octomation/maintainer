package exec

import (
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/pkg/unsafe"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// configure(cmd) -> setup flags, setup run

func Contribution(cnf *config.Tool) Runner {
	return func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
		delta := unsafe.ReturnBool(cmd.Flags().GetBool("delta"))
		short := unsafe.ReturnBool(cmd.Flags().GetBool("short"))
		target := unsafe.ReturnUint(cmd.Flags().GetUint("target"))

		// data provisioning
		opts, err := ParseDate(args, FallbackDate(args), 5)
		if err != nil {
			return err
		}

		scope := xtime.RangeByWeeks(opts.Value, opts.Weeks, opts.Half).ExpandRight(time.Now().UTC())
		chm, err := service.ContributionHeatMap(cmd.Context(), scope)
		if err != nil {
			return err
		}

		suggestion := contribution.Suggest(chm, scope.Base(), scope.To(), target)
		opts.Value = suggestion.Day
		area := contribution.LookupRange(opts)
		data := contribution.HistogramByWeekday(chm.Subset(area), false)

		// data presentation
		return view.Suggest(cmd, area, data, view.SuggestOption{
			Suggest: suggestion,
			Current: chm[suggestion.Day],
			Delta:   delta,
			Short:   short,
		})
	}
}
