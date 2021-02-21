package contribution

import (
	"sort"
	"time"
)

// HeatMap contains how many contributions were in a specific time.
type HeatMap map[time.Time]int

// Count returns how many contributions were in the specific time.
func (chm HeatMap) Count(ts time.Time) int {
	return chm[ts]
}

// SetCount sets how many contributions were to the specific time.
func (chm HeatMap) SetCount(ts time.Time, count int) {
	chm[ts] = count
}

type histogramByCountRow struct {
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
func HistogramByCount(chm HeatMap) []histogramByCountRow {
	h := make([]histogramByCountRow, 0, 8)
	m := make(map[int]int)

	for _, count := range chm {
		idx, found := m[count]
		if !found {
			idx = len(h)
			h = append(h, histogramByCountRow{Count: count})
			m[count] = idx
		}
		h[idx].Frequency++
	}

	sort.Slice(h, func(i, j int) bool { return h[i].Count < h[j].Count })
	return h
}

type histogramByDateRow struct {
	Date string
	Sum  int
}

// HistogramByDate returns the sum of the number of contributions grouped by date.
// The first value is a date in the specified format, and the second is a sum.
// The result is sorted by the first value.
//
//  2006-01-02 #
//  2006-01-04 ###
//  2006-01-05 ##
//  2006-02-01 #
//
//  2006-01    ######
//  2006-02    #
//
func HistogramByDate(chm HeatMap, format string) []histogramByDateRow {
	h := make([]histogramByDateRow, 0, 8)
	m := make(map[string]int)

	for ts, count := range chm {
		date := ts.Format(format)
		idx, found := m[date]
		if !found {
			idx = len(h)
			h = append(h, histogramByDateRow{Date: date})
			m[date] = idx
		}
		h[idx].Sum += count
	}

	sort.Slice(h, func(i, j int) bool { return h[i].Date < h[j].Date })
	return h
}
