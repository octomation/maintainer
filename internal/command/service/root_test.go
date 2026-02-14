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
	"strconv"
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
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "the plain address has no data: "+address)
		req.Host = web.LabelHost(label(t, port)) + ":" + port
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
	require.Len(t, addresses, 2, "listen replaces both defaults; the port takes the other family too")
	host, _, err := net.SplitHostPort(addresses[0])
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", host)
	host, _, err = net.SplitHostPort(addresses[1])
	require.NoError(t, err)
	assert.Equal(t, "::1", host)

	ctx, cancel = context.WithCancel(context.Background())
	addresses, _, done = start(t, ctx, "--listen", "[::1]:0")
	cancel()
	require.NoError(t, <-done)
	require.Len(t, addresses, 2)
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

// labelPath is the label file of port in the isolated state folder.
func labelPath(port string) string {
	return filepath.Join(os.Getenv("XDG_STATE_HOME"), "maintainer", "service", port+".label")
}

// label reads the label of the run on port.
func label(t *testing.T, port string) string {
	t.Helper()
	raw, err := os.ReadFile(labelPath(port))
	require.NoError(t, err)
	value := strings.TrimSpace(string(raw))
	require.True(t, web.ValidLabel(value), value)
	return value
}

// getHost sends GET path to address with host as Host and returns the status.
func getHost(t *testing.T, address, host, path string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+address+path, http.NoBody)
	require.NoError(t, err)
	req.Host = host
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	return resp.StatusCode
}

func TestRunLabelFile(t *testing.T) {
	isolate(t)
	ctx, cancel := context.WithCancel(context.Background())
	addresses, _, before, done := startLogged(t, ctx, "--listen", "127.0.0.1:0")
	_, port, err := net.SplitHostPort(addresses[0])
	require.NoError(t, err)
	first := label(t, port)
	info, err := os.Stat(labelPath(port))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	info, err = os.Stat(filepath.Dir(labelPath(port)))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	assert.NotContains(t, strings.Join(before, ""), first, "the label is in no log line")
	assert.Equal(t, http.StatusOK, getHost(t, addresses[0], web.LabelHost(first)+":"+port, "/api/v0/catalog"))

	// A second instance on the same port fails and leaves the file be.
	second := service.New()
	second.SetArgs([]string{"run", "--listen", addresses[0]})
	second.SetOut(io.Discard)
	var stderr strings.Builder
	second.SetErr(&stderr)
	err = second.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), addresses[0])
	assert.NotContains(t, err.Error()+stderr.String(), first)
	assert.Equal(t, first, label(t, port))

	// A clean stop removes the file.
	cancel()
	require.NoError(t, <-done)
	_, err = os.Stat(labelPath(port))
	assert.True(t, os.IsNotExist(err), "a clean stop removes the label file")

	// A restart on the same port makes a new label; the old one is dead.
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	_, _, done = start(t, ctx, "--listen", addresses[0])
	again := label(t, port)
	assert.NotEqual(t, first, again)
	assert.Equal(t, http.StatusMisdirectedRequest, getHost(t, addresses[0], web.LabelHost(first)+":"+port, "/api/v0/catalog"))
	assert.Equal(t, http.StatusOK, getHost(t, addresses[0], web.LabelHost(again)+":"+port, "/api/v0/catalog"))
	assert.Equal(t, http.StatusUnauthorized, getHost(t, addresses[0], "maintainer.localhost:"+port, "/api/v0/catalog"))
	assert.Equal(t, http.StatusOK, getHost(t, addresses[0], "maintainer.localhost:"+port, "/status"))
	cancel()
	require.NoError(t, <-done)
}

func TestRunKeepsAReplacementLabel(t *testing.T) {
	isolate(t)
	ctx, cancel := context.WithCancel(context.Background())
	addresses, _, done := start(t, ctx, "--listen", "127.0.0.1:0")
	_, port, err := net.SplitHostPort(addresses[0])
	require.NoError(t, err)
	label(t, port)
	const replacement = "fedcba9876543210fedcba9876543210"
	require.NoError(t, os.WriteFile(labelPath(port), []byte(replacement+"\n"), 0o600))
	cancel()
	require.NoError(t, <-done)
	assert.Equal(t, replacement, label(t, port), "the file of another run stays")
}

func TestRunNeedsTheOtherFamily(t *testing.T) {
	isolate(t)
	busy, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "[::1]:0")
	require.NoError(t, err)
	defer busy.Close()
	_, port, err := net.SplitHostPort(busy.Addr().String())
	require.NoError(t, err)
	cmd := service.New()
	cmd.SetArgs([]string{"run", "--listen", "127.0.0.1:" + port})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err = cmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[::1]:"+port)
	_, err = os.Stat(labelPath(port))
	assert.True(t, os.IsNotExist(err), "no label file appears")
}

// runOpen runs `maintainer service open` with args and returns stdout.
func runOpen(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := service.New()
	cmd.SilenceUsage = true // as under the maintainer root command
	cmd.SetArgs(append([]string{"open"}, args...))
	var stdout strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	err := cmd.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestOpenPrint(t *testing.T) {
	isolate(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addresses, _, done := start(t, ctx, "--listen", "127.0.0.1:0")
	_, port, err := net.SplitHostPort(addresses[0])
	require.NoError(t, err)

	out, err := runOpen(t, "--listen", addresses[0], "--print")
	require.NoError(t, err)
	want := web.LabelURL(label(t, port), mustAtoi(t, port))
	assert.Equal(t, want+"\n", out)

	// The printed address opens the hub with data.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, want+"api/v0/catalog", http.NoBody)
	require.NoError(t, err)
	resp, err := (&http.Client{Transport: &http.Transport{DisableKeepAlives: true}}).Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Contains(t, []int{http.StatusOK, http.StatusAccepted}, resp.StatusCode)

	// The address comes from service.toml when --listen is not given.
	dir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "maintainer")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "service.toml"), []byte("listen = [\""+addresses[0]+"\"]\n"), 0o600))
	out, err = runOpen(t, "--print")
	require.NoError(t, err)
	assert.Equal(t, want+"\n", out)
	cancel()
	require.NoError(t, <-done)

	// Stopped: no file, so the hub is not running.
	_, err = runOpen(t, "--listen", addresses[0], "--print")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestOpenRefusesAnotherListener(t *testing.T) {
	isolate(t)
	const planted = "0123456789abcdef0123456789abcdef"
	other := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"proof":"`+strings.Repeat("0", 64)+`"}}`)
	})}
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = other.Serve(listener) }()
	defer other.Close()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, web.WriteLabel(filepath.Dir(labelPath(port)), mustAtoi(t, port), planted))

	out, err := runOpen(t, "--listen", listener.Addr().String(), "--print")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "another program listens on "+listener.Addr().String())
	assert.NotContains(t, err.Error(), planted)
	assert.Empty(t, out)

	// A file with a label but nothing on the port: not running.
	require.NoError(t, other.Close())
	_, err = runOpen(t, "--listen", listener.Addr().String(), "--print")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// A file that holds no label is refused before any request.
	require.NoError(t, os.WriteFile(labelPath(port), []byte("../x\n"), 0o600))
	_, err = runOpen(t, "--listen", listener.Addr().String(), "--print")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	for _, bad := range []string{"localhost:1", "127.0.0.1:0"} {
		_, err = runOpen(t, "--listen", bad, "--print")
		var coder exit.Coder
		if assert.True(t, errors.As(err, &coder), bad) {
			assert.Equal(t, exit.User, coder.ExitCode(), bad)
		}
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	require.NoError(t, err)
	return n
}
