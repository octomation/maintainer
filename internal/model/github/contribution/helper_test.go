package contribution_test

import (
	"os"
	"strconv"
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

func TestLookupRange(t *testing.T) {
	const name = "testdata/kamilsk.2021.html"

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	chm := BuildHeatMap(doc)

	t.Run("issue#124: correct centering", func(t *testing.T) {
		opts := DateOptions{
			Value: xtime.UTC().Year(2021).Month(time.January).Day(30).Time(),
			Weeks: 3, Half: true,
		}
		scope := LookupRange(opts).Until(time.Now())
		schedule, target := xtime.Everyday(xtime.Hours(5, 19, 0)), uint(5)
		suggestion := Suggest(chm, scope.Since(opts.Value), schedule, target)

		opts.Value = suggestion.Time
		scope = LookupRange(opts)
		assert.Equal(t, "2021-01-17", scope.From().Format(xtime.DateOnly))
		assert.Equal(t, "2021-02-06", scope.To().Format(xtime.DateOnly))
	})
}

func TestYearRange(t *testing.T) {
	const name = "testdata/kamilsk.1986.html"
	// The committed fixture was captured in 2025. The live healthcheck supplies
	// the capture year when it refreshes fixtures, without weakening this check.
	expectedYear := 2025
	if value := os.Getenv("MAINTAINER_TESTDATA_YEAR"); value != "" {
		var err error
		expectedYear, err = strconv.Atoi(value)
		require.NoError(t, err)
	}

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	min, max := YearRange(doc)
	assert.Equal(t, 2011, min)
	assert.Equal(t, expectedYear, max)
}

func load(t testing.TB, name string) *goquery.Document {
	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	return doc
}
