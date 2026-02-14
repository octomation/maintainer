// Package web serves the local maintainer service over HTTP on loopback.
package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
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

// Listen binds every address and, for each port among them, 127.0.0.1 and
// [::1] on that port too: the label host resolves to both, so the service
// holds both before it serves data, and no other program can answer on the
// other family. Addresses with port 0 share the one port the first of them
// gets. With optional set, an address that the system does not have, such as
// [::1] with IPv6 turned off, is passed to skip and left out; an added
// loopback address is always optional. Any other failure, such as a port that
// another program holds, closes what was bound and is an error that names the
// address; so is no address bound at all.
func Listen(ctx context.Context, addresses []string, optional bool, skip func(address string, err error)) ([]net.Listener, error) {
	type target struct {
		ip       string
		optional bool
	}
	var ports []string
	groups := map[string][]target{}
	for _, address := range addresses {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		if _, ok := groups[port]; !ok {
			ports = append(ports, port)
		}
		groups[port] = append(groups[port], target{ip: host, optional: optional})
	}
	for _, port := range ports {
		for _, loopback := range []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback} {
			known := false
			for _, t := range groups[port] {
				if ip := net.ParseIP(t.ip); ip != nil && ip.Equal(loopback) {
					known = true
				}
			}
			if !known {
				groups[port] = append(groups[port], target{ip: loopback.String(), optional: true})
			}
		}
	}
	var listeners []net.Listener
	closeAll := func() {
		for _, listener := range listeners {
			_ = listener.Close()
		}
	}
	for _, port := range ports {
		actual := port
		for _, t := range groups[port] {
			address := net.JoinHostPort(t.ip, actual)
			listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
			if err != nil {
				if t.optional && unavailable(err) {
					skip(address, err)
					continue
				}
				closeAll()
				return nil, err
			}
			listeners = append(listeners, listener)
			if tcp, ok := listener.Addr().(*net.TCPAddr); ok && actual == "0" {
				actual = strconv.Itoa(tcp.Port)
			}
		}
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
// stopping, if not nil, runs once on a clean stop (ctx ended), before any
// listener is released, while the service still holds its addresses; after a
// failure of a listener it does not run, since that listener is already gone.
func Serve(ctx context.Context, listeners []net.Listener, handler http.Handler, grace time.Duration, stopping func()) error {
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
	if err == nil && stopping != nil {
		stopping()
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
