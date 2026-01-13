package status_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/command/status"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	statussvc "go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestOfflineOutputAndConfigDiscovery(t *testing.T) {
	root := t.TempDir()
	configDir := t.TempDir()
	stateDir := filepath.Join(t.TempDir(), "state")
	t.Setenv("XDG_STATE_HOME", stateDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	t.Setenv("GITHUB_TOKEN", "")
	configPath := filepath.Join(configDir, "maintainer", "fetch.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o700))
	require.NoError(t, os.WriteFile(configPath, []byte("[defaults]\nroot = '"+root+"'\n[profiles.primary]\ntoken_env = 'UNRESOLVED_TOKEN'\n"), 0o600))
	for _, format := range []string{"auto", "plain", "json"} {
		cmd := status.New()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"--format=" + format})
		require.NoError(t, cmd.ExecuteContext(context.Background()))
		assert.NotContains(t, out.String(), "\x1b")
		if format == "json" {
			var rows []statussvc.Row
			require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
			assert.NotNil(t, rows)
			assert.Empty(t, rows)
		} else {
			assert.Contains(t, out.String(), "0 checkouts")
		}
	}
	_, err := os.Stat(stateDir)
	assert.True(t, os.IsNotExist(err), "status must not create state/lock files")
	// Explicit empty config bypasses even a malformed discovered config.
	require.NoError(t, os.WriteFile(configPath, []byte("invalid ["), 0o600))
	cmd := status.New()
	cmd.SetArgs([]string{"--config=", "--root", root, "--format=json"})
	cmd.SetOut(new(bytes.Buffer))
	require.NoError(t, cmd.Execute())
}

func TestInvalidOptions(t *testing.T) {
	for _, args := range [][]string{{"--format=wat"}, {"--concurrency=0"}, {"--concurrency=-1"}, {"--timeout=-1s"}, {"--format=tui", "--config=", "--root", t.TempDir()}} {
		cmd := status.New()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs(args)
		err := cmd.Execute()
		require.Error(t, err)
		var coded exit.Coder
		require.ErrorAs(t, err, &coded)
		assert.Equal(t, 2, coded.ExitCode())
	}
}

func TestErrorRowsDoNotHideSuccessfulCheckouts(t *testing.T) {
	root := t.TempDir()
	_, err := git.PlainInit(filepath.Join(root, "public/acme/present"), false)
	require.NoError(t, err)
	statePath := filepath.Join(t.TempDir(), "state.json")
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "missing", Path: filepath.Join(root, "public/acme/missing")})
	require.NoError(t, state.NewStore(afero.NewOsFs(), statePath, nil).Save(st))
	configPath := filepath.Join(t.TempDir(), "fetch.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("[defaults]\nroot = '"+root+"'\nstate_file = '"+statePath+"'\n"), 0o600))
	cmd := status.New()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--config", configPath, "--format=json"})
	require.ErrorContains(t, cmd.Execute(), "could not inspect 1 checkout")
	var rows []statussvc.Row
	require.NoError(t, json.Unmarshal(output.Bytes(), &rows))
	require.Len(t, rows, 2)
	assert.Equal(t, "error", rows[0].Status)
	assert.Equal(t, "unborn", rows[1].Status)
}
