package contribution_test

import (
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
)

func TestYearRange(t *testing.T) {
	const name = "testdata/kamilsk.1986.html"

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	min, max := YearRange(doc)
	assert.Equal(t, 2011, min)
	assert.Equal(t, 2023, max)
}
