package service_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/command/service"
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

// start runs the service with args and returns the addresses it reports.
func start(t *testing.T, ctx context.Context, args ...string) ([]string, string, <-chan error) {
	t.Helper()
	stderr, logs := io.Pipe()
	cmd := service.New()
	cmd.SetArgs(append([]string{"run"}, args...))
	cmd.SetOut(io.Discard)
	cmd.SetErr(logs)
	done := make(chan error, 1)
	go func() { done <- cmd.ExecuteContext(ctx); logs.Close() }()
	line, err := bufio.NewReader(stderr).ReadString('\n')
	require.NoError(t, err)
	go func() { _, _ = io.Copy(io.Discard, stderr) }()
	match := regexp.MustCompile(`^maintainer service: (.+) \(URL: (http://maintainer\.localhost:\d+/)\)\n$`).FindStringSubmatch(line)
	require.Len(t, match, 3, line)
	return strings.Split(match[1], ", "), match[2], done
}

func TestRunServesEveryAddressUntilCancelled(t *testing.T) {
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
