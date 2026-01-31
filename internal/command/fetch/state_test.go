package fetch_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/fetch"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func execute(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := New(nil)
	cmd.SilenceUsage = true // as under the maintainer root command
	var out, errw bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errw)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errw.String(), err
}

func writeState(t *testing.T, path string, records ...state.Record) {
	t.Helper()
	st := state.New()
	for _, rec := range records {
		st.Upsert(rec)
	}
	raw, err := json.Marshal(st)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, raw, 0o600))
}

func stateRecords(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var st state.State
	require.NoError(t, json.Unmarshal(raw, &st))
	return len(st.Repos)
}

// --config="" means the same for the state commands as for fetch: no config
// discovery and the default state location; an omitted flag keeps discovery.
func TestStateCommandsHonorConfigSelection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "none"))
	other := filepath.Join(dir, "other.json")
	writeState(t, other, state.Record{ID: 1, OwnerLogin: "a", Name: "b", Path: filepath.Join(dir, "gone")})
	cnf := filepath.Join(dir, "fetch.toml")
	require.NoError(t, os.WriteFile(cnf, []byte("[defaults]\nstate_file = '"+other+"'\n[profiles.primary]\ntoken_env = 'GITHUB_TOKEN'\n"), 0o600))

	for name, discover := range map[string]func(t *testing.T){
		"environment": func(t *testing.T) { t.Setenv("MAINTAINER_FETCH_CONFIG", cnf) },
		"working directory": func(t *testing.T) {
			t.Setenv("MAINTAINER_FETCH_CONFIG", "")
			t.Chdir(dir)
		},
	} {
		t.Run(name, func(t *testing.T) {
			discover(t)
			out, _, err := execute(t, "state", "show")
			require.NoError(t, err)
			assert.Contains(t, out, `"owner_login": "a"`, "an omitted --config keeps discovery")

			out, _, err = execute(t, "--config=", "state", "show")
			require.NoError(t, err)
			var st state.State
			require.NoError(t, json.Unmarshal([]byte(out), &st))
			assert.Empty(t, st.Repos, "--config=\"\" shows the default store")

			out, _, err = execute(t, "--config=", "state", "prune")
			require.NoError(t, err)
			assert.Equal(t, "pruned 0 record(s)\n", out)
			assert.Equal(t, 1, stateRecords(t, other), "--config=\"\" never prunes the configured store")
		})
	}
}

// prune forgets only records whose path does not exist; any other failure
// keeps the record, names its path and reason, and exits 1.
func TestStatePruneKeepsUncheckedRecords(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions do not apply to root")
	}
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "xdg"))
	locked := filepath.Join(dir, "locked")
	present := filepath.Join(dir, "present")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "repo"), 0o755))
	require.NoError(t, os.MkdirAll(present, 0o755))
	require.NoError(t, os.Chmod(locked, 0))
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	path := filepath.Join(dir, "state.json")
	writeState(t, path,
		state.Record{ID: 1, OwnerLogin: "a", Name: "gone", Path: filepath.Join(dir, "gone")},
		state.Record{ID: 2, OwnerLogin: "a", Name: "locked", Path: filepath.Join(locked, "repo")},
		state.Record{ID: 3, OwnerLogin: "a", Name: "present", Path: present},
	)
	cnf := filepath.Join(dir, "fetch.toml")
	require.NoError(t, os.WriteFile(cnf, []byte("[defaults]\nstate_file = '"+path+"'\n"), 0o600))

	out, errs, err := execute(t, "--config", cnf, "state", "prune")
	require.Error(t, err)
	var coded exit.Coder
	assert.False(t, errors.As(err, &coded), "an unchecked record is a state error, exit 1")
	assert.Contains(t, err.Error(), "kept 1 record(s)")
	assert.Equal(t, "pruned 1 record(s)\n", out)
	assert.Contains(t, errs, "prune: id=1 a/gone (missing")
	assert.Contains(t, errs, "prune: id=2 a/locked kept, cannot check "+filepath.Join(locked, "repo"))
	assert.Contains(t, errs, "permission denied")
	require.NoError(t, os.Chmod(locked, 0o755))
	assert.Equal(t, 2, stateRecords(t, path))
}
