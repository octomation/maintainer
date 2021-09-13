package run

import (
	"context"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
)

func ContributionDiff(
	ctx context.Context,
	src, dst ContributionSource,
	printer Printer,
) error {
	base, err := src.Fetch(ctx)
	if err != nil {
		return err
	}

	head, err := dst.Fetch(ctx)
	if err != nil {
		return err
	}

	return view.ContributionDiff(printer, base.Diff(head), src.Location(), dst.Location())
}
