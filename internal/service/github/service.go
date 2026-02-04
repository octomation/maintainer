package github

import (
	"io"
	"net/http"
	"time"

	"github.com/google/go-github/v93/github"
)

// concurrentRequests is the most GraphQL requests in flight at once:
// a year of contributions takes a request a month, and a burst of them
// is unwelcome.
const concurrentRequests = 4

// New returns a new GitHub service.
func New(client *http.Client) *Service {
	srv := &Service{now: time.Now, slots: make(chan struct{}, concurrentRequests), warn: io.Discard}
	// WithHTTPClient never fails for a non-enterprise client.
	srv.client, _ = github.NewClient(github.WithHTTPClient(client))

	return srv
}

type Service struct {
	client *github.Client
	now    func() time.Time
	slots  chan struct{}

	cache string    // the directory of cached contributions, none if empty
	warn  io.Writer // where the contributions without a date are reported
}

// WithCache keeps the contribution events of past fetches in the directory,
// so a month whose total is unchanged is not fetched again.
func (srv *Service) WithCache(dir string) *Service {
	srv.cache = dir
	return srv
}

// WithWarnings reports to w what the service cannot show, such as
// the contributions the API counts but gives no date for.
func (srv *Service) WithWarnings(w io.Writer) *Service {
	srv.warn = w
	return srv
}
