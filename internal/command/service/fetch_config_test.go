package service_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/command/service"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

const homeNotice = "maintainer service: notice: fetch.toml in the home folder is not read"

// fetchConfigs puts a fetch.toml in home, which the service never reads, and
// a valid one in the user folder when user is set; it returns both paths.
func fetchConfigs(t *testing.T, user bool) (inHome, inUserFolder string) {
	t.Helper()
	dir := isolate(t)
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	home := os.Getenv("HOME")
	inHome = filepath.Join(home, "fetch.toml")
	require.NoError(t, os.WriteFile(inHome, []byte("[defaults]\nroot = '"+filepath.Join(home, "planted")+"'\n"), 0o600))
	inUserFolder = filepath.Join(dir, "fetch.toml")
	if user {
		require.NoError(t, os.MkdirAll(dir, 0o700))
		require.NoError(t, os.WriteFile(inUserFolder, []byte("[defaults]\nroot = '"+filepath.Join(home, "mine")+"'\n"), 0o600))
	}
	return inHome, inUserFolder
}

// startOnce starts the service, stops it, and returns the lines it logged
// before it listened.
func startOnce(t *testing.T) []string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	_, _, before, done := startLogged(t, ctx, "--listen", "127.0.0.1:0")
	cancel()
	require.NoError(t, <-done)
	return before
}

// Next to the user folder config, a fetch.toml in home gives the notice.
func TestRunNoticesFetchConfigInHome(t *testing.T) {
	inHome, _ := fetchConfigs(t, true)
	logs := strings.Join(startOnce(t), "")
	assert.Contains(t, logs, homeNotice)
	assert.Contains(t, logs, `set config = "`+inHome+`" under [fetch] in service.toml`)
}

// A fetch config the service names, in service.toml or the environment,
// works as before and gives no notice.
func TestRunNamedFetchConfigGivesNoNotice(t *testing.T) {
	inHome, _ := fetchConfigs(t, false)
	t.Setenv("MAINTAINER_FETCH_CONFIG", inHome)
	assert.NotContains(t, strings.Join(startOnce(t), ""), "notice:")

	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	dir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "maintainer")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "service.toml"), []byte("[fetch]\nconfig = \"fetch.toml\"\n"), 0o600))
	assert.NotContains(t, strings.Join(startOnce(t), ""), "notice:")
}

// A fetch.toml in home and no other fetch config stop the start with a usage
// error before the service listens.
func TestRunStopsOnFetchConfigInHomeAlone(t *testing.T) {
	fetchConfigs(t, false)
	cmd := service.New()
	cmd.SetArgs([]string{"run", "--listen", "127.0.0.1:0"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := cmd.ExecuteContext(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fetch config found: fetch.toml in the home folder is not read")
	var coder exit.Coder
	require.True(t, errors.As(err, &coder))
	assert.Equal(t, exit.User, coder.ExitCode())
}
