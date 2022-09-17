package contribution

import (
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Diff(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	cmd.Args = cobra.ExactArgs(2)

	isYear := regexp.MustCompile(`^\d{4}$`)
	wrap := func(err error, arg string) error {
		return fmt.Errorf(
			"please provide the argument in format YYYY, e.g., 2006: %w",
			fmt.Errorf("invalid argument %q: %w", arg, err),
		)
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))

		var base ContributionSource
		if input := args[0]; isYear.MatchString(input) {
			year, err := time.Parse(xtime.YearOnly, input)
			if err != nil {
				return wrap(err, input)
			}

			base = contribution.NewUpstreamSource(service, year)
		} else {
			base = contribution.NewFileSource(cnf.FS, input)
		}

		var head ContributionSource
		if input := args[1]; isYear.MatchString(input) {
			year, err := time.Parse(xtime.YearOnly, input)
			if err != nil {
				return wrap(err, input)
			}

			head = contribution.NewUpstreamSource(service, year)
		} else {
			head = contribution.NewFileSource(cnf.FS, input)
		}
		ctx := cmd.Context()

		src, err := base.Fetch(ctx)
		if err != nil {
			return err
		}

		dst, err := head.Fetch(ctx)
		if err != nil {
			return err
		}

		return view.ContributionDiff(cmd, src.Diff(dst), base.Location(), head.Location())
	}

	return cmd
}
