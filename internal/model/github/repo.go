package github

import (
	"strings"

	"go.octolab.org/toolset/maintainer/internal/model/git"
)

const (
	suf = ".git"
	sep = "/"
)

// Repository represents a GitHub repository.
type Repository struct {
	*git.Remote
	ID     int64
	Labels []Label
}

// OwnerAndName returns an owner and repository name.
//
// The naive implementation to proof of concept.
func (repo *Repository) OwnerAndName() (string, string) {
	parts := strings.Split(
		strings.TrimSuffix(
			repo.Remote.URL.Path, // TODO:unsafe
			suf,
		),
		sep,
	)
	return parts[0], parts[1] // TODO:unsafe
}
