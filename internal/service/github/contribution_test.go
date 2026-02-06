//go:build integration

package github_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/toolkit/config"

	xhttp "go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func TestService_ContributionHeatMap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	day := xtime.UTC().Year(2013).Month(time.December).Day(12).Time()
	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	chm, err := service.ContributionHeatMap(ctx, xtime.RangeByYears(day, 1, false))
	require.NoError(t, err)
	require.Equal(t, uint(7), chm.Count(day))
}

func TestService_FetchContributions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	chm, err := service.FetchContributions(ctx, "kamilsk", 2013)
	require.NoError(t, err)
	assert.Len(t, chm, 365)
	assert.Equal(t, uint(7), chm.Count(xtime.UTC().Year(2013).Month(time.December).Day(12).Time()))
}

// TestService_FetchContributions_ProfileDays covers MAIN-159: the calendar of
// the API put the reviews of these days on the dates of the account's time
// zone, 0, 30 and 15 in Tokyo, while the profile shows 15 a day. The year is
// 2024, whose contributions no longer change, unlike the ones still filled in.
func TestService_FetchContributions_ProfileDays(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	chm, err := service.FetchContributions(ctx, "kamilsk", 2024)
	require.NoError(t, err)
	assert.Len(t, chm, 366)
	for _, day := range []int{28, 29, 30} {
		assert.Equal(t, uint(15), chm.Count(xtime.UTC().Year(2024).Month(time.January).Day(day).Time()), "January %d", day)
	}
	var total uint
	for _, count := range chm {
		total += count
	}
	assert.Equal(t, uint(5910), total)
}

// TestService_FetchContributions_Consistency covers issue#457: the profile HTML
// was served from caches of different ages, so a day's count walked backwards
// between two calls; a consistent source never does that.
func TestService_FetchContributions_Consistency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	year := time.Now().UTC().Year()
	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	seen, err := service.FetchContributions(ctx, "kamilsk", year)
	require.NoError(t, err)

	for attempt := 1; attempt <= 2; attempt++ {
		actual, err := service.FetchContributions(ctx, "kamilsk", year)
		require.NoError(t, err)
		for ts, count := range seen {
			require.GreaterOrEqual(t, actual.Count(ts), count,
				"attempt %d, %s", attempt, ts.Format(xtime.DateOnly))
		}
		seen = actual
	}
}
