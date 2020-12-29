package labels

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/entity"
)

type Provider interface {
	RepositoryWithLabels(context.Context, ...entity.RepositoryURN) ([]entity.Repository, error)
}
