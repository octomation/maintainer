package github

import (
	"context"
	"time"

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
	ContributionHeatMap(context.Context, time.Time) (map[time.Time]int, error)

	Labels(context.Context, github.Remote) (github.LabelSet, error)
	PatchLabels(context.Context, github.LabelSet, string) (github.LabelSet, error)
	UpdateLabels(context.Context, github.Remote, github.LabelSet) error
}
