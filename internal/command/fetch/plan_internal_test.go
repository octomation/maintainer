package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	githubsvc "go.octolab.org/toolset/maintainer/internal/service/github"
)

// fakeGitHub lists one public repository of acme and answers checks by id
// and lookups by name; it sends no request.
type fakeGitHub struct{ snap githubsvc.RepoSnapshot }

func (f fakeGitHub) List(_ context.Context, p githubsvc.Profile) (githubsvc.Discovery, error) {
	snap := f.snap
	snap.SourceProfile = p.Name
	return githubsvc.Discovery{Profile: p.Name, Snapshots: []githubsvc.RepoSnapshot{snap}}, nil
}

func (f fakeGitHub) ConfirmByID(context.Context, githubsvc.Profile, int64) (githubsvc.RepoSnapshot, error) {
	return f.snap, nil
}

func (f fakeGitHub) ResolveByName(context.Context, githubsvc.Profile, string, string) (githubsvc.RepoSnapshot, error) {
	return f.snap, nil
}

// A run in a folder with a planted fetch.toml plans from the user config:
// the clone goes under the user's root, never under the planted one, and the
// planted file gets the notice.
func TestPlanUsesUserConfigNotWorkingDirectory(t *testing.T) {
	base := t.TempDir()
	work := filepath.Join(base, "work")
	userRoot := filepath.Join(base, "mine")
	planted := filepath.Join(base, "planted")
	require.NoError(t, os.MkdirAll(work, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(work, "fetch.toml"), []byte(
		"[workspace]\nroot = '"+planted+"'\npath = 'evil/{{.Repo}}'\n"+
			"[profiles.hostile]\ntoken_env = 'HOSTILE_TOKEN'\ninclude_owners = ['acme']\n"), 0o600))
	user := filepath.Join(base, "config", "maintainer", "fetch.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(user), 0o700))
	require.NoError(t, os.WriteFile(user, []byte(
		"[workspace]\nroot = '"+userRoot+"'\n"+
			"[profiles.user]\ntoken_env = 'USER_TOKEN'\ninclude_owners = ['acme']\n"), 0o600))
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(base, "state"))
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("HOSTILE_TOKEN", "hostile")
	t.Setenv("USER_TOKEN", "user")
	t.Chdir(work)

	previous := newDiscoverer
	t.Cleanup(func() { newDiscoverer = previous })
	newDiscoverer = func() gitHub {
		return fakeGitHub{snap: githubsvc.RepoSnapshot{
			ID: 1, Owner: "acme", Name: "svc", Visibility: githubsvc.Public, DefaultBranch: "main",
			HTTPSCloneURL: "https://github.com/acme/svc.git", SSHCloneURL: "git@github.com:acme/svc.git",
		}}
	}

	cmd := New(nil)
	var out, errs bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errs)
	cmd.SetArgs([]string{"--format=json"})
	require.NoError(t, cmd.Execute(), errs.String())

	assert.Contains(t, errs.String(), "notice: fetch.toml in the working directory is not read")
	assert.Contains(t, errs.String(), "config: loaded "+user)
	var plan struct {
		Actions []struct {
			Kind   string `json:"kind"`
			ToPath string `json:"to_path"`
			Path   string `json:"path"`
		} `json:"actions"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
	require.Len(t, plan.Actions, 1, out.String())
	assert.Equal(t, "clone", plan.Actions[0].Kind)
	want := filepath.Join(userRoot, "public", "acme", "svc")
	assert.Contains(t, []string{plan.Actions[0].Path, plan.Actions[0].ToPath}, want, out.String())
	assert.False(t, strings.Contains(out.String(), planted), "the planted root is never planned")
	_, err := os.Stat(planted)
	assert.True(t, os.IsNotExist(err))
}

// A run whose implicit root is a working directory reached through a symlink
// plans as in the folder the symlink names; a root named as a symlink is
// still refused.
func TestImplicitRootThroughSymlink(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	real := filepath.Join(base, "real")
	link := filepath.Join(base, "link")
	require.NoError(t, os.MkdirAll(real, 0o700))
	require.NoError(t, os.Symlink("real", link))
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(base, "state"))
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	t.Setenv("GITHUB_TOKEN", "token")

	previous := newDiscoverer
	t.Cleanup(func() { newDiscoverer = previous })
	newDiscoverer = func() gitHub {
		return fakeGitHub{snap: githubsvc.RepoSnapshot{
			ID: 1, Owner: "acme", Name: "svc", Visibility: githubsvc.Public, DefaultBranch: "main",
			HTTPSCloneURL: "https://github.com/acme/svc.git", SSHCloneURL: "git@github.com:acme/svc.git",
		}}
	}
	run := func(dir string, args ...string) (string, error) {
		t.Chdir(dir)
		cmd := New(nil)
		var out, errs bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errs)
		cmd.SetArgs(append([]string{"--format=json"}, args...))
		err := cmd.Execute()
		var plan struct {
			Actions json.RawMessage `json:"actions"`
		}
		if err == nil {
			require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
		}
		return string(plan.Actions), err
	}

	want, err := run(real, "--config=", "--owner", "acme")
	require.NoError(t, err)
	assert.Contains(t, want, filepath.Join(real, "public", "acme", "svc"))
	got, err := run(link, "--config=", "--owner", "acme")
	require.NoError(t, err)
	assert.Equal(t, want, got)

	cnf := filepath.Join(base, "fetch.toml")
	require.NoError(t, os.WriteFile(cnf, []byte("[workspace]\nroot = '"+link+"'\n"+
		"[profiles.user]\ntoken_env = 'GITHUB_TOKEN'\ninclude_owners = ['acme']\n"), 0o600))
	_, err = run(base, "--config", cnf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "symlink roots are not followed")
}
