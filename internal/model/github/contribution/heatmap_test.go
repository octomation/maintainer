package contribution_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestDayDiff_Delta(t *testing.T) {
	tests := map[string]struct {
		diff     DayDiff
		expected int64
	}{
		"grown to max":     {DayDiff{Before: 0, After: math.MaxInt64}, math.MaxInt64},
		"dropped from max": {DayDiff{Before: math.MaxInt64, After: 0}, -math.MaxInt64},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.diff.Delta())
		})
	}
}

func TestHeatMap_Count(t *testing.T) {
	chm := load(t, "testdata/kamilsk.2019.json")
	ts := xtime.UTC().Year(2019).Month(time.November).Day(13).Time()
	assert.Equal(t, uint(3), chm.Count(ts), "2019-11-13")
	assert.Equal(t, uint(2), chm.Count(ts.AddDate(0, 1, 0)), "2019-12-13")
	assert.Equal(t, uint(3), chm.Count(ts.AddDate(0, 1, 14)), "2019-12-27")
}

func TestHeatMap_Diff(t *testing.T) {
	Jan2022 := xtime.UTC().Year(2022).Month(time.January)
	newYearsEve := xtime.UTC().Year(2021).Month(time.December).Day(31).Time()
	newYear := Jan2022.Day(1).Time()
	sunday, monday, tuesday := Jan2022.Day(2).Time(), Jan2022.Day(3).Time(), Jan2022.Day(4).Time()
	require.Equal(t, time.Sunday, sunday.Weekday())
	require.Equal(t, time.Monday, monday.Weekday())

	tests := []struct {
		name     string
		base     HeatMap
		head     HeatMap
		expected []DayDiff
		deltas   []int64
	}{
		{
			name:     "changed Sunday",
			base:     HeatMap{newYear: 1, sunday: 0, monday: 3},
			head:     HeatMap{newYear: 1, sunday: 4, monday: 3},
			expected: []DayDiff{{Day: sunday, Before: 0, After: 4, InBase: true, InHead: true}},
			deltas:   []int64{+4},
		},
		{
			name:     "changed Monday",
			base:     HeatMap{sunday: 0, monday: 3, tuesday: 5},
			head:     HeatMap{sunday: 0, monday: 7, tuesday: 5},
			expected: []DayDiff{{Day: monday, Before: 3, After: 7, InBase: true, InHead: true}},
			deltas:   []int64{+4},
		},
		{
			name: "changes across the year boundary",
			base: HeatMap{newYearsEve: 0, newYear: 0, sunday: 1},
			head: HeatMap{newYearsEve: 3, newYear: 2, sunday: 1},
			expected: []DayDiff{
				{Day: newYearsEve, Before: 0, After: 3, InBase: true, InHead: true},
				{Day: newYear, Before: 0, After: 2, InBase: true, InHead: true},
			},
			deltas: []int64{+3, +2},
		},
		{
			name:     "decrease is negative",
			base:     HeatMap{tuesday: 5},
			head:     HeatMap{tuesday: 2},
			expected: []DayDiff{{Day: tuesday, Before: 5, After: 2, InBase: true, InHead: true}},
			deltas:   []int64{-3},
		},
		{
			name:     "day only in base",
			base:     HeatMap{newYearsEve: 2, newYear: 1},
			head:     HeatMap{newYear: 1},
			expected: []DayDiff{{Day: newYearsEve, Before: 2, After: 0, InBase: true, InHead: false}},
			deltas:   []int64{-2},
		},
		{
			name:     "day only in head",
			base:     HeatMap{newYear: 1},
			head:     HeatMap{newYearsEve: 2, newYear: 1},
			expected: []DayDiff{{Day: newYearsEve, Before: 0, After: 2, InBase: false, InHead: true}},
			deltas:   []int64{+2},
		},
		{
			name:     "missing day counts as zero",
			base:     HeatMap{newYearsEve: 0, newYear: 1},
			head:     HeatMap{newYear: 1, sunday: 0},
			expected: nil,
			deltas:   nil,
		},
		{
			name:     "identical inputs",
			base:     load(t, "testdata/kamilsk.2021.json"),
			head:     load(t, "testdata/kamilsk.2021.json"),
			expected: nil,
			deltas:   nil,
		},
		{
			name: "sorted by day",
			base: HeatMap{tuesday: 5, newYearsEve: 1, monday: 3},
			head: HeatMap{monday: 1, sunday: 2, newYear: 4},
			expected: []DayDiff{
				{Day: newYearsEve, Before: 1, After: 0, InBase: true, InHead: false},
				{Day: newYear, Before: 0, After: 4, InBase: false, InHead: true},
				{Day: sunday, Before: 0, After: 2, InBase: false, InHead: true},
				{Day: monday, Before: 3, After: 1, InBase: true, InHead: true},
				{Day: tuesday, Before: 5, After: 0, InBase: true, InHead: false},
			},
			deltas: []int64{-1, +4, +2, -2, -5},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diff := test.base.Diff(test.head)
			require.Len(t, diff, len(test.expected))
			for i, day := range diff {
				assert.Equal(t, test.expected[i], day)
				assert.Equal(t, test.deltas[i], day.Delta())
			}
		})
	}
}

func TestHeatMap_Only(t *testing.T) {
	Sep2026 := xtime.UTC().Year(2026).Month(time.September)
	day := func(d int) time.Time { return Sep2026.Day(d).Time() }

	tests := map[string]struct {
		left, right HeatMap
		expected    []time.Time
	}{
		"same days": {HeatMap{day(1): 1, day(2): 0}, HeatMap{day(1): 3, day(2): 0}, []time.Time{}},
		"only left, sorted": {
			HeatMap{day(3): 0, day(1): 1, day(2): 2, day(5): 0},
			HeatMap{day(2): 2},
			[]time.Time{day(1), day(3), day(5)},
		},
		"only right":  {HeatMap{day(1): 1}, HeatMap{day(1): 1, day(2): 0}, []time.Time{}},
		"empty left":  {HeatMap{}, HeatMap{day(1): 1}, []time.Time{}},
		"empty right": {HeatMap{day(2): 0, day(1): 1}, HeatMap{}, []time.Time{day(1), day(2)}},
		"both empty":  {nil, nil, []time.Time{}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.left.Only(test.right))
		})
	}
}
