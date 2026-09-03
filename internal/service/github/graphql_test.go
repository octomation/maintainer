package github_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func TestService_FetchContributionsFromGraphQL(t *testing.T) {
	const payload = `{"data":{"user":{"contributionsCollection":{"contributionCalendar":{"weeks":[
		{"contributionDays":[
			{"date":"2026-08-30","contributionCount":25},
			{"date":"2026-08-31","contributionCount":23}
		]},
		{"contributionDays":[{"date":"2026-09-01","contributionCount":17}]}
	]}}}}}`

	var request struct {
		Query     string `json:"query"`
		Variables struct {
			Login string `json:"login"`
			From  string `json:"from"`
			To    string `json:"to"`
		} `json:"variables"`
	}
	service := serve(t, func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/graphql", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.NoError(t, json.NewDecoder(req.Body).Decode(&request))
		_, _ = io.WriteString(rw, payload)
	})

	now := xtime.UTC().Year(2026).Month(time.September).Day(2).Hour(12).Minute(7).Second(31).Nanosecond(500).Time()
	service.WithClock(func() time.Time { return now })

	chm, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
	require.NoError(t, err)

	assert.Contains(t, request.Query, "contributionCalendar")
	assert.Equal(t, "kamilsk", request.Variables.Login)
	// The current year is bounded by the present second, so every call asks
	// for a range GitHub has not cached yet (MAIN-126).
	assert.Equal(t, "2026-01-01T00:00:00Z", request.Variables.From)
	assert.Equal(t, "2026-09-02T12:07:31Z", request.Variables.To)

	require.Len(t, chm, 3)

	t.Run("past year", func(t *testing.T) {
		_, err := service.FetchContributions(context.Background(), "kamilsk", 2025)
		require.NoError(t, err)
		// The collection must span the whole year and no more: GitHub rejects
		// a range longer than one year.
		assert.Equal(t, "2025-01-01T00:00:00Z", request.Variables.From)
		assert.Equal(t, "2025-12-31T23:59:59Z", request.Variables.To)
	})
	assert.Equal(t, uint(25), chm.Count(xtime.UTC().Year(2026).Month(time.August).Day(30).Time()))
	assert.Equal(t, uint(23), chm.Count(xtime.UTC().Year(2026).Month(time.August).Day(31).Time()))
	assert.Equal(t, uint(17), chm.Count(xtime.UTC().Year(2026).Month(time.September).Day(1).Time()))
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

	t.Run("malformed date", func(t *testing.T) {
		service := serve(t, func(rw http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(rw, `{"data":{"user":{"contributionsCollection":{"contributionCalendar":{
				"weeks":[{"contributionDays":[{"date":"31.08.2026","contributionCount":23}]}]
			}}}}}`)
		})

		_, err := service.FetchContributions(context.Background(), "kamilsk", 2026)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `parse contribution date "31.08.2026"`)
	})
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
