package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	exitpkg "go.octolab.org/toolset/maintainer/internal/pkg/exit"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// gitIn runs Git in dir without the caller's configuration.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

// worktreeState captures what a move must leave as it was in a linked
// worktree: its branch, HEAD, index and files.
func worktreeState(t *testing.T, dir string) map[string]string {
	t.Helper()
	index, err := os.ReadFile(gitIn(t, dir, "rev-parse", "--path-format=absolute", "--git-path", "index"))
	require.NoError(t, err)
	snapshot := map[string]string{
		"branch": gitIn(t, dir, "branch", "--show-current"),
		"HEAD":   gitIn(t, dir, "rev-parse", "HEAD"),
		"refs":   gitIn(t, dir, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads"),
		"index":  string(index),
	}
	require.NoError(t, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(dir, path)
			snapshot["file "+rel] = string(data)
		}
		return nil
	}))
	return snapshot
}

// trackedWithWorktrees makes a tracked checkout of acme/old with one linked
// worktree per name, outside the workspace, and returns their paths.
func trackedWithWorktrees(t *testing.T, checkout string, store *state.Store, names ...string) []string {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(checkout), 0o755))
	gitIn(t, filepath.Dir(checkout), "init", "-q", "-b", "main", checkout)
	require.NoError(t, os.WriteFile(filepath.Join(checkout, "f"), []byte("base"), 0o644))
	gitIn(t, checkout, "add", "f")
	gitIn(t, checkout, "-c", "user.name=t", "-c", "user.email=t@e", "commit", "-q", "-m", "init")
	gitIn(t, checkout, "remote", "add", "origin", "https://github.com/acme/old.git")
	outside := t.TempDir()
	var worktrees []string
	for _, name := range names {
		worktree := filepath.Join(outside, name)
		gitIn(t, checkout, "worktree", "add", "-q", worktree, "-b", name)
		require.NoError(t, os.WriteFile(filepath.Join(worktree, "f"), []byte("changed in "+name), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(worktree, "untracked"), []byte(name), 0o644))
		real, err := filepath.EvalSymlinks(worktree)
		require.NoError(t, err)
		worktrees = append(worktrees, real)
	}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: checkout,
		RemoteURL: "https://github.com/acme/old.git", CloneURL: "https", SourceProfile: "p"})
	require.NoError(t, store.Save(st))
	return worktrees
}

func TestService_MoveKeepsLinkedWorktreesUsable(t *testing.T) {
	root := t.TempDir()
	old, renamed := filepath.Join(root, "public/acme/old"), filepath.Join(root, "public/acme/new")
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	worktrees := trackedWithWorktrees(t, old, store, "feature")
	before := worktreeState(t, worktrees[0])

	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public, SourceProfile: "p"}
	var out, errw bytes.Buffer
	run := func(format string, apply bool) error {
		out.Reset()
		errw.Reset()
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
			Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: renamedRepository{snapshot: snap},
			GitSync: noNetworkFetch{gitsvc.NewSync()}, Reporter: NewReporter(&out, &errw, format, 0, false),
		})
		require.NoError(t, err)
		return svc.Run(context.Background(), apply)
	}

	require.NoError(t, run(FormatHuman, false))
	assert.Contains(t, out.String(), "+ reconnect linked worktree "+worktrees[0])
	require.NoError(t, run(FormatJSON, false))
	var plan struct {
		Actions []struct {
			Kind      string   `json:"kind"`
			Worktrees []string `json:"worktrees"`
		} `json:"actions"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &plan))
	require.NotEmpty(t, plan.Actions)
	assert.Equal(t, "move", plan.Actions[0].Kind)
	assert.Equal(t, worktrees, plan.Actions[0].Worktrees)

	require.NoError(t, run(FormatJSON, true), errw.String())
	gitIn(t, worktrees[0], "status", "--short")
	assert.Equal(t, before, worktreeState(t, worktrees[0]))
	assert.Equal(t, filepath.Join(must(filepath.EvalSymlinks(renamed)), ".git"), gitIn(t, worktrees[0], "rev-parse", "--path-format=absolute", "--git-common-dir"))
}

func TestService_MoveNamesTheWorktreesItCannotBringBack(t *testing.T) {
	root := t.TempDir()
	old, renamed := filepath.Join(root, "public/acme/old"), filepath.Join(root, "public/acme/new")
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	worktrees := trackedWithWorktrees(t, old, store, "gone", "kept", "lost")
	require.NoError(t, os.RemoveAll(worktrees[0]))
	require.NoError(t, os.RemoveAll(worktrees[2]))
	fetched := &countingFetch{Sync: gitsvc.NewSync()}

	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public, SourceProfile: "p"}
	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: renamedRepository{snapshot: snap},
		GitSync: fetched, Reporter: NewReporter(&out, &errw, FormatHuman, 0, false),
	})
	require.NoError(t, err)
	err = svc.Run(context.Background(), true)
	var coded exitpkg.Coder
	require.True(t, errors.As(err, &coded), "%v", err)
	assert.Equal(t, exitpkg.Partial, coded.ExitCode())
	assert.Contains(t, errw.String(), "linked worktree "+worktrees[0]+" is not usable after the move")
	assert.Contains(t, errw.String(), "linked worktree "+worktrees[2]+" is not usable after the move")
	assert.NotContains(t, errw.String(), worktrees[1]+" is not usable")
	assert.Contains(t, out.String(), "errors=2", "each worktree is an error of its own")
	assert.Equal(t, []string{renamed}, fetched.paths, "the moved checkout is still refreshed")

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, renamed, loaded.Repos[0].Path, "the move stands")
	assert.Equal(t, "https://github.com/acme/new.git", loaded.Repos[0].RemoteURL)
	gitIn(t, worktrees[1], "status", "--short")
}

func TestService_MoveReportsAWorktreeItCannotFind(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads an unreadable file")
	}
	root := t.TempDir()
	old, renamed := filepath.Join(root, "public/acme/old"), filepath.Join(root, "public/acme/new")
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	worktrees := trackedWithWorktrees(t, old, store, "healthy", "hidden")
	require.NoError(t, os.Chmod(filepath.Join(old, ".git", "worktrees", "hidden", "gitdir"), 0))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(renamed, ".git", "worktrees", "hidden", "gitdir"), 0o644) })

	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public, SourceProfile: "p"}
	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: renamedRepository{snapshot: snap},
		GitSync: noNetworkFetch{gitsvc.NewSync()}, Reporter: NewReporter(&out, &errw, FormatHuman, 0, false),
	})
	require.NoError(t, err)
	err = svc.Run(context.Background(), true)
	var coded exitpkg.Coder
	require.True(t, errors.As(err, &coded), "%v", err)
	assert.Equal(t, exitpkg.Partial, coded.ExitCode())
	assert.Contains(t, errw.String(), "worktrees/hidden cannot be found")
	assert.Contains(t, out.String(), "errors=1")
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, renamed, loaded.Repos[0].Path, "the move stands")
	gitIn(t, worktrees[0], "status", "--short")
}

// countingFetch records the fetches of the real Git adapter instead of
// running them.
type countingFetch struct {
	*gitsvc.Sync
	paths []string
}

func (f *countingFetch) Fetch(_ context.Context, path string, _ gitsvc.Auth) error {
	f.paths = append(f.paths, path)
	return nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
