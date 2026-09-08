// Package status wires the offline repository inventory command.
package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	statussvc "go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// New returns a command that inspects all local checkouts without credentials.
func New() *cobra.Command {
	var configPath, root, format string
	var owners []string
	var concurrency int
	var timeout time.Duration
	cmd := &cobra.Command{
		SilenceUsage: true,
		Use:          "status", Short: "inspect local repository branches, changes and divergence",
		Long: "Inspect local checkouts using fetch configuration and state. No network\n" +
			"access or fetch is performed; divergence uses cached upstream refs.\n" +
			"In a terminal: arrows/jk select rows, Tab selects a column, s sorts,\n" +
			"Shift+s appends sorting, 0 resets sorting, / filters, q quits.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if format != "auto" && format != "plain" && format != "json" && format != "tui" {
				return exit.WithUser(fmt.Errorf("unknown format %q (want auto|plain|json|tui)", format))
			}
			if concurrency < 0 || (cmd.Flags().Changed("concurrency") && concurrency == 0) || timeout < 0 {
				return exit.WithUser(fmt.Errorf("concurrency must be positive and timeout non-negative"))
			}
			fs := afero.NewOsFs()
			home, _ := os.UserHomeDir()
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			path, _ := (config.FetchConfigLookup{Explicit: configPath, ExplicitSet: cmd.Flags().Changed("config"),
				Getenv: os.Getenv, WorkDir: cwd, Home: home}).Resolve(fs)
			cnf, err := config.LoadFetch(fs, path)
			if err != nil {
				return exit.WithUser(err)
			}
			if cmd.Flags().Changed("root") {
				cnf.Defaults.Root = root
			}
			if err := cnf.Validate(); err != nil {
				return exit.WithUser(err)
			}
			statePath := cnf.Defaults.StateFile
			if statePath == "" {
				statePath = state.DefaultPath(os.Getenv, home)
			}
			// Save uses atomic rename, so reading requires neither a lock file nor
			// creation of the state directory. Status never persists anything.
			st, err := state.NewStore(fs, statePath, nil).Load()
			if err != nil {
				return err
			}
			cap := concurrency
			if cap == 0 {
				cap = cnf.Defaults.Concurrency
			}
			ctx := cmd.Context()
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			in, inputFile := cmd.InOrStdin().(*os.File)
			out, outputFile := cmd.OutOrStdout().(*os.File)
			terminal := inputFile && outputFile && term.IsTerminal(int(in.Fd())) && term.IsTerminal(int(out.Fd()))
			mode := format
			if mode == "auto" {
				mode = "plain"
				if terminal && os.Getenv("TERM") != "dumb" && os.Getenv("TERM") != "" {
					mode = "tui"
				}
			}
			if mode == "tui" && !terminal {
				return exit.WithUser(fmt.Errorf("--format=tui requires terminal input and output"))
			}
			if mode == "tui" {
				fmt.Fprintln(cmd.ErrOrStderr(), "Reading local repository status…")
			}
			rows, err := statussvc.Collect(ctx, cnf, st, home, cwd, owners, cap)
			if err != nil {
				return err
			}
			switch mode {
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				err = enc.Encode(rows)
			case "tui":
				err = statussvc.Interactive(ctx, in, out, rows)
			default:
				err = statussvc.Plain(cmd.OutOrStdout(), rows)
			}
			if err != nil {
				return err
			}
			failed := 0
			for _, row := range rows {
				if row.Error != "" {
					failed++
				}
			}
			if failed > 0 {
				return fmt.Errorf("could not inspect %d checkout(s); see error rows", failed)
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&configPath, "config", "", "fetch config path (--config=\"\" disables discovery)")
	flags.StringVar(&root, "root", "", "override the checkout scan root")
	flags.StringVar(&format, "format", "auto", "output format: auto|plain|json|tui")
	flags.StringSliceVar(&owners, "owner", nil, "limit rows to these owners (repeatable)")
	flags.IntVar(&concurrency, "concurrency", 0, "parallel Git inspection cap (default: fetch config)")
	flags.DurationVar(&timeout, "timeout", 0, "wall-clock budget (0 = unlimited)")
	return cmd
}
