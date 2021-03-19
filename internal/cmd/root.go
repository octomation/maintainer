package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/cmd/github"
	"go.octolab.org/toolset/maintainer/internal/cmd/golang"
	"go.octolab.org/toolset/maintainer/internal/cmd/hub"
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

	githubToken := os.Getenv("GITHUB_TOKEN")

	command.AddCommand(
		github.New(githubToken),
		golang.New(),
		hub.New(githubToken),
		makefile.New(),
	)

	return &command
}
