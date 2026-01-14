package status

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortCycleAndPriority(t *testing.T) {
	keys := cycleSort(nil, BranchColumn, false)
	assert.Equal(t, []sortKey{{Column: BranchColumn}}, keys)
	keys = cycleSort(keys, StatusColumn, true)
	keys = cycleSort(keys, ChangesColumn, true)
	assert.Equal(t, []sortKey{{Column: BranchColumn}, {Column: StatusColumn}, {Column: ChangesColumn}}, keys)
	keys = cycleSort(keys, StatusColumn, true)
	assert.Equal(t, []sortKey{{Column: BranchColumn}, {Column: StatusColumn, Desc: true}, {Column: ChangesColumn}}, keys)
	keys = cycleSort(keys, StatusColumn, true)
	assert.Equal(t, []sortKey{{Column: BranchColumn}, {Column: ChangesColumn}}, keys)
	keys = cycleSort(keys, StatusColumn, true)
	assert.Equal(t, StatusColumn, keys[2].Column, "a removed key re-enters at the end")
	keys = cycleSort(keys, ChangesColumn, false)
	assert.Equal(t, []sortKey{{Column: ChangesColumn, Desc: true}}, keys, "plain activation replaces the whole chain")
	assert.Empty(t, cycleSort(keys, ChangesColumn, false))
}

func TestSortUsesWholeSnapshotAndNumericValues(t *testing.T) {
	rows := []Row{
		{Repository: "z", Path: "/z", Branch: "main", Added: 12, Behind: 100},
		{Repository: "a", Path: "/a", Branch: "dev", Added: 3, Ahead: 2},
		{Repository: "b", Path: "/b", Branch: "main", Added: 2, Behind: 10},
		{Repository: "error", Path: "/error", Error: "cannot inspect"},
	}
	before := slices.Clone(rows)
	assert.Equal(t, []int{2, 1, 0, 3}, queryRows(rows, []sortKey{{Column: ChangesColumn}}, ""))
	assert.Equal(t, []int{0, 1, 2, 3}, queryRows(rows, []sortKey{{Column: ChangesColumn, Desc: true}}, ""))
	assert.Equal(t, []int{1, 2, 0, 3}, queryRows(rows, []sortKey{{Column: StatusColumn}}, ""))
	assert.Equal(t, []int{0, 2, 1, 3}, queryRows(rows, []sortKey{{Column: StatusColumn, Desc: true}}, ""))
	// The first key wins; later keys sort only rows tied under all earlier keys.
	assert.Equal(t, []int{1, 0, 2}, queryRows(rows[:3], []sortKey{{Column: BranchColumn}, {Column: ChangesColumn, Desc: true}}, ""))
	assert.Equal(t, []int{0, 1, 2}, queryRows(rows[:3], []sortKey{{Column: ChangesColumn, Desc: true}, {Column: BranchColumn}}, ""))
	assert.Equal(t, before, rows)
}

func TestFuzzyFilterAndSortComposition(t *testing.T) {
	rows := []Row{
		{Repository: "acme/maintainer", Path: "/work/maintainer", Branch: "main", Ahead: 12},
		{Repository: "kamilsk/dotfiles", Path: "/home/me/.dotfiles", Branch: "main", Ahead: 2},
		{Repository: "acme/docs", Path: "/work/docs", Branch: "feature", Ahead: 3},
		{Repository: "команда/заметки", Path: "/work/заметки", Branch: "main"},
	}
	assert.Equal(t, []int{1}, queryRows(rows, nil, "DTF MAIN"), "case-insensitive, non-contiguous and across fields")
	assert.Equal(t, []int{1}, queryRows(rows, nil, ".dtf"), "search includes external paths")
	assert.Equal(t, []int{3}, queryRows(rows, nil, "змт"), "Unicode subsequences")
	assert.Empty(t, queryRows(rows, nil, "no-such-repository"))
	assert.Equal(t, []int{0, 1, 3}, queryRows(rows, []sortKey{{Column: StatusColumn, Desc: true}}, "main"))
	assert.Len(t, queryRows(rows, nil, "  "), 4)
	// With no explicit keys the strongest match wins, not lexical repository order.
	ranked := queryRows([]Row{{Repository: "a/dxxxxoxxxxt", Path: "/a"}, {Repository: "z/dot", Path: "/z"}}, nil, "dot")
	require.Len(t, ranked, 2)
	assert.Equal(t, 1, ranked[0])
}

func TestSortTiesAndNonLineChanges(t *testing.T) {
	rows := []Row{{Repository: "same", Path: "/b", Untracked: 12}, {Repository: "same", Path: "/a", Untracked: 2}}
	assert.Equal(t, []int{1, 0}, queryRows(rows, []sortKey{{Column: ChangesColumn}}, ""))
	assert.Equal(t, []int{0, 1}, queryRows(rows, []sortKey{{Column: ChangesColumn, Desc: true}}, ""))
	assert.Equal(t, []int{1, 0}, queryRows(rows, []sortKey{{Column: RepositoryColumn}}, ""))
	assert.Equal(t, []int{1, 0}, queryRows(rows, nil, ""))
}

func TestSortAndFilterDisplayedPath(t *testing.T) {
	rows := []Row{
		{Repository: "a", Path: "/home/me/.dotfiles", displayPath: "~/.dotfiles"},
		{Repository: "z", Path: "/home/me/work/public/acme/tool", displayPath: "public/acme/tool"},
	}
	assert.Equal(t, []int{1, 0}, queryRows(rows, []sortKey{{Column: PathColumn}}, ""))
	assert.Equal(t, []int{0, 1}, queryRows(rows, []sortKey{{Column: PathColumn, Desc: true}}, ""))
	assert.Equal(t, []int{0}, queryRows(rows, nil, "~/.dotfiles"))
	assert.Equal(t, []int{0}, queryRows(rows, nil, "/home/me/.dotfiles"))
}

func TestSortAndFilterPushLock(t *testing.T) {
	rows := []Row{
		{Repository: "z", PushLocked: true},
		{Repository: "b"},
		{Repository: "a", PushLocked: true, OrphanReason: "duplicate-pin"},
		{Repository: "c"},
	}
	assert.Equal(t, []int{1, 3, 2, 0}, queryRows(rows, []sortKey{{Column: LockColumn}}, ""))
	assert.Equal(t, []int{2, 0, 1, 3}, queryRows(rows, []sortKey{{Column: LockColumn, Desc: true}}, ""))
	assert.Equal(t, []int{0, 2, 3, 1}, queryRows(rows, []sortKey{{Column: LockColumn, Desc: true}, {Column: RepositoryColumn, Desc: true}}, ""))
	assert.Equal(t, []int{2, 0}, queryRows(rows, nil, "🔒"))
}
