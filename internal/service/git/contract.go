package git

import "github.com/go-git/go-git/v5"

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test

// Repository represents a Git repository.
type Repository interface {
	Remotes() ([]*git.Remote, error)
}
