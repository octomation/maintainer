package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/cmd/github"
	"go.octolab.org/toolset/maintainer/internal/cmd/makefile"
)

// New returns the new root command.
func New() *cobra.Command {
	command := cobra.Command{
		Args: cobra.NoArgs,

		Use:   "maintainer",
		Short: "maintainer is an indispensable assistant to Open Source contribution",
		Long:  "Maintainer is an indispensable assistant to Open Source contribution.",

		SilenceErrors: false,
		SilenceUsage:  true,
	}

	command.AddCommand(
		github.New(os.Getenv("GITHUB_TOKEN")),
		makefile.New(),
	)

	return &command
}
