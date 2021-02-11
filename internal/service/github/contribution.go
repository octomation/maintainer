package github

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/pkg/url"
)

var overview = url.MustParse("https://github.com?tab=overview")

func (srv *service) ContributionHeatMap(
	ctx context.Context,
	since time.Time,
) (contribution.HeatMap, error) {
	u, _, err := srv.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	src := overview.SetPath(u.GetLogin()).AddQueryParam("from", since.Format("2006-01-02")).String()
	req, err := http.NewGetRequestWithContext(ctx, src)
	if err != nil {
		return nil, err
	}

	resp, err := srv.client.Client().Do(req)
	if err != nil {
		return nil, err
	}
	defer safe.Close(resp.Body, unsafe.Ignore)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	min, max := contributionRange(doc)
	if expected := since.Year(); expected < min || max < expected {
		return nil, fmt.Errorf("no contribution in the %d year", expected)
	}
	chm := contributionHeatMap(doc)
	return chm, nil
}

func contributionRange(doc *goquery.Document) (int, int) {
	cr := make([]string, 0, 4)
	doc.Find("div.js-profile-timeline-year-list a.js-year-link").
		Each(func(_ int, node *goquery.Selection) {
			cr = append(cr, node.Text())
		})

	switch len(cr) {
	case 0:
		return 0, 0
	case 1:
		single, _ := strconv.Atoi(cr[0])
		return single, single
	default:
		sort.Strings(cr)
		min, _ := strconv.Atoi(cr[0])
		max, _ := strconv.Atoi(cr[len(cr)-1])
		return min, max
	}
}

func contributionHeatMap(doc *goquery.Document) contribution.HeatMap {
	chm := make(contribution.HeatMap)
	doc.Find("svg.js-calendar-graph-svg rect.ContributionCalendar-day").
		Each(func(_ int, node *goquery.Selection) {
			d, _ := time.Parse("2006-01-02", node.AttrOr("data-date", ""))
			c, _ := strconv.Atoi(node.AttrOr("data-level", ""))
			chm.Set(d, c)
		})
	return chm
}
