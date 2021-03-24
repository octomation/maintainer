package github

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/model/git"
	"go.octolab.org/toolset/maintainer/internal/model/github"
)

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test

// Git represents a Git service.
type Git interface {
	Remotes() (git.Remotes, error)
}

// GitHub represents a GitHub service.
type GitHub interface {
	Labels(context.Context, github.Remote) (*github.LabelSet, error)
}
