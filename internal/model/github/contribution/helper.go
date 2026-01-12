package contribution

import (
	"time"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// MinYear is the earliest year of a date the contribution commands accept.
// GitHub has no contributions long before it, and a range built around
// an earlier date could reach the zero time instant, which ranges reject.
const MinYear = 1970

type DateOptions struct {
	Value time.Time
	Weeks int
	Half  bool
}

func LookupRange(opts DateOptions) xtime.Range {
	return xtime.GregorianWeeks(opts.Value.UTC(), opts.Weeks, opts.Half)
}
