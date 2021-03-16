package provider

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

// FallbackTo returns a fallback mechanism to handle
// case when git.ErrRepositoryNotExists is occurred.
// If it happens it returns a stub of git.Repository.
// If it impossible it raises a panic.
func FallbackTo(remote *string) interface {
	Apply(*git.Repository, error) *git.Repository
} {
	return fallback(func(repo *git.Repository, err error) *git.Repository {
		if err == nil {
			return repo
		}

		if !errors.Is(err, git.ErrRepositoryNotExists) {
			panic(err)
		}

		repo, err = git.Init(memory.NewStorage(), nil)
		if err != nil {
			panic(err)
		}

		// TODO:naive
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name:  "origin",
			URLs:  []string{*remote},
			Fetch: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
		})
		if err != nil {
			panic(err)
		}

		return repo
	})
}

type fallback func(*git.Repository, error) *git.Repository

func (fn fallback) Apply(repo *git.Repository, err error) *git.Repository {
	return fn(repo, err)
}
