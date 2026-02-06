package github_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func day(year int, month time.Month, d int) time.Time {
	return xtime.UTC().Year(year).Month(month).Day(d).Time()
}

func TestService_FetchContributionsPlacement(t *testing.T) {
	fake := &calendar{
		zone:     9 * time.Hour, // the account is in Tokyo
		pageSize: 2,             // every list of more than two events takes pages
		commits:  map[string]uint{"2026-08-30": 25, "2026-01-01": 3},
		reviews: []event{
			{"2025-12-31T20:00:00Z", "PR_E"}, // January 1 in Tokyo, but the day of 2025
			{"2026-01-03T20:06:00Z", "PR_A"},
			{"2026-04-30T19:21:00Z", "PR_F"}, // in the collection of May
			{"2026-03-01T10:00:00Z", "PR_A"}, // the year counts the latest review of PR_A only
			{"2026-08-31T23:30:00Z", "PR_B"}, // already September 1 in Tokyo
			{"2026-08-31T01:00:00Z", ""},     // a review without a pull request counts as is
			{"2026-08-31T02:00:00Z", "PR_C"},
			{"2026-08-31T03:00:00Z", "PR_D"},
		},
		issues: []event{{"2026-09-01T18:00:00Z", ""}, {"2026-09-01T23:59:59Z", ""}, {"2026-08-30T00:00:00Z", ""}},
	}
	service := serve(t, fake.handle(t))
	now := xtime.UTC().Year(2026).Month(time.September).Day(2).Hour(12).Minute(7).Second(31).Nanosecond(500).Time()
	service.WithClock(func() time.Time { return now })

	chm, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
	require.NoError(t, err)

	// Every day from January 1 up to today, and no later one.
	assert.Len(t, chm, 245)
	assert.Equal(t, map[string]uint{
		"2026-01-01": 3,
		"2026-03-01": 1,
		"2026-04-30": 1,
		"2026-08-30": 26,
		"2026-08-31": 4,
		"2026-09-01": 2,
	}, nonzero(chm))

	// The current year is bounded by the present second, so every call asks
	// for a range GitHub has not cached yet (issue#457).
	totals := fake.requests("totals")
	require.Len(t, totals, 1)
	assert.Equal(t, "2026-01-01T00:00:00Z", totals[0]["from0"])
	assert.Equal(t, "2026-01-31T23:59:59Z", totals[0]["to0"])
	assert.Equal(t, "2026-09-01T00:00:00Z", totals[0]["from8"])
	assert.Equal(t, "2026-09-02T12:07:31Z", totals[0]["to8"])
	// The day before the year, the day after today, and no month after now.
	assert.Equal(t, "2025-12-31T00:00:00Z", totals[0]["from9"])
	assert.Equal(t, "2025-12-31T23:59:59Z", totals[0]["to9"])
	assert.Equal(t, "2026-09-03T00:00:00Z", totals[0]["from10"])
	assert.Equal(t, "2026-09-03T23:59:59Z", totals[0]["to10"])
	assert.NotContains(t, totals[0], "from11")
	// A month without contributions is not fetched.
	assert.ElementsMatch(t, []string{"2026-01", "2026-03", "2026-05", "2026-08", "2026-09"}, fake.months())
	assert.NotEmpty(t, fake.requests("next"), "the lists take pages")

	t.Run("past year", func(t *testing.T) {
		chm, err := service.FetchContributions(context.Background(), "kamilsk", 2025)
		require.NoError(t, err)
		assert.Len(t, chm, 365)
		assert.Equal(t, map[string]uint{"2025-12-31": 1}, nonzero(chm), "the review of the day after the year")
		totals := fake.requests("totals")
		last := totals[len(totals)-1]
		assert.Equal(t, "2025-12-01T00:00:00Z", last["from11"])
		assert.Equal(t, "2025-12-31T23:59:59Z", last["to11"])
		assert.Equal(t, "2026-01-01T00:00:00Z", last["from13"])
		assert.Equal(t, "2026-01-01T23:59:59Z", last["to13"])
	})
}

func TestService_FetchContributionsToday(t *testing.T) {
	// The current year ends now, but a collection ends with a day of the
	// account's zone: in Tokyo, today ends at 15:00Z, and a later review is
	// in the collection of tomorrow.
	tests := map[string]struct {
		zone  time.Duration
		now   time.Time
		after string // the day after the days of the year
	}{
		"ahead of UTC": {
			zone:  9 * time.Hour,
			now:   xtime.UTC().Year(2026).Month(time.November).Day(7).Hour(18).Minute(2).Second(51).Time(),
			after: "2026-11-08",
		},
		"behind UTC": {
			zone:  -8 * time.Hour,
			now:   xtime.UTC().Year(2026).Month(time.November).Day(7).Hour(18).Minute(2).Second(51).Time(),
			after: "2026-11-08",
		},
		"on the last day of the year": {
			zone:  9 * time.Hour,
			now:   xtime.UTC().Year(2026).Month(time.December).Day(31).Hour(18).Minute(2).Second(51).Time(),
			after: "2027-01-01",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			today := test.now.Format(xtime.DateOnly)
			fake := &calendar{
				zone:    test.zone,
				reviews: []event{{today + "T14:58:50Z", "PR_A"}, {today + "T16:30:44Z", "PR_B"}},
			}
			service := serve(t, fake.handle(t)).WithCache(t.TempDir())
			service.WithClock(func() time.Time { return test.now })
			fetch := func() map[string]uint {
				t.Helper()
				fake.reset()
				chm, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
				require.NoError(t, err)
				return nonzero(chm)
			}

			assert.Equal(t, map[string]uint{today: 2}, fetch())
			totals := fake.requests("totals")
			require.Len(t, totals, 1)
			// The months up to now, the day before the year and the day after.
			after := int(test.now.Month()) + 1
			assert.Equal(t, test.after+"T00:00:00Z", totals[0][fmt.Sprintf("from%d", after)])
			assert.Equal(t, test.after+"T23:59:59Z", totals[0][fmt.Sprintf("to%d", after)])
			assert.NotContains(t, totals[0], fmt.Sprintf("from%d", after+1))

			// The next run sees a new review and fetches only the range it is in.
			fake.reviews = append(fake.reviews, event{today + "T17:56:46Z", "PR_C"})
			assert.Equal(t, map[string]uint{today: 3}, fetch())
			assert.Len(t, fake.requests("month"), 1)
		})
	}
}

func TestService_FetchContributionsCache(t *testing.T) {
	fake := &calendar{
		commits: map[string]uint{"2025-03-10": 5},
		reviews: []event{{"2025-07-04T12:00:00Z", "PR_A"}, {"2025-11-05T12:00:00Z", "PR_A"}},
	}
	dir := t.TempDir()
	service := serve(t, fake.handle(t)).WithCache(dir)
	service.WithClock(func() time.Time { return day(2026, time.October, 31) })
	fetch := func() map[string]uint {
		t.Helper()
		fake.reset()
		chm, err := service.FetchContributions(context.Background(), "kamilsk", 2025)
		require.NoError(t, err)
		return nonzero(chm)
	}

	expected := map[string]uint{"2025-03-10": 5, "2025-11-05": 1}
	assert.Equal(t, expected, fetch())
	assert.ElementsMatch(t, []string{"2025-03", "2025-07", "2025-11"}, fake.months())

	t.Run("unchanged totals", func(t *testing.T) {
		assert.Equal(t, expected, fetch())
		assert.Len(t, fake.requests("totals"), 1)
		assert.Empty(t, fake.months(), "no month is fetched again")
	})

	t.Run("a changed month", func(t *testing.T) {
		fake.issues = append(fake.issues, event{"2025-07-20T08:00:00Z", ""})
		assert.Equal(t, map[string]uint{"2025-03-10": 5, "2025-07-20": 1, "2025-11-05": 1}, fetch())
		assert.Equal(t, []string{"2025-07"}, fake.months())
		// The cached July still takes part in the review of the year.
		fake.reviews = append(fake.reviews, event{"2025-12-01T12:00:00Z", "PR_A"})
		assert.Equal(t, map[string]uint{"2025-03-10": 5, "2025-07-20": 1, "2025-12-01": 1}, fetch())
		assert.Equal(t, []string{"2025-12"}, fake.months())
	})

	t.Run("a corrupt cache", func(t *testing.T) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "kamilsk", "2025.json"), []byte("{"), 0o600))
		assert.Equal(t, map[string]uint{"2025-03-10": 5, "2025-07-20": 1, "2025-12-01": 1}, fetch())
		assert.ElementsMatch(t, []string{"2025-03", "2025-07", "2025-11", "2025-12"}, fake.months())
	})

	t.Run("an unwritable cache", func(t *testing.T) {
		service := serve(t, fake.handle(t)).WithCache(filepath.Join(dir, "kamilsk", "2025.json", "file"))
		service.WithClock(func() time.Time { return day(2026, time.October, 31) })
		_, err := service.FetchContributions(context.Background(), "kamilsk", 2025)
		require.NoError(t, err)
	})
}

func TestService_FetchContributionsUndated(t *testing.T) {
	fake := &calendar{
		commits: map[string]uint{"2023-06-10": 5, "2023-07-04": 2},
		reviews: []event{{"2023-06-15T12:00:00Z", "PR_A"}},
		undated: map[string]uint{"2023-06": 36, "2023-09": 1},
	}
	var warnings bytes.Buffer
	service := serve(t, fake.handle(t)).WithWarnings(&warnings)
	service.WithClock(func() time.Time { return day(2026, time.October, 31) })

	chm, err := service.FetchContributions(context.Background(), "kamilsk", 2023)
	require.NoError(t, err)
	// The months with contributions the API does not date take its calendar.
	assert.Equal(t, map[string]uint{"2023-06-01": 36, "2023-06-10": 5, "2023-06-15": 1, "2023-07-04": 2, "2023-09-01": 1}, nonzero(chm))
	assert.Len(t, fake.requests("days"), 2)
	// Only a month with reviews may be a day off.
	assert.Equal(t, "warn: the GitHub API gives no date for some contributions of 2023-06, "+
		"such as ones in private repositories the token cannot read, "+
		"so the days there are those of its calendar, where a review may be a day off\n", warnings.String())
}

func TestService_FetchContributionsLaggingMonth(t *testing.T) {
	// The lists of the current month may lag behind its total: its calendar
	// is used for a run, but not kept for the next one.
	fake := &calendar{commits: map[string]uint{"2026-10-10": 2}, undated: map[string]uint{"2026-10": 3}}
	service := serve(t, fake.handle(t)).WithCache(t.TempDir())
	service.WithClock(func() time.Time { return xtime.UTC().Year(2026).Month(time.October).Day(31).Hour(1).Time() })

	for range 2 {
		fake.reset()
		chm, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
		require.NoError(t, err)
		assert.Equal(t, map[string]uint{"2026-10-01": 3, "2026-10-10": 2}, nonzero(chm))
		assert.Equal(t, []string{"2026-10"}, fake.months())
	}
}

func TestService_FetchContributionsIncomplete(t *testing.T) {
	tests := map[string]struct {
		fake     *calendar
		expected string
	}{
		"a list ends before its total": {
			fake:     &calendar{issues: []event{{"2025-02-01T10:00:00Z", ""}}, lose: "issueContributions"},
			expected: "issueContributions has 0 of 1 events",
		},
		"commits in more than 100 repositories": {
			fake:     &calendar{commits: map[string]uint{"2025-02-01": 3}, hidden: 2},
			expected: "the commits of 1 repositories are 3 of 5",
		},
		"commits outside the month": {
			fake:     &calendar{commits: map[string]uint{"2025-02-01": 3}, shift: 31 * xtime.Day},
			expected: "commits of 2025-03-04 are outside 2025-02",
		},
		"malformed date": {
			fake:     &calendar{issues: []event{{"01.02.2025", ""}}, month: "2025-02"},
			expected: `parse contribution date "01.02.2025"`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			service := serve(t, test.fake.handle(t))
			service.WithClock(func() time.Time { return day(2026, time.October, 31) })

			_, err := service.FetchContributions(context.Background(), "kamilsk", 2025)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `fetch contributions of "kamilsk" for 2025-02`)
			assert.Contains(t, err.Error(), test.expected)
		})
	}
}

func TestService_FetchContributionsGraphQLErrors(t *testing.T) {
	t.Run("payload errors", func(t *testing.T) {
		service := serve(t, func(rw http.ResponseWriter, _ *http.Request) {
			// GitHub reports a GraphQL failure as 200 OK with an errors array.
			_, _ = io.WriteString(rw, `{"data":{"user":null},"errors":[
				{"type":"NOT_FOUND","message":"Could not resolve to a User with the login of 'ghost'."}
			]}`)
		})

		_, err := service.FetchContributions(context.Background(), "ghost", 2026)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `fetch contributions of "ghost" for 2026`)
		assert.Contains(t, err.Error(), "NOT_FOUND: Could not resolve to a User")
	})

	t.Run("unexpected status", func(t *testing.T) {
		service := serve(t, func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(rw, `{"message":"Bad credentials"}`)
		})

		_, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "401 Unauthorized")
		assert.Contains(t, err.Error(), "Bad credentials")
	})
}

func TestService_FetchContributionsWithoutCalendar(t *testing.T) {
	tests := map[string]string{
		"no data":       `{"data":null}`,
		"no user":       `{"data":{"user":null}}`,
		"no collection": `{"data":{"user":{"m0":null}}}`,
		"no calendar":   `{"data":{"user":{"m0":{"contributionCalendar":null}}}}`,
		"no total":      `{"data":{"user":{"m0":{"contributionCalendar":{}}}}}`,
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			service := serve(t, func(rw http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(rw, payload)
			})

			chm, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
			require.Error(t, err)
			assert.Nil(t, chm)
			assert.Contains(t, err.Error(), `fetch contributions of "kamilsk" for 2026`)
			assert.Contains(t, err.Error(), "no contribution calendar")
		})
	}
}

func TestService_ContributionHeatMapYears(t *testing.T) {
	now := xtime.UTC().Year(2026).Month(time.September).Day(25).Hour(12).Time()

	tests := map[string]struct {
		scope xtime.Range
		years []int
	}{
		"before GitHub": {
			scope: xtime.RangeByYears(xtime.UTC().Year(1990).Time(), 19, false),
			years: []int{
				1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 1998, 1999,
				2000, 2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009,
			},
		},
		"long before GitHub": {
			scope: xtime.RangeByYears(xtime.UTC().Year(1990).Time(), 0, false),
			years: []int{1990},
		},
		"after the current year": {
			scope: xtime.NewRange(
				xtime.UTC().Year(2025).Month(time.June).Time(),
				xtime.UTC().Year(2028).Month(time.June).Time(),
			),
			years: []int{2025, 2026},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &calendar{}
			service := serve(t, fake.handle(t))
			service.WithClock(func() time.Time { return now })

			chm, err := service.ContributionHeatMap(context.Background(), test.scope)
			require.NoError(t, err)

			var years []int
			for _, vars := range fake.requests("totals") {
				from, err := time.Parse(time.RFC3339, vars["from0"])
				require.NoError(t, err)
				years = append(years, from.Year())
			}
			sort.Ints(years)
			assert.Equal(t, test.years, years)
			assert.Empty(t, fake.months(), "a year without contributions takes one request")
			for _, year := range test.years {
				first := day(year, time.January, 1)
				if test.scope.Contains(first) {
					assert.Contains(t, chm, first, "a day of every requested year")
				}
			}
		})
	}
}

func TestService_ContributionHeatMapLimit(t *testing.T) {
	const limit = 4 // the same as the service has

	var inFlight, peak, total atomic.Int32
	queued := make(chan struct{})
	var once sync.Once
	fake := &calendar{}
	handle := fake.handle(t)
	service := serve(t, func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/user" {
			handle(rw, req)
			return
		}

		current := inFlight.Add(1)
		for {
			if old := peak.Load(); current <= old || peak.CompareAndSwap(old, current) {
				break
			}
		}
		total.Add(1)
		// hold the requests until the limit is reached, so a burst would show
		if current >= limit {
			once.Do(func() { close(queued) })
		}
		select {
		case <-queued:
		case <-time.After(5 * time.Second):
			t.Error("the requests are not concurrent")
		}
		inFlight.Add(-1)
		handle(rw, req)
	})
	service.WithClock(func() time.Time { return day(2026, time.September, 25) })

	scope := xtime.RangeByYears(xtime.UTC().Year(1990).Time(), 19, false)
	_, err := service.ContributionHeatMap(context.Background(), scope)
	require.NoError(t, err)

	assert.Equal(t, int32(20), total.Load())
	assert.Equal(t, int32(limit), peak.Load(), "no more requests at once than the limit")
}

// event is a single contribution at a moment, of a pull request for a review.
type event struct{ at, pr string }

// calendar answers the contribution queries of the service the way GitHub
// does: the totals of the months, and the events of a month by pages, with
// one review a pull request in a range, the latest.
type calendar struct {
	zone     time.Duration   // the offset of the account's days from UTC
	pageSize int             // nodes of a page, 100 without it
	commits  map[string]uint // by day
	reviews  []event
	issues   []event
	undated  map[string]uint // calendar totals without events, by month

	lose   string        // the list that loses its last node
	hidden uint          // commits counted in the total, but in no repository
	shift  time.Duration // moves the days of commits
	month  string        // the month a malformed event is in

	mu   sync.Mutex
	seen []map[string]string // the variables of the requests, with "kind"
}

func (c *calendar) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen = nil
}

func (c *calendar) requests(kind string) []map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []map[string]string
	for _, vars := range c.seen {
		if vars["kind"] == kind {
			out = append(out, vars)
		}
	}
	return out
}

// months returns the months whose events are fetched, as YYYY-MM.
func (c *calendar) months() []string {
	var out []string
	for _, vars := range c.requests("month") {
		out = append(out, vars["from"][:7])
	}
	return out
}

func (c *calendar) handle(t *testing.T) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/user" {
			_, _ = io.WriteString(rw, `{"login":"kamilsk"}`)
			return
		}

		var request struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if !assert.NoError(t, json.NewDecoder(req.Body).Decode(&request)) {
			return
		}
		vars := make(map[string]string, len(request.Variables))
		for k, v := range request.Variables {
			vars[k] = fmt.Sprint(v)
		}
		switch {
		case strings.Contains(request.Query, "contributionDays"):
			vars["kind"] = "days"
		case strings.Contains(request.Query, "contributionCalendar"):
			vars["kind"] = "totals"
		case strings.Contains(request.Query, "commitContributionsByRepository"):
			vars["kind"] = "month"
		default:
			vars["kind"] = "next"
		}
		c.mu.Lock()
		c.seen = append(c.seen, vars)
		c.mu.Unlock()

		var data map[string]any
		switch vars["kind"] {
		case "totals":
			data = map[string]any{}
			for i := 0; ; i++ {
				from, present := vars[fmt.Sprintf("from%d", i)]
				if !present {
					break
				}
				total := c.total(t, from, vars[fmt.Sprintf("to%d", i)])
				data[fmt.Sprintf("m%d", i)] = map[string]any{"contributionCalendar": map[string]any{"totalContributions": total}}
			}
		case "days":
			var days []any
			from, err := time.Parse(time.RFC3339, vars["from"])
			require.NoError(t, err)
			for d := xtime.TruncateToDay(from); d.Format(xtime.DateOnly) <= vars["to"][:10]; d = d.Add(xtime.Day) {
				date := d.Format(xtime.DateOnly)
				days = append(days, map[string]any{"date": date, "contributionCount": c.total(t, date+"T00:00:00Z", date+"T23:59:59Z")})
			}
			data = map[string]any{"contributionsCollection": map[string]any{
				"contributionCalendar": map[string]any{"weeks": []any{map[string]any{"contributionDays": days}}},
			}}
		case "month":
			collection := map[string]any{}
			var count uint
			var nodes []map[string]any
			for _, d := range c.days(t, vars["from"], vars["to"]) {
				ts, err := time.Parse(xtime.DateOnly, d)
				require.NoError(t, err)
				nodes = append(nodes, map[string]any{
					"occurredAt":  ts.Add(8*time.Hour + c.shift).Format(time.RFC3339),
					"commitCount": c.commits[d],
				})
				count += c.commits[d]
			}
			collection["totalCommitContributions"] = count + c.hidden
			collection["commitContributionsByRepository"] = []any{
				map[string]any{"contributions": map[string]any{"pageInfo": map[string]any{"hasNextPage": false}, "nodes": nodes}},
			}
			for _, field := range []string{"pullRequestReviewContributions", "issueContributions", "pullRequestContributions", "repositoryContributions"} {
				collection[field] = c.page(t, field, vars, 0)
			}
			data = map[string]any{"contributionsCollection": collection}
		case "next":
			offset, err := strconv.Atoi(vars["after"])
			require.NoError(t, err)
			collection := map[string]any{}
			for _, field := range []string{"pullRequestReviewContributions", "issueContributions", "pullRequestContributions", "repositoryContributions"} {
				if strings.Contains(request.Query, field) {
					collection[field] = c.page(t, field, vars, offset)
				}
			}
			data = map[string]any{"contributionsCollection": collection}
		}
		assert.NoError(t, json.NewEncoder(rw).Encode(map[string]any{"data": map[string]any{"user": data}}))
	}
}

// in tells if the moment is on the days of the range in the account's zone:
// the days whose dates from and to have in UTC.
func (c *calendar) in(t *testing.T, at, from, to string) bool {
	ts, err := time.Parse(time.RFC3339, at)
	require.NoError(t, err)
	date := ts.Add(c.zone).UTC().Format(xtime.DateOnly)
	return date >= from[:10] && date <= to[:10]
}

func (c *calendar) days(t *testing.T, from, to string) []string {
	var out []string
	for d := range c.commits {
		if c.in(t, d+"T08:00:00Z", from, to) {
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

// events returns the events of the list in the range, newest first.
func (c *calendar) events(t *testing.T, field, from, to string) []event {
	var source []event
	switch field {
	case "pullRequestReviewContributions":
		latest := map[string]event{}
		for _, e := range c.reviews {
			if !c.in(t, e.at, from, to) {
				continue
			}
			if e.pr == "" {
				source = append(source, e)
				continue
			}
			if prev, present := latest[e.pr]; !present || e.at > prev.at {
				latest[e.pr] = e
			}
		}
		for _, e := range latest {
			source = append(source, e)
		}
	case "issueContributions":
		for _, e := range c.issues {
			if c.month != "" {
				if strings.HasPrefix(from, c.month) {
					source = append(source, e)
				}
				continue
			}
			if c.in(t, e.at, from, to) {
				source = append(source, e)
			}
		}
	}
	sort.Slice(source, func(i, j int) bool { return source[i].at > source[j].at })
	return source
}

func (c *calendar) page(t *testing.T, field string, vars map[string]string, offset int) map[string]any {
	all := c.events(t, field, vars["from"], vars["to"])
	size := c.pageSize
	if size == 0 {
		size = 100
	}
	end := min(offset+size, len(all))
	nodes := []any{}
	listed := all[offset:end]
	if c.lose == field && end == len(all) && len(listed) > 0 {
		listed = listed[:len(listed)-1]
	}
	for _, e := range listed {
		n := map[string]any{"occurredAt": e.at}
		if field == "pullRequestReviewContributions" {
			n["pullRequest"] = nil
			if e.pr != "" {
				n["pullRequest"] = map[string]any{"id": e.pr}
			}
		}
		nodes = append(nodes, n)
	}
	return map[string]any{
		"totalCount": len(all),
		"pageInfo":   map[string]any{"hasNextPage": end < len(all), "endCursor": strconv.Itoa(end)},
		"nodes":      nodes,
	}
}

// total is the calendar total of the range: the events the lists have in it,
// and the contributions without events.
func (c *calendar) total(t *testing.T, from, to string) uint {
	var total uint
	for _, d := range c.days(t, from, to) {
		total += c.commits[d]
	}
	total += c.hidden * uint(len(c.days(t, from, to)))
	for _, field := range []string{"pullRequestReviewContributions", "issueContributions"} {
		total += uint(len(c.events(t, field, from, to)))
	}
	if month := from[:7]; from[8:10] == "01" {
		total += c.undated[month]
	}
	return total
}

func nonzero(chm map[time.Time]uint) map[string]uint {
	out := make(map[string]uint)
	for ts, count := range chm {
		if count > 0 {
			out[ts.Format(xtime.DateOnly)] = count
		}
	}
	return out
}

// serve returns a service whose every request lands on the handler.
func serve(t *testing.T, handler http.HandlerFunc) *github.Service {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	return github.New(&http.Client{Transport: redirect{target}})
}

// redirect routes a request to the target host, whatever host it asks for.
type redirect struct{ target *url.URL }

func (r redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	proxy := req.Clone(req.Context())
	proxy.URL.Scheme, proxy.URL.Host = r.target.Scheme, r.target.Host
	proxy.Host = r.target.Host

	return http.DefaultTransport.RoundTrip(proxy)
}
