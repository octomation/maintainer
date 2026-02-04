package contribution_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
)

// calendarTransport answers as GitHub does for an account whose January 2025
// has 15 commits a day, except 14 on January 3.
type calendarTransport struct{}

func (calendarTransport) count(day time.Time) uint {
	if day.Format(time.DateOnly) == "2025-01-03" {
		return 14
	}
	return 15
}

// days returns the days of January 2025 in the range of the dates of from and to.
func (c calendarTransport) days(from, to string) []time.Time {
	var out []time.Time
	for day := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC); day.Month() == time.January; day = day.AddDate(0, 0, 1) {
		if date := day.Format(time.DateOnly); date >= from[:10] && date <= to[:10] {
			out = append(out, day)
		}
	}
	return out
}

func (c calendarTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("Authorization") != "Bearer test-token" {
		return nil, fmt.Errorf("missing API authentication")
	}

	body := `{"login":"traveller"}`
	if r.URL.Path == "/graphql" {
		var request struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return nil, err
		}
		user := map[string]any{}
		if strings.Contains(request.Query, "contributionCalendar") {
			for i := 0; ; i++ {
				from, present := request.Variables[fmt.Sprintf("from%d", i)]
				if !present {
					break
				}
				var total uint
				for _, day := range c.days(from, request.Variables[fmt.Sprintf("to%d", i)]) {
					total += c.count(day)
				}
				user[fmt.Sprintf("m%d", i)] = map[string]any{"contributionCalendar": map[string]any{"totalContributions": total}}
			}
		} else {
			var (
				total uint
				nodes []any
			)
			for _, day := range c.days(request.Variables["from"], request.Variables["to"]) {
				nodes = append(nodes, map[string]any{"occurredAt": day.Add(8 * time.Hour).Format(time.RFC3339), "commitCount": c.count(day)})
				total += c.count(day)
			}
			empty := map[string]any{"totalCount": 0, "pageInfo": map[string]any{"hasNextPage": false}, "nodes": []any{}}
			user["contributionsCollection"] = map[string]any{
				"totalCommitContributions": total,
				"commitContributionsByRepository": []any{map[string]any{"contributions": map[string]any{
					"pageInfo": map[string]any{"hasNextPage": false}, "nodes": nodes,
				}}},
				"pullRequestReviewContributions": empty,
				"issueContributions":             empty,
				"pullRequestContributions":       empty,
				"repositoryContributions":        empty,
			}
		}
		raw, err := json.Marshal(map[string]any{"data": map[string]any{"user": user}})
		if err != nil {
			return nil, err
		}
		body = string(raw)
	}
	return &http.Response{
		StatusCode: http.StatusOK, Status: "200 OK", Header: make(http.Header),
		Body: io.NopCloser(strings.NewReader(body)), Request: r,
	}, nil
}

func TestCommandsKeepTheDayWhenTravelling(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	original := time.Local
	t.Cleanup(func() { time.Local = original })

	var previous string
	for _, zone := range []string{"UTC", "Europe/Moscow", "Asia/Tokyo", "America/Los_Angeles"} {
		t.Run(zone, func(t *testing.T) {
			var err error
			time.Local, err = time.LoadLocation(zone)
			require.NoError(t, err)
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: calendarTransport{}})

			for _, suggest := range []bool{false, true} {
				var out, diagnostic bytes.Buffer
				cmd := &cobra.Command{Use: "test"}
				cnf := &config.Tool{Token: "test-token"}
				if suggest {
					cmd = Suggest(cmd, cnf)
					cmd.SetArgs([]string{"--target=15", "2025-01-03T18:00:00Z/+0"})
				} else {
					cmd = Lookup(cmd, cnf)
					cmd.SetArgs([]string{"2025-01-03/+0"})
				}
				cmd.SetOut(&out)
				cmd.SetErr(&diagnostic)
				require.NoError(t, cmd.ExecuteContext(ctx))

				if suggest {
					// A late UTC moment is the next day in Tokyo: printed in UTC,
					// it stays on the day the calendar needs.
					value := out.String()
					require.True(t, strings.HasPrefix(value, "2025-01-03T18:"), value)
					require.True(t, strings.HasSuffix(value, "Z"), value)
					require.Contains(t, diagnostic.String(), "[14]")
					continue
				}
				if previous != "" {
					require.Equal(t, previous, diagnostic.String())
				}
				previous = diagnostic.String()
			}
		})
	}
}
