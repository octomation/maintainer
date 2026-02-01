package status

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// Plain, the TUI and the HTTP API build their table from the same exported
// functions: for every sort chain and query, the order of Query, the cells of
// Cells and the summary of Summarize are what Plain prints.
func TestExportedAdapterMatchesPlain(t *testing.T) {
	rows := []Row{
		{Repository: "acme/tool", Path: "/w/acme/tool", Branch: "main", DefaultBranch: "main", Added: 3, Deleted: 1, Changed: 2, Status: "synced", Upstream: "origin/main"},
		{Repository: "acme/lib", Path: "/w/acme/lib", Branch: "feature", DefaultBranch: "main", Ahead: 2, Status: "ahead", PushLocked: true},
		{Repository: "zeta/app", Path: "/w/zeta/app", Status: "error", Error: "git status: boom"},
		{Repository: "acme/old", Path: "/w/acme/old", Branch: "main", Status: "synced", OrphanReason: "duplicate-pin", ActivePath: "/w/acme/tool"},
		{Repository: "ghost", Path: "/w/ghost", Status: StatusUnresolvedPin, placeholder: true},
	}
	for _, tc := range []struct {
		name  string
		keys  []SortKey
		query string
	}{
		{"default", nil, ""},
		{"by status desc", []SortKey{{Column: StatusColumn, Desc: true}}, ""},
		{"chain", []SortKey{{Column: LockColumn, Desc: true}, {Column: RepositoryColumn}}, ""},
		{"query", nil, "acme"},
		{"query sorted", []SortKey{{Column: PathColumn}}, "acme"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			visible := Query(rows, tc.keys, tc.query)
			shown := make([]Row, len(visible))
			for i, index := range visible {
				shown[i] = rows[index]
			}
			var out bytes.Buffer
			require.NoError(t, Plain(&out, shown))
			lines := strings.Split(out.String(), "\n")
			assert.Contains(t, lines[0], Summarize(shown).String())
			assert.Contains(t, lines[0], plural(Checkouts(shown), "checkout"))
			table := newTable(shown)
			body := lines[4:] // the title, a blank line, the header and its rule
			line := 0
			for i, row := range shown {
				assert.Equal(t, table.line(i+1), body[line], "row %d", i)
				for _, cell := range Cells(row) {
					assert.Contains(t, body[line], cell, "cell of row %d", i)
				}
				line++
				if row.OrphanReason != "" {
					line++ // path:
					if row.ActivePath != "" {
						line++ // active:
					}
					if row.RemoteCheckedAt != nil {
						line++ // GitHub check
					}
				}
			}
		})
	}
	assert.Equal(t, Summary{Checkouts: 4, Uncommitted: 1, Ahead: 1, Locked: 1, Orphan: 1, Errors: 1}, Summarize(rows))
	assert.Equal(t, "1 uncommitted · 1 ahead · 0 behind · 1 locked · 1 orphan · 1 error", Summarize(rows).String())
}

func TestParseColumn(t *testing.T) {
	for name, want := range map[string]Column{"repository": RepositoryColumn, "Branch": BranchColumn, "UNCOMMITTED": ChangesColumn, "status": StatusColumn, "lock": LockColumn, "path": PathColumn} {
		got, ok := ParseColumn(name)
		assert.True(t, ok, name)
		assert.Equal(t, want, got, name)
	}
	_, ok := ParseColumn("changes")
	assert.False(t, ok)
	assert.Equal(t, []string{"Repository", "Branch", "Uncommitted", "Status", "Lock", "Path"}, ColumnNames())
}

// Discover gives the same rows as Collect, in the same order, without the
// Git columns; placeholders and orphans are flagged in both.
func TestDiscoverMatchesCollectWithoutGit(t *testing.T) {
	root := t.TempDir()
	duplicate := checkout(t, root)
	pin := checkout(t, t.TempDir())
	write(t, filepath.Join(pin, "text"), "active\n")
	cnf := &config.Fetch{Defaults: config.Defaults{Root: root, Path: config.DefaultPath},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}, {Match: config.RepoMatch{Owner: "acme", Name: "ghost"}, Path: "{{.Visibility}}/ghost"}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: duplicate, Visibility: "private"})
	st.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "missing", Path: filepath.Join(root, "public/acme/missing")})
	d, err := Discover(context.Background(), cnf, st, root, root, nil)
	require.NoError(t, err)
	assert.Equal(t, root, d.Root())
	discovered := d.Rows()
	collected, err := d.Inspect(context.Background(), 2)
	require.NoError(t, err)
	full, err := Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	assert.Equal(t, full, collected)
	require.Len(t, discovered, len(collected))
	for i := range discovered {
		assert.Equal(t, collected[i].Path, discovered[i].Path, i)
		assert.Equal(t, collected[i].Repository, discovered[i].Repository, i)
		assert.Equal(t, collected[i].OrphanReason, discovered[i].OrphanReason, i)
		assert.Equal(t, collected[i].IsPlaceholder(), discovered[i].IsPlaceholder(), i)
		assert.Equal(t, collected[i].DisplayPath(), discovered[i].DisplayPath(), i)
		assert.Empty(t, discovered[i].Branch, "no Git without Inspect")
	}
	var placeholders, orphans, clean int
	for _, row := range discovered {
		switch {
		case row.IsPlaceholder():
			placeholders++
		case row.OrphanReason != "":
			orphans++
		case row.Error == "":
			clean++
		}
	}
	// The visibility comes from the state record, without GitHub.
	for _, row := range discovered {
		if row.Path == pin {
			assert.Equal(t, "private", row.Visibility)
		}
	}
	assert.Equal(t, 1, placeholders, "the unresolved pin")
	assert.Equal(t, 1, orphans, "the duplicate pin")
	assert.Equal(t, 2, clean, "the pinned checkout and the missing one: a missing path is an error only after Inspect")
}
