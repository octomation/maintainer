// Package service wires the `maintainer service` command group: the local
// maintainer service on loopback.
package service

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/service/web"
)

// Port is the port of the service; indexit owns 17840, and AGENTS.md records
// the ports of the local services.
const Port = 17841

// Listen holds the default addresses: maintainer.localhost resolves to both.
var Listen = []string{
	net.JoinHostPort("127.0.0.1", strconv.Itoa(Port)),
	net.JoinHostPort("::1", strconv.Itoa(Port)),
}

// New returns the `maintainer service` command group.
func New() *cobra.Command {
	command := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "service",
		Short: "run the local maintainer service",
	}
	command.AddCommand(runCommand())
	return command
}

func runCommand() *cobra.Command {
	var listen []string
	command := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "run",
		Short: "serve the local maintainer service on loopback",
		Long: "Serves the local maintainer service over HTTP on loopback, at\n" +
			web.URL(Port) + " by default: GET and HEAD of / answer with a page that\n" +
			"says the service is running. It runs as your user and listens on\n" +
			"127.0.0.1 and ::1; a default address the system lacks, as ::1 with IPv6\n" +
			"turned off, is skipped. --listen replaces the defaults and may repeat.\n" +
			"SIGINT or SIGTERM stops it once the responses in flight are sent.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, address := range listen {
				host, port, err := net.SplitHostPort(address)
				if err != nil {
					return exit.WithUser(fmt.Errorf("--listen: %w", err))
				}
				if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
					return exit.WithUser(fmt.Errorf("--listen must use a literal loopback IP address, not %q", host))
				}
				// A number only: an empty port or a service name would bind
				// another port than the one written; 0 picks a free one.
				if _, err := strconv.ParseUint(port, 10, 16); err != nil {
					return exit.WithUser(fmt.Errorf("--listen must use a port number from 0 to 65535, not %q", port))
				}
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			stderr := cmd.ErrOrStderr()
			listeners, err := web.Listen(ctx, listen, !cmd.Flags().Changed("listen"), func(address string, err error) {
				fmt.Fprintf(stderr, "maintainer service: skipping %s: %v\n", address, err)
			})
			if err != nil {
				return err
			}
			addresses := make([]string, 0, len(listeners))
			for _, listener := range listeners {
				addresses = append(addresses, listener.Addr().String())
			}
			fmt.Fprintf(stderr, "maintainer service: %s (URL: %s)\n", strings.Join(addresses, ", "), web.URLFor(listeners[0].Addr()))
			return web.Serve(ctx, listeners, web.Handler(), 5*time.Second)
		},
	}
	command.Flags().StringArrayVar(&listen, "listen", Listen, "loopback address to listen on, repeatable; replaces the defaults")
	return command
}
