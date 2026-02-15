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

// gitHub is what a run asks GitHub: the listings, a check by id and a lookup
// by name.
type gitHub interface {
	githubsvc.Discoverer
	githubsvc.Confirmer
	githubsvc.NameResolver
}

// newDiscoverer builds the GitHub side of a run; tests replace it.
var newDiscoverer = func() gitHub { return githubsvc.NewRESTDiscoverer(githubsvc.DefaultClientFactory) }

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
	token       string // the retired --token; never read
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
			"actions (clone, fetch, move, update-remote, adopt). Existing checkouts\n" +
			"and local work are retained.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, opts)
		},
	}

	set := command.PersistentFlags()
	set.StringVar(&opts.config, "config", "", "fetch config path (--config=\"\" disables discovery)")
	set.StringSliceVar(&opts.profiles, "profile", nil, "limit this run to these profiles: repositories they list or track; others are left alone (repeatable)")
	set.StringSliceVar(&opts.owners, "owner", nil, "limit this run to repositories of these owners; others are left alone (repeatable)")
	set.StringVar(&opts.format, "format", fetchsvc.FormatHuman, "output format: human|json")
	set.IntVar(&opts.concurrency, "concurrency", 0, "parallel discovery/clone cap (0 = from config)")
	set.DurationVar(&opts.timeout, "timeout", 0, "wall-clock budget for the whole command")
	set.CountVarP(&opts.verbose, "verbose", "v", "increase verbosity: -v each action, -vv discovery, confirmations and saves, -vvv each Git operation")
	set.BoolVarP(&opts.quiet, "quiet", "q", false, "suppress everything below error on stderr")
	set.StringVar(&opts.token, "token", "", "retired: set GITHUB_TOKEN or token_env of a profile")
	_ = set.MarkHidden("token")

	command.Flags().BoolVar(&opts.apply, "apply", false, "execute the plan (default: plan only)")

	command.AddCommand(
		configCommand(opts),
		stateCommand(opts),
	)
	return command
}

func run(cmd *cobra.Command, opts *options) error {
	if opts.format != fetchsvc.FormatHuman && opts.format != fetchsvc.FormatJSON {
		return exit.WithUser(fmt.Errorf("unknown format %q (want human|json)", opts.format))
	}
	if opts.concurrency < 0 || opts.timeout < 0 {
		return exit.WithUser(fmt.Errorf("concurrency and timeout must not be negative"))
	}
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

	discoverer := newDiscoverer()
	reporter := fetchsvc.NewReporter(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts.format, opts.verbose, opts.quiet)

	svc, err := fetchsvc.NewService(fcfg, profiles, home, cwd, concurrency, fetchsvc.Deps{
		Store:      store,
		Discoverer: discoverer,
		Confirmer:  discoverer,
		Resolver:   discoverer,
		GitSync:    gitsvc.NewSync(),
		Reporter:   reporter,
		Selection:  fetchsvc.Selection{Owners: opts.owners, Profiles: opts.profiles},
	})
	if err != nil {
		return err
	}
	return svc.Run(ctx, opts.apply)
}

// configPath resolves the fetch config the same way for every fetch command:
// an explicit --config="" disables discovery, while an omitted flag keeps it.
// The working directory is never read: a fetch config there gives a notice,
// or a usage error when no other config is found. An empty result means no
// config file.
func configPath(cmd *cobra.Command, fs afero.Fs, opts *options) (string, error) {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	path, notice, err := config.FetchConfigLookup{
		Explicit:    opts.config,
		ExplicitSet: cmd.Flags().Changed("config"),
		Getenv:      os.Getenv,
		WorkDir:     cwd,
		Home:        home,
	}.Find(fs)
	if err != nil {
		return "", exit.WithUser(err)
	}
	if notice != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "notice: %s\n", notice)
	}
	return path, nil
}

// loadConfig resolves and loads the fetch config, then validates it (§4).
func loadConfig(cmd *cobra.Command, fs afero.Fs, opts *options) (*config.Fetch, error) {
	path, err := configPath(cmd, fs, opts)
	if err != nil {
		return nil, err
	}
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
