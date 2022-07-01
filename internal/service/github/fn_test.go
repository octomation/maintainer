package github_test

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	. "go.octolab.org/toolset/maintainer/internal/service/github"
)

func TestContributionRange(t *testing.T) {
	const name = "testdata/contribution/kamilsk.1986.html"

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	min, max := ContributionRange(doc)
	assert.Equal(t, 2011, min)
	assert.Equal(t, 2023, max)
}

func TestContributionHeatMap(t *testing.T) {
	const name = "testdata/contribution/kamilsk.2013.html"

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	ts := xtime.Year(2013).Month(time.November).Day(13).Location(time.UTC).Time()
	chm := ContributionHeatMap(doc)
	assert.Equal(t, 1, chm.Count(ts))                   // 2013-11-13
	assert.Equal(t, 0, chm.Count(ts.AddDate(0, 1, 0)))  // 2013-12-13
	assert.Equal(t, 2, chm.Count(ts.AddDate(0, 1, 14))) // 2013-12-27
}
