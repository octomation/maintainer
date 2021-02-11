package golang

import "github.com/spf13/cobra"

func New() *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "go",
		Short: "Go manager",
		Long:  "Go manager.",
	}

	command.AddCommand(
		NewVanityCommand(),
	)

	return &command
}
