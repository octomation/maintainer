package contribution

import (
	"sort"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

type DateOptions struct {
	Value time.Time
	Weeks int
	Half  bool
}

func LookupRange(opts DateOptions) xtime.Range {
	return ShiftRange(xtime.RangeByWeeks(opts.Value, opts.Weeks, opts.Half)).ExcludeFuture()
}

func ShiftRange(r xtime.Range) xtime.Range {
	if r.Base().Weekday() == time.Sunday {
		return r.Shift(6 * xtime.Day)
	}
	return r.Shift(-xtime.Day)
}

func YearRange(doc *goquery.Document) (int, int) {
	cr := make([]string, 0, 4)
	doc.Find("div.js-profile-timeline-year-list a.js-year-link").
		Each(func(_ int, node *goquery.Selection) { cr = append(cr, node.Text()) })

	switch len(cr) {
	case 0:
		return 0, 0
	case 1:
		single, err := strconv.Atoi(cr[0])
		if err != nil {
			panic(ContentError{
				error:   err,
				Content: cr[0],
			})
		}
		return single, single
	default:
		sort.Strings(cr)
		min, err := strconv.Atoi(cr[0])
		if err != nil {
			panic(ContentError{
				error:   err,
				Content: cr[0],
			})
		}
		max, err := strconv.Atoi(cr[len(cr)-1])
		if err != nil {
			panic(ContentError{
				error:   err,
				Content: cr[len(cr)-1],
			})
		}
		return min, max
	}
}
