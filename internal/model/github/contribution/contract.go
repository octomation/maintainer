package contribution

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type Contributor interface {
	ContributionHeatMap(context.Context, time.Range) (HeatMap, error)
}
