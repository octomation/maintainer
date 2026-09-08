package status

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestOrphanKeepsDirtyStatusAndIsSearchable(t *testing.T) {
	root := t.TempDir()
	old, pin := checkout(t, filepath.Join(root, "public/acme")), checkout(t, t.TempDir())
	write(t, filepath.Join(old, "text"), "old copy work\n")
	write(t, filepath.Join(old, "untracked"), "do not lose\n")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: pin, PinnedPath: pin, PreviousPaths: []string{old}})
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	indexes := queryRows(rows, nil, "orphan duplicate-pin")
	require.Len(t, indexes, 1)
	row := rows[indexes[0]]
	assert.Equal(t, old, row.Path)
	assert.Equal(t, pin, row.ActivePath)
	assert.False(t, row.Pinned)
	assert.Equal(t, "synced", row.Status)
	assert.Empty(t, row.Error)
	assert.Equal(t, 1, row.Added)
	assert.Equal(t, 2, row.Deleted)
	assert.Equal(t, 1, row.Untracked)
	var plain bytes.Buffer
	require.NoError(t, Plain(&plain, rows))
	assert.Contains(t, plain.String(), "orphan · synced")
	assert.NotContains(t, plain.String(), "duplicate-pin")
	assert.Contains(t, plain.String(), "path: "+old)
	assert.Contains(t, plain.String(), "active: "+pin)
	assert.Contains(t, plain.String(), "+1/-2")
	m := newScreen([]Row{row})
	m.Update(tea.WindowSizeMsg{Width: 240, Height: 16})
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "orphan · synced")
	assert.NotContains(t, view, "duplicate-pin")
	assert.Contains(t, view, "active: "+pin)
}

func TestCachedRemoteOrphanRetainsTrackingStatus(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	row := Row{Repository: "acme/tool", Path: "/work/tool", Status: "behind", Behind: 12, OrphanReason: "remote-gone", RemoteCheckedAt: &now}
	assert.Equal(t, "orphan · behind 12", cells(row)[3])
	var out bytes.Buffer
	require.NoError(t, Plain(&out, []Row{row}))
	assert.Contains(t, out.String(), "GitHub check (cached): 2026-09-08T12:00:00Z")
}
