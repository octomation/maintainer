package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// concurrentYears is the most calendar years requested at once: the span
// is bounded by the commands, but a burst of requests is still unwelcome.
const concurrentYears = 4

// contributionCalendar requests the daily contribution counts of a user.
// GitHub caps the span of a single collection at one year, which is why
// ContributionHeatMap fans out per calendar year.
const contributionCalendar = `query ($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
      contributionCalendar {
        weeks {
          contributionDays {
            date
            contributionCount
          }
        }
      }
    }
  }
}`

func (srv *Service) ContributionHeatMap(
	ctx context.Context,
	scope xtime.Range,
) (contribution.HeatMap, error) {
	u, _, err := srv.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}

	chm := make(contribution.HeatMap)
	merge := func() func(contribution.HeatMap, error) error {
		var mu sync.Mutex
		return func(src contribution.HeatMap, err error) error {
			if err != nil {
				return err
			}

			mu.Lock()
			for ts, count := range src {
				chm.SetCount(ts, count)
			}
			mu.Unlock()
			return nil
		}
	}()

	group, cascade := errgroup.WithContext(ctx)
	group.SetLimit(concurrentYears)
	min, max := scope.From().UTC().Year(), scope.To().UTC().Year()
	// There are no contributions after the current year, so the heat map
	// has no days of those years, and they are not requested. A year before
	// GitHub is requested all the same: its days are known to have none.
	if year := srv.now().UTC().Year(); max > year {
		max = year
	}
	for i, user := min, u.GetLogin(); i <= max; i++ {
		year := i
		group.Go(func() error { return merge(srv.FetchContributions(cascade, user, year)) })
	}

	err = group.Wait()
	return chm.Subset(scope), err
}

// FetchContributions returns the contribution heat map of the user for the
// whole year, including the days with no contributions at all.
func (srv *Service) FetchContributions(
	ctx context.Context,
	user string, year int,
) (contribution.HeatMap, error) {
	// Pointers tell a missing calendar from an empty one: GitHub may answer
	// without the user or the calendar, and it is not a year without contributions.
	var response *struct {
		User *struct {
			Contributions *struct {
				Calendar *struct {
					Weeks []struct {
						Days []struct {
							Date  string `json:"date"`
							Count uint   `json:"contributionCount"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	from := xtime.UTC().Year(year).Time()
	to := xtime.UTC().Year(year).Month(time.December).Day(31).Hour(23).Minute(59).Second(59).Time()
	// GitHub serves a repeated query from several caches of different ages
	// (issue#457), so the current year is bounded by this very second: a range
	// no cache has seen yet is always resolved from the source. Past years
	// are immutable, and a cached answer is as good as a fresh one.
	if now := srv.now().UTC(); now.After(from) && now.Before(to) {
		to = now.Truncate(time.Second)
	}
	vars := map[string]any{
		"login": user,
		"from":  from.Format(time.RFC3339),
		"to":    to.Format(time.RFC3339),
	}
	if err := srv.queryGraphQL(ctx, contributionCalendar, vars, &response); err != nil {
		return nil, fmt.Errorf("fetch contributions of %q for %d: %w", user, year, err)
	}

	if response == nil || response.User == nil || response.User.Contributions == nil ||
		response.User.Contributions.Calendar == nil || len(response.User.Contributions.Calendar.Weeks) == 0 {
		return nil, fmt.Errorf("fetch contributions of %q for %d: the response has no contribution calendar", user, year)
	}

	chm := make(contribution.HeatMap, 366)
	for _, week := range response.User.Contributions.Calendar.Weeks {
		for _, day := range week.Days {
			// An expected format is date-only, so the parsed value is
			// midnight UTC, as the heat map requires.
			ts, err := time.Parse(xtime.DateOnly, day.Date)
			if err != nil {
				return nil, fmt.Errorf("parse contribution date %q: %w", day.Date, err)
			}
			chm.SetCount(ts, day.Count)
		}
	}
	return chm, nil
}
