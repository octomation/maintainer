package github

import (
	"strings"

	"go.octolab.org/toolset/maintainer/internal/model/git"
	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

const (
	suf = ".git"
	sep = "/"
)

// Remote represents a connection to a remote GitHub repository.
type Remote git.Remote

// ID returns "{owner}/{repo}" as a remote identifier.
func (remote Remote) ID() string {
	assert.True(func() bool { return remote.URL != nil })
	return strings.TrimSuffix(remote.URL.Path, suf)
}

// OwnerAndName returns an owner and repository name.
func (remote Remote) OwnerAndName() (string, string) {
	parts := strings.Split(remote.ID(), sep)

	assert.True(func() bool { return len(parts) == 2 })
	return parts[0], parts[1]
}

// Repository represents a GitHub repository.
type Repository struct {
	Remote
	ID     int64
	Labels []Label
}
