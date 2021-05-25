package github

import (
	"net/http"

	"github.com/google/go-github/v39/github"
)

// New returns a new GitHub service.
func New(client *http.Client) *service {
	srv := new(service)
	srv.client = github.NewClient(client)

	return srv
}

type service struct {
	client *github.Client
}
