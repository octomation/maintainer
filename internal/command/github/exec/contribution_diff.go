package exec

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/run"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/config/flag"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func ContributionDiff(cnf *config.Tool) Runner {
	return func(cmd *cobra.Command, args []string) error {
		service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
		date := time.TruncateToYear(time.Now().UTC())

		// input validation: files{params}, date(year){args}
		var baseSource, headSource string
		dst, err := flag.Adopt(cmd.Flags()).GetFile("base")
		if err != nil {
			return err
		}
		if dst == nil {
			return fmt.Errorf("please provide a base file by `--base` parameter")
		}
		baseSource = dst.Name()

		src, err := flag.Adopt(cmd.Flags()).GetFile("head")
		if err != nil {
			return err
		}
		if src == nil && len(args) == 0 {
			return fmt.Errorf("please provide a compared file by `--head` parameter or year in args")
		}
		if src != nil && len(args) > 0 {
			return fmt.Errorf("please omit `--head` or argument, only one of them is allowed")
		}
		if len(args) == 1 {
			var err error
			wrap := func(err error) error {
				return fmt.Errorf(
					"please provide argument in format YYYY, e.g., 2006: %w",
					fmt.Errorf("invalid argument %q: %w", args[0], err),
				)
			}

			switch input := args[0]; len(input) {
			case len(time.RFC3339Year):
				date, err = time.Parse(time.RFC3339Year, input)
			default:
				err = fmt.Errorf("unsupported format")
			}
			if err != nil {
				return wrap(err)
			}
			headSource = fmt.Sprintf("upstream:year(%s)", date.Format(time.RFC3339Year))
		} else {
			headSource = src.Name()
		}

		return run.ContributionDiff(cmd.Context(), service, cmd, date, dst, src, baseSource, headSource)
	}
}
