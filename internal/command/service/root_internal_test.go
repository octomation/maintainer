package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/status"
)

func TestOriginName(t *testing.T) {
	config := "[core]\n\tbare = false\n[remote \"origin\"]\n\turl = git@github.com:octomation/maintainer.git\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n"

	t.Run("a checkout", func(t *testing.T) {
		repo := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, ".git", "config"), []byte(config), 0o644))
		assert.Equal(t, "octomation/maintainer", originName(repo))
	})

	t.Run("a linked worktree", func(t *testing.T) {
		main := t.TempDir()
		gitDir := filepath.Join(main, ".git")
		wtDir := filepath.Join(gitDir, "worktrees", "feature")
		require.NoError(t, os.MkdirAll(wtDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte(config), 0o644))
		// The worktree's own directory has no config: commondir points up.
		require.NoError(t, os.WriteFile(filepath.Join(wtDir, "commondir"), []byte("../..\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(wtDir, "HEAD"), []byte("ref: refs/heads/feature\n"), 0o644))
		worktree := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: "+wtDir+"\n"), 0o644))
		assert.Equal(t, "octomation/maintainer", originName(worktree))
	})

	t.Run("no origin", func(t *testing.T) {
		repo := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, ".git", "config"), []byte("[core]\n\tbare = false\n"), 0o644))
		assert.Equal(t, "", originName(repo))
		assert.Equal(t, "", originName(t.TempDir()))
	})
}

func TestMirrorRootsCarryVisibility(t *testing.T) {
	mirror := func() string {
		path := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(path, ".github", "issues"), 0o755))
		return path
	}
	private, unknown := mirror(), mirror()
	roots := mirrorRoots([]status.Row{
		{ID: 1, Repository: "acme/secret", Path: private, Visibility: "private"},
		{ID: 2, Repository: "acme/open", Path: unknown},
	})
	require.Len(t, roots, 2)
	assert.Equal(t, "private", roots[0].Visibility)
	assert.Equal(t, "", roots[1].Visibility)
}
