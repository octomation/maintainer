package contribution

import (
	"sort"
	"time"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type HistogramByCountRow struct {
	Count, Frequency uint
}

type hbc = HistogramByCountRow

type orderByCount []HistogramByCountRow

func (list orderByCount) Len() int           { return len(list) }
func (list orderByCount) Less(i, j int) bool { return list[i].Count < list[j].Count }
func (list orderByCount) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

func OrderByCount(in []HistogramByCountRow) sort.Interface { return orderByCount(in) }

type orderByFrequency []HistogramByCountRow

func (list orderByFrequency) Len() int           { return len(list) }
func (list orderByFrequency) Less(i, j int) bool { return list[i].Frequency < list[j].Frequency }
func (list orderByFrequency) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

func OrderByFrequency(in []HistogramByCountRow) sort.Interface { return orderByFrequency(in) }

// HistogramByCount returns the distribution of amount contributions.
//
//	1 #
//	3 #####
//	4 ##
//	7 ###
func HistogramByCount(chm HeatMap, order ...func([]hbc) sort.Interface) []HistogramByCountRow {
	h := make([]hbc, 0, 8)
	m := make(map[uint]int)

	for _, count := range chm {
		idx, found := m[count]
		if !found {
			idx = len(h)
			h = append(h, hbc{Count: count})
			m[count] = idx
		}
		h[idx].Frequency++
	}

	for _, fn := range order {
		sort.Sort(fn(h))
	}
	return h
}

type HistogramByDateRow struct {
	Date string
	Sum  uint
}

type hbd = HistogramByDateRow

// HistogramByDate returns the sum of the number of contributions grouped by date.
// The first value is a date in the specified format, and the second is a sum.
// The result is sorted by the first value.
//
//	format: time.RFC3339Day
//		2022-01-02 #
//		2022-01-04 ###
//		2022-01-05 ##
//		2022-02-01 #
//
//	format: time.RFC3339Month
//		2022-01    ######
//		2022-02    #
func HistogramByDate(chm HeatMap, format string) []HistogramByDateRow {
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

type HistogramByWeekdayRow struct {
	Day time.Time
	Sum uint
}

type hbw = HistogramByWeekdayRow

// HistogramByWeekday returns the sum of the number of contributions grouped by day of week.
// The first value is a date in the specified format, and the second is a sum.
// The result is sorted by the first value.
//
//	grouped: false
//		Monday  #
//		Tuesday ###
//		Friday  ##
//		Monday  #
//
//	grouped: true
//		Monday  ##
//		Tuesday ###
//		Friday  ##
func HistogramByWeekday(chm HeatMap, grouped bool) []HistogramByWeekdayRow {
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
		h[idx].Sum += chm.Count(ts)
	}

	return h
}
