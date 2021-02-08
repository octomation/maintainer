package github_test

import (
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
)

func TestContributionHeatMap_selector(t *testing.T) {
	f, err := os.Open("./testdata/github.kamilsk.2013.html")
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	data := make(map[string]string)
	doc.Find("svg.js-calendar-graph-svg rect.ContributionCalendar-day").
		Each(func(_ int, node *goquery.Selection) {
			data[node.AttrOr("data-date", "")] = node.AttrOr("data-level", "")
		})
	assert.Len(t, data, 365)
	assert.Equal(t, data["2013-11-13"], "1")
	assert.Equal(t, data["2013-12-13"], "0")
	assert.Equal(t, data["2013-12-27"], "2")
}
