package golang

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"gopkg.in/yaml.v2"

	"go.octolab.org/toolset/maintainer/internal/entity/golang"
	"go.octolab.org/toolset/maintainer/internal/entity/golang/vanity"
)

func NewVanityCommand() *cobra.Command {
	const (
		file = "modules.yml"
		host = "go.octolab.org"
	)

	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "vanity",
		Short: "vanity URL manager",
		Long:  "Vanity URL manager.",
	}

	command.AddCommand(
		&cobra.Command{
			Args:  cobra.NoArgs,
			Use:   "dump",
			Short: "dump vanity URLs",
			Long:  "Dump vanity URLs",
			RunE: func(cmd *cobra.Command, args []string) error {
				file, err := os.Open(file)
				if err != nil {
					return err
				}
				defer safe.Close(file, unsafe.Ignore)

				var modules []golang.Module
				if err := yaml.NewDecoder(file).Decode(&modules); err != nil {
					return err
				}

				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				return vanity.New(host, afero.NewOsFs()).PublishAt(wd, modules)
			},
		},
	)

	return &command
}
