package contribution

import (
	"time"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type DateOptions struct {
	Value time.Time
	Weeks int
	Half  bool
}

func LookupRange(opts DateOptions) xtime.Range {
	return xtime.GregorianWeeks(opts.Value.UTC(), opts.Weeks, opts.Half)
}
