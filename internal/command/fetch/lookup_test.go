package fetch_test

import (
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

const notice = "fetch.toml in the working directory is not read"

// workspace prepares a working directory with a fetch.toml that the run does
// not name, a git checkout or not, and an empty user folder; it returns the
// working directory and the path of the user config.
func workspace(t *testing.T, checkout bool) (dir, user string) {
	t.Helper()
	base := t.TempDir()
	dir = filepath.Join(base, "work")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	if checkout {
		_, err := git.PlainInit(dir, false)
		require.NoError(t, err)
	}
	hostile := "[defaults]\nroot = '" + filepath.Join(base, "planted") + "'\n" +
		"[profiles.hostile]\ntoken_env = 'HOSTILE_TOKEN'\ninclude_owners = ['evil']\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fetch.toml"), []byte(hostile), 0o600))
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(base, "state"))
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("HOSTILE_TOKEN", "")
	t.Setenv("USER_TOKEN", "")
	t.Chdir(dir)
	return dir, filepath.Join(base, "config", "maintainer", "fetch.toml")
}

func writeUserConfig(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	body := "[profiles.user]\ntoken_env = 'USER_TOKEN'\ninclude_owners = ['acme']\n"
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
}

func requireUsage(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var coded exit.Coder
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, exit.User, coded.ExitCode())
}

// A fetch.toml in the working directory, a checkout or an unpacked archive,
// never steers fetch: the user config wins and the file gets a notice.
func TestWorkingDirectoryConfigIsNotRead(t *testing.T) {
	for name, checkout := range map[string]bool{"archive": false, "checkout": true} {
		t.Run(name, func(t *testing.T) {
			_, user := workspace(t, checkout)
			writeUserConfig(t, user)

			out, errs, err := execute(t, "config", "validate")
			require.NoError(t, err)
			assert.Contains(t, out, "ok: "+user)
			assert.Contains(t, errs, notice)

			// The run plans from the user config: its profile is the one
			// that lacks a token, the planted one is never seen.
			_, errs, err = execute(t)
			requireUsage(t, err)
			assert.Contains(t, err.Error(), `profile "user"`)
			assert.NotContains(t, err.Error(), "hostile")
			assert.Contains(t, errs, notice)
		})
	}
}

// With no other config, the file in the working directory stops the run
// instead of the single-run mode; naming it with --config works as before.
func TestWorkingDirectoryConfigAloneStops(t *testing.T) {
	dir, _ := workspace(t, false)

	for _, args := range [][]string{nil, {"--owner", "acme"}, {"config", "validate"}, {"state", "show"}} {
		out, _, err := execute(t, args...)
		requireUsage(t, err)
		assert.Contains(t, err.Error(), "no fetch config found: "+notice, args)
		assert.Contains(t, err.Error(), "--config "+filepath.Join(dir, "fetch.toml"), args)
		assert.Empty(t, out, args)
	}
	_, err := os.Stat(os.Getenv("XDG_STATE_HOME"))
	assert.True(t, os.IsNotExist(err), "a stopped run touches no state")

	out, errs, err := execute(t, "--config", "./fetch.toml", "config", "validate")
	require.NoError(t, err)
	assert.Contains(t, out, "ok: ./fetch.toml")
	assert.NotContains(t, errs, notice)
	_, errs, err = execute(t, "--config", "./fetch.toml")
	requireUsage(t, err)
	assert.Contains(t, err.Error(), `profile "hostile"`)
	assert.NotContains(t, errs, notice)
}

// --config="" keeps the single-run mode: no notice and no stop.
func TestEmptyConfigIgnoresWorkingDirectory(t *testing.T) {
	workspace(t, false)
	_, errs, err := execute(t, "--config=", "--owner", "acme")
	requireUsage(t, err)
	assert.EqualError(t, err, "a token is required (set $GITHUB_TOKEN), even for public-only discovery")
	assert.NotContains(t, errs, notice)
}

// config init writes the template to the user folder, keeps an existing file
// without --force, and the next run finds it.
func TestConfigInitWritesUserFile(t *testing.T) {
	_, user := workspace(t, false)
	require.NoError(t, os.Remove("fetch.toml"))

	out, _, err := execute(t, "config", "init")
	require.NoError(t, err)
	assert.Equal(t, "wrote "+user+"\n", out)
	_, err = os.Stat("fetch.toml")
	assert.True(t, os.IsNotExist(err), "init writes nothing to the working directory")

	_, _, err = execute(t, "config", "init")
	requireUsage(t, err)
	assert.Contains(t, err.Error(), "pass --force to overwrite")
	_, _, err = execute(t, "config", "init", "--force")
	require.NoError(t, err)

	out, _, err = execute(t, "config", "validate")
	require.NoError(t, err)
	assert.Contains(t, out, "ok: "+user)

	other := filepath.Join(t.TempDir(), "custom.toml")
	out, _, err = execute(t, "--config", other, "config", "init")
	require.NoError(t, err)
	assert.Equal(t, "wrote "+other+"\n", out)
}
