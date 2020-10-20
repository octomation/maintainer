package github

import (
	"context"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"go.octolab.org/toolset/maintainer/internal/cmd/github/labels"
	"go.octolab.org/toolset/maintainer/internal/github"
)

func New(token string) *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "github",
		Short: "GitHub manager",
		Long:  "GitHub manager for all OctoLab's projects.",
	}

	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := oauth2.NewClient(context.TODO(), source)

	command.AddCommand(
		labels.New(github.New(client)),
	)

	return &command
}
