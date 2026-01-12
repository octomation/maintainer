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

	set := command.PersistentFlags()
	set.String("token", "", "personal access token")

	fn.Must(
		func() error {
			return cnf.Bind(func(v *viper.Viper) error {
				return v.BindEnv("GITHUB_TOKEN")
			})
		},
		func() error {
			return cnf.Bind(func(v *viper.Viper) error {
				return v.BindPFlag("github_token", set.Lookup("token"))
			})
		},
	)

	command.AddCommand(
		Contribution(cnf),
	)

	return &command
}
