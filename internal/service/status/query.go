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
	LockColumn
	PathColumn
)

var columnNames = []string{"Repository", "Branch", "Uncommitted", "Status", "Lock", "Path"}

// ColumnNames returns the header names in column order.
func ColumnNames() []string { return slices.Clone(columnNames) }

// ParseColumn finds a column by its header name, case-insensitively.
func ParseColumn(name string) (Column, bool) {
	for i, known := range columnNames {
		if strings.EqualFold(known, name) {
			return Column(i), true
		}
	}
	return 0, false
}

// SortKey is one key of a sort chain: a column and its direction.
type SortKey struct {
	Column Column
	Desc   bool
}

// cycleSort preserves the precedence of existing keys in additive mode.
// A plain activation replaces the whole chain, while still cycling the
// activated column's current direction: absent -> asc -> desc -> absent.
func cycleSort(keys []SortKey, column Column, additive bool) []SortKey {
	index := slices.IndexFunc(keys, func(key SortKey) bool { return key.Column == column })
	next := SortKey{Column: column}
	remove := false
	if index >= 0 {
		next.Desc = !keys[index].Desc
		remove = keys[index].Desc
	}
	if !additive {
		if remove {
			return nil
		}
		return []SortKey{next}
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

// Query returns indexes into an immutable snapshot: the rows every term of
// the fuzzy query matches, in the order of the sort chain. Fuzzy relevance
// orders an unsorted search; explicit sort keys take precedence over
// relevance; unknown numbers stay last in either direction; the repository
// and the path break ties. The TUI, Plain and the HTTP API share it.
func Query(rows []Row, keys []SortKey, query string) []int {
	type candidate struct{ index, score int }
	candidates := make([]candidate, 0, len(rows))
	terms := strings.Fields(query)
	for i, row := range rows {
		score, matches := 0, true
		fields := append(Cells(row), safeText(row.Path), safeText(row.Upstream), safeText(row.Commit), safeText(row.ActivePath), safeText(row.OrphanReason))
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
	case LockColumn:
		if a.PushLocked == b.PushLocked {
			return 0
		}
		if a.PushLocked {
			return 1
		}
		return -1
	case PathColumn:
		return strings.Compare(a.pathCell(), b.pathCell())
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
