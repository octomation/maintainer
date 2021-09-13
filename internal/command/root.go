package command

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github"
	"go.octolab.org/toolset/maintainer/internal/command/golang"
	"go.octolab.org/toolset/maintainer/internal/command/makefile"
	"go.octolab.org/toolset/maintainer/internal/config"
)

// New returns the new root command.
func New() *cobra.Command {
	var cnf config.Tool

	command := cobra.Command{
		Args: cobra.NoArgs,

		Use:   "maintainer",
		Short: "assists with Open Source contribution",
		Long:  "An indispensable assistant for Open Source contribution.",

		PersistentPreRunE: func(*cobra.Command, []string) error {
			// TODO:feature home dir and specific config
			return cnf.Load(afero.NewOsFs())
		},

		SilenceErrors: false,
		SilenceUsage:  true,
	}

	command.AddCommand(
		github.New(&cnf),
		golang.New(),
		makefile.New(),
	)

	return &command
}
