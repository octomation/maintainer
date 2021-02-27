package contribution

import (
	"sort"
	"time"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// HeatMap contains how many contributions have been made in a time.
type HeatMap map[time.Time]int

// Count returns how many contributions have been made in the specified time.
func (chm HeatMap) Count(ts time.Time) int {
	return chm[ts]
}

// SetCount sets how many contributions have been made to the specified time.
func (chm HeatMap) SetCount(ts time.Time, count int) {
	chm[ts] = count
}

// Subset returns a subset of contribution heatmap in the provided time range.
func (chm HeatMap) Subset(scope xtime.Range) HeatMap {
	subset := make(HeatMap)

	for ts, count := range chm {
		if scope.Contains(ts) {
			subset[ts] = count
		}
	}

	return subset
}

type hbc struct {
	Count, Frequency int
}

// HistogramByCount returns the distribution of amount contributions.
// The first value is an amount, and the second is a frequency.
// The result is sorted by the first value.
//
//  1 #
//  3 #####
//  4 ##
//  7 ###
//
func HistogramByCount(chm HeatMap) []hbc {
	h := make([]hbc, 0, 8)
	m := make(map[int]int)

	for _, count := range chm {
		idx, found := m[count]
		if !found {
			idx = len(h)
			h = append(h, hbc{Count: count})
			m[count] = idx
		}
		h[idx].Frequency++
	}

	sort.Slice(h, func(i, j int) bool { return h[i].Count < h[j].Count })
	return h
}

type hbd struct {
	Date string
	Sum  int
}

// HistogramByDate returns the sum of the number of contributions grouped by date.
// The first value is a date in the specified format, and the second is a sum.
// The result is sorted by the first value.
//
//  format: time.RFC3339Day
//  	2022-01-02 #
//  	2022-01-04 ###
//  	2022-01-05 ##
//  	2022-02-01 #
//
//  format: time.RFC3339Month
//  	2022-01    ######
//  	2022-02    #
//
func HistogramByDate(chm HeatMap, format string) []hbd {
	h := make([]hbd, 0, 8)
	m := make(map[string]int)

	for ts, count := range chm {
		date := ts.Format(format)
		idx, found := m[date]
		if !found {
			idx = len(h)
			h = append(h, hbd{Date: date})
			m[date] = idx
		}
		h[idx].Sum += count
	}

	sort.Slice(h, func(i, j int) bool { return h[i].Date < h[j].Date })
	return h
}

type hbw struct {
	Day time.Time
	Sum int
}

// HistogramByWeekday returns the sum of the number of contributions grouped by day of week.
// The first value is a date in the specified format, and the second is a sum.
// The result is sorted by the first value.
//
//  grouped: false
//  	Monday  #
//  	Tuesday ###
//  	Friday  ##
//  	Monday  #
//
//  grouped: true
//  	Monday  ##
//  	Tuesday ###
//  	Friday  ##
//
func HistogramByWeekday(chm HeatMap, grouped bool) []hbw {
	f := make([]time.Time, 0, len(chm))
	for ts := range chm {
		f = append(f, ts)
	}
	sort.Slice(f, func(i, j int) bool { return f[i].Before(f[j]) })
	h := make([]hbw, 0, 8)
	m := make(map[time.Weekday]int)

	var prev time.Time
	for _, ts := range f {
		weekday := ts.Weekday()
		current := xtime.TruncateToDay(ts)

		idx, found := m[weekday]
		if !found || (!grouped && !prev.Equal(current)) {
			idx = len(h)
			h = append(h, hbw{Day: current})
			m[weekday] = idx

			prev = current
		}
		h[idx].Sum += chm[ts]
	}

	return h
}
