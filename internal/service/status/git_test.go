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

// Inspect inherits the process environment, so fixtures and assertions
// must not depend on the developer's or the CI runner's Git configuration.
func TestMain(m *testing.M) {
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_COUNT"} {
		_ = os.Unsetenv(key)
	}
	_ = os.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	_ = os.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	os.Exit(m.Run())
}

// gitOutput runs Git with a fixed identity and no hooks; callers decide whether failure is expected.
func gitOutput(t *testing.T, path string, args ...string) (string, error) {
	t.Helper()
	argv := append([]string{"-C", path, "-c", "user.name=Test", "-c", "user.email=test@example.org", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"}, args...)
	out, err := exec.CommandContext(t.Context(), "git", argv...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runGit(t *testing.T, path string, args ...string) string {
	t.Helper()
	out, err := gitOutput(t, path, args...)
	require.NoError(t, err, "%s", out)
	return out
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
	assert.Equal(t, "error", row.Status)
	assert.True(t, strings.HasPrefix(row.Error, "git status: fatal: "), row.Error)
	assert.Contains(t, row.Error, "index")
	assert.True(t, row.PushLocked)
	assert.Equal(t, "acme/tool", row.Repository, "identity is read when git status fails")
}

func TestInspectPushLockConfigError(t *testing.T) {
	path := checkout(t, t.TempDir())
	write(t, filepath.Join(path, ".git", "config"), "[invalid config")
	row := Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "error", row.Status)
	assert.True(t, strings.HasPrefix(row.Error, "git status: fatal: bad config"), row.Error)
	assert.NotContains(t, row.Error, "git config", "the lock probe keeps the first failure")
}

func TestInspectErrorQuotesGit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken")
	require.NoError(t, os.MkdirAll(path, 0o755))
	write(t, filepath.Join(path, ".git"), "gitdir: /nonexistent/x\n")
	row := Inspect(context.Background(), Row{Path: path})
	assert.Equal(t, "error", row.Status)
	// The wording depends on the Git version; the step, the message and the
	// exit status do not.
	assert.True(t, strings.HasPrefix(row.Error, "git status: fatal: "), row.Error)
	assert.True(t, strings.HasSuffix(row.Error, " (exit status 128)"), row.Error)
	assert.Equal(t, "fatal: one two", oneLine([]byte("\n  fatal: one\n\n\ttwo  \n")))
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
	assert.Zero(t, row.Added)
	assert.Zero(t, row.Deleted)
	runGit(t, path, "commit", "-m", "rename")
	runGit(t, path, "checkout", "-b", "conflict")
	write(t, filepath.Join(path, "renamed\nfile"), "other\n")
	runGit(t, path, "commit", "-am", "other")
	runGit(t, path, "checkout", "main")
	write(t, filepath.Join(path, "renamed\nfile"), "local\n")
	runGit(t, path, "commit", "-am", "local")
	out, err := gitOutput(t, path, "merge", "conflict")
	require.Error(t, err, "%s", out)
	require.Contains(t, out, "CONFLICT", "merge must stop on a content conflict")
	runGit(t, path, "rev-parse", "-q", "--verify", "MERGE_HEAD")
	row = Inspect(context.Background(), Row{Path: path})
	assert.Empty(t, row.Error)
	assert.Equal(t, 1, row.Conflicts)
	assert.Equal(t, 1, row.Changed)
}

func TestInspectRenames(t *testing.T) {
	type counts struct{ changed, added, deleted, binary, untracked int }
	move := func(from, to string) func(*testing.T, string) {
		return func(t *testing.T, path string) { runGit(t, path, "mv", "--", from, to) }
	}
	for _, tc := range []struct {
		name  string
		setup []func(*testing.T, string)
		want  counts
	}{
		{name: "pure rename", setup: []func(*testing.T, string){move("text", "renamed")},
			want: counts{changed: 1}},
		{name: "binary rename", setup: []func(*testing.T, string){move("binary", "logo")},
			want: counts{changed: 1, binary: 1}},
		{name: "text and binary renames", setup: []func(*testing.T, string){move("text", "main"), move("binary", "logo")},
			want: counts{changed: 2, binary: 1}},
		{name: "staged rename with edit", setup: []func(*testing.T, string){move("text", "renamed"), func(t *testing.T, path string) {
			write(t, filepath.Join(path, "renamed"), "one\ntwo\nthree\n")
			runGit(t, path, "add", "renamed")
		}}, want: counts{changed: 1, added: 1}},
		{name: "staged rename edited in worktree", setup: []func(*testing.T, string){move("text", "renamed"), func(t *testing.T, path string) {
			write(t, filepath.Join(path, "renamed"), "one\ntwo\nthree\n")
		}}, want: counts{changed: 1, added: 1}},
		{name: "new paths with newline and tab", setup: []func(*testing.T, string){move("text", "renamed\nwith\ttab"), move("binary", "bin\tary\nlogo")},
			want: counts{changed: 2, binary: 1}},
		{name: "new paths that look like records", setup: []func(*testing.T, string){move("text", "u old\nfile"), move("binary", "1 x")},
			want: counts{changed: 2, binary: 1}},
		{name: "original paths that look like records", setup: []func(*testing.T, string){func(t *testing.T, path string) {
			write(t, filepath.Join(path, "u old\nfile"), "alpha\nbeta\n")
			write(t, filepath.Join(path, "1\t2\tx"), "gamma\ndelta\n")
			write(t, filepath.Join(path, "? y"), "epsilon\nzeta\n")
			write(t, filepath.Join(path, "-\t-\tz"), "\x00zz")
			runGit(t, path, "add", ".")
			runGit(t, path, "commit", "-m", "hostile")
		}, move("u old\nfile", "a"), move("1\t2\tx", "b"), move("? y", "c"), move("-\t-\tz", "d")},
			want: counts{changed: 4, binary: 1}},
		// Git cannot pair a deletion with a file that is not in the index.
		{name: "moved without add", setup: []func(*testing.T, string){func(t *testing.T, path string) {
			require.NoError(t, os.Rename(filepath.Join(path, "text"), filepath.Join(path, "moved")))
		}}, want: counts{changed: 1, deleted: 2, untracked: 1}},
		{name: "moved with intent to add", setup: []func(*testing.T, string){func(t *testing.T, path string) {
			require.NoError(t, os.Rename(filepath.Join(path, "text"), filepath.Join(path, "moved")))
			runGit(t, path, "add", "--intent-to-add", "moved")
		}}, want: counts{changed: 1}},
		// Below the similarity threshold a rename is a deletion and an addition.
		{name: "rename rewritten beyond similarity", setup: []func(*testing.T, string){move("text", "renamed"), func(t *testing.T, path string) {
			write(t, filepath.Join(path, "renamed"), "alpha\nbeta\ngamma\n")
			runGit(t, path, "add", "renamed")
		}}, want: counts{changed: 2, added: 3, deleted: 2}},
		// The index keeps the rename, the worktree content no longer matches it.
		{name: "staged rename rewritten in worktree", setup: []func(*testing.T, string){move("text", "renamed"), func(t *testing.T, path string) {
			write(t, filepath.Join(path, "renamed"), "alpha\nbeta\ngamma\n")
		}}, want: counts{changed: 1, added: 3, deleted: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := checkout(t, t.TempDir())
			for _, setup := range tc.setup {
				setup(t, path)
			}
			row := Inspect(context.Background(), Row{Path: path})
			require.Empty(t, row.Error)
			assert.Equal(t, tc.want, counts{row.Changed, row.Added, row.Deleted, row.Binary, row.Untracked})
			assert.Zero(t, row.Conflicts)
		})
	}
}

func TestInspectRenamesIgnoreRenameConfig(t *testing.T) {
	path := checkout(t, t.TempDir())
	runGit(t, path, "mv", "text", "main")
	runGit(t, path, "mv", "binary", "logo")
	write(t, filepath.Join(path, "copy"), "one\ntwo\n")
	runGit(t, path, "add", "copy")
	for _, key := range []string{"status.renames", "diff.renames"} {
		for _, value := range []string{"false", "true", "copies"} {
			t.Run(key+"="+value, func(t *testing.T) {
				runGit(t, path, "config", key, value)
				defer runGit(t, path, "config", "--unset", key)
				row := Inspect(context.Background(), Row{Path: path})
				require.Empty(t, row.Error)
				assert.Equal(t, 3, row.Changed)
				assert.Equal(t, 2, row.Added)
				assert.Zero(t, row.Deleted)
				assert.Equal(t, 1, row.Binary)
			})
		}
	}
}

func TestCountStatus(t *testing.T) {
	const oid = "0123456789012345678901234567890123456789"
	ordinary := "1 .M N... 100644 100644 100644 " + oid + " " + oid + " "
	renamed := "2 R. N... 100644 100644 100644 " + oid + " " + oid + " R100 "
	for _, tc := range []struct {
		name                          string
		raw                           string
		changed, untracked, conflicts int
		branch, commit                string
	}{
		{name: "empty"},
		{name: "branch", raw: "# branch.oid " + oid + "\x00# branch.head main\x00", branch: "main", commit: oid},
		{name: "ordinary with newline", raw: ordinary + "a\nb\x00", changed: 1},
		{name: "untracked that looks like records", raw: "? u x\n1 y\x00", untracked: 1},
		{name: "rename", raw: renamed + "new\x00old\x00", changed: 1},
		{name: "rename with newline and tab", raw: renamed + "new\nname\x00old\tname\x00", changed: 1},
		{name: "original like conflict", raw: renamed + "new\x00u old\nfile\x00", changed: 1},
		{name: "original like ordinary", raw: renamed + "new\x001 x\x00", changed: 1},
		{name: "original like rename", raw: renamed + "new\x002 x\x00" + renamed + "b\x00a\x00", changed: 2},
		{name: "original like untracked", raw: renamed + "new\x00? y\x00", changed: 1},
		{name: "original like header", raw: "# branch.head main\x00" + renamed + "new\x00# branch.head evil\x00", changed: 1, branch: "main"},
		{name: "conflict", raw: "u UU N... 100644 100644 100644 100644 " + oid + " " + oid + " " + oid + " f\x00", changed: 1, conflicts: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var row Row
			countStatus(&row, tc.raw)
			assert.Equal(t, Row{Changed: tc.changed, Untracked: tc.untracked, Conflicts: tc.conflicts, Branch: tc.branch, Commit: tc.commit}, row)
		})
	}
}

func TestCountLines(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		raw                    string
		added, deleted, binary int
		err                    string
	}{
		{name: "empty"},
		{name: "ordinary with tab", raw: "3\t1\ta\tb\x00", added: 3, deleted: 1},
		{name: "binary", raw: "-\t-\tlogo\x00", binary: 1},
		{name: "pure rename", raw: "0\t0\t\x00old\x00new\x00"},
		{name: "rename with changes", raw: "2\t1\t\x00old\x00new\x00", added: 2, deleted: 1},
		{name: "binary rename", raw: "-\t-\t\x00logo.png\x00icon.png\x00", binary: 1},
		{name: "rename with newline and tab", raw: "1\t0\t\x00old\nname\x00new\tname\x00", added: 1},
		{name: "paths that look like records", raw: "0\t0\t\x005\t5\tx\x00-\t-\ty\x001\t1\tz\x00", added: 1, deleted: 1},
		{name: "rename between records", raw: "1\t0\ta\x000\t0\t\x00u old\nfile\x00b\x00-\t-\tc\x00", added: 1, binary: 1},
		{name: "missing fields", raw: "1\t0\x00", err: "invalid Git numstat"},
		{name: "truncated rename", raw: "1\t0\t\x00old\x00", err: "invalid Git numstat"},
		{name: "rename without paths", raw: "1\t0\t\x00\x00\x00", err: "invalid Git numstat"},
		{name: "invalid counts", raw: "x\t0\ta\x00", err: "invalid Git line counts"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var row Row
			err := countLines(&row, tc.raw)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, Row{Added: tc.added, Deleted: tc.deleted, Binary: tc.binary}, row)
		})
	}
}

func TestInspectAgainRecomputes(t *testing.T) {
	ctx := context.Background()
	path := checkout(t, t.TempDir())
	runGit(t, path, "checkout", "-b", "remote-work")
	write(t, filepath.Join(path, "remote"), "remote\n")
	runGit(t, path, "add", "remote")
	runGit(t, path, "commit", "-m", "remote")
	runGit(t, path, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, path, "checkout", "main")
	write(t, filepath.Join(path, "local"), "local\n")
	runGit(t, path, "add", "local")
	runGit(t, path, "commit", "-m", "local")
	write(t, filepath.Join(path, "text"), "one\nthree\n")
	write(t, filepath.Join(path, "new"), "x\n")
	runGit(t, path, "config", "remote.origin.pushurl", "no_push")
	first := Inspect(ctx, Row{ID: 1, Path: path, displayPath: "repo", Pinned: true})
	require.Empty(t, first.Error)
	assert.Equal(t, [6]int{1, 1, 1, 1, 1, 1}, [6]int{first.Changed, first.Untracked, first.Added, first.Deleted, first.Ahead, first.Behind})
	assert.True(t, first.PushLocked)
	assert.Equal(t, first, Inspect(ctx, first))
	// A failed inspection carries nothing from the previous one, and a row
	// that failed reads as a fresh one once the checkout is repaired.
	index := filepath.Join(path, ".git", "index")
	saved, err := os.ReadFile(index)
	require.NoError(t, err)
	write(t, index, "invalid index")
	failed := Inspect(ctx, first)
	require.Equal(t, "error", failed.Status)
	assert.Equal(t, Row{ID: 1, Repository: "acme/tool", Path: path, DefaultBranch: "main", Status: "error", Error: failed.Error,
		PushLocked: true, Pinned: true, displayPath: "repo"}, failed)
	assert.Equal(t, failed, Inspect(ctx, failed))
	require.NoError(t, os.WriteFile(index, saved, 0o600))
	assert.Equal(t, first, Inspect(ctx, failed))
	write(t, filepath.Join(path, ".git", "config"), "[invalid config")
	broken := Inspect(ctx, first)
	assert.Equal(t, "error", broken.Status)
	assert.False(t, broken.PushLocked, "an unreadable configuration shows no lock")
}

func TestRemoteName(t *testing.T) {
	tests := []struct {
		raw, want string
	}{
		{"git@github.com:octomation/maintainer.git\n", "octomation/maintainer"},
		{"github.com:octomation/maintainer", "octomation/maintainer"},
		{"ssh://git@github.com:2222/octomation/maintainer.git", "octomation/maintainer"},
		{"https://github.com/octomation/maintainer.git", "octomation/maintainer"},
		{"https://github.com/octomation/maintainer/", "octomation/maintainer"},
		{"https://gitlab.com/group/subgroup/project.git", "group/subgroup/project"},
		{"git@[::1]:octomation/maintainer.git", "octomation/maintainer"},
		{"  first.last@github-work:octomation/maintainer.git  ", "octomation/maintainer"},

		// No host, the file scheme or a single segment names no repository.
		{"", ""},
		{"/srv/git/octomation/maintainer.git", ""},
		{"../octomation/maintainer", ""},
		{"file:///srv/git/octomation/maintainer.git", ""},
		{"file://host/srv/octomation/maintainer.git", ""},
		{`C:\src\octomation\maintainer`, ""},
		{"C:/src/octomation/maintainer", ""},
		{"https://github.com/maintainer.git", ""},
		{"svn://github.com/octomation/maintainer", ""},
		{"ssh://git@github.com:octomation/maintainer.git", ""},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			assert.Equal(t, tc.want, RemoteName(tc.raw))
		})
	}
}
