package fetch_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/fetch"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

func TestNew(t *testing.T) {
	cmd := New(nil)
	require.NotNil(t, cmd)
	assert.Equal(t, "fetch", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	// persistent flags from §3.1 are present.
	for _, name := range []string{"config", "profile", "owner", "format", "concurrency", "timeout", "verbose", "quiet"} {
		assert.NotNilf(t, cmd.PersistentFlags().Lookup(name), "missing persistent flag --%s", name)
	}
	// --apply is fetch-local, not persistent (§3.2).
	assert.NotNil(t, cmd.Flags().Lookup("apply"))

	// the help promises only what the run does: existing checkouts are kept,
	// and --owner/--profile narrow the discovery of new repositories alone.
	assert.NotContains(t, cmd.Long, "never deletes")
	assert.Contains(t, cmd.Long, "Existing checkouts\nand local work are retained.")
	for _, name := range []string{"owner", "profile"} {
		usage := cmd.PersistentFlags().Lookup(name).Usage
		assert.Containsf(t, usage, "discover new repositories", "--%s names what it narrows", name)
		assert.Containsf(t, usage, "existing checkouts are still processed", "--%s names what it keeps", name)
	}

	// subcommands.
	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	assert.True(t, names["config"], "config subcommand registered")
	assert.True(t, names["state"], "state subcommand registered")
}

func TestInvalidOptions(t *testing.T) {
	// a run that gets past the checks must not find a real config or token.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	t.Setenv("GITHUB_TOKEN", "")

	for _, test := range []struct {
		args    []string
		message string
	}{
		{[]string{"--format=jsno"}, `unknown format "jsno" (want human|json)`},
		{[]string{"--format=JSON"}, `unknown format "JSON" (want human|json)`},
		{[]string{"--concurrency=-1"}, "concurrency and timeout must not be negative"},
		{[]string{"--timeout=-1s"}, "concurrency and timeout must not be negative"},
		// 0 keeps the configured concurrency: the run goes on to the config.
		{[]string{"--concurrency=0", "--timeout=0"}, "not configured"},
	} {
		cmd := New(nil)
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs(append([]string{"--config="}, test.args...))
		err := cmd.Execute()
		require.Errorf(t, err, "%v", test.args)
		var coded exit.Coder
		require.ErrorAsf(t, err, &coded, "%v", test.args)
		assert.Equalf(t, exit.User, coded.ExitCode(), "%v", test.args)
		assert.ErrorContainsf(t, err, test.message, "%v", test.args)
	}
}
