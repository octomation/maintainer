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
	require.Len(t, rows, 2)
	assert.Equal(t, "error", rows[0].Status)
	assert.Equal(t, pin, rows[1].Path)
	assert.True(t, rows[1].Pinned)
	assert.Equal(t, 1, rows[1].Added)
	assert.Equal(t, 2, rows[1].Deleted)
	rows, err = Collect(context.Background(), cnf, st, root, root, []string{"another"}, 2)
	require.NoError(t, err)
	assert.Empty(t, rows)
	// An absent pin must not silently read the stale duplicate.
	cnf.Repos[0].Path = filepath.Join(root, "absent-pin")
	rows, err = Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "error", rows[1].Status)
	assert.True(t, rows[1].Pinned)
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
	rows, err = Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, pin, rows[0].Path)
	assert.True(t, rows[0].Pinned)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = Collect(ctx, cnf, st, root, root, nil, 2)
	assert.ErrorIs(t, err, context.Canceled)
}
