package github

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// The calendar of the GraphQL API puts a pull request review on its date in
// the time zone of the token owner's account, and it moves all the history
// when that zone changes, while the profile puts it on its UTC date. So the
// days are built from the contribution events of each month, placed as the
// profile places them (MAIN-159):
//
//   - a review on its UTC date, one per pull request a year, the latest;
//   - commits on the days GitHub gives them, the dates of the commits;
//   - an opened issue or pull request and a created repository on its UTC date.
//
// The API counts some contributions without giving their events, such as
// the ones in private repositories the token cannot read. A month with them
// takes the days of the calendar of the API instead: commits are on their
// days there, but a review may be a day off.
//
// A collection covers whole days in the account's zone: the days whose
// dates from and to have in UTC. So the months of a year do not overlap,
// but an event of a UTC year can be in the collection of a day next to it,
// such as a review on December 31 at 20:00Z in Tokyo, and the year takes
// the last day before it and the first day after it as well. The current
// year ends now, which may be past the end of today in the account's zone,
// so it takes the day after today instead: in Tokyo, a review today at
// 20:00Z is in the collection of tomorrow. The totals of the collections
// tell which of them have changed since their events were cached.

// contributionPage is the most nodes GitHub returns for a page of a list.
const contributionPage = 100

// contributionMonth requests the contribution events of a month: commits by
// repository and day, and the first page of every list of single events.
const contributionMonth = `query ($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
      totalCommitContributions
      commitContributionsByRepository(maxRepositories: 100) {
        contributions(first: 100) {
          pageInfo { hasNextPage }
          nodes { occurredAt commitCount }
        }
      }
      %s
    }
  }
}`

// contributionDays requests the days of the calendar of the API, for a month
// whose events do not add up to its total.
const contributionDays = `query ($login: String!, $from: DateTime!, $to: DateTime!) {
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

// contributionNext requests the next page of a list of single events.
const contributionNext = `query ($login: String!, $from: DateTime!, $to: DateTime!, $after: String!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
      %s
    }
  }
}`

// eventList is a list of single contribution events of a collection.
type eventList struct {
	field  string // the list in contributionsCollection
	nodes  string // the fields of a node
	review bool   // a node is a review of a pull request
}

func (list eventList) selection(after string) string {
	return fmt.Sprintf("%s(first: %d%s) { totalCount pageInfo { hasNextPage endCursor } nodes { %s } }",
		list.field, contributionPage, after, list.nodes)
}

var eventLists = []eventList{
	{field: "pullRequestReviewContributions", nodes: "occurredAt pullRequest { id }", review: true},
	{field: "issueContributions", nodes: "occurredAt"},
	{field: "pullRequestContributions", nodes: "occurredAt"},
	{field: "repositoryContributions", nodes: "occurredAt"},
}

// monthEvents are the contribution events of a month, as the cache keeps them.
type monthEvents struct {
	// Total is the calendar total of the month the events were fetched for.
	Total uint `json:"total"`
	// Commits are counted by the day GitHub gives them.
	Commits map[string]uint `json:"commits,omitempty"`
	// Others are issues, pull requests and repositories by their UTC day.
	Others map[string]uint `json:"others,omitempty"`
	// Reviews are all the reviews of the month: the year keeps one
	// a pull request.
	Reviews []review `json:"reviews,omitempty"`
	// Calendar are the days of the calendar of the API, for a month whose
	// events do not add up to its total; they replace the events.
	Calendar map[string]uint `json:"calendar,omitempty"`
}

type review struct {
	At time.Time `json:"at"`
	// PR is the node ID of the pull request, empty if GitHub gives none.
	PR string `json:"pr,omitempty"`
}

// dated is how many contributions of the total the events of the month date.
func (events monthEvents) dated() uint {
	var count uint
	for _, n := range events.Commits {
		count += n
	}
	for _, n := range events.Others {
		count += n
	}
	return count + uint(len(events.Reviews))
}

// complete tells if the events account for the total of the month, with
// the calendar if they do not add up to it. The lists of the current month
// may lag behind its total for a while, so its calendar is not kept.
func (events monthEvents) complete(month span) bool {
	if events.Calendar != nil {
		return !month.open
	}
	return month.edge || events.dated() >= events.Total
}

// span is a month of a year, cut at now in the current one, or a day next to
// the days of the year at their edge.
type span struct {
	from, to time.Time
	edge     bool // a day next to the days of the year
	open     bool // the current month, cut at now
}

func (s span) label() string {
	if s.edge {
		return s.from.Format(xtime.DateOnly)
	}
	return s.from.Format("2006-01")
}

func (srv *Service) ContributionHeatMap(
	ctx context.Context,
	scope xtime.Range,
) (contribution.HeatMap, error) {
	u, _, err := srv.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}

	chm := make(contribution.HeatMap)
	var (
		mu      sync.Mutex
		inexact []string
	)

	group, cascade := errgroup.WithContext(ctx)
	min, max := scope.From().UTC().Year(), scope.To().UTC().Year()
	// There are no contributions after the current year, so the heat map
	// has no days of those years, and they are not requested. A year before
	// GitHub is requested all the same: its days are known to have none.
	if year := srv.now().UTC().Year(); max > year {
		max = year
	}
	for i, user := min, u.GetLogin(); i <= max; i++ {
		year := i
		group.Go(func() error {
			src, months, err := srv.fetchYear(cascade, user, year)
			if err != nil {
				return err
			}

			mu.Lock()
			for ts, count := range src {
				chm.SetCount(ts, count)
			}
			inexact = append(inexact, months...)
			mu.Unlock()
			return nil
		})
	}

	err = group.Wait()
	if err == nil {
		srv.warnInexact(inexact)
	}
	return chm.Subset(scope), err
}

// FetchContributions returns the contribution heat map of the user for the
// whole year, including the days with no contributions at all.
func (srv *Service) FetchContributions(
	ctx context.Context,
	user string, year int,
) (contribution.HeatMap, error) {
	chm, inexact, err := srv.fetchYear(ctx, user, year)
	if err != nil {
		return nil, err
	}
	srv.warnInexact(inexact)
	return chm, nil
}

// fetchYear returns the heat map of the year and the months whose reviews
// may be a day off, as YYYY-MM.
func (srv *Service) fetchYear(
	ctx context.Context,
	user string, year int,
) (contribution.HeatMap, []string, error) {
	from := xtime.UTC().Year(year).Time()
	to := xtime.UTC().Year(year).Month(time.December).Day(31).Hour(23).Minute(59).Second(59).Time()
	// GitHub serves a repeated query from several caches of different ages
	// (issue#457), so the current year is bounded by this very second: a range
	// no cache has seen yet is always resolved from the source. Past years
	// are immutable, and a cached answer is as good as a fresh one.
	if now := srv.now().UTC(); now.After(from) && now.Before(to) {
		to = now.Truncate(time.Second)
	}
	var months []span
	for start := from; !start.After(to); start = start.AddDate(0, 1, 0) {
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		if end.After(to) {
			end = to
		}
		months = append(months, span{from: start, to: end, open: end.Equal(to) && to.Before(start.AddDate(0, 1, 0).Add(-time.Second))})
	}
	// The days next to the days of the year hold the events of the year that
	// are on them in the account's zone: the day before the year, and the day
	// after its last day, which is tomorrow in the current year.
	last := xtime.TruncateToDay(to)
	months = append(months,
		span{from: from.AddDate(0, 0, -1), to: from.Add(-time.Second), edge: true},
		span{from: last.Add(xtime.Day), to: last.Add(2*xtime.Day - time.Second), edge: true},
	)

	totals, err := srv.monthTotals(ctx, user, months)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch contributions of %q for %d: %w", user, year, err)
	}

	cached := srv.loadYear(user, year)
	events := make(map[string]monthEvents, len(months))
	var (
		mu      sync.Mutex
		fetched bool
	)
	group, cascade := errgroup.WithContext(ctx)
	for i, month := range months {
		key, total := month.label(), totals[i]
		if entry, present := cached[key]; total == 0 || present && entry.Total == total && entry.complete(month) {
			if total == 0 {
				entry = monthEvents{}
			}
			mu.Lock()
			events[key] = entry
			mu.Unlock()
			continue
		}

		fetched = true
		group.Go(func() error {
			entry, err := srv.fetchMonth(cascade, user, month, total)
			if err != nil {
				return fmt.Errorf("fetch contributions of %q for %s: %w", user, key, err)
			}

			mu.Lock()
			events[key] = entry
			mu.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	if fetched {
		srv.storeYear(user, year, events)
	}

	chm, inexact, err := placeEvents(from, to, months, events)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch contributions of %q for %d: %w", user, year, err)
	}
	return chm, inexact, nil
}

// monthTotals requests the calendar total of every month at once.
func (srv *Service) monthTotals(ctx context.Context, user string, months []span) ([]uint, error) {
	var query strings.Builder
	vars := map[string]any{"login": user}
	query.WriteString("query ($login: String!")
	for i := range months {
		fmt.Fprintf(&query, ", $from%d: DateTime!, $to%d: DateTime!", i, i)
	}
	query.WriteString(") {\n  user(login: $login) {\n")
	for i, month := range months {
		fmt.Fprintf(&query,
			"    m%d: contributionsCollection(from: $from%d, to: $to%d) { contributionCalendar { totalContributions } }\n",
			i, i, i)
		vars[fmt.Sprintf("from%d", i)] = month.from.Format(time.RFC3339)
		vars[fmt.Sprintf("to%d", i)] = month.to.Format(time.RFC3339)
	}
	query.WriteString("  }\n}")

	// Pointers tell a missing total from a zero one: GitHub may answer
	// without the user or the calendar, and it is not a month without contributions.
	var response *struct {
		User map[string]*struct {
			Calendar *struct {
				Total *uint `json:"totalContributions"`
			} `json:"contributionCalendar"`
		} `json:"user"`
	}
	if err := srv.queryGraphQL(ctx, query.String(), vars, &response); err != nil {
		return nil, err
	}
	if response == nil || response.User == nil {
		return nil, fmt.Errorf("the response has no contribution calendar")
	}

	totals := make([]uint, len(months))
	for i, month := range months {
		collection := response.User[fmt.Sprintf("m%d", i)]
		if collection == nil || collection.Calendar == nil || collection.Calendar.Total == nil {
			return nil, fmt.Errorf("the response has no contribution calendar for %s", month.label())
		}
		totals[i] = *collection.Calendar.Total
	}
	return totals, nil
}

// eventPage is a page of a list of single events.
type eventPage struct {
	Total    *int `json:"totalCount"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
	Nodes []struct {
		OccurredAt  string `json:"occurredAt"`
		PullRequest *node  `json:"pullRequest"`
	} `json:"nodes"`
}

// node is a reference to a node of the graph.
type node struct {
	ID string `json:"id"`
}

// fetchMonth requests all the contribution events of the month and checks
// that they add up to the totals GitHub gives for them. If they do not add up
// to the calendar total of the month, it requests the days of the calendar.
func (srv *Service) fetchMonth(ctx context.Context, user string, month span, total uint) (monthEvents, error) {
	vars := map[string]any{
		"login": user,
		"from":  month.from.Format(time.RFC3339),
		"to":    month.to.Format(time.RFC3339),
	}
	selections := make([]string, 0, len(eventLists))
	for _, list := range eventLists {
		selections = append(selections, list.selection(""))
	}

	collection, err := srv.queryCollection(ctx, fmt.Sprintf(contributionMonth, strings.Join(selections, "\n      ")), vars)
	if err != nil {
		return monthEvents{}, err
	}

	events := monthEvents{Total: total, Commits: make(map[string]uint), Others: make(map[string]uint)}
	if err := events.addCommits(collection, month); err != nil {
		return monthEvents{}, err
	}

	for _, list := range eventLists {
		var (
			page eventPage
			seen int
		)
		raw := collection[list.field]
		for {
			if err := decodeField(raw, list.field, &page); err != nil {
				return monthEvents{}, err
			}
			if page.Total == nil {
				return monthEvents{}, fmt.Errorf("the response has no total of %s", list.field)
			}
			for _, event := range page.Nodes {
				if err := events.add(list, event.OccurredAt, event.PullRequest, month); err != nil {
					return monthEvents{}, err
				}
			}
			seen += len(page.Nodes)
			if !page.PageInfo.HasNextPage {
				break
			}
			if len(page.Nodes) == 0 || page.PageInfo.EndCursor == "" {
				return monthEvents{}, fmt.Errorf("%s has a next page, but no cursor to it", list.field)
			}

			next := map[string]any{"after": page.PageInfo.EndCursor}
			for k, v := range vars {
				next[k] = v
			}
			collection, err := srv.queryCollection(ctx, fmt.Sprintf(contributionNext, list.selection(", after: $after")), next)
			if err != nil {
				return monthEvents{}, err
			}
			raw, page = collection[list.field], eventPage{}
		}
		// An answer that ends before its total would miss days silently.
		if seen != *page.Total {
			return monthEvents{}, fmt.Errorf("%s has %d of %d events", list.field, seen, *page.Total)
		}
	}

	// A day at the edge is a day of another month, which takes the calendar
	// if it needs to: a month of the year next to it, or the month of
	// tomorrow once tomorrow comes.
	if month.edge || events.dated() >= total {
		return events, nil
	}
	events.Calendar, err = srv.calendarDays(ctx, vars)
	if err != nil {
		return monthEvents{}, err
	}
	return events, nil
}

// calendarDays requests the days of the calendar of the API in the range.
func (srv *Service) calendarDays(ctx context.Context, vars map[string]any) (map[string]uint, error) {
	collection, err := srv.queryCollection(ctx, contributionDays, vars)
	if err != nil {
		return nil, err
	}
	var calendar *struct {
		Weeks []struct {
			Days []struct {
				Date  string `json:"date"`
				Count uint   `json:"contributionCount"`
			} `json:"contributionDays"`
		} `json:"weeks"`
	}
	if err := decodeField(collection["contributionCalendar"], "contributionCalendar", &calendar); err != nil {
		return nil, err
	}
	if calendar == nil || len(calendar.Weeks) == 0 {
		return nil, fmt.Errorf("the response has no contribution calendar")
	}

	days := make(map[string]uint, 31)
	for _, week := range calendar.Weeks {
		for _, day := range week.Days {
			if _, err := time.Parse(xtime.DateOnly, day.Date); err != nil {
				return nil, fmt.Errorf("parse contribution date %q: %w", day.Date, err)
			}
			days[day.Date] = day.Count
		}
	}
	return days, nil
}

// queryCollection executes the query and returns the fields of the
// contributionsCollection of the user.
func (srv *Service) queryCollection(ctx context.Context, query string, vars map[string]any) (map[string]json.RawMessage, error) {
	var response *struct {
		User *struct {
			Collection map[string]json.RawMessage `json:"contributionsCollection"`
		} `json:"user"`
	}
	if err := srv.queryGraphQL(ctx, query, vars, &response); err != nil {
		return nil, err
	}
	if response == nil || response.User == nil || response.User.Collection == nil {
		return nil, fmt.Errorf("the response has no contributions")
	}
	return response.User.Collection, nil
}

func decodeField(raw json.RawMessage, field string, dst any) error {
	if len(raw) == 0 || string(raw) == "null" {
		return fmt.Errorf("the response has no %s", field)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode %s: %w", field, err)
	}
	return nil
}

func (events *monthEvents) addCommits(collection map[string]json.RawMessage, month span) error {
	var (
		total  *uint
		byRepo []struct {
			Contributions struct {
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
				Nodes []struct {
					OccurredAt string `json:"occurredAt"`
					Count      uint   `json:"commitCount"`
				} `json:"nodes"`
			} `json:"contributions"`
		}
	)
	if err := decodeField(collection["totalCommitContributions"], "totalCommitContributions", &total); err != nil {
		return err
	}
	if err := decodeField(collection["commitContributionsByRepository"], "commitContributionsByRepository", &byRepo); err != nil {
		return err
	}

	var count uint
	for _, repo := range byRepo {
		// A month has fewer days than a page has nodes.
		if repo.Contributions.PageInfo.HasNextPage {
			return fmt.Errorf("the commits of a repository have more days than %d", contributionPage)
		}
		for _, day := range repo.Contributions.Nodes {
			// The moment is midnight of the commits' day in some zone,
			// such as 08:00Z, so the day is the date as written.
			ts, err := time.Parse(time.RFC3339, day.OccurredAt)
			if err != nil {
				return fmt.Errorf("parse contribution date %q: %w", day.OccurredAt, err)
			}
			date := ts.Format(xtime.DateOnly)
			if date < month.from.Add(-xtime.Day).Format(xtime.DateOnly) || date > month.to.Add(xtime.Day).Format(xtime.DateOnly) {
				return fmt.Errorf("commits of %s are outside %s", date, month.label())
			}
			events.Commits[date] += day.Count
			count += day.Count
		}
	}
	if total == nil || count != *total {
		var expected uint
		if total != nil {
			expected = *total
		}
		return fmt.Errorf("the commits of %d repositories are %d of %d; the rest is in more than %d repositories",
			len(byRepo), count, expected, contributionPage)
	}
	return nil
}

func (events *monthEvents) add(list eventList, occurredAt string, pr *node, month span) error {
	ts, err := time.Parse(time.RFC3339, occurredAt)
	if err != nil {
		return fmt.Errorf("parse contribution date %q: %w", occurredAt, err)
	}
	ts = ts.UTC()
	// The days of the account's zone are less than a day away from UTC ones.
	if ts.Before(xtime.TruncateToDay(month.from).Add(-xtime.Day)) || !ts.Before(xtime.TruncateToDay(month.to).Add(2*xtime.Day)) {
		return fmt.Errorf("a contribution at %s is outside %s", occurredAt, month.label())
	}

	if list.review {
		var id string
		if pr != nil {
			id = pr.ID
		}
		events.Reviews = append(events.Reviews, review{At: ts, PR: id})
		return nil
	}
	events.Others[ts.Format(xtime.DateOnly)]++
	return nil
}

// placeEvents puts the events of the spans on the days of the year from
// to, as the profile does, and returns the months that take the calendar of
// the API and have reviews, which may be a day off there. An event of a day
// outside the year belongs to the year next to it.
func placeEvents(from, to time.Time, spans []span, events map[string]monthEvents) (contribution.HeatMap, []string, error) {
	chm := make(contribution.HeatMap, 366)
	for day := from; !day.After(to); day = day.Add(xtime.Day) {
		chm.SetCount(day, 0)
	}
	add := func(date string, count uint) error {
		day, err := time.Parse(xtime.DateOnly, date)
		if err != nil {
			return fmt.Errorf("parse contribution date %q: %w", date, err)
		}
		if _, present := chm[day]; present {
			chm[day] += count
		}
		return nil
	}

	var inexact []string
	latest := make(map[string]time.Time)
	for _, month := range spans {
		events := events[month.label()]
		if events.Calendar != nil {
			for date, count := range events.Calendar {
				if err := add(date, count); err != nil {
					return nil, nil, err
				}
			}
			if len(events.Reviews) > 0 {
				inexact = append(inexact, month.label())
			}
			continue
		}
		for date, count := range events.Commits {
			if err := add(date, count); err != nil {
				return nil, nil, err
			}
		}
		for date, count := range events.Others {
			if err := add(date, count); err != nil {
				return nil, nil, err
			}
		}
		for _, r := range events.Reviews {
			if _, present := chm[xtime.TruncateToDay(r.At.UTC())]; !present {
				continue
			}
			if r.PR == "" {
				chm[xtime.TruncateToDay(r.At.UTC())]++
				continue
			}
			if at, present := latest[r.PR]; !present || r.At.After(at) {
				latest[r.PR] = r.At
			}
		}
	}
	// A pull request reviewed more than once in a year counts once, on the
	// day of the latest review, as the calendar of the year counts it.
	for _, at := range latest {
		chm[xtime.TruncateToDay(at.UTC())]++
	}
	return chm, inexact, nil
}

func (srv *Service) warnInexact(months []string) {
	if len(months) == 0 {
		return
	}
	sort.Strings(months)
	fmt.Fprintf(srv.warn, "warn: the GitHub API gives no date for some contributions of %s, "+
		"such as ones in private repositories the token cannot read, "+
		"so the days there are those of its calendar, where a review may be a day off\n", strings.Join(months, ", "))
}
