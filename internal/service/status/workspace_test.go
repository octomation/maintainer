package status

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestWorkspaceBroadPinDuplicatesAndExplicitSelection(t *testing.T) {
	root := t.TempDir()
	managed := checkout(t, filepath.Join(root, "public/acme"))
	pin := checkout(t, filepath.Join(root, "prototyping"))
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"prototyping"}}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	for _, row := range rows {
		assert.Contains(t, row.Error, "multiple checkouts")
	}
	cnf.Repos = []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}
	rows, err = Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, managed, rows[1].Path)
	assert.Equal(t, "duplicate-pin", rows[1].OrphanReason)
	assert.Equal(t, pin, rows[0].Path)
	assert.NotEqual(t, managed, rows[0].Path)
	assert.True(t, rows[0].Pinned)
	assert.Empty(t, rows[0].Error)
}

func TestWorkspaceRemovedPinDoesNotReenterThroughStateOrTemplate(t *testing.T) {
	root := t.TempDir()
	path := checkout(t, filepath.Join(root, "public/acme"))
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path, PinnedPath: path, PinSource: "workspace"})
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}}
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "out-of-scope", rows[0].OrphanReason)
	assert.False(t, rows[0].Pinned)
	cnf.Workspace.Pins = []string{"public/acme"}
	rows, err = Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, rows[0].Pinned)
	assert.Equal(t, "workspace", st.Repos[0].PinSource, "status must not mutate state")
}

func TestWorkspaceMissingPinIsVisible(t *testing.T) {
	root := t.TempDir()
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"missing"}}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, rows[0].Pinned)
	assert.Equal(t, "error", rows[0].Status)
}
