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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	day := xtime.UTC().Year(2013).Month(time.December).Day(12).Time()
	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	chm, err := service.ContributionHeatMap(ctx, xtime.RangeByYears(day, 1, false))
	require.NoError(t, err)
	require.Equal(t, uint(7), chm.Count(day))
}

func TestService_FetchContributions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	chm, err := service.FetchContributions(ctx, "kamilsk", 2013)
	require.NoError(t, err)
	assert.Len(t, chm, 365)
	assert.Equal(t, uint(7), chm.Count(xtime.UTC().Year(2013).Month(time.December).Day(12).Time()))
}

// TestService_FetchContributions_Consistency covers MAIN-126: the profile HTML
// was served from caches of different ages, so a day's count walked backwards
// between two calls; a consistent source never does that.
func TestService_FetchContributions_Consistency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	year := time.Now().UTC().Year()
	service := github.New(xhttp.TokenSourcedClient(ctx, config.Secret(os.Getenv("GITHUB_TOKEN"))))
	seen, err := service.FetchContributions(ctx, "kamilsk", year)
	require.NoError(t, err)

	for attempt := 1; attempt <= 5; attempt++ {
		actual, err := service.FetchContributions(ctx, "kamilsk", year)
		require.NoError(t, err)
		for ts, count := range seen {
			require.GreaterOrEqual(t, actual.Count(ts), count,
				"attempt %d, %s", attempt, ts.Format(xtime.DateOnly))
		}
		seen = actual
	}
}
