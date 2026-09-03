package vanity

import (
	"path"
	"sort"
)

// temporary solution for MAIN-6 (issue#6)
func fill(prefix string, packages []string) []string {
	if prefix == "" {
		return packages
	}

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
