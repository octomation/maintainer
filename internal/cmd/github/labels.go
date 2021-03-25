package github

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"gopkg.in/yaml.v2"

	model "go.octolab.org/toolset/maintainer/internal/model/github"
	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
)

// Labels returns a command to work with GitHub labels.
func Labels(git Git, github GitHub) *cobra.Command {
	var (
		input  string
		output string
	)

	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "labels",
		Short: "manage repository labels",
		Long:  "Manage repository labels.",
	}

	{
		patch := cobra.Command{
			Args:  cobra.ExactArgs(1),
			Use:   "patch",
			Short: "patch repository labels",
			Long:  "Patch repository labels.",
			RunE: func(cmd *cobra.Command, args []string) error {
				var current model.LabelSet

				in := cmd.InOrStdin()
				if input != "" {
					f, err := os.Open(input)
					if err != nil {
						return err
					}
					defer safe.Close(f, unsafe.Ignore)
					in = f
				}

				if err := yaml.NewDecoder(in).Decode(&current); err != nil {
					return err
				}

				out := cmd.OutOrStdout()
				if output != "" {
					f, err := os.Open(output)
					if err != nil {
						return err
					}
					defer safe.Close(f, unsafe.Ignore)
					out = f
				}

				assert.True(func() bool { return len(args) == 1 })
				patched, err := github.PatchLabels(cmd.Context(), current, args[0])
				if err != nil {
					return err
				}

				sort.Sort(model.SortLabelsByName(patched))
				return yaml.NewEncoder(out).Encode(patched)
			},
		}
		flags := patch.Flags()
		flags.StringVar(&input, "input", "", "input file with labels [stdin]")
		flags.StringVar(&output, "output", "", "output file to store labels [stdout]")
		command.AddCommand(&patch)
	}

	{
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

				sort.Sort(model.SortLabelsByName(labels))
				return yaml.NewEncoder(cmd.OutOrStdout()).Encode(labels)
			},
		}
		command.AddCommand(&pull)
	}

	{
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
		command.AddCommand(&push)
	}

	{
		sync := cobra.Command{
			Args:  cobra.NoArgs,
			Use:   "sync",
			Short: "sync repository labels",
			Long:  "Sync repository labels.",
			RunE: func(cmd *cobra.Command, args []string) error {
				push, args, err := cmd.Parent().Find([]string{"push"})
				if err != nil {
					return err
				}
				if err := push.RunE(push, args); err != nil {
					return err
				}

				pull, args, err := cmd.Parent().Find([]string{"pull"})
				if err != nil {
					return err
				}
				return pull.RunE(cmd, args)
			},
		}
		command.AddCommand(&sync)
	}

	return &command
}
