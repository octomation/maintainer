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

	patch := cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   "patch",
		Short: "patch repository labels",
		Long:  "Patch repository labels.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	pull := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "pull",
		Short: "pull repository labels",
		Long:  "Pull repository labels.",
		RunE: func(cmd *cobra.Command, args []string) error {
			remotes, err := git.Remotes()
			if err != nil {
				return fmt.Errorf("cannot specify repository: %w", err)
			}

			remote, found := remotes.GitHub()
			if !found {
				return fmt.Errorf("cannot find GitHub repository")
			}
			src := model.Remote(remote)

			labels, err := github.Labels(cmd.Context(), src)
			if err != nil {
				return fmt.Errorf("cannot fetch repository labels: %w", err)
			}

			return yaml.NewEncoder(cmd.OutOrStdout()).Encode(labels)
		},
	}

	push := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "push",
		Short: "push repository labels",
		Long:  "Push repository labels.",
		RunE: func(cmd *cobra.Command, args []string) error {
			remotes, err := git.Remotes()
			if err != nil {
				return fmt.Errorf("cannot specify repository: %w", err)
			}

			remote, found := remotes.GitHub()
			if !found {
				return fmt.Errorf("cannot find GitHub repository")
			}

			_ = remote
			return nil
		},
	}

	command.AddCommand(&pull, &patch, &push)

	return &command
}
