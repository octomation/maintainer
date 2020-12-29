package github

import (
	"net/http"

	"github.com/google/go-github/v33/github"
)

func New(client *http.Client) *manager {
	manager := &manager{client: github.NewClient(client)}
	return manager
}

type manager struct {
	client *github.Client
}
