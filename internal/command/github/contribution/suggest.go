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

// jitterLimit is the most a random jitter shifts a suggestion by.
const jitterLimit = time.Hour

func Suggest(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	var (
		delta  bool
		short  bool
		target uint
	)
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Short = "Suggest a moment to contribute"
	cmd.Long = `Suggests the first moment between the anchor and now, both inclusive,
on a day with fewer contributions than its week requires: the most a day
of the week has, but not less than --target. A moment is within the
schedule, 05:00-19:00 UTC every day, and a random jitter of less than an
hour moves it neither after now nor after the schedule. The table of the
weeks around the suggestion marks its day with "*" in stderr, and stdout
gets only the timestamp, e.g., to pass it to git commit --date.

The argument is date[/weeks], the same as lookup has: the date is empty,
git, now, YYYY, YYYY-MM, YYYY-MM-DD, or RFC 3339, the weeks add to the week
of the date: +N after it, -N before it, or N/2, rounded down, on each side,
so an even N shows N+1 weeks; 5 without them, five weeks. The date is
the anchor; YYYY and YYYY-MM are anchored at the first day. An empty date
or git is the author date of HEAD in the current repository, or in the one
GIT_DIR points to; outside a repository, or in one without commits, it is
now. N is at most 53, so the table shows up to 54 weeks, and the year is not
before 1970.

It fails with a non-zero exit code if the anchor is in the future,
or if there is no suitable moment up to now.
` + tokenNote
	cmd.Example = `  git commit --date="$(maintainer github contribution suggest)"
  maintainer github contribution suggest --delta 2013-11-20
  maintainer github contribution suggest --target=5 2013/+10
  maintainer github contribution suggest --short 2013-11/-10`
	cmd.Flags().BoolVar(&delta, "delta", false, "prints the suggestion relative to now, e.g., -3d2H0M0S")
	cmd.Flags().BoolVar(&short, "short", false, "hides the table")
	cmd.Flags().UintVar(&target, "target", 5, "the minimum contributions a day requires")
	// TODO:configure setup from flags
	// TODO:extend support Location
	schedule := xtime.Everyday(xtime.Hours(5, 19, 0)) // TODO:extend UTC correction

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// The moment is printed to the second, so the window is whole seconds.
		now := time.Now().Truncate(time.Second)
		if target == 0 {
			return fmt.Errorf("please provide a positive target, e.g., --target=5")
		}

		// data provisioning
		anchor, err := FallbackDate(args, now)
		if err != nil {
			return err
		}
		opts, scope, err := SuggestScope(args, anchor, now)
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

		suggestion := contribution.Suggest(chm, opts.Value, now, schedule, target)
		if suggestion.Time.IsZero() {
			return fmt.Errorf(
				"nothing to suggest between %s and %s: there is no day short of the target within the schedule",
				opts.Value.Format(time.RFC3339), now.Format(time.RFC3339),
			)
		}
		suggestion.Time = Moment(suggestion.Time, now, schedule, jitter.FullRandom())

		// data presentation
		if !short {
			opts.Value = suggestion.Time // reuse options
			area, err := contribution.LookupRange(opts).ExcludeFuture(now)
			if err != nil {
				return err
			}

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
			cmd.Print(Datetime(suggestion.Time.Local(), now))
		} else {
			cmd.Print(suggestion.Time.Local().Format(time.RFC3339))
		}
		cmd.PrintErrf(", %d → %d\n", suggestion.Actual, suggestion.Target)
		return nil
	}

	return cmd
}

// SuggestScope returns the lookup of the argument in format date[/weeks],
// where an empty or git date is the anchor and no weeks mean five centered
// on it, and the data a suggestion needs: from the beginning of the lookup
// up to now, because any moment between the anchor and now may be suggested.
func SuggestScope(
	args []string,
	anchor, now time.Time,
) (contribution.DateOptions, xtime.Range, error) {
	opts, err := ParseDate(args, anchor, 5, now)
	if err != nil {
		return opts, xtime.Range{}, err
	}
	opts.Value = ceilSecond(opts.Value)
	if opts.Value.After(now) {
		return opts, xtime.Range{}, fmt.Errorf(
			"the anchor %s is in the future, now is %s: nothing to suggest between them",
			opts.Value.Format(time.RFC3339), now.Format(time.RFC3339),
		)
	}

	scope, err := contribution.LookupRange(opts).ExcludeFuture(now)
	if err != nil {
		return opts, xtime.Range{}, err
	}
	// The lookup may end before now, but the data must not: a day without
	// data looks like a gap, whatever contributions it has.
	return opts, scope.Until(now), nil
}

// Jitter shifts the suggested time by a random duration, which is less than
// the jitter limit and keeps the time neither after now nor after the end
// of the schedule interval the time lies in.
func Jitter(t, now time.Time, hours xtime.Schedule, random jitter.Transformation) time.Time {
	end := hours.End(t)
	if end.IsZero() {
		return t
	}
	if margin := min(jitterLimit, now.Sub(t), end.Sub(t)); margin > 0 {
		return t.Add(random.Apply(margin))
	}
	return t
}

// Moment jitters the suggested time and drops fractional seconds:
// the moment is printed to the second, and its delta from now must agree.
// It stays within the window, because the anchor and now are whole seconds.
func Moment(t, now time.Time, hours xtime.Schedule, random jitter.Transformation) time.Time {
	return Jitter(t, now, hours, random).Truncate(time.Second)
}

// ceilSecond rounds the time up to a whole second: a moment is printed
// to the second, which must not move it before the anchor.
func ceilSecond(t time.Time) time.Time {
	if floor := t.Truncate(time.Second); !floor.Equal(t) {
		return floor.Add(time.Second)
	}
	return t
}
