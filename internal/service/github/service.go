package github

import (
	"net/http"
	"time"

	"github.com/google/go-github/v91/github"
)

// New returns a new GitHub service.
func New(client *http.Client) *Service {
	srv := &Service{now: time.Now}
	// WithHTTPClient never fails for a non-enterprise client.
	srv.client, _ = github.NewClient(github.WithHTTPClient(client))

	return srv
}

type Service struct {
	client *github.Client
	now    func() time.Time
}
