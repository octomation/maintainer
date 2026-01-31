package status

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestCollectPinnedAndMissing(t *testing.T) {
	root := t.TempDir()
	duplicate := checkout(t, root)
	pin := checkout(t, t.TempDir())
	write(t, filepath.Join(pin, "text"), "active\n")
	cnf := &config.Fetch{Defaults: config.Defaults{Root: root, Path: config.DefaultPath},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: duplicate})
	st.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "missing", Path: filepath.Join(root, "public/acme/missing")})
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, "error", rows[0].Status)
	assert.Equal(t, "duplicate-pin", rows[1].OrphanReason)
	assert.Equal(t, pin, rows[2].Path)
	assert.True(t, rows[2].Pinned)
	assert.Equal(t, 1, rows[2].Added)
	assert.Equal(t, 2, rows[2].Deleted)
	rows, err = Collect(context.Background(), cnf, st, root, root, []string{"another"}, 2)
	require.NoError(t, err)
	assert.Empty(t, rows)
	// An absent pin must not silently read the stale duplicate.
	cnf.Repos[0].Path = filepath.Join(root, "absent-pin")
	rows, err = Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, "error", rows[1].Status)
	assert.True(t, rows[1].Pinned)
	assert.Equal(t, "duplicate-pin", rows[2].OrphanReason)
}

func TestCollectIDIgnoreAndLiteralPinWithoutState(t *testing.T) {
	root := t.TempDir()
	path := checkout(t, root)
	cnf := &config.Fetch{Defaults: config.Defaults{Root: root, Path: config.DefaultPath},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Ignore: true}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path})
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	assert.Empty(t, rows)
	pin := checkout(t, t.TempDir())
	cnf.Repos[0] = config.Repo{Match: config.RepoMatch{ID: 1}, Path: pin}
	rows, err = Collect(context.Background(), cnf, state.New(), filepath.Dir(pin), root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, pin, rows[0].Path)
	assert.True(t, rows[0].Pinned)
	assert.Equal(t, "~/repo", rows[0].displayPath)
	raw, err := json.Marshal(rows[0])
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"path":"`+pin+`"`)
	assert.NotContains(t, string(raw), "~/repo")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = Collect(ctx, cnf, st, root, root, nil, 2)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCollectCheckoutsWithoutIdentityAreNotDuplicates(t *testing.T) {
	root := t.TempDir()
	var paths []string
	for _, owner := range []string{"alice", "bob", "carol", "dave", "erin", "frank"} {
		path := filepath.Join(root, "public", owner, "tool")
		require.NoError(t, os.MkdirAll(path, 0o755))
		runGit(t, path, "init", "-b", "main")
		paths = append(paths, path)
	}
	runGit(t, paths[1], "remote", "add", "upstream", "https://github.com/acme/tool.git")
	runGit(t, paths[2], "remote", "add", "origin", paths[0])
	runGit(t, paths[3], "remote", "add", "origin", "file://"+paths[0])
	runGit(t, paths[4], "remote", "add", "origin", "file://localhost/acme/tool.git")
	runGit(t, paths[5], "remote", "add", "origin", "file://localhost/acme/tool.git")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 6)
	for _, row := range rows {
		assert.Equal(t, "tool", row.Repository)
		assert.Equal(t, "unborn", row.Status, row.Path)
		assert.Empty(t, row.Error, row.Path)
	}
	// An ID-only pin without origin is named after its directory too and
	// selects no other checkout of that name.
	pin := filepath.Join(t.TempDir(), "tool")
	require.NoError(t, os.MkdirAll(pin, 0o755))
	runGit(t, pin, "init", "-b", "main")
	cnf.Repos = []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}
	rows, err = Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 7)
	for _, row := range rows {
		assert.Equal(t, "unborn", row.Status, row.Path)
		assert.Empty(t, row.Error, row.Path)
		assert.Empty(t, row.OrphanReason, row.Path)
		assert.Equal(t, row.Path == pin, row.Pinned, row.Path)
	}
}

func TestCollectOrphanIsNotDuplicate(t *testing.T) {
	old, root := t.TempDir(), t.TempDir()
	remembered := checkout(t, filepath.Join(old, "public/acme"))
	clone := checkout(t, filepath.Join(root, "public/acme"))
	write(t, filepath.Join(clone, "text"), "one\n")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}}
	st := state.New()
	st.Upsert(state.Record{ID: 7, OwnerLogin: "acme", Name: "tool", Path: remembered})
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	byPath := map[string]Row{rows[0].Path: rows[0], rows[1].Path: rows[1]}
	assert.Equal(t, "out-of-scope", byPath[remembered].OrphanReason)
	assert.Equal(t, "synced", byPath[remembered].Status)
	assert.Empty(t, byPath[remembered].Error)
	assert.Empty(t, byPath[clone].OrphanReason)
	assert.Equal(t, "synced", byPath[clone].Status)
	assert.Equal(t, "main", byPath[clone].Branch)
	assert.Equal(t, 1, byPath[clone].Deleted)
	assert.Empty(t, byPath[clone].Error)
	assert.Equal(t, int64(7), byPath[clone].ID)
}
