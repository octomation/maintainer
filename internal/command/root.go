package command

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	cli "go.octolab.org/toolkit/cli/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/fetch"
	"go.octolab.org/toolset/maintainer/internal/command/github"
	"go.octolab.org/toolset/maintainer/internal/command/golang"
	"go.octolab.org/toolset/maintainer/internal/command/makefile"
	"go.octolab.org/toolset/maintainer/internal/command/status"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/buildinfo"
	"go.octolab.org/toolset/maintainer/skills"
)

// New returns the new root command with the whole command tree. Call
// buildinfo.Set first: the version and skill commands report what it holds.
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
		fetch.New(&cnf),
		status.New(),
		github.New(&cnf),
		golang.New(),
		makefile.New(),
		skillCommand(skills.Maintainer()),
		cli.NewVersionCommand(buildinfo.Version, buildinfo.Date, buildinfo.Commit, config.Features...),
	)

	return &command
}
