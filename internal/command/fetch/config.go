package fetch

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

// configCommand implements `maintainer fetch config {init,validate}` (§3).
func configCommand(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "manage the fetch configuration",
		Args:  cobra.NoArgs,
	}

	var force bool
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "write a documented fetch.toml template",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := opts.config
			if path == "" {
				path = "fetch.toml"
			}
			fs := afero.NewOsFs()
			if ok, _ := afero.Exists(fs, path); ok && !force {
				return exit.WithUser(fmt.Errorf("%q already exists; pass --force to overwrite", path))
			}
			if err := afero.WriteFile(fs, path, []byte(config.FetchConfigTemplate), 0o644); err != nil {
				return fmt.Errorf("write template: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
			return nil
		},
	}
	initCmd.Flags().BoolVar(&force, "force", false, "overwrite an existing file")

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "parse and structurally check the fetch config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fs := afero.NewOsFs()
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
			if path == "" {
				return exit.WithUser(fmt.Errorf("no fetch config found to validate"))
			}
			fcfg, err := config.LoadFetch(fs, path)
			if err != nil {
				return exit.WithUser(err)
			}
			if err := fcfg.Validate(); err != nil {
				return exit.WithUser(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: %s (%d profiles, %d owners, %d repo rules)\n",
				path, len(fcfg.Profiles), len(fcfg.Owners), len(fcfg.Repos))
			return nil
		},
	}

	command.AddCommand(initCmd, validateCmd)
	return command
}
