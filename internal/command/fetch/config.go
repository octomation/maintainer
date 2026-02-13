package fetch

import (
	"fmt"
	"os"
	"path/filepath"

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
		Short: "write a documented fetch.toml template to the user folder or --config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := opts.config
			if path == "" {
				home, _ := os.UserHomeDir()
				dir := config.FetchConfigLookup{Getenv: os.Getenv, Home: home}.UserDir()
				if dir == "" {
					return exit.WithUser(fmt.Errorf("no user folder for the config: set HOME or XDG_CONFIG_HOME, or pass --config"))
				}
				path = filepath.Join(dir, "fetch.toml")
			}
			fs := afero.NewOsFs()
			if ok, _ := afero.Exists(fs, path); ok && !force {
				return exit.WithUser(fmt.Errorf("%q already exists; pass --force to overwrite", path))
			}
			if err := fs.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return fmt.Errorf("create the config folder: %w", err)
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
			path, err := configPath(cmd, fs, opts)
			if err != nil {
				return err
			}
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
			// A run reports a broken per-repo path for its repository only.
			if err := fcfg.ValidateRepoPaths(); err != nil {
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
