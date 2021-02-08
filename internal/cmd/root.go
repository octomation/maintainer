package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"go.octolab.org/unsafe"

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

	diff := cobra.Command{
		Use:   "diff",
		Short: "compare files line by line",
		Long:  "Compare files line by line.",

		RunE: func(cmd *cobra.Command, args []string) error {
			proxy := exec.Command("diff", args...)
			proxy.Env = os.Environ()
			proxy.Stdin = cmd.InOrStdin()
			proxy.Stdout = cmd.OutOrStdout()
			proxy.Stderr = cmd.ErrOrStderr()

			unsafe.Ignore(proxy.Run())
			return nil
		},

		DisableFlagParsing: true,
	}

	githubToken := os.Getenv("GITHUB_TOKEN")

	command.AddCommand(
		&diff,
		github.New(githubToken),
		golang.New(),
		hub.New(githubToken),
		makefile.New(),
	)

	return &command
}
