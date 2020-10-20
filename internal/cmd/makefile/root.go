package makefile

import "github.com/spf13/cobra"

func New() *cobra.Command {
	command := cobra.Command{
		Use:   "makefile",
		Short: "makefiles manager",
		Long:  "Makefiles manager.",
		Args:  cobra.NoArgs,
	}

	command.AddCommand(
		NewBuildCommand(),
	)

	return &command
}
