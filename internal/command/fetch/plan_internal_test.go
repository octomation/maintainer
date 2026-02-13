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
