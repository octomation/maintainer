package github

import (
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/management"
	"go.octolab.org/toolset/maintainer/internal/config"
)

func Management(cnf *config.Tool) *cobra.Command {
	cmd := cobra.Command{
		Use:   "management",
		Short: "GitHub management tools",
		Long:  "Allows to manage GitHub repositories and issues.",
	}

	// $ maintainer github management issue
	//
	// Fetches all issues from the current GitHub repository.
	//
	cmd.AddCommand(management.Issue(&cobra.Command{Use: "issue"}, cnf))

	return &cmd
}
