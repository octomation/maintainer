package github

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	model "go.octolab.org/toolset/maintainer/internal/model/github"
)

// Labels returns a command to work with GitHub labels.
func Labels(git Git, github GitHub) *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "labels",
		Short: "manage repository labels",
		Long:  "Manage repository labels.",
	}

	dump := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "dump",
		Short: "dump repository labels",
		Long:  "Dump repository labels.",
		RunE: func(cmd *cobra.Command, args []string) error {
			remotes, err := git.Remotes()
			if err != nil {
				return fmt.Errorf("cannot specify repository: %w", err)
			}

			remote, found := remotes.GitHub()
			if !found {
				return fmt.Errorf("cannot find GitHub repository")
			}

			labels, err := github.Labels(cmd.Context(), model.GitHub(remote))
			if err != nil {
				return fmt.Errorf("cannot fetch repository labels: %w", err)
			}

			return yaml.NewEncoder(cmd.OutOrStdout()).Encode(labels)
		},
	}

	command.AddCommand(&dump)

	return &command
}
