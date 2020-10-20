package makefile

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

func NewBuildCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "build",
		Short: "build Makefiles",
		Long:  "Build Makefiles.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cancel := context.WithCancel(context.TODO())
			defer cancel()

			makefiles := make(Makefiles, 0, 8)
			for _, name := range args {
				makefiles = append(makefiles, Makefile(name))
			}
			if err := makefiles.CompileTo(distributionDir); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	return &command
}
