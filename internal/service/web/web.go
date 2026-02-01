// Package web serves the local maintainer service over HTTP on loopback.
package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Host is the name of the service: .localhost names resolve to loopback.
const Host = "maintainer.localhost"

// URL returns the address of the service on port.
func URL(port int) string { return "http://" + Host + ":" + strconv.Itoa(port) + "/" }

// URLFor returns the address to open for a listener: Host when it listens on
// 127.0.0.1 or ::1, which the name resolves to, the literal address otherwise.
func URLFor(addr net.Addr) string {
	if tcp, ok := addr.(*net.TCPAddr); ok && (tcp.IP.Equal(net.IPv4(127, 0, 0, 1)) || tcp.IP.Equal(net.IPv6loopback)) {
		return URL(tcp.Port)
	}
	return "http://" + addr.String() + "/"
}

// Page is what the service answers at its root: it has nothing else to serve
// yet, so the page only says that it is running.
const Page = "<!doctype html>\n<html lang=\"en\">\n<head><meta charset=\"utf-8\"><title>maintainer service</title></head>\n<body><p>Service is running.</p></body>\n</html>\n"

// Handler answers GET and HEAD of / with Page. Every other path is not found.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", time.Time{}, strings.NewReader(Page))
	})
}

// Listen binds every address. With optional set, an address that the system
// does not have, such as [::1] with IPv6 turned off, is passed to skip and left
// out; any other failure, or no address bound at all, is an error.
func Listen(ctx context.Context, addresses []string, optional bool, skip func(address string, err error)) ([]net.Listener, error) {
	var listeners []net.Listener
	closeAll := func() {
		for _, listener := range listeners {
			_ = listener.Close()
		}
	}
	for _, address := range addresses {
		listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
		if err != nil {
			if optional && unavailable(err) {
				skip(address, err)
				continue
			}
			closeAll()
			return nil, err
		}
		listeners = append(listeners, listener)
	}
	if len(listeners) == 0 {
		return nil, fmt.Errorf("no address to listen on: %v", addresses)
	}
	return listeners, nil
}

// unavailable reports whether err means the system has no such address or
// address family, rather than, say, a port already in use.
func unavailable(err error) bool {
	return errors.Is(err, syscall.EADDRNOTAVAIL) || errors.Is(err, syscall.EAFNOSUPPORT) ||
		errors.Is(err, syscall.EPROTONOSUPPORT)
}

// Serve answers on every listener with handler until ctx ends, then shuts down
// gracefully: it stops accepting connections and waits up to grace for the
// responses in flight. A failure of a listener ends it with that error.
func Serve(ctx context.Context, listeners []net.Listener, handler http.Handler, grace time.Duration) error {
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	served := make(chan error, len(listeners))
	for _, listener := range listeners {
		go func() { served <- server.Serve(listener) }()
	}
	var err error
	pending := len(listeners)
	select {
	case err = <-served:
		pending--
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	if e := server.Shutdown(shutdown); e != nil {
		_ = server.Close()
		err = errors.Join(err, fmt.Errorf("shutdown: %w", e))
	}
	for ; pending > 0; pending-- {
		if e := <-served; !errors.Is(e, http.ErrServerClosed) {
			err = errors.Join(err, e)
		}
	}
	return err
}
