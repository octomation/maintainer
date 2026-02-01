package service_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/command/service"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/service/web"
)

func TestRunDefaults(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1:17841", "[::1]:17841"}, service.Listen)
	run, _, err := service.New().Find([]string{"run"})
	require.NoError(t, err)
	flag := run.Flags().Lookup("listen")
	require.NotNil(t, flag)
	assert.Equal(t, "[127.0.0.1:17841,[::1]:17841]", flag.DefValue)
	assert.Contains(t, run.Long, "http://maintainer.localhost:17841/")
	require.NotNil(t, run.Flags().Lookup("config"))
}

func TestRunRejectsNonLoopbackAddresses(t *testing.T) {
	isolate(t)
	for _, listen := range []string{"0.0.0.0:17841", "localhost:17841", "192.0.2.1:17841", ":17841", "127.0.0.1", "[::]:17841",
		"127.0.0.1:", "127.0.0.1:65536", "127.0.0.1:-1", "127.0.0.1:http", "[::1]:+80"} {
		cmd := service.New()
		cmd.SetArgs([]string{"run", "--listen", "127.0.0.1:0", "--listen", listen})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		// A regression would start serving: the deadline ends it instead of the test.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := cmd.ExecuteContext(ctx)
		cancel()
		var coder exit.Coder
		if assert.True(t, errors.As(err, &coder), listen) {
			assert.Equal(t, exit.User, coder.ExitCode(), listen)
		}
	}
}

// isolate gives the service an empty home, so a test neither reads the
// collection of whoever runs it nor writes into their config directory, and
// returns the directory service.toml lives in.
func isolate(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	return filepath.Join(home, ".config", "maintainer")
}

// start runs the service with args and returns the addresses it reports.
func start(t *testing.T, ctx context.Context, args ...string) ([]string, string, <-chan error) {
	t.Helper()
	addresses, url, _, done := startLogged(t, ctx, args...)
	return addresses, url, done
}

// startLogged is start that also returns the lines logged before the
// addresses.
func startLogged(t *testing.T, ctx context.Context, args ...string) ([]string, string, []string, <-chan error) {
	t.Helper()
	stderr, logs := io.Pipe()
	cmd := service.New()
	cmd.SetArgs(append([]string{"run"}, args...))
	cmd.SetOut(io.Discard)
	cmd.SetErr(logs)
	done := make(chan error, 1)
	go func() { done <- cmd.ExecuteContext(ctx); logs.Close() }()
	reader := bufio.NewReader(stderr)
	pattern := regexp.MustCompile(`^maintainer service: (.+) \(URL: (http://maintainer\.localhost:\d+/)\)\n$`)
	var before []string
	for {
		line, err := reader.ReadString('\n')
		require.NoError(t, err, strings.Join(before, ""))
		if match := pattern.FindStringSubmatch(line); match != nil {
			go func() { _, _ = io.Copy(io.Discard, reader) }()
			return strings.Split(match[1], ", "), match[2], before, done
		}
		before = append(before, line)
	}
}

func TestRunServesEveryAddressUntilCancelled(t *testing.T) {
	isolate(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addresses, url, done := start(t, ctx, "--listen", "127.0.0.1:0", "--listen", "[::1]:0")
	require.Len(t, addresses, 2)
	_, port, err := net.SplitHostPort(addresses[0])
	require.NoError(t, err)
	assert.Equal(t, "http://maintainer.localhost:"+port+"/", url)
	// Keep-alive is off: the transport may dial a spare connection that never
	// sends a request, and the graceful shutdown would wait for it.
	transport := &http.Transport{DisableKeepAlives: true}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for _, address := range addresses {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/", http.NoBody)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err, address)
		resp.Body.Close()
		assert.Equal(t, http.StatusFound, resp.StatusCode, address)
		assert.Equal(t, "/status", resp.Header.Get("Location"), address)
		for _, page := range web.Pages {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+page, http.NoBody)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err, address)
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode, address+page)
			assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"), address+page)
			assert.Contains(t, string(body), "<div id=\"app\">", address+page)
			assert.Contains(t, string(body), "/assets/", address+page)
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/api/v0/catalog", http.NoBody)
		require.NoError(t, err)
		resp, err = client.Do(req)
		require.NoError(t, err, address)
		resp.Body.Close()
		assert.Contains(t, []int{http.StatusOK, http.StatusAccepted}, resp.StatusCode, address)
	}
	cancel()
	assert.NoError(t, <-done)
}

func TestRunRejectsMissingExplicitConfig(t *testing.T) {
	isolate(t)
	cmd := service.New()
	cmd.SetArgs([]string{"run", "--listen", "127.0.0.1:0", "--config", filepath.Join(t.TempDir(), "missing.toml")})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := cmd.ExecuteContext(ctx)
	require.Error(t, err)
	var coder exit.Coder
	if assert.True(t, errors.As(err, &coder)) {
		assert.Equal(t, exit.User, coder.ExitCode())
	}
}

func TestRunReportsBusyAddress(t *testing.T) {
	isolate(t)
	ctx, cancel := context.WithCancel(context.Background())
	addresses, _, done := start(t, ctx, "--listen", "127.0.0.1:0")
	defer func() { cancel(); <-done }()

	second := service.New()
	second.SetArgs([]string{"run", "--listen", addresses[0]})
	second.SetOut(io.Discard)
	second.SetErr(io.Discard)
	err := second.ExecuteContext(context.Background())
	require.Error(t, err)
	var coder exit.Coder
	assert.False(t, errors.As(err, &coder), "a busy address is not a usage error")
}

func TestRunWritesTheDefaultConfigOnce(t *testing.T) {
	dir := isolate(t)
	path := filepath.Join(dir, "service.toml")
	for _, first := range []bool{true, false} {
		ctx, cancel := context.WithCancel(context.Background())
		_, _, before, done := startLogged(t, ctx, "--listen", "127.0.0.1:0")
		cancel()
		require.NoError(t, <-done)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, config.ServiceTemplate, data)
		wrote := "maintainer service: wrote the default config " + path + "\n"
		if first {
			assert.Equal(t, []string{wrote}, before)
		} else {
			assert.NotContains(t, before, wrote, "the second start leaves the file be")
		}
	}
}

func TestRunListensWhereServiceTomlSays(t *testing.T) {
	dir := isolate(t)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "service.toml"), []byte("listen = [\"127.0.0.1:0\"]\n"), 0o600))

	ctx, cancel := context.WithCancel(context.Background())
	addresses, _, done := start(t, ctx)
	cancel()
	require.NoError(t, <-done)
	require.Len(t, addresses, 1, "listen replaces both defaults")
	host, _, err := net.SplitHostPort(addresses[0])
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", host)

	ctx, cancel = context.WithCancel(context.Background())
	addresses, _, done = start(t, ctx, "--listen", "[::1]:0")
	cancel()
	require.NoError(t, <-done)
	require.Len(t, addresses, 1)
	host, _, err = net.SplitHostPort(addresses[0])
	require.NoError(t, err)
	assert.Equal(t, "::1", host, "--listen replaces listen")
}

func TestRunRejectsBadListenInServiceToml(t *testing.T) {
	for _, body := range []string{"listen = [\"localhost:17841\"]\n", "listen = []\n"} {
		dir := isolate(t)
		require.NoError(t, os.MkdirAll(dir, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "service.toml"), []byte(body), 0o600))
		cmd := service.New()
		cmd.SetArgs([]string{"run"})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := cmd.ExecuteContext(ctx)
		cancel()
		var coder exit.Coder
		if assert.True(t, errors.As(err, &coder), body) {
			assert.Equal(t, exit.User, coder.ExitCode(), body)
		}
	}
}
