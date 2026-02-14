package web_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/web"
)

func TestServe(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/slow" {
				close(entered)
				<-release
			}
			_, _ = io.WriteString(w, "done")
		})
		listener := listen(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		served := make(chan error, 1)
		go func() { served <- web.Serve(ctx, []net.Listener{listener}, handler, 5*time.Second, nil) }()

		url := "http://" + listener.Addr().String()
		assert.Equal(t, "done", get(t, url+"/"))

		answered := make(chan string, 1)
		go func() { answered <- get(t, url+"/slow") }()
		<-entered
		cancel()
		select {
		case err := <-served:
			t.Fatalf("Serve returned before the response in flight: %v", err)
		case <-time.After(100 * time.Millisecond):
		}
		close(release)
		assert.Equal(t, "done", <-answered)
		require.NoError(t, <-served)

		conn, err := (&net.Dialer{Timeout: time.Second}).DialContext(context.Background(), "tcp", listener.Addr().String())
		if err == nil {
			conn.Close()
		}
		assert.Error(t, err, "no new connections after the shutdown")
	})

	t.Run("shutdown deadline", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		defer close(release)
		handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			close(entered)
			<-release
		})
		listener := listen(t)
		ctx, cancel := context.WithCancel(context.Background())
		served := make(chan error, 1)
		go func() { served <- web.Serve(ctx, []net.Listener{listener}, handler, 50*time.Millisecond, nil) }()
		go func() {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+listener.Addr().String()+"/", http.NoBody)
			if resp, err := http.DefaultClient.Do(req); err == nil {
				resp.Body.Close()
			}
		}()
		<-entered
		cancel()
		err := <-served
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("several listeners", func(t *testing.T) {
		first, second := listen(t), listen(t)
		ctx, cancel := context.WithCancel(context.Background())
		served := make(chan error, 1)
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "page") })
		go func() { served <- web.Serve(ctx, []net.Listener{first, second}, handler, time.Second, nil) }()
		for _, listener := range []net.Listener{first, second} {
			assert.Equal(t, "page", get(t, "http://"+listener.Addr().String()+"/"))
		}
		cancel()
		require.NoError(t, <-served)
	})

	t.Run("stopping runs while the addresses are held", func(t *testing.T) {
		listener := listen(t)
		address := listener.Addr().String()
		ctx, cancel := context.WithCancel(context.Background())
		var held bool
		calls := 0
		stopping := func() {
			calls++
			// Nothing else can bind the address yet.
			other, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", address)
			if err == nil {
				_ = other.Close()
			}
			held = err != nil
		}
		served := make(chan error, 1)
		go func() {
			served <- web.Serve(ctx, []net.Listener{listener}, http.NotFoundHandler(), time.Second, stopping)
		}()
		cancel()
		require.NoError(t, <-served)
		assert.Equal(t, 1, calls)
		assert.True(t, held, "stopping ran before the listener was released")
		again, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", address)
		require.NoError(t, err, "released after")
		require.NoError(t, again.Close())
	})

	t.Run("listener failure", func(t *testing.T) {
		healthy, broken := listen(t), listen(t)
		require.NoError(t, broken.Close())
		stopped := false
		err := web.Serve(context.Background(), []net.Listener{healthy, broken}, http.NotFoundHandler(), time.Second, func() { stopped = true })
		assert.Error(t, err)
		assert.False(t, stopped, "an error exit does not run stopping: the label file stays")
		assert.NotErrorIs(t, err, http.ErrServerClosed)
		_, dial := (&net.Dialer{Timeout: time.Second}).DialContext(context.Background(), "tcp", healthy.Addr().String())
		assert.Error(t, dial, "the healthy listener is closed too")
	})
}

func TestListen(t *testing.T) {
	ctx := context.Background()
	// ::2 is not assigned to any interface, as ::1 is not with IPv6 turned off.
	missing := "[::2]:0"

	t.Run("optional address the system lacks", func(t *testing.T) {
		var skipped []string
		listeners, err := web.Listen(ctx, []string{"127.0.0.1:0", missing}, true, func(address string, err error) {
			assert.Error(t, err)
			skipped = append(skipped, address)
		})
		require.NoError(t, err)
		require.Len(t, listeners, 2, "127.0.0.1 and the added ::1")
		port := listeners[0].Addr().(*net.TCPAddr).Port
		assert.Equal(t, []string{net.JoinHostPort("::2", strconv.Itoa(port))}, skipped)
		for _, listener := range listeners {
			require.NoError(t, listener.Close())
		}
	})

	t.Run("explicit address the system lacks", func(t *testing.T) {
		_, err := web.Listen(ctx, []string{"127.0.0.1:0", missing}, false, func(string, error) { t.Fatal("no skip") })
		assert.Error(t, err)
	})

	t.Run("busy port is not skipped", func(t *testing.T) {
		busy := listen(t)
		defer busy.Close()
		_, err := web.Listen(ctx, []string{busy.Addr().String()}, true, func(string, error) { t.Fatal("no skip") })
		assert.Error(t, err)
	})

	t.Run("both loopback families for each port", func(t *testing.T) {
		for _, address := range []string{"127.0.0.1:0", "[::1]:0"} {
			listeners, err := web.Listen(ctx, []string{address}, false, func(string, error) { t.Fatal("no skip") })
			require.NoError(t, err, address)
			require.Len(t, listeners, 2, address)
			first, second := listeners[0].Addr().(*net.TCPAddr), listeners[1].Addr().(*net.TCPAddr)
			assert.Equal(t, first.Port, second.Port, "one port for both families")
			ips := []string{first.IP.String(), second.IP.String()}
			assert.ElementsMatch(t, []string{"127.0.0.1", "::1"}, ips, address)
			for _, listener := range listeners {
				require.NoError(t, listener.Close())
			}
		}
	})

	t.Run("the other family taken stops the start and names it", func(t *testing.T) {
		busy, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "[::1]:0")
		require.NoError(t, err)
		defer busy.Close()
		port := strconv.Itoa(busy.Addr().(*net.TCPAddr).Port)
		_, err = web.Listen(ctx, []string{"127.0.0.1:" + port}, false, func(string, error) { t.Fatal("no skip") })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "[::1]:"+port)
		again, err := web.Listen(ctx, []string{"127.0.0.1:" + port}, true, func(string, error) {})
		require.Error(t, err, "a taken address is not skipped as missing")
		assert.Nil(t, again)
		free, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:"+port)
		require.NoError(t, err, "127.0.0.1 is released")
		require.NoError(t, free.Close())
	})

	t.Run("a later failure closes the earlier listeners", func(t *testing.T) {
		busy := listen(t)
		defer busy.Close()
		free := listen(t)
		address := free.Addr().String()
		require.NoError(t, free.Close())
		_, err := web.Listen(ctx, []string{address, busy.Addr().String()}, false, func(string, error) {})
		require.Error(t, err)
		again, err := web.Listen(ctx, []string{address}, false, func(string, error) {})
		require.NoError(t, err, "the first address is free again")
		for _, listener := range again {
			require.NoError(t, listener.Close())
		}
	})
}

func TestURL(t *testing.T) {
	assert.Equal(t, "http://maintainer.localhost:17841/", web.URL(17841))
	assert.Equal(t, "http://maintainer.localhost:8080/", web.URLFor(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}))
	assert.Equal(t, "http://maintainer.localhost:8080/", web.URLFor(&net.TCPAddr{IP: net.IPv6loopback, Port: 8080}))
	// maintainer.localhost does not resolve to other loopback addresses.
	assert.Equal(t, "http://127.0.0.2:8080/", web.URLFor(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 8080}))
}

func listen(t *testing.T) net.Listener {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return listener
}

func get(t *testing.T, url string) string {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if !assert.NoError(t, err) {
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
	if !assert.NoError(t, err) {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return string(body)
}
