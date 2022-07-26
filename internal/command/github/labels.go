package github

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"gopkg.in/yaml.v2"

	"go.octolab.org/toolset/maintainer/internal/config"
	model "go.octolab.org/toolset/maintainer/internal/model/github"
	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
	"go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/git/provider"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// Labels returns a command to work with GitHub labels.
func Labels(cnf *config.Tool) *cobra.Command {
	const (
		rootCommand  = "labels"
		patchCommand = "patch"
		pullCommand  = "pull"
		pushCommand  = "push"
		syncCommand  = "sync"
	)

	var (
		input  string
		output string
	)

	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   rootCommand,
		Short: "manage repository labels",
		Long:  "Manage repository labels.",
	}

	{
		patch := cobra.Command{
			Args:  cobra.ExactArgs(1),
			Use:   patchCommand,
			Short: "patch repository labels",
			Long:  "Patch repository labels.",
			RunE: func(cmd *cobra.Command, args []string) error {
				service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))

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
				patched, err := service.PatchLabels(cmd.Context(), current, args[0])
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
			Use:   pullCommand,
			Short: "pull repository labels",
			Long:  "Pull repository labels.",
			RunE: func(cmd *cobra.Command, args []string) error {
				service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
				source := git.New(provider.FallbackTo(cnf.Remote).Apply(provider.Current()))

				remotes, err := source.Remotes()
				if err != nil {
					return fmt.Errorf("cannot specify repository: %w", err)
				}

				remote, found := remotes.GitHub()
				if !found {
					return fmt.Errorf("cannot find GitHub repository")
				}
				src := model.Remote(remote)

				labels, err := service.Labels(cmd.Context(), src)
				if err != nil {
					return fmt.Errorf("cannot fetch repository labels: %w", err)
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

				sort.Sort(model.SortLabelsByName(labels))
				return yaml.NewEncoder(out).Encode(labels)
			},
		}

		flags := pull.Flags()
		flags.StringVar(&output, "output", "", "output file to store labels [stdout]")

		command.AddCommand(&pull)
	}

	{
		push := cobra.Command{
			Args:  cobra.NoArgs,
			Use:   pushCommand,
			Short: "push repository labels",
			Long:  "Push repository labels.",
			RunE: func(cmd *cobra.Command, args []string) error {
				service := github.New(http.TokenSourcedClient(cmd.Context(), cnf.Token))
				source := git.New(provider.FallbackTo(cnf.Remote).Apply(provider.Current()))

				var patched model.LabelSet

				in := cmd.InOrStdin()
				if input != "" {
					f, err := os.Open(input)
					if err != nil {
						return err
					}
					defer safe.Close(f, unsafe.Ignore)
					in = f
				}

				if err := yaml.NewDecoder(in).Decode(&patched); err != nil {
					return err
				}

				remotes, err := source.Remotes()
				if err != nil {
					return fmt.Errorf("cannot specify repository: %w", err)
				}

				remote, found := remotes.GitHub()
				if !found {
					return fmt.Errorf("cannot find GitHub repository")
				}
				src := model.Remote(remote)

				return service.UpdateLabels(cmd.Context(), src, patched)
			},
		}

		flags := push.Flags()
		flags.StringVar(&input, "input", "", "input file with labels [stdin]")

		command.AddCommand(&push)
	}

	{
		sync := cobra.Command{
			Args:  cobra.NoArgs,
			Use:   syncCommand,
			Short: "sync repository labels",
			Long:  "Sync repository labels.",
			RunE: func(cmd *cobra.Command, _ []string) error {
				push, args, err := cmd.Parent().Find([]string{pushCommand})
				if err != nil {
					return err
				}
				if err := push.RunE(cmd, args); err != nil {
					return err
				}

				pull, args, err := cmd.Parent().Find([]string{pullCommand})
				if err != nil {
					return err
				}
				return pull.RunE(cmd, args)
			},
		}

		flags := sync.Flags()
		flags.StringVar(&input, "input", "", "input file with patched labels [stdin]")
		flags.StringVar(&output, "output", "", "output file to store labels [stdout]")

		command.AddCommand(&sync)
	}

	return &command
}
