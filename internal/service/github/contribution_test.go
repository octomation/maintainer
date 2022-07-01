//go:build integration

package github_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func TestService_FetchContributions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	service := github.New(http.DefaultClient)
	doc, err := service.FetchContributions(ctx, "kamilsk", 2013)
	require.NoError(t, err)

	selectors := []string{
		"svg.js-calendar-graph-svg rect.ContributionCalendar-day",
		"svg.js-calendar-graph-svg .ContributionCalendar-day",
		".js-calendar-graph-svg rect.ContributionCalendar-day",
		".js-calendar-graph-svg .ContributionCalendar-day",
	}
	for _, selector := range selectors {
		assert.Equal(t, 365, doc.Find(selector).Length())
	}
}
