package labels

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/entity"
)

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test

type Provider interface {
	RepositoryWithLabels(context.Context, ...entity.RepositoryURN) ([]entity.Repository, error)
}
