package makefile

import "github.com/spf13/cobra"

func New() *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "makefile",
		Short: "makefiles manager",
		Long:  "Makefiles manager.",
	}

	command.AddCommand(
		NewBuildCommand(),
	)

	return &command
}
