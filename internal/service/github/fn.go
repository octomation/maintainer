package github

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"

	"github.com/PuerkitoBio/goquery"
)

type contentError struct {
	error
	Content string
}

var contributionCount = regexp.MustCompile(`^\d+`)

func ContributionHeatMap(doc *goquery.Document) contribution.HeatMap {
	chm := make(contribution.HeatMap)
	doc.Find("svg.js-calendar-graph-svg rect.ContributionCalendar-day").
		Each(func(_ int, node *goquery.Selection) {
			// data-count="0"
			// data-count="2"
			count, has := node.Attr("data-count")
			if !has {
				// No contributions on January 2, 2006
				// 2 contributions on January 2, 2006
				count = contributionCount.FindString(node.Text())
				if count == "" {
					count = "0"
				}
			}
			c, err := strconv.Atoi(count)
			if err != nil {
				html, _ := node.Html()
				panic(contentError{
					error:   fmt.Errorf("invalid count value: %w", err),
					Content: html,
				})
			}

			// data-date="2006-01-02"
			date := node.AttrOr("data-date", "")
			d, err := time.Parse(xtime.RFC3339Day, date)
			if err != nil {
				html, _ := node.Html()
				panic(contentError{
					error:   fmt.Errorf("invalid date value: %w", err),
					Content: html,
				})
			}

			chm.SetCount(d, c)
		})
	return chm
}

func ContributionRange(doc *goquery.Document) (int, int) {
	cr := make([]string, 0, 4)
	doc.Find("div.js-profile-timeline-year-list a.js-year-link").
		Each(func(_ int, node *goquery.Selection) { cr = append(cr, node.Text()) })

	switch len(cr) {
	case 0:
		return 0, 0
	case 1:
		single, err := strconv.Atoi(cr[0])
		if err != nil {
			panic(contentError{
				error:   err,
				Content: cr[0],
			})
		}
		return single, single
	default:
		sort.Strings(cr)
		min, err := strconv.Atoi(cr[0])
		if err != nil {
			panic(contentError{
				error:   err,
				Content: cr[0],
			})
		}
		max, err := strconv.Atoi(cr[len(cr)-1])
		if err != nil {
			panic(contentError{
				error:   err,
				Content: cr[len(cr)-1],
			})
		}
		return min, max
	}
}
