package git

import (
	giturls "github.com/whilp/git-urls"

	"go.octolab.org/toolset/maintainer/internal/model/git"
)

// Remotes returns a list with all the remotes.
func (srv *service) Remotes() (git.Remotes, error) {
	list, err := srv.repo.Remotes()
	if err != nil {
		return nil, err
	}

	result := make([]git.Remote, 0, len(list))
	for _, remote := range list {
		config := remote.Config()
		link, err := giturls.Parse(config.URLs[0]) // TODO:unsafe
		if err != nil {
			return nil, err
		}

		result = append(result, git.Remote{
			Name: config.Name,
			URL:  link,
		})
	}
	return result, nil
}
