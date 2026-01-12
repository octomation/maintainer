package contribution

import (
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func Lookup(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	cmd.Short = "Show the contribution calendar around a date"
	cmd.Long = `Shows the contribution calendar of the token owner as a table:
a column per week, a row per day, and the distribution of the shown days
by the count of contributions under it.

The argument is date[/weeks]. The date is empty, git, now, YYYY, YYYY-MM,
YYYY-MM-DD, or RFC 3339; YYYY and YYYY-MM are the anchor at the first day,
not the whole year or month. An empty date or git is the author date of HEAD
in the current repository, or in the one GIT_DIR points to; outside
a repository, or in one without commits, it is now. The weeks add to the week
of the date: +N after it, -N before it, or N/2, rounded down, on each side,
so an even N shows N+1 weeks; without them it is -1, two weeks. N is at most
53, so the table shows up to 54 weeks, the year is not before 1970, and
a period in the future is an error. Days after now are shown as "?".

Weeks start on Sunday, and the week that contains January 1 is #01.
` + tokenNote
	cmd.Example = `  maintainer github contribution lookup
  maintainer github contribution lookup 2013-12-03/9
  maintainer github contribution lookup 2022/+10
  maintainer github contribution lookup now/-4`
	cmd.Args = cobra.MaximumNArgs(1)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// dependencies and defaults
		now := time.Now()

		// data provisioning
		anchor, err := FallbackDate(args, now)
		if err != nil {
			return err
		}
		scope, err := LookupScope(args, anchor, now)
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
		TableView(cmd, chm, scope)
		return nil
	}

	return cmd
}

// LookupScope returns the weeks the lookup shows: the argument in format
// date[/weeks], where an empty or git date is the anchor and no weeks mean
// the previous one and the current one, cut at now.
func LookupScope(args []string, anchor, now time.Time) (xtime.Range, error) {
	opts, err := ParseDate(args, anchor, -1, now)
	if err != nil {
		return xtime.Range{}, err
	}
	return contribution.LookupRange(opts).ExcludeFuture(now)
}
