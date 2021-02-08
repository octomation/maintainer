package git

import "net/url"

const (
	bitbucket = "bitbucket.org"
	github    = "github.com"
	gitlab    = "gitlab.com"
	origin    = "origin"
	mirror    = "mirror"
)

// Remote represents a connection to a remote repository.
type Remote struct {
	Name string
	URL  *url.URL
}

// Remotes represents a list of Remote with extra methods.
type Remotes []Remote

// Bitbucket finds a Remote related to the Bitbucket.
//
// The naive implementation to proof of concept.
func (list Remotes) Bitbucket() (Remote, bool) {
	for _, remote := range list {
		// TODO:naive:unsafe
		if remote.Name == mirror && remote.URL.Host == bitbucket {
			return remote, true
		}
	}
	return Remote{}, false
}

// GitHub finds a Remote related to the GitHub.
//
// The naive implementation to proof of concept.
func (list Remotes) GitHub() (Remote, bool) {
	for _, remote := range list {
		// TODO:naive:unsafe
		if remote.Name == origin && remote.URL.Host == github {
			return remote, true
		}
	}
	return Remote{}, false
}

// GitLab finds a Remote related to the GitLab.
//
// The naive implementation to proof of concept.
func (list Remotes) GitLab() (Remote, bool) {
	for _, remote := range list {
		// TODO:naive:unsafe
		if remote.Name == mirror && remote.URL.Host == gitlab {
			return remote, true
		}
	}
	return Remote{}, false
}
