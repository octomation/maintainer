package web_test

import (
	"context"
	"io"
	"net"
	"net/http"
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
		go func() { served <- web.Serve(ctx, []net.Listener{listener}, handler, 5*time.Second) }()

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
		go func() { served <- web.Serve(ctx, []net.Listener{listener}, handler, 50*time.Millisecond) }()
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
		go func() { served <- web.Serve(ctx, []net.Listener{first, second}, handler, time.Second) }()
		for _, listener := range []net.Listener{first, second} {
			assert.Equal(t, "page", get(t, "http://"+listener.Addr().String()+"/"))
		}
		cancel()
		require.NoError(t, <-served)
	})

	t.Run("listener failure", func(t *testing.T) {
		healthy, broken := listen(t), listen(t)
		require.NoError(t, broken.Close())
		err := web.Serve(context.Background(), []net.Listener{healthy, broken}, http.NotFoundHandler(), time.Second)
		assert.Error(t, err)
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
		require.Len(t, listeners, 1)
		assert.Equal(t, []string{missing}, skipped)
		require.NoError(t, listeners[0].Close())
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

	t.Run("nothing bound", func(t *testing.T) {
		_, err := web.Listen(ctx, []string{missing}, true, func(string, error) {})
		assert.Error(t, err)
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
		require.NoError(t, again[0].Close())
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
