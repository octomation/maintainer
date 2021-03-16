package git

import "net/url"

const (
	origin = "origin"
	github = "github.com"
)

// Remote represents a connection to a remote repository.
type Remote struct {
	Name string
	URL  *url.URL
}

// Remotes represents a list of Remote with extra methods.
type Remotes []*Remote

// GitHub finds a Remote related to the GitHub.
//
// The naive implementation to proof of concept.
func (list Remotes) GitHub() (*Remote, bool) {
	for _, remote := range list {
		if remote.Name == origin && remote.URL.Host == github {
			return remote, true
		}
	}
	return nil, false
}
