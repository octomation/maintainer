package contribution_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
)

func TestHeatMap_Histogram(t *testing.T) {
	chm := make(contribution.HeatMap)
	chm.Set(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1)
	chm.Set(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1)
	chm.Set(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3)
	chm.Set(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1)
	chm.Set(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2)
	chm.Set(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8)
	chm.Set(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7)
	chm.Set(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1)

	expected := [][2]int{
		{1, 4},
		{2, 1},
		{3, 1},
		{7, 1},
		{8, 1},
	}
	assert.Equal(t, expected, chm.Histogram())
}
