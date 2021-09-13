package run

import (
	"context"
	"encoding/json"
	"os"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func ContributionDiff(
	ctx context.Context,
	service Contributor,
	printer Printer,

	date time.Time,
	src, dst *os.File,
	baseSource, headSource string,
) error {
	var (
		base contribution.HeatMap
		head contribution.HeatMap
	)
	if err := json.NewDecoder(dst).Decode(&base); err != nil {
		return err
	}
	if src != nil {
		if err := json.NewDecoder(src).Decode(&head); err != nil {
			return err
		}
	} else {
		var err error
		scope := time.RangeByYears(date, 0, false).ExcludeFuture()
		head, err = service.ContributionHeatMap(ctx, scope)
		if err != nil {
			return err
		}
	}

	return view.ContributionDiff(printer, base.Diff(head), baseSource, headSource)
}
