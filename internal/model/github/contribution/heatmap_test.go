package contribution_test

import (
	"os"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestBuildHeatMap(t *testing.T) {
	const name = "testdata/kamilsk.2019.html"

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	chm := BuildHeatMap(doc)
	ts := xtime.UTC().Year(2019).Month(time.November).Day(13).Time()
	assert.Equal(t, uint(3), chm.Count(ts), "2019-11-13")
	assert.Equal(t, uint(2), chm.Count(ts.AddDate(0, 1, 0)), "2019-12-13")
	assert.Equal(t, uint(3), chm.Count(ts.AddDate(0, 1, 14)), "2019-12-27")
}
