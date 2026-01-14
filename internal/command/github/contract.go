package github

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/model/git"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test
//go:generate goimports -local go.octolab.org/toolset/maintainer -w mocks_test.go

// Git represents a Git service.
type Git interface {
	Remotes() (git.Remotes, error)
}

// GitHub represents a GitHub service.
type GitHub interface {
	ContributionHeatMap(context.Context, time.Range) (contribution.HeatMap, error)
}
