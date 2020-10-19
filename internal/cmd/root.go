package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/cmd/github"
)

// New returns the new root command.
func New() *cobra.Command {
	command := cobra.Command{
		Use:   "maintainer",
		Short: "maintainer is an indispensable assistant to Open Source contribution",
		Long:  "Maintainer is an indispensable assistant to Open Source contribution.",

		Args: cobra.NoArgs,

		SilenceErrors: false,
		SilenceUsage:  true,
	}

	command.AddCommand(
		github.New(os.Getenv("GITHUB_TOKEN")),
	)

	return &command
}
