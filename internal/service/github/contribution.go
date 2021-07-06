package github

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.octolab.org/safe"
	"go.octolab.org/toolkit/protocol/http/header"
	"go.octolab.org/unsafe"
	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xhttp "go.octolab.org/toolset/maintainer/internal/pkg/http"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/pkg/url"
)

var overview = url.MustParse("https://github.com?tab=overview")

func (srv *service) ContributionHeatMap(
	ctx context.Context,
	scope xtime.Range,
) (contribution.HeatMap, error) {
	u, _, err := srv.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	chm := make(contribution.HeatMap)
	merge := func() func(*goquery.Document, error) error {
		var mu sync.Mutex
		return func(doc *goquery.Document, err error) error {
			if err != nil {
				return err
			}

			mu.Lock()
			for ts, count := range contributionHeatMap(doc) {
				chm[ts] = count
			}
			mu.Unlock()
			return nil
		}
	}()

	group, cascade := errgroup.WithContext(ctx)
	min, max := scope.From().Year(), scope.To().Year()
	for i, user := min, u.GetLogin(); i <= max; i++ {
		year := i
		group.Go(func() error { return merge(fetchContributions(cascade, user, year)) })
	}

	err = group.Wait()
	return chm.Subset(scope), err
}

func fetchContributions(ctx context.Context, user string, year int) (*goquery.Document, error) {
	src := overview.
		SetPath(user).
		AddQueryParam("from", time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format(xtime.RFC3339Day)).
		String()
	req, err := xhttp.NewGetRequestWithContext(ctx, src)
	if err != nil {
		return nil, err
	}
	req.Header.Set(header.CacheControl, "no-cache")

	// TODO:debt use srv.client.Client() instead
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer safe.Close(resp.Body, unsafe.Ignore)

	return goquery.NewDocumentFromReader(resp.Body)
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
			count := node.AttrOr("data-count", "")
			c, err := strconv.Atoi(count)
			if err != nil {
				panic(fmt.Errorf("invalid count value: %s", count))
			}
			if c == 0 {
				return
			}

			date := node.AttrOr("data-date", "")
			d, err := time.Parse(xtime.RFC3339Day, date)
			if err != nil {
				panic(fmt.Errorf("invalid date value: %s", date))
			}

			chm.SetCount(d, c)
		})
	return chm
}
