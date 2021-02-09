package github

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
)

func TestContributionRange(t *testing.T) {
	f, err := os.Open("testdata/github.kamilsk.1986.html")
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	min, max := contributionRange(doc)
	assert.Equal(t, 2011, min)
	assert.Equal(t, 2022, max)
}

func TestContributionHeatMap(t *testing.T) {
	f, err := os.Open("testdata/github.kamilsk.2013.html")
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
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

	chm := contributionHeatMap(doc)
	assert.Equal(t, 1, chm[time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC)])
	assert.Equal(t, 0, chm[time.Date(2013, 12, 13, 0, 0, 0, 0, time.UTC)])
	assert.Equal(t, 2, chm[time.Date(2013, 12, 27, 0, 0, 0, 0, time.UTC)])
}
