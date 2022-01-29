package github

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
)

var update = flag.Bool("update", false, "update testdata")

func TestContributionRange(t *testing.T) {
	const name = "testdata/github.kamilsk.1986.html"

	if *update {
		d, err := fetchContributions(context.TODO(), "kamilsk", 1986)
		require.NoError(t, err)

		h, err := d.Html()
		require.NoError(t, err)

		require.NoError(t, os.Truncate(name, 0))
		require.NoError(t, os.WriteFile(name, []byte(h), 0666))
	}

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	min, max := contributionRange(doc)
	assert.Equal(t, 2011, min)
	assert.Equal(t, 2023, max)
}

func TestContributionHeatMap(t *testing.T) {
	const name = "testdata/github.kamilsk.2013.html"

	if *update {
		d, err := fetchContributions(context.TODO(), "kamilsk", 2013)
		require.NoError(t, err)

		h, err := d.Html()
		require.NoError(t, err)

		require.NoError(t, os.Truncate(name, 0))
		require.NoError(t, os.WriteFile(name, []byte(h), 0666))
	}

	f, err := os.Open(name)
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

	ts := time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC)
	chm := contributionHeatMap(doc)
	assert.Equal(t, 1, chm.Count(ts))                   // 2013-11-13
	assert.Equal(t, 0, chm.Count(ts.AddDate(0, 1, 0)))  // 2013-12-13
	assert.Equal(t, 2, chm.Count(ts.AddDate(0, 1, 14))) // 2013-12-27
}
