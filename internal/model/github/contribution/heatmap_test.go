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

func TestHeatMap_Subset(t *testing.T) {
	Nov2013 := xtime.UTC().Year(2013).Month(time.November)

	chm := make(HeatMap)
	chm.SetCount(Nov2013.Day(13).Time(), 1)
	chm.SetCount(Nov2013.Day(20).Time(), 1)
	chm.SetCount(Nov2013.Day(21).Time(), 3)
	chm.SetCount(Nov2013.Day(24).Time(), 1)
	chm.SetCount(Nov2013.Day(25).Time(), 2)
	chm.SetCount(Nov2013.Day(26).Time(), 8)
	chm.SetCount(Nov2013.Day(28).Time(), 7)
	chm.SetCount(Nov2013.Day(29).Time(), 1)

	t.Run("one week", func(t *testing.T) {
		ts := Nov2013.Day(20).Time()
		subset := chm.Subset(xtime.RangeByWeeks(ts, 0, false))
		assert.Len(t, subset, 3)
	})

	t.Run("one week behind", func(t *testing.T) {
		ts := Nov2013.Day(20).Time()
		subset := chm.Subset(xtime.RangeByWeeks(ts, -1, false))
		assert.Len(t, subset, 4)
	})

	t.Run("one week ahead", func(t *testing.T) {
		ts := Nov2013.Day(20).Time()
		subset := chm.Subset(xtime.RangeByWeeks(ts, 1, false))
		assert.Len(t, subset, 7)
	})

	t.Run("one week around", func(t *testing.T) {
		ts := Nov2013.Day(20).Time()
		subset := chm.Subset(xtime.RangeByWeeks(ts, 1, true))
		assert.Len(t, subset, 3)
	})

	t.Run("three weeks around", func(t *testing.T) {
		ts := Nov2013.Day(20).Time()
		subset := chm.Subset(xtime.RangeByWeeks(ts, 3, true))
		assert.Len(t, subset, 8)
	})
}

func TestBuildHeatMap(t *testing.T) {
	const name = "testdata/kamilsk.2013.html"

	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)

	ts := xtime.UTC().Year(2013).Month(time.November).Day(13).Time()
	chm := BuildHeatMap(doc)
	assert.Equal(t, 1, chm.Count(ts))                   // 2013-11-13
	assert.Equal(t, 0, chm.Count(ts.AddDate(0, 1, 0)))  // 2013-12-13
	assert.Equal(t, 2, chm.Count(ts.AddDate(0, 1, 14))) // 2013-12-27
}
