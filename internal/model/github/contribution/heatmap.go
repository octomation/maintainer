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
//  format: "2006-01-02"
//  	2006-01-02 #
//  	2006-01-04 ###
//  	2006-01-05 ##
//  	2006-02-01 #
//
//  format: "2006-01"
//  	2006-01    ######
//  	2006-02    #
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
	Day time.Weekday
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
	h := make([]hbw, 0, 8)
	m := make(map[time.Weekday]int)

	f := make([]time.Time, 0, len(chm))
	for ts := range chm {
		f = append(f, ts)
	}
	sort.Slice(f, func(i, j int) bool { return f[i].Before(f[j]) })

	var pd, py int
	for _, ts := range f {
		day := ts.Weekday()
		cd, cy := ts.YearDay(), ts.Year()
		idx, found := m[day]
		if !found || (!grouped && (pd != cd || py != cy)) {
			idx = len(h)
			h = append(h, hbw{Day: day})
			m[day] = idx

			pd, py = cd, cy
		}
		h[idx].Sum += chm[ts]
	}

	sort.Slice(h, func(i, j int) bool { return h[i].Day < h[j].Day })
	return h
}
