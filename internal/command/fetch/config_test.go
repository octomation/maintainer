package fetch_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/fetch"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

// Every fetch command that loads the config refuses an unknown key as bad
// config, before it resolves a token or touches the state.
func TestUnknownConfigKeyIsBadConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("GITHUB_TOKEN", "")
	for _, file := range []string{"fetch.toml", "fetch.yaml"} {
		path := filepath.Join(dir, file)
		body := "[profiles.primary]\ntoken_env = \"GITHUB_TOKEN\"\ninclude_owner = [\"acme\"]\n"
		if file == "fetch.yaml" {
			body = "profiles:\n  primary:\n    token_env: GITHUB_TOKEN\n    include_owner: [acme]\n"
		}
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
		for _, args := range [][]string{
			{"--config", path},
			{"--config", path, "--apply"},
			{"--config", path, "config", "validate"},
			{"--config", path, "state", "show"},
			{"--config", path, "state", "prune"},
		} {
			cmd := New(nil)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(args)
			err := cmd.Execute()
			require.Error(t, err, args)
			var coded exit.Coder
			require.ErrorAs(t, err, &coded)
			assert.Equal(t, exit.User, coded.ExitCode(), args)
			assert.Contains(t, err.Error(), path)
			assert.Contains(t, err.Error(), "unknown key include_owner in [profiles.primary]; did you mean include_owners?")
		}
	}
	_, err := os.Stat(filepath.Join(dir, "xdg"))
	assert.True(t, os.IsNotExist(err), "a rejected config must not create the state directory")
}

// validate refuses a per-repo path that does not parse; a run reports it for
// its repository only.
func TestValidateRejectsBrokenRepoPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "xdg"))
	path := filepath.Join(dir, "fetch.toml")
	body := "[[repos]]\nmatch = { owner = \"acme\", name = \"broken\" }\npath = \"~/ext/{{.Owner\"\n"
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	cmd := New(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", path, "config", "validate"})
	err := cmd.Execute()
	require.Error(t, err)
	var coded exit.Coder
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, exit.User, coded.ExitCode())
	assert.EqualError(t, err, "repos[0].path of acme/broken is not a valid template: template: path:1: unclosed action")
	assert.NotContains(t, out.String(), "ok:")
}

// config init writes only a config that validate accepts at the same path:
// a TOML path gets the template, any other extension is a usage error that
// leaves no file.
func TestConfigInitWritesLoadableFormat(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", filepath.Join(t.TempDir(), "xdg"))
	for _, tc := range []struct {
		file string
		ok   bool
	}{
		{"fetch.toml", true},
		{"fetch.TOML", true},
		{"fetch.yaml", false},
		{"fetch.yml", false},
		{"fetch.json", false},
		{"fetch", false},
	} {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.file)
			run := func(args ...string) error {
				cmd := New(nil)
				var out bytes.Buffer
				cmd.SetOut(&out)
				cmd.SetErr(&out)
				cmd.SetArgs(append([]string{"--config", path}, args...))
				return cmd.Execute()
			}
			err := run("config", "init")
			if tc.ok {
				require.NoError(t, err)
				require.NoError(t, run("config", "validate"))
				return
			}
			require.Error(t, err)
			var coded exit.Coder
			require.ErrorAs(t, err, &coded)
			assert.Equal(t, exit.User, coded.ExitCode())
			assert.Contains(t, err.Error(), "want .toml")
			assert.Contains(t, err.Error(), `"`+filepath.Ext(tc.file)+`"`)
			_, statErr := os.Stat(path)
			assert.True(t, os.IsNotExist(statErr), "a refused init must leave no file")
		})
	}
}
