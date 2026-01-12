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
	cmd.Short = "Compare the contributions of two snapshots"
	cmd.Long = `Compares the base and the head day by day and shows every day that
differs: the count in the base, the count in the head, and the signed
difference, e.g., +4 or -3. A day one of them lacks counts as zero there
and is shown as "-", so such a day is listed only if it has contributions. A line under
the table names the days only one of them covers.

Each argument is a year, YYYY, to fetch its contributions now, or a file
made by snapshot. Two files need no token; a year does:
` + tokenNote
	cmd.Example = `  maintainer github contribution diff /tmp/snap.01.2013.json /tmp/snap.02.2013.json
  maintainer github contribution diff /tmp/snap.2013.json 2013`

	isYear := regexp.MustCompile(`^\d{4}$`)
	wrap := func(err error, arg string) error {
		return fmt.Errorf(
			"please provide the argument in format YYYY, e.g., 2006: %w",
			fmt.Errorf("invalid argument %q: %w", arg, err),
		)
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// input validation: a year or a snapshot file, a year needs GitHub data
		years := make(map[int]time.Time, len(args))
		for i, input := range args {
			if !isYear.MatchString(input) {
				continue
			}
			year, err := time.Parse(xtime.YearOnly, input)
			if err != nil {
				return wrap(err, input)
			}
			years[i] = year
		}

		var service *github.Service
		if len(years) > 0 {
			if err := RequireToken(cnf); err != nil {
				return err
			}
			service = github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
		}
		source := func(i int) ContributionSource {
			if year, is := years[i]; is {
				return contribution.NewUpstreamSource(service, year)
			}
			return contribution.NewFileSource(cnf.FS, args[i])
		}
		base, head := source(0), source(1)
		ctx := cmd.Context()

		src, err := base.Fetch(ctx)
		if err != nil {
			return err
		}

		dst, err := head.Fetch(ctx)
		if err != nil {
			return err
		}

		return view.ContributionDiff(cmd, src.Diff(dst), src.Only(dst), dst.Only(src), base.Location(), head.Location())
	}

	return cmd
}
