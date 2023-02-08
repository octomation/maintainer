package github

import (
	"net/http"

	"github.com/google/go-github/v68/github"
)

// New returns a new GitHub service.
func New(client *http.Client) *Service {
	srv := new(Service)
	srv.client = github.NewClient(client)

	return srv
}

type Service struct {
	client *github.Client
}
