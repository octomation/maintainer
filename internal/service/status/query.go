package status

import (
	"cmp"
	"slices"
	"strings"

	"github.com/sahilm/fuzzy"
)

// Column is a visible sortable column; values also define header order.
type Column int

const (
	RepositoryColumn Column = iota
	BranchColumn
	ChangesColumn
	StatusColumn
)

var columnNames = []string{"Repository", "Branch", "Uncommitted", "Status"}

type sortKey struct {
	Column Column
	Desc   bool
}

// cycleSort preserves the precedence of existing keys in additive mode.
// A plain activation replaces the whole chain, while still cycling the
// activated column's current direction: absent -> asc -> desc -> absent.
func cycleSort(keys []sortKey, column Column, additive bool) []sortKey {
	index := slices.IndexFunc(keys, func(key sortKey) bool { return key.Column == column })
	next := sortKey{Column: column}
	remove := false
	if index >= 0 {
		next.Desc = !keys[index].Desc
		remove = keys[index].Desc
	}
	if !additive {
		if remove {
			return nil
		}
		return []sortKey{next}
	}
	result := slices.Clone(keys)
	switch {
	case index < 0:
		result = append(result, next)
	case remove:
		result = slices.Delete(result, index, index+1)
	default:
		result[index] = next
	}
	return result
}

// queryRows returns indexes into an immutable snapshot. Fuzzy relevance orders
// an unsorted search; explicit sort keys take precedence over relevance.
func queryRows(rows []Row, keys []sortKey, query string) []int {
	type candidate struct{ index, score int }
	candidates := make([]candidate, 0, len(rows))
	terms := strings.Fields(query)
	for i, row := range rows {
		score, matches := 0, true
		fields := append(cells(row), safeText(row.Path), safeText(row.Upstream), safeText(row.Commit))
		for _, term := range terms {
			found := fuzzy.Find(term, fields)
			if len(found) == 0 {
				matches = false
				break
			}
			score += found[0].Score
		}
		if matches {
			candidates = append(candidates, candidate{i, score})
		}
	}
	slices.SortStableFunc(candidates, func(a, b candidate) int {
		left, right := rows[a.index], rows[b.index]
		for _, key := range keys {
			// Unknown numeric values remain last in either direction.
			if key.Column == ChangesColumn || key.Column == StatusColumn {
				if (left.Error != "") != (right.Error != "") {
					if left.Error != "" {
						return 1
					}
					return -1
				}
			}
			order := compareColumn(left, right, key.Column)
			if key.Desc {
				order = -order
			}
			if order != 0 {
				return order
			}
		}
		if len(keys) == 0 && len(terms) > 0 && a.score != b.score {
			return cmp.Compare(b.score, a.score)
		}
		if order := strings.Compare(strings.ToLower(left.Repository), strings.ToLower(right.Repository)); order != 0 {
			return order
		}
		return strings.Compare(left.Path, right.Path)
	})
	result := make([]int, len(candidates))
	for i, candidate := range candidates {
		result[i] = candidate.index
	}
	return result
}

func compareColumn(a, b Row, column Column) int {
	switch column {
	case RepositoryColumn:
		return strings.Compare(strings.ToLower(a.Repository), strings.ToLower(b.Repository))
	case BranchColumn:
		return strings.Compare(strings.ToLower(a.Branch), strings.ToLower(b.Branch))
	case ChangesColumn:
		return compareNumbers([]int{a.Added + a.Deleted, a.Changed, a.Untracked, a.Binary, a.Conflicts},
			[]int{b.Added + b.Deleted, b.Changed, b.Untracked, b.Binary, b.Conflicts})
	case StatusColumn:
		if order := compareNumbers([]int{a.Ahead + a.Behind, a.Ahead, a.Behind}, []int{b.Ahead + b.Behind, b.Ahead, b.Behind}); order != 0 {
			return order
		}
		return strings.Compare(a.Status, b.Status)
	default:
		return 0
	}
}

func compareNumbers(a, b []int) int {
	for i := range a {
		if order := cmp.Compare(a[i], b[i]); order != 0 {
			return order
		}
	}
	return 0
}
