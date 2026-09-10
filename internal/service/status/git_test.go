package status

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runGit(t *testing.T, path string, args ...string) string {
	t.Helper()
	argv := append([]string{"-C", path, "-c", "user.name=Test", "-c", "user.email=test@example.org", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"}, args...)
	cmd := exec.Command("git", argv...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", out)
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
}

func checkout(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "repo")
	require.NoError(t, os.MkdirAll(path, 0o755))
	runGit(t, path, "init", "-b", "main")
	write(t, filepath.Join(path, "text"), "one\ntwo\n")
	write(t, filepath.Join(path, "binary"), "\x00abc")
	runGit(t, path, "add", ".")
	runGit(t, path, "commit", "-m", "base")
	runGit(t, path, "remote", "add", "origin", "https://github.com/acme/tool.git")
	runGit(t, path, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, path, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, path, "branch", "--set-upstream-to=origin/main")
	return path
}

func TestInspectChangesAndDivergence(t *testing.T) {
	path := checkout(t, t.TempDir())
	runGit(t, path, "checkout", "-b", "remote-work")
	write(t, filepath.Join(path, "remote"), "remote\n")
	runGit(t, path, "add", ".")
	runGit(t, path, "commit", "-m", "remote")
	runGit(t, path, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, path, "checkout", "main")
	write(t, filepath.Join(path, "local"), "local\n")
	runGit(t, path, "add", ".")
	runGit(t, path, "commit", "-m", "local")
	write(t, filepath.Join(path, "text"), "one\nstaged\n")
	runGit(t, path, "add", "text")
	write(t, filepath.Join(path, "text"), "one\nchanged\nthree\nfour\n")
	write(t, filepath.Join(path, "binary"), "\x00def")
	write(t, filepath.Join(path, "untracked\nwith\ttabs"), "not counted as lines\n")
	before, err := os.ReadFile(filepath.Join(path, ".git", "index"))
	require.NoError(t, err)
	row := Inspect(context.Background(), Row{Path: path})
	assert.Empty(t, row.Error)
	assert.Equal(t, "acme/tool", row.Repository)
	assert.Equal(t, "main", row.Branch)
	assert.Equal(t, "main", row.DefaultBranch)
	assert.Equal(t, "origin/main", row.Upstream)
	assert.Equal(t, "diverged", row.Status)
	assert.Equal(t, 1, row.Ahead)
	assert.Equal(t, 1, row.Behind)
	assert.Equal(t, 3, row.Added)
	assert.Equal(t, 1, row.Deleted)
	assert.Equal(t, 1, row.Untracked)
	assert.Equal(t, 1, row.Binary)
	assert.Equal(t, 2, row.Changed)
	after, err := os.ReadFile(filepath.Join(path, ".git", "index"))
	require.NoError(t, err)
	assert.Equal(t, before, after, "status must not refresh the index")
}

func TestInspectBranchStates(t *testing.T) {
	path := checkout(t, t.TempDir())
	row := Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "synced", row.Status)
	runGit(t, path, "update-ref", "-d", "refs/remotes/origin/main")
	row = Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "upstream gone", row.Status)
	assert.Empty(t, row.Error)
	runGit(t, path, "branch", "--unset-upstream")
	row = Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "no upstream", row.Status)
	runGit(t, path, "checkout", "--detach")
	row = Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "detached", row.Status)
	assert.NotEmpty(t, row.Commit)
	row = Inspect(context.Background(), Row{Path: filepath.Join(path, "missing")})
	assert.Equal(t, "error", row.Status)
	assert.NotEmpty(t, row.Error)
}

func TestInspectPushLock(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, tc := range []struct {
		name, remote string
		urls         []string
		locked       bool
	}{
		{name: "no pushurl", remote: "origin"},
		{name: "origin locked", remote: "origin", urls: []string{"no_push"}, locked: true},
		{name: "other remote locked", remote: "backup", urls: []string{"no_push"}, locked: true},
		{name: "multiple pushurls", remote: "origin", urls: []string{"no_push", "https://github.com/acme/tool.git"}, locked: true},
		{name: "custom pushurl", remote: "origin", urls: []string{"https://github.com/acme/tool.git"}},
		{name: "exact marker only", remote: "origin", urls: []string{"no_push_extra", "/tmp/no_push", "NO_PUSH", "no_push\n"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := checkout(t, t.TempDir())
			if tc.remote != "origin" {
				runGit(t, path, "remote", "add", tc.remote, "https://github.com/acme/backup.git")
			}
			for _, url := range tc.urls {
				runGit(t, path, "config", "--add", "remote."+tc.remote+".pushurl", url)
			}
			before, err := os.ReadFile(filepath.Join(path, ".git", "config"))
			require.NoError(t, err)
			row := Inspect(context.Background(), Row{Path: path})
			assert.Empty(t, row.Error)
			assert.Equal(t, tc.locked, row.PushLocked)
			assert.Equal(t, "synced", row.Status)
			assert.Equal(t, "acme/tool", row.Repository)
			after, err := os.ReadFile(filepath.Join(path, ".git", "config"))
			require.NoError(t, err)
			assert.Equal(t, before, after, "inspection must not change push configuration")
			if tc.locked {
				runGit(t, path, "config", "--unset-all", "remote."+tc.remote+".pushurl")
				row = Inspect(context.Background(), Row{Path: path})
				assert.Empty(t, row.Error)
				assert.False(t, row.PushLocked)
			}
		})
	}
}

func TestInspectPushLockIncludesAndWorktree(t *testing.T) {
	path := checkout(t, t.TempDir())
	include := filepath.Join(t.TempDir(), "lock.config")
	write(t, include, "[remote \"origin\"]\n\tpushurl = no_push\n")
	runGit(t, path, "config", "include.path", include)
	row := Inspect(context.Background(), Row{Path: path})
	assert.Empty(t, row.Error)
	assert.True(t, row.PushLocked)
	worktree := filepath.Join(t.TempDir(), "linked")
	runGit(t, path, "worktree", "add", "-b", "feature", worktree)
	row = Inspect(context.Background(), Row{Path: worktree})
	assert.Empty(t, row.Error)
	assert.True(t, row.PushLocked)
	runGit(t, path, "config", "--unset", "include.path")
	runGit(t, path, "config", "extensions.worktreeConfig", "true")
	runGit(t, worktree, "config", "--worktree", "remote.origin.pushurl", "no_push")
	row = Inspect(context.Background(), Row{Path: worktree})
	assert.Empty(t, row.Error)
	assert.True(t, row.PushLocked)
	assert.False(t, Inspect(context.Background(), Row{Path: path}).PushLocked)
	// A later Git failure must retain a lock that was successfully read.
	write(t, filepath.Join(path, ".git", "index"), "invalid index")
	runGit(t, path, "config", "remote.origin.pushurl", "no_push")
	row = Inspect(context.Background(), Row{Path: path})
	assert.NotEmpty(t, row.Error)
	assert.True(t, row.PushLocked)
}

func TestInspectPushLockConfigError(t *testing.T) {
	path := checkout(t, t.TempDir())
	write(t, filepath.Join(path, ".git", "config"), "[invalid config")
	row := Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "error", row.Status)
	assert.Contains(t, row.Error, "git config:")
}

func TestInspectUnbornAndWorktree(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	write(t, filepath.Join(root, "staged"), "one\ntwo\n")
	runGit(t, root, "add", ".")
	write(t, filepath.Join(root, "untracked"), "new\n")
	row := Inspect(context.Background(), Row{Path: root})
	assert.Empty(t, row.Error)
	assert.Equal(t, "unborn", row.Status)
	assert.Equal(t, 2, row.Added)
	assert.Equal(t, 1, row.Untracked)
	assert.Empty(t, row.Commit)
	path := checkout(t, t.TempDir())
	worktree := filepath.Join(t.TempDir(), "linked")
	runGit(t, path, "worktree", "add", "-b", "feature", worktree)
	row = Inspect(context.Background(), Row{Path: worktree})
	assert.Empty(t, row.Error)
	assert.Equal(t, "feature", row.Branch)
	assert.Equal(t, "acme/tool", row.Repository)
}

func TestInspectRenameAndConflict(t *testing.T) {
	path := checkout(t, t.TempDir())
	runGit(t, path, "mv", "text", "renamed\nfile")
	row := Inspect(context.Background(), Row{Path: path})
	assert.Empty(t, row.Error)
	assert.Equal(t, 1, row.Changed)
	assert.Equal(t, 2, row.Added)
	assert.Equal(t, 2, row.Deleted)
	runGit(t, path, "commit", "-m", "rename")
	runGit(t, path, "checkout", "-b", "conflict")
	write(t, filepath.Join(path, "renamed\nfile"), "other\n")
	runGit(t, path, "commit", "-am", "other")
	runGit(t, path, "checkout", "main")
	write(t, filepath.Join(path, "renamed\nfile"), "local\n")
	runGit(t, path, "commit", "-am", "local")
	cmd := exec.Command("git", "-C", path, "-c", "core.hooksPath=/dev/null", "merge", "conflict")
	require.Error(t, cmd.Run())
	row = Inspect(context.Background(), Row{Path: path})
	assert.Empty(t, row.Error)
	assert.Equal(t, 1, row.Conflicts)
	assert.Equal(t, 1, row.Changed)
}
