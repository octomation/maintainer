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
