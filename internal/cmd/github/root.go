package github

import (
	"context"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/git/provider"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// New returns a command to work with GitHub.
func New(token string) *cobra.Command {
	var (
		source = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.TODO(), source)

		remote string
	)

	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "github",
		Short: "GitHub manager",
		Long:  "GitHub manager for all OctoLab's projects.",
	}

	set := command.PersistentFlags()
	set.StringVar(&remote, "remote", "", "a connection to a remote repository")

	gitService := git.New(
		provider.
			FallbackTo(&remote). // TODO:naive delay init git by sync.Once
			Apply(provider.Default()),
	)
	githubService := github.New(client)

	command.AddCommand(
		Contribution(githubService),
		Labels(gitService, githubService),
	)

	return &command
}
