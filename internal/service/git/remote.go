package git

import (
	"go.octolab.org/toolset/maintainer/internal/model/git"
	"go.octolab.org/toolset/maintainer/internal/pkg/giturl"
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
		link, err := giturl.Parse(config.URLs[0]) // TODO:unsafe
		if err != nil {
			// a remote helper or an unknown transport must not hide the other remotes
			continue
		}

		result = append(result, git.Remote{
			Name: config.Name,
			URL:  link,
		})
	}
	return result, nil
}
