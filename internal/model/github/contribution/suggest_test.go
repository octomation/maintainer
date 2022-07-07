package contribution_test

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestSuggest(t *testing.T) {
	tests := map[string]struct {
		// input
		chm   HeatMap
		start time.Time
		end   time.Time
		basis int

		// output
		expected HistogramByWeekdayRow
	}{
		"issue#68: missed zero": {
			golden(t, "issue-68.golden.json"),
			xtime.UTC().Year(2021).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.September).Day(11).Time(),
				Sum: 7,
			},
		},
		"full week with some distribution": {
			golden(t, "issue-68.golden.json"),
			xtime.UTC().Year(2021).Month(time.September).Day(15).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.September).Day(12).Time(),
				Sum: 12,
			},
		},
		"week without contributions": {
			golden(t, "issue-68.golden.json"),
			xtime.UTC().Year(2021).Month(time.October).Day(7).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(3).Time(),
				Sum: 5,
			},
		},
		"week with gaps": {
			golden(t, "issue-68.golden.json"),
			xtime.UTC().Year(2021).Month(time.October).Day(16).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(11).Time(),
				Sum: 8,
			},
		},
		"empty contribution heatmap": {
			make(HeatMap),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			time.Now().UTC(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(3).Time(),
				Sum: 5,
			},
		},
		"no range": {
			make(HeatMap),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
			5,

			HistogramByWeekdayRow{
				Day: xtime.UTC().Year(2021).Month(time.October).Day(5).Time(),
				Sum: 5,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Suggest(test.chm, test.start, test.end, test.basis))
		})
	}
}

func golden(t testing.TB, name string) HeatMap {
	src := NewFileSource(afero.NewBasePathFs(afero.NewOsFs(), "testdata"), name)
	chm, err := src.Fetch(context.Background())
	require.NoError(t, err)
	return chm
}
