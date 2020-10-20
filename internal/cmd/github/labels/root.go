package labels

import "github.com/spf13/cobra"

func New(provider Provider) *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "labels",
		Short: "manage labels",
		Long:  "Manage labels.",
	}

	command.AddCommand(
		NewCompareCommand(provider),
		NewListCommand(provider),
	)

	return &command
}
