package github

import (
	"context"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/pkg/url"
)

var overview = url.MustParse("https://github.com?tab=overview")

func (srv *service) ContributionHeatMap(
	ctx context.Context,
	since time.Time,
) (map[time.Time]int, error) {
	const layout = "2006-01-02"

	u, _, err := srv.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	src := overview.SetPath(u.GetLogin()).AddQueryParam("from", since.Format(layout)).String()
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

	chm := make(map[time.Time]int)
	doc.Find("svg.js-calendar-graph-svg rect.ContributionCalendar-day").
		Each(func(_ int, node *goquery.Selection) {
			d, _ := time.Parse(layout, node.AttrOr("data-date", ""))
			c, _ := strconv.Atoi(node.AttrOr("data-level", ""))
			chm[d] = c
		})
	return chm, nil
}
