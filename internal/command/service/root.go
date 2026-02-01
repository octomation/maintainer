// Package service wires the `maintainer service` command group: the local
// maintainer service on loopback.
package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/service/hub"
	"go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/service/web"
	"go.octolab.org/toolset/maintainer/internal/state"
	assets "go.octolab.org/toolset/maintainer/web"
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
	var configPath string
	command := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "run",
		Short: "serve the local maintainer service on loopback",
		Long: "Serves the local maintainer service over HTTP on loopback, at\n" +
			web.URL(Port) + " by default: the pages /status, /inbox and /tablo show\n" +
			"the checkouts of the collection and the issues mirrored in their\n" +
			".github/issues and .github/roadmap, over the JSON API under /api/v0.\n" +
			"It reads service.toml from $XDG_CONFIG_HOME/maintainer, else\n" +
			"~/.config/maintainer, and writes a commented default there when there is\n" +
			"none (--config names another file, which must exist), and the fetch\n" +
			"config and state as `maintainer status` does, with the home directory as\n" +
			"the working directory. It runs as your user and listens on 127.0.0.1 and\n" +
			"::1; a default address the system lacks, as ::1 with IPv6 turned off, is\n" +
			"skipped. listen in service.toml replaces the defaults, and --listen, which\n" +
			"may repeat, replaces both. SIGINT or SIGTERM stops it once the responses\n" +
			"in flight are sent. Do not run the mirror's export or push, or another\n" +
			"editor of the issue files, while triaging in the browser.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, address := range listen {
				if err := config.CheckLoopback(address); err != nil {
					return exit.WithUser(fmt.Errorf("--listen %w", err))
				}
			}
			stderr := cmd.ErrOrStderr()
			runtime, err := load(configPath, cmd.Flags().Changed("config"), func(path string, err error) {
				if err != nil {
					fmt.Fprintf(stderr, "maintainer service: cannot write the default config %s: %v\n", path, err)
					return
				}
				fmt.Fprintf(stderr, "maintainer service: wrote the default config %s\n", path)
			})
			if err != nil {
				return exit.WithUser(err)
			}
			// --listen replaces listen in service.toml, which replaces the
			// defaults; only a default address may be missing on the system.
			optional := !cmd.Flags().Changed("listen")
			if optional && runtime.service.Listen != nil {
				listen, optional = runtime.service.Listen, false
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			listeners, err := web.Listen(ctx, listen, optional, func(address string, err error) {
				fmt.Fprintf(stderr, "maintainer service: skipping %s: %v\n", address, err)
			})
			if err != nil {
				return err
			}
			addresses := make([]string, 0, len(listeners))
			addrs := make([]net.Addr, 0, len(listeners))
			for _, listener := range listeners {
				addresses = append(addresses, listener.Addr().String())
				addrs = append(addrs, listener.Addr())
			}
			fmt.Fprintf(stderr, "maintainer service: %s (URL: %s)\n", strings.Join(addresses, ", "), web.URLFor(listeners[0].Addr()))
			if runtime.service.Source != "" {
				fmt.Fprintf(stderr, "maintainer service: config %s\n", runtime.service.Source)
			}
			if runtime.fetch.Source != "" {
				fmt.Fprintf(stderr, "maintainer service: fetch config %s\n", runtime.fetch.Source)
			}
			report := func(err error) { fmt.Fprintf(stderr, "maintainer service: index: %v\n", err) }
			runtime.index.Start(ctx, report)
			runtime.checkouts.Start(ctx)
			handler, err := web.NewHandler(web.Options{Addrs: addrs, Assets: assets.Dist(), Index: runtime.index, Checkouts: runtime.checkouts})
			if err != nil {
				return err
			}
			return web.Serve(ctx, listeners, handler, 5*time.Second)
		},
	}
	command.Flags().StringArrayVar(&listen, "listen", Listen, "loopback address to listen on, repeatable; replaces the defaults")
	command.Flags().StringVar(&configPath, "config", "", "service config path (default: $XDG_CONFIG_HOME/maintainer/service.toml)")
	return command
}

// runtime is what the service runs on: the configs, the index and the
// checkouts cache.
type runtime struct {
	service   *config.Service
	fetch     *config.Fetch
	index     *hub.Index
	checkouts *web.Checkouts
}

// load reads service.toml, the fetch config and the state with the home
// directory as the working directory, and prepares the index and the
// checkouts cache. State is read once per collection through Store.Load.
// Without --config and without a file at the default path, it writes the
// commented default there first and tells wrote; a failure leaves the
// defaults in effect.
func load(configPath string, explicit bool, wrote func(path string, err error)) (*runtime, error) {
	fs := afero.NewOsFs()
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home directory: %w", err)
	}
	lookup := config.ServiceConfigLookup{Explicit: configPath, ExplicitSet: explicit, Getenv: os.Getenv, Home: home}
	servicePath, _ := lookup.Resolve(fs)
	if explicit {
		servicePath = config.ResolveAgainst(home, configPath)
	}
	if servicePath == "" {
		if path := lookup.Default(); path != "" {
			created, err := config.EnsureServiceFile(path)
			switch {
			case err != nil:
				wrote(path, err)
			case created:
				wrote(path, nil)
				servicePath = path
			default:
				// Another process wrote it between the lookup and now.
				servicePath = path
			}
		}
	}
	service, err := config.LoadService(fs, servicePath)
	if err != nil {
		return nil, err
	}
	fetchPath, named := service.FetchConfigPath(fs, os.Getenv, home)
	if named {
		if _, err := fs.Stat(fetchPath); err != nil {
			return nil, fmt.Errorf("fetch config %q: %w", fetchPath, err)
		}
	}
	fetch, err := config.LoadFetch(fs, fetchPath)
	if err != nil {
		return nil, err
	}
	if err := fetch.Validate(); err != nil {
		return nil, err
	}
	statePath := fetch.Defaults.StateFile
	if statePath == "" {
		statePath = state.DefaultPath(os.Getenv, home)
	} else {
		statePath = config.ResolveAgainst(home, statePath)
	}
	store := state.NewStore(fs, statePath, nil)
	discover := func(ctx context.Context) (*status.Discovery, error) {
		st, err := store.Load()
		if err != nil {
			return nil, err
		}
		return status.Discover(ctx, fetch, st, home, home, nil)
	}
	roots := func(ctx context.Context) ([]hub.Root, error) {
		d, err := discover(ctx)
		if err != nil {
			return nil, err
		}
		return mirrorRoots(d.Rows()), nil
	}
	weights := service.Weights()
	repoWeight := func(r hub.Root) float64 { return service.RepoWeight(r.ID, r.Owner, r.Name) }
	checkouts := web.NewCheckouts(func(ctx context.Context) ([]status.Row, error) {
		d, err := discover(ctx)
		if err != nil {
			return nil, err
		}
		return d.Inspect(ctx, fetch.Defaults.Concurrency)
	})
	return &runtime{service: service, fetch: fetch, index: hub.New(roots, weights, repoWeight), checkouts: checkouts}, nil
}

// mirrorRoots selects the mirror roots among the discovered rows: a real
// checkout row without an error, so placeholders, orphans and historical
// paths grant no write authority. A row that state does not identify learns
// owner/name from its origin URL in .git/config.
func mirrorRoots(rows []status.Row) []hub.Root {
	var roots []hub.Root
	for _, row := range rows {
		if row.IsPlaceholder() || row.OrphanReason != "" || row.Error != "" {
			continue
		}
		if !hub.HasMirror(row.Path) {
			continue
		}
		owner, name, ok := strings.Cut(row.Repository, "/")
		if !ok {
			if origin := originName(row.Path); origin != "" {
				owner, name, ok = strings.Cut(origin, "/")
			}
		}
		if !ok || owner == "" || name == "" {
			continue
		}
		roots = append(roots, hub.Root{Path: row.Path, Owner: owner, Name: name, ID: row.ID, Visibility: row.Visibility})
	}
	return roots
}

// originName reads the origin URL of a checkout from its Git config, without
// running Git, and returns its owner/name. A linked worktree has a .git file
// that points into .git/worktrees/<name>, whose commondir names the
// repository's own directory, where config lives.
func originName(path string) string {
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil && !info.IsDir() {
		raw, err := os.ReadFile(gitDir)
		if err != nil {
			return ""
		}
		dir := strings.TrimSpace(strings.TrimPrefix(string(raw), "gitdir:"))
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(path, dir)
		}
		gitDir = dir
	}
	if raw, err := os.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
		common := strings.TrimSpace(string(raw))
		if !filepath.IsAbs(common) {
			common = filepath.Join(gitDir, common)
		}
		gitDir = common
	}
	raw, err := os.ReadFile(filepath.Join(gitDir, "config"))
	if err != nil {
		return ""
	}
	inOrigin := false
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			inOrigin = line == `[remote "origin"]`
			continue
		}
		if inOrigin {
			if value, ok := strings.CutPrefix(line, "url"); ok {
				value = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(value), "="))
				return status.RemoteName(value)
			}
		}
	}
	return ""
}
