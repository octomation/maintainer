package contribution

import (
	"sort"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

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
