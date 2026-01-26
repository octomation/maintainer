package git_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
)

// UpdateRemote inherits the process environment, so fixtures and assertions
// must not depend on the developer's or the CI runner's Git configuration.
func TestMain(m *testing.M) {
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_CONFIG", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_COUNT"} {
		_ = os.Unsetenv(key)
	}
	_ = os.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	_ = os.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	os.Exit(m.Run())
}

// gitValue runs Git in path and returns its trimmed output.
func gitValue(t *testing.T, path string, args ...string) string {
	t.Helper()
	out, err := exec.CommandContext(t.Context(), "git", append([]string{"-C", path}, args...)...).CombinedOutput()
	require.NoError(t, err, "%s", out)
	return strings.TrimSpace(string(out))
}

// configured creates an empty repository, appends extra to its local config
// and returns the repository path with the resulting config text.
func configured(t *testing.T, extra string) (string, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "checkout")
	_, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	file := filepath.Join(dir, ".git", "config")
	base, err := os.ReadFile(file)
	require.NoError(t, err)
	config := string(base) + extra
	require.NoError(t, os.WriteFile(file, []byte(config), 0o644))
	return dir, config
}

func localConfigText(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	require.NoError(t, err)
	return string(raw)
}

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

func TestSync_EmptyRemoteLifecycle(t *testing.T) {
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

	// Repeated reconciliation of the still-empty upstream is a successful no-op.
	require.NoError(t, sync.Fetch(context.Background(), work, gitsvc.Auth{Transport: "https"}))
	info, err = sync.Inspect(work)
	require.NoError(t, err)
	assert.Empty(t, info.HeadShort)

	// Once the first branch appears upstream, the next fetch discovers it.
	local, err := gogit.PlainOpen(work)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("first push"), 0o644))
	wt, err := local.Worktree()
	require.NoError(t, err)
	_, err = wt.Add("README.md")
	require.NoError(t, err)
	hash, err := wt.Commit("first commit", &gogit.CommitOptions{
		Author: &object.Signature{Name: "t", Email: "t@e", When: time.Unix(2, 0)},
	})
	require.NoError(t, err)
	head, err := local.Head()
	require.NoError(t, err)
	require.NoError(t, local.Push(&gogit.PushOptions{RefSpecs: []gitconfig.RefSpec{
		gitconfig.RefSpec(head.Name().String() + ":" + head.Name().String()),
	}}))

	remoteRef := plumbing.NewRemoteReferenceName("origin", head.Name().Short())
	_ = local.Storer.RemoveReference(remoteRef)
	require.NoError(t, sync.Fetch(context.Background(), work, gitsvc.Auth{Transport: "https"}))
	fetched, err := local.Reference(remoteRef, true)
	require.NoError(t, err)
	assert.Equal(t, hash, fetched.Hash())
}

func TestSync_FetchKeepsRepositoryErrors(t *testing.T) {
	work := filepath.Join(t.TempDir(), "checkout")
	repo, err := gogit.PlainInit(work, false)
	require.NoError(t, err)
	missing := filepath.Join(t.TempDir(), "missing-upstream.git")
	_, err = repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{missing}})
	require.NoError(t, err)

	err = gitsvc.NewSync().Fetch(context.Background(), work, gitsvc.Auth{Transport: "https"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetch "+work)
}

func TestSync_InspectAndUpdateRemotePreservePushURLs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "locked")
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	const original = "git@github.com:acme/original.git"
	_, err = repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{original}})
	require.NoError(t, err)

	cfg, err := repo.Config()
	require.NoError(t, err)
	cfg.Raw.Section("remote").Subsection("origin").SetOption("pushurl", "no_push")
	require.NoError(t, repo.SetConfig(cfg))

	sync := gitsvc.NewSync()
	info, err := sync.Inspect(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{original}, info.Origins)
	assert.Equal(t, []string{"no_push"}, info.PushURLs)

	const renamed = "git@github.com:acme/renamed.git"
	require.NoError(t, sync.UpdateRemote(dir, renamed))
	info, err = sync.Inspect(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{renamed}, info.Origins)
	assert.Equal(t, []string{"no_push"}, info.PushURLs, "update_remote removed the push lock")
}

func TestSync_UpdateRemoteKeepsTheRestOfTheConfig(t *testing.T) {
	// The insteadOf rule matches origin's URL, before and after the update.
	dir, before := configured(t, `# a line comment
[commit]
	gpgsign
[remote "origin"]
	url = https://github.com/acme/old.git
	fetch = +refs/heads/*:refs/remotes/origin/*
	pushurl = no_push
[remote "upstream"]
	url = https://github.com/upstream/tool.git ; a trailing comment
	fetch = +refs/heads/*:refs/remotes/upstream/*
	pushurl = DISABLE
[url "git@github.com:"]
	insteadOf = https://github.com/
`)
	const renamed = "https://github.com/acme/new.git"
	sync := gitsvc.NewSync()
	require.NoError(t, sync.UpdateRemote(dir, renamed))

	want := strings.Replace(before, "\turl = https://github.com/acme/old.git\n", "\turl = "+renamed+"\n", 1)
	require.NotEqual(t, before, want)
	assert.Equal(t, want, localConfigText(t, dir), "only origin's url line may change")
	assert.Equal(t, renamed, gitValue(t, dir, "config", "--local", "--get-all", "remote.origin.url"))
	assert.Equal(t, "git@github.com:acme/new.git", gitValue(t, dir, "remote", "get-url", "origin"), "insteadOf no longer applies")
	assert.Equal(t, "no_push", gitValue(t, dir, "config", "--get-all", "remote.origin.pushurl"), "update_remote removed the push lock")
	assert.Equal(t, "https://github.com/upstream/tool.git", gitValue(t, dir, "config", "--get-all", "remote.upstream.url"))
	assert.Equal(t, "DISABLE", gitValue(t, dir, "config", "--get-all", "remote.upstream.pushurl"))
	assert.Equal(t, "true", gitValue(t, dir, "config", "--type=bool", "commit.gpgsign"))

	info, err := sync.Inspect(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{renamed}, info.Origins)
	assert.Equal(t, []string{"no_push"}, info.PushURLs)

	require.NoError(t, sync.UpdateRemote(dir, renamed))
	assert.Equal(t, want, localConfigText(t, dir))
}

func TestSync_UpdateRemoteLeavesACorrectURLAsWritten(t *testing.T) {
	const canonical = "git@github.com:acme/tool.git"
	dir, before := configured(t, `[remote "origin"]
	url = `+canonical+` # set by hand
	fetch = +refs/heads/*:refs/remotes/origin/*
`)
	require.NoError(t, gitsvc.NewSync().UpdateRemote(dir, canonical))
	assert.Equal(t, before, localConfigText(t, dir))
}

func TestSync_UpdateRemoteRequiresOrigin(t *testing.T) {
	dir, before := configured(t, `[remote "upstream"]
	url = https://github.com/upstream/tool.git
`)
	err := gitsvc.NewSync().UpdateRemote(dir, "https://github.com/acme/tool.git")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `has no "origin" remote`)
	assert.Equal(t, before, localConfigText(t, dir), "update_remote created origin")
}

// Inspect sees only the checkout's own config file, so existence, the no-op
// check and the read-back ignore global and included values.
func TestSync_UpdateRemoteReadsOnlyTheLocalConfig(t *testing.T) {
	const renamed = "https://github.com/acme/new.git"
	global := filepath.Join(t.TempDir(), "gitconfig")
	require.NoError(t, os.WriteFile(global, []byte("[remote \"origin\"]\n\turl = https://github.com/acme/global.git\n"), 0o644))

	t.Run("origin only in the global config", func(t *testing.T) {
		t.Setenv("GIT_CONFIG_GLOBAL", global)
		dir, before := configured(t, "")
		err := gitsvc.NewSync().UpdateRemote(dir, renamed)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `has no "origin" remote`)
		assert.Equal(t, before, localConfigText(t, dir))
	})

	t.Run("global value next to a local origin", func(t *testing.T) {
		t.Setenv("GIT_CONFIG_GLOBAL", global)
		dir, before := configured(t, "[remote \"origin\"]\n\turl = "+renamed+"\n")
		require.NoError(t, gitsvc.NewSync().UpdateRemote(dir, renamed))
		assert.Equal(t, before, localConfigText(t, dir))
	})

	t.Run("included value next to a local origin", func(t *testing.T) {
		dir, before := configured(t, "[include]\n\tpath = "+global+"\n[remote \"origin\"]\n\turl = https://github.com/acme/old.git\n")
		require.NoError(t, gitsvc.NewSync().UpdateRemote(dir, renamed))
		assert.Equal(t, strings.Replace(before, "acme/old.git", "acme/new.git", 1), localConfigText(t, dir))
	})
}

func TestSync_UpdateRemoteTakesTheURLAsAValue(t *testing.T) {
	dir, _ := configured(t, "[remote \"origin\"]\n\turl = https://github.com/acme/old.git\n")
	require.NoError(t, gitsvc.NewSync().UpdateRemote(dir, "--add"))
	assert.Equal(t, "--add", gitValue(t, dir, "config", "--local", "--get-all", "remote.origin.url"))
}

func TestSync_UpdateRemoteStaysInTheCheckout(t *testing.T) {
	const origin = "[remote \"origin\"]\n\turl = https://github.com/acme/old.git\n"
	const renamed = "https://github.com/acme/new.git"

	t.Run("hook variables", func(t *testing.T) {
		dir, before := configured(t, origin)
		other, untouched := configured(t, origin)
		gitDir := filepath.Join(other, ".git")
		t.Setenv("GIT_DIR", gitDir)
		t.Setenv("GIT_COMMON_DIR", gitDir)
		t.Setenv("GIT_CONFIG", filepath.Join(gitDir, "config"))

		require.NoError(t, gitsvc.NewSync().UpdateRemote(dir, renamed))
		assert.Equal(t, strings.Replace(before, "acme/old.git", "acme/new.git", 1), localConfigText(t, dir))
		assert.Equal(t, untouched, localConfigText(t, other))
	})

	t.Run("no repository of its own", func(t *testing.T) {
		parent, untouched := configured(t, origin)
		nested := filepath.Join(parent, "nested")
		require.NoError(t, os.Mkdir(nested, 0o755))

		require.Error(t, gitsvc.NewSync().UpdateRemote(nested, renamed))
		assert.Equal(t, untouched, localConfigText(t, parent))
		_, err := os.Stat(filepath.Join(nested, ".git"))
		assert.True(t, os.IsNotExist(err))
	})
}

func TestSync_UpdateRemoteReportsGitFailures(t *testing.T) {
	const origin = "[remote \"origin\"]\n\turl = https://github.com/acme/old.git\n"
	const renamed = "https://github.com/acme/new.git"

	t.Run("locked config", func(t *testing.T) {
		dir, before := configured(t, origin)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "config.lock"), nil, 0o644))

		err := gitsvc.NewSync().UpdateRemote(dir, renamed)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not lock config file")
		assert.Equal(t, before, localConfigText(t, dir))
	})

	t.Run("value not written", func(t *testing.T) {
		gitPath, err := exec.LookPath("git")
		require.NoError(t, err)
		bin := t.TempDir()
		// A git that reports success for the write but does not perform it.
		script := "#!/bin/sh\nfor arg do [ \"$arg\" = --replace-all ] && exit 0; done\nexec '" + gitPath + "' \"$@\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755))
		t.Setenv("PATH", bin)

		dir, _ := configured(t, origin)
		err = gitsvc.NewSync().UpdateRemote(dir, renamed)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "after the update")
	})
}

func TestSync_CloneRefusesExistingTargetWithoutModification(t *testing.T) {
	origin := originRepo(t)
	target := filepath.Join(t.TempDir(), "existing")
	require.NoError(t, os.Mkdir(target, 0o755))
	marker := filepath.Join(target, "local-only.txt")
	require.NoError(t, os.WriteFile(marker, []byte("must survive"), 0o644))

	err := gitsvc.NewSync().Clone(context.Background(), gitsvc.CloneOptions{
		URL: origin, Path: target, Auth: gitsvc.Auth{Transport: "https"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists; refusing to modify it")
	content, readErr := os.ReadFile(marker)
	require.NoError(t, readErr)
	assert.Equal(t, "must survive", string(content))
}

func TestSync_MoveMissingSourceErrors(t *testing.T) {
	sync := gitsvc.NewSync()
	err := sync.Move(filepath.Join(t.TempDir(), "nope"), filepath.Join(t.TempDir(), "dst"))
	assert.Error(t, err)
}
