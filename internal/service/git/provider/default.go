package provider

import "github.com/go-git/go-git/v5"

// Default opens a Git repository from the working directory.
// It walks parent directories until found a .git directory.
// It returns git.ErrRepositoryNotExists if the working directory
// doesn't contain a valid repository.
func Default() (*git.Repository, error) {
	opt := new(git.PlainOpenOptions)
	opt.DetectDotGit = true

	return git.PlainOpenWithOptions("", opt)
}
