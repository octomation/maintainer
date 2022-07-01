package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xhttp "go.octolab.org/toolset/maintainer/internal/pkg/http"
	xheader "go.octolab.org/toolset/maintainer/internal/pkg/http/header"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/pkg/url"
)

var overview = url.MustParse("https://github.com?tab=overview")

func (srv *Service) ContributionHeatMap(
	ctx context.Context,
	scope xtime.Range,
) (contribution.HeatMap, error) {
	u, _, err := srv.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}

	chm := make(contribution.HeatMap)
	merge := func() func(*goquery.Document, error) error {
		var mu sync.Mutex
		return func(doc *goquery.Document, err error) error {
			if err != nil {
				return err
			}

			mu.Lock()
			for ts, count := range ContributionHeatMap(doc) {
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
		group.Go(func() error { return merge(srv.FetchContributions(cascade, user, year)) })
	}

	err = group.Wait()
	return chm.Subset(scope), err
}

func (srv *Service) FetchContributions(
	ctx context.Context,
	user string, year int,
) (*goquery.Document, error) {
	src := overview.
		SetPath(user).
		AddQueryParam("from", xtime.Year(year).Location(time.UTC).Format(xtime.RFC3339Day)).
		String()
	req, err := xhttp.NewGetRequestWithContext(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("build contributions request: %w", err)
	}
	xheader.Set(req.Header).NoCache()

	resp, err := srv.client.Client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("send contributions request: %w", err)
	}
	defer safe.Close(resp.Body, unsafe.Ignore)

	return goquery.NewDocumentFromReader(resp.Body)
}
