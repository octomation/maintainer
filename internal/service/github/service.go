package github

import (
	"net/http"

	"github.com/google/go-github/v88/github"
)

// New returns a new GitHub service.
func New(client *http.Client) *Service {
	srv := new(Service)
	// WithHTTPClient never fails for a non-enterprise client.
	srv.client, _ = github.NewClient(github.WithHTTPClient(client))

	return srv
}

type Service struct {
	client *github.Client
}
