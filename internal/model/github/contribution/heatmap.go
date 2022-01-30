package contribution

import (
	"sort"

	"go.octolab.org/toolset/maintainer/internal/pkg/time"
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
func (chm HeatMap) Subset(scope time.Range) HeatMap {
	subset := make(HeatMap)

	for ts, count := range chm {
		if scope.Contains(ts) {
			subset[ts] = count
		}
	}

	return subset
}

// Diff calculates the difference between two heatmaps.
func (chm HeatMap) Diff(src HeatMap) HeatMap {
	diff := make(HeatMap)

	keys := make(map[time.Time]struct{}, len(chm)+len(src))
	for ts := range chm {
		keys[ts] = struct{}{}
	}
	for ts := range src {
		keys[ts] = struct{}{}
	}
	for ts := range keys {
		if delta := src[ts] - chm[ts]; delta != 0 {
			diff[ts] = delta
		}
	}

	return diff
}

// From returns minimum time of the heatmap, otherwise the zero time instant.
func (chm HeatMap) From() time.Time {
	var min time.Time
	for ts := range chm {
		if ts.Before(min) || min.IsZero() {
			min = ts
		}
	}
	return min
}

// To returns maximum time of the heatmap, otherwise the zero time instant.
func (chm HeatMap) To() time.Time {
	var max time.Time
	for ts := range chm {
		if ts.After(max) {
			max = ts
		}
	}
	return max
}

// Range returns time range of the heatmap, otherwise the zero time range instant.
func (chm HeatMap) Range() time.Range {
	return time.NewRange(chm.From(), chm.To())
}

type HistogramByCountRow struct {
	Count, Frequency int
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

	for _, fn := range order {
		sort.Sort(fn(h))
	}
	return h
}

type HistogramByDateRow struct {
	Date string
	Sum  int
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
	Sum int
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
		current := time.TruncateToDay(ts)

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

// Suggest finds a week with gaps in the contribution heatmap
// and returns an appropriate day to contribute.
func Suggest(
	chm HeatMap,
	start time.Time,
	end time.Time,
	basis int,
) HistogramByWeekdayRow {
	defaults := HistogramByWeekdayRow{
		Day: start,
		Sum: basis,
	}

	for t := start; t.Before(end); t = t.Add(time.Week) {
		week := time.RangeByWeeks(t, 0, false).Shift(-time.Day) // shift Sunday
		data := HistogramByCount(chm.Subset(week), OrderByCount)
		sunday := week.From()

		// good week: no gaps and enough contributions
		if len(data) == 1 && data[0].Count >= defaults.Sum {
			continue
		}

		// bad week: no contributions
		if len(data) == 0 {
			return HistogramByWeekdayRow{Day: sunday, Sum: defaults.Sum}
		}

		// otherwise, we choose the maximum amount of contributions
		// it's the last element in the histogram because it's sorted by count ASC
		count := data[len(data)-1].Count
		if count < defaults.Sum {
			count = defaults.Sum
		}
		// and try to find an appropriate day to contribute
		day := sunday
		for i := time.Sunday; i <= time.Saturday; i++ {
			if chm[day] != count {
				break
			}
			day = day.Add(time.Day)
		}
		return HistogramByWeekdayRow{Day: day, Sum: count}
	}

	return defaults
}
