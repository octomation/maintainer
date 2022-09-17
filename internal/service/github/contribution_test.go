//go:build integration

package github_test

import (
	"context"
	"net/http"
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

	selectors := []string{
		"svg.js-calendar-graph-svg rect.ContributionCalendar-day",
		"svg.js-calendar-graph-svg .ContributionCalendar-day",
		".js-calendar-graph-svg rect.ContributionCalendar-day",
		".js-calendar-graph-svg .ContributionCalendar-day",
	}
	service := github.New(http.DefaultClient)
	doc, err := service.FetchContributions(ctx, "kamilsk", 2013)
	require.NoError(t, err)
	for _, selector := range selectors {
		assert.Equal(t, 365, doc.Find(selector).Length())
	}
}
