package contribution

import (
	"sort"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"

	"go.octolab.org/toolset/maintainer/internal/pkg/errors"
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

type Year = int

func YearRange(doc *goquery.Document) (Year, Year) {
	cr := make([]string, 0, 4)
	doc.Find("div.js-profile-timeline-year-list a.js-year-link").
		Each(func(_ int, node *goquery.Selection) { cr = append(cr, node.Text()) })

	switch len(cr) {
	case 0:
		return 0, 0
	case 1:
		single, err := strconv.Atoi(cr[0])
		if err != nil {
			panic(errors.ContentError(err, cr[0]))
		}
		return single, single
	default:
		sort.Strings(cr)
		min, err := strconv.Atoi(cr[0])
		if err != nil {
			panic(errors.ContentError(err, cr[0]))
		}
		max, err := strconv.Atoi(cr[len(cr)-1])
		if err != nil {
			panic(errors.ContentError(err, cr[len(cr)-1]))
		}
		return min, max
	}
}
