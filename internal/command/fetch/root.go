// Package fetch wires the `maintainer fetch` command group: a state-reconciling
// fetcher that discovers GitHub repositories across owners and materialises a
// local checkout tree (fetch plan §3).
package fetch

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	fetchsvc "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	githubsvc "go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// options holds the persistent flag values for the fetch group.
type options struct {
	config      string
	profiles    []string
	owners      []string
	format      string
	concurrency int
	timeout     time.Duration
	verbose     int
	quiet       bool
	token       string
	apply       bool
}

// New returns the `maintainer fetch` command group (§3).
func New(_ *config.Tool) *cobra.Command {
	opts := new(options)

	command := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "fetch",
		Short: "reconcile local GitHub checkouts",
		Long: "Discovers GitHub repositories across owners and reconciles a local\n" +
			"checkout tree. Plan-only by default; --apply performs non-destructive\n" +
			"actions (clone, fetch, move, update-remote, adopt). It never deletes.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, opts)
		},
	}

	set := command.PersistentFlags()
	set.StringVar(&opts.config, "config", "", "fetch config path (--config=\"\" disables discovery)")
	set.StringSliceVar(&opts.profiles, "profile", nil, "limit work to these profiles (repeatable)")
	set.StringSliceVar(&opts.owners, "owner", nil, "limit work to these owners (repeatable)")
	set.StringVar(&opts.format, "format", fetchsvc.FormatHuman, "output format: human|json")
	set.IntVar(&opts.concurrency, "concurrency", 0, "parallel discovery/clone cap (0 = from config)")
	set.DurationVar(&opts.timeout, "timeout", 0, "wall-clock budget for the whole command")
	set.CountVarP(&opts.verbose, "verbose", "v", "increase verbosity (-v, -vv, -vvv)")
	set.BoolVarP(&opts.quiet, "quiet", "q", false, "suppress everything below error on stderr")
	set.StringVar(&opts.token, "token", "", "personal access token for the default/first profile")

	command.Flags().BoolVar(&opts.apply, "apply", false, "execute the plan (default: plan only)")

	command.AddCommand(
		configCommand(opts),
		stateCommand(opts),
	)
	return command
}

func run(cmd *cobra.Command, opts *options) error {
	if opts.quiet && opts.verbose > 0 {
		return exit.WithUser(fmt.Errorf("--quiet and --verbose are mutually exclusive"))
	}

	ctx := cmd.Context()
	if opts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.timeout)
		defer cancel()
	}

	fs := afero.NewOsFs()
	fcfg, err := loadConfig(cmd, fs, opts)
	if err != nil {
		return err
	}

	profiles, err := resolveProfiles(fs, fcfg, opts, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if fcfg.Source != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "config: loaded %s\n", fcfg.Source)
	}

	store, err := newStore(fs, fcfg)
	if err != nil {
		return err
	}

	concurrency := fcfg.Defaults.Concurrency
	if opts.concurrency > 0 {
		concurrency = opts.concurrency
	}

	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	discoverer := githubsvc.NewRESTDiscoverer(githubsvc.DefaultClientFactory)
	reporter := fetchsvc.NewReporter(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts.format, opts.verbose, opts.quiet)

	svc, err := fetchsvc.NewService(fcfg, profiles, home, cwd, concurrency, fetchsvc.Deps{
		Store:      store,
		Discoverer: discoverer,
		Confirmer:  discoverer,
		Resolver:   discoverer,
		GitSync:    gitsvc.NewSync(),
		Reporter:   reporter,
	})
	if err != nil {
		return err
	}
	return svc.Run(ctx, opts.apply)
}

// loadConfig resolves and loads the fetch config, then validates it (§4).
func loadConfig(cmd *cobra.Command, fs afero.Fs, opts *options) (*config.Fetch, error) {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	lookup := config.FetchConfigLookup{
		Explicit:    opts.config,
		ExplicitSet: cmd.Flags().Changed("config"),
		Getenv:      os.Getenv,
		WorkDir:     cwd,
		Home:        home,
	}
	path, _ := lookup.Resolve(fs)

	fcfg, err := config.LoadFetch(fs, path)
	if err != nil {
		return nil, exit.WithUser(err)
	}
	if err := fcfg.Validate(); err != nil {
		return nil, exit.WithUser(err)
	}
	return fcfg, nil
}

// newStore builds the state store at the configured/default path with a flock.
func newStore(fs afero.Fs, fcfg *config.Fetch) (*state.Store, error) {
	path := fcfg.Defaults.StateFile
	if path == "" {
		home, _ := os.UserHomeDir()
		path = state.DefaultPath(os.Getenv, home)
	}
	lock, err := state.FileLock(path)
	if err != nil {
		return nil, exit.WithUser(err)
	}
	return state.NewStore(fs, path, lock), nil
}
