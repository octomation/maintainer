package contribution

import (
	"sort"
	"time"
)

// HeatMap contains how many contributions were in a specific time.
type HeatMap map[time.Time]int

// Count returns how many contributions were in the specific time.
func (chm HeatMap) Count(ts time.Time) int {
	if chm == nil {
		return 0
	}

	return chm[ts]
}

// Set sets how many contributions were to the specific time.
func (chm HeatMap) Set(ts time.Time, count int) {
	if chm == nil {
		return
	}

	chm[ts] = count
}

// Histogram returns distribution of amount of contributions.
// The first value is the amount, and the second is frequency.
// The result is sorted by the first value.
func (chm HeatMap) Histogram() [][2]int {
	if chm == nil {
		return nil
	}

	histogram := make([][2]int, 0, 8)
	calc := make(map[int]int)
	for _, count := range chm {
		idx, found := calc[count]
		if !found {
			idx = len(histogram)
			histogram = append(histogram, [2]int{count, 0})
			calc[count] = idx
		}
		histogram[idx][1]++
	}
	sort.Slice(histogram, func(i, j int) bool { return histogram[i][0] < histogram[j][0] })
	return histogram
}
