package git_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
)

// originRepo creates a local non-bare repo with one commit and returns its path.
func originRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "origin")
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644))
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = wt.Add("README.md")
	require.NoError(t, err)
	_, err = wt.Commit("init", &gogit.CommitOptions{
		Author: &object.Signature{Name: "t", Email: "t@e", When: time.Unix(1, 0)},
	})
	require.NoError(t, err)
	return dir
}

func TestSync_CloneInspectMoveUpdateFetch(t *testing.T) {
	origin := originRepo(t)
	sync := gitsvc.NewSync()
	ctx := context.Background()

	// local clone needs no credentials; https + empty token resolves to nil auth.
	work := filepath.Join(t.TempDir(), "clone")
	require.NoError(t, sync.Clone(ctx, gitsvc.CloneOptions{
		URL: origin, Path: work, Auth: gitsvc.Auth{Transport: "https"},
	}))

	info, err := sync.Inspect(work)
	require.NoError(t, err)
	require.Len(t, info.Origins, 1)
	assert.Equal(t, origin, info.Origins[0])
	assert.NotEmpty(t, info.HeadShort)

	// fetch from the local origin is a no-op but must not error.
	require.NoError(t, sync.Fetch(ctx, work, gitsvc.Auth{Transport: "https"}))

	// move renames the directory the fetcher owns.
	moved := filepath.Join(t.TempDir(), "moved")
	require.NoError(t, sync.Move(work, moved))
	_, statErr := os.Stat(work)
	assert.True(t, os.IsNotExist(statErr))

	// update_remote rewrites origin's URL in place.
	const canonical = "https://github.com/acme/configs.git"
	require.NoError(t, sync.UpdateRemote(moved, canonical))
	info, err = sync.Inspect(moved)
	require.NoError(t, err)
	require.Len(t, info.Origins, 1)
	assert.Equal(t, canonical, info.Origins[0])
}

func TestSync_CloneEmptyRepo(t *testing.T) {
	// An origin with no commits (created but never pushed).
	origin := filepath.Join(t.TempDir(), "empty")
	_, err := gogit.PlainInit(origin, true) // bare, zero refs
	require.NoError(t, err)

	sync := gitsvc.NewSync()
	work := filepath.Join(t.TempDir(), "clone")
	// Must succeed, mirroring `git clone` of an empty repo.
	require.NoError(t, sync.Clone(context.Background(), gitsvc.CloneOptions{
		URL: origin, Path: work, Auth: gitsvc.Auth{Transport: "https"},
	}))

	// The clone is tracked: a .git exists and origin points at the URL.
	info, err := sync.Inspect(work)
	require.NoError(t, err)
	require.Len(t, info.Origins, 1)
	assert.Equal(t, origin, info.Origins[0])
	assert.Empty(t, info.HeadShort) // no commits yet
}

func TestSync_MoveMissingSourceErrors(t *testing.T) {
	sync := gitsvc.NewSync()
	err := sync.Move(filepath.Join(t.TempDir(), "nope"), filepath.Join(t.TempDir(), "dst"))
	assert.Error(t, err)
}
