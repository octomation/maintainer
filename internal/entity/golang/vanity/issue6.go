package vanity

import (
	"path"
	"sort"
)

// temporary solution for issue#6
func fill(prefix string, packages []string) []string {
	index := map[string]struct{}{}
	for _, pkg := range packages {
		index[pkg] = struct{}{}

		for pkg != prefix && pkg != "" {
			pkg = path.Dir(pkg)
			index[pkg] = struct{}{}
		}
	}

	filled := make([]string, 0, len(index))
	for pkg := range index {
		filled = append(filled, pkg)
	}

	sort.Strings(filled)
	return filled
}
