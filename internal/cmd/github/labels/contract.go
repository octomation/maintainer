package labels

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/entity/github"
)

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test

type Provider interface {
	RepositoryWithLabels(context.Context, ...github.RepositoryURN) ([]github.Repository, error)
}
