package exec

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/run"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func ContributionDiff(cnf *config.Tool) Runner {
	isYear := regexp.MustCompile(`^\d{4}$`)
	wrap := func(err error, input string) error {
		return fmt.Errorf(
			"please provide the argument in format YYYY, e.g., 2006: %w",
			fmt.Errorf("invalid argument %q: %w", input, err),
		)
	}

	// input validation:
	//  - Args: cobra.ExactArgs(2)
	//  - base{file|date(year)} head{file|date(year)}
	return func(cmd *cobra.Command, args []string) error {
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))

		var base run.ContributionSource
		if input := args[0]; isYear.MatchString(input) {
			year, err := time.Parse(time.RFC3339Year, input)
			if err != nil {
				return wrap(err, input)
			}

			base = contribution.NewUpstreamSource(service, year)
		} else {
			base = contribution.NewFileSource(cnf.FS, input)
		}

		var head run.ContributionSource
		if input := args[1]; isYear.MatchString(input) {
			year, err := time.Parse(time.RFC3339Year, input)
			if err != nil {
				return wrap(err, input)
			}

			head = contribution.NewUpstreamSource(service, year)
		} else {
			head = contribution.NewFileSource(cnf.FS, input)
		}

		return run.ContributionDiff(cmd.Context(), base, head, cmd)
	}
}
