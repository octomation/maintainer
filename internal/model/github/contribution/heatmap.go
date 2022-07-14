package contribution

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"

	"github.com/PuerkitoBio/goquery"
)

// HeatMap contains how many contributions have been made in a time.
type HeatMap map[time.Time]uint

// Count returns how many contributions have been made in the specified time.
func (chm HeatMap) Count(ts time.Time) uint {
	return chm[ts]
}

// SetCount sets how many contributions have been made to the specified time.
func (chm HeatMap) SetCount(ts time.Time, count uint) {
	chm[ts] = count
}

// Subset returns a subset of contribution heatmap in the provided time range.
// TODO:perf improve algorithm
func (chm HeatMap) Subset(scope xtime.Range) HeatMap {
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
func (chm HeatMap) Range() xtime.Range {
	return xtime.NewRange(chm.From(), chm.To())
}

var counter = regexp.MustCompile(`^\d+`)

func BuildHeatMap(doc *goquery.Document) HeatMap {
	chm := make(HeatMap)
	doc.Find("svg.js-calendar-graph-svg rect.ContributionCalendar-day").
		Each(func(_ int, node *goquery.Selection) {
			// data-count="0"
			// data-count="2"
			count, has := node.Attr("data-count")
			if !has {
				// No contributions on January 2, 2006
				// 2 contributions on January 2, 2006
				count = counter.FindString(node.Text())
				if count == "" {
					count = "0"
				}
			}
			c, err := strconv.ParseUint(count, 10, 0)
			if err != nil {
				html, _ := node.Html()
				panic(ContentError{
					error:   fmt.Errorf("invalid count value: %w", err),
					Content: html,
				})
			}

			// data-date="2006-01-02"
			date := node.AttrOr("data-date", "")
			d, err := time.Parse(xtime.RFC3339Day, date)
			if err != nil {
				html, _ := node.Html()
				panic(ContentError{
					error:   fmt.Errorf("invalid date value: %w", err),
					Content: html,
				})
			}

			chm.SetCount(d, uint(c))
		})
	return chm
}
