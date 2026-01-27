package fetch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/fetch"
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
