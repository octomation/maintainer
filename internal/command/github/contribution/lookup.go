package contribution

import (
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Lookup(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	cmd.Args = cobra.MaximumNArgs(1)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))

		// data provisioning
		opts, err := ParseDate(args, FallbackDate(args), -1)
		if err != nil {
			return err
		}

		scope := contribution.LookupRange(opts)
		chm, err := service.ContributionHeatMap(cmd.Context(), scope)
		if err != nil {
			return err
		}

		// data presentation
		data := contribution.HistogramByWeekday(chm, false)
		return view.Lookup(cmd, scope, data)
	}

	return cmd
}
