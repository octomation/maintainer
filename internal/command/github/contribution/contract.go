package contribution

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
)

type ContributionSource interface {
	Location() string
	Fetch(context.Context) (contribution.HeatMap, error)
}
