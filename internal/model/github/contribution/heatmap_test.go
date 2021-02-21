package contribution_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
)

func TestHistogramByCount(t *testing.T) {
	chm := make(HeatMap)
	chm.SetCount(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3)
	chm.SetCount(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2)
	chm.SetCount(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8)
	chm.SetCount(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7)
	chm.SetCount(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1)

	expected := map[int]int{
		1: 4,
		2: 1,
		3: 1,
		7: 1,
		8: 1,
	}

	histogram := HistogramByCount(chm)
	assert.Len(t, histogram, len(expected))
	for i, row := range histogram {
		assert.Equal(t, expected[row.Count], row.Frequency, i)
	}
}

func TestHistogramByDate(t *testing.T) {
	chm := make(HeatMap)
	chm.SetCount(time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 20, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 21, 0, 0, 0, 0, time.UTC), 3)
	chm.SetCount(time.Date(2013, 11, 24, 0, 0, 0, 0, time.UTC), 1)
	chm.SetCount(time.Date(2013, 11, 25, 0, 0, 0, 0, time.UTC), 2)
	chm.SetCount(time.Date(2013, 11, 26, 0, 0, 0, 0, time.UTC), 8)
	chm.SetCount(time.Date(2013, 11, 28, 0, 0, 0, 0, time.UTC), 7)
	chm.SetCount(time.Date(2013, 11, 29, 0, 0, 0, 0, time.UTC), 1)

	t.Run("grouped by day", func(t *testing.T) {
		expected := map[string]int{
			"2013-11-13": 1,
			"2013-11-20": 1,
			"2013-11-21": 3,
			"2013-11-24": 1,
			"2013-11-25": 2,
			"2013-11-26": 8,
			"2013-11-28": 7,
			"2013-11-29": 1,
		}

		histogram := HistogramByDate(chm, "2006-01-02")
		assert.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})

	t.Run("grouped by month", func(t *testing.T) {
		expected := map[string]int{
			"2013-11": 24,
		}

		histogram := HistogramByDate(chm, "2006-01")
		assert.Len(t, histogram, len(expected))
		for i, row := range histogram {
			assert.Equal(t, expected[row.Date], row.Sum, i)
		}
	})
}
