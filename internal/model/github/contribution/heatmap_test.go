package contribution_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestHeatMap_Count(t *testing.T) {
	chm := load(t, "testdata/kamilsk.2019.json")
	ts := xtime.UTC().Year(2019).Month(time.November).Day(13).Time()
	assert.Equal(t, uint(3), chm.Count(ts), "2019-11-13")
	assert.Equal(t, uint(2), chm.Count(ts.AddDate(0, 1, 0)), "2019-12-13")
	assert.Equal(t, uint(3), chm.Count(ts.AddDate(0, 1, 14)), "2019-12-27")
}
