package run

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type Contributor interface {
	ContributionHeatMap(context.Context, time.Range) (contribution.HeatMap, error)
}

type Printer interface {
	Println(...interface{})
}
