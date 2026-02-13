package github

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.octolab.org/fn"

	"go.octolab.org/toolset/maintainer/internal/config"
)

// New returns a command set to work with GitHub.
func New(cnf *config.Tool) *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "github",
		Short: "GitHub manager",
		Long:  "Allows to work with GitHub repositories and automate routine.",
	}

	// The retired --token stays known to the parser, so that the root
	// command refuses it as a usage error; its value is never read.
	set := command.PersistentFlags()
	set.String("token", "", "retired: set GITHUB_TOKEN")
	fn.Must(
		func() error { return set.MarkHidden("token") },
		func() error {
			return cnf.Bind(func(v *viper.Viper) error {
				return v.BindEnv("GITHUB_TOKEN")
			})
		},
	)

	command.AddCommand(
		Contribution(cnf),
	)

	return &command
}
