package golang

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"gopkg.in/yaml.v2"

	"go.octolab.org/toolset/maintainer/internal/model/golang"
	"go.octolab.org/toolset/maintainer/internal/model/golang/vanity"
)

func NewVanityCommand() *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "vanity",
		Short: "vanity URL manager",
		Long:  "Vanity URL manager.",
	}

	var (
		file string
		host string
	)
	flags := command.PersistentFlags()
	flags.StringVarP(&file, "file", "f", "modules.yml", "file with modules")
	flags.StringVar(&host, "host", "go.octolab.org", "host for vanity url")

	command.AddCommand(
		&cobra.Command{
			Args:  cobra.MaximumNArgs(1),
			Use:   "build",
			Short: "build vanity URLs",
			Long:  "Build vanity URLs",
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

				var dir string
				if len(args) == 1 {
					dir = args[0]
				} else {
					wd, err := os.Getwd()
					if err != nil {
						return err
					}
					dir = wd
				}

				return vanity.New(host, afero.NewOsFs()).PublishAt(dir, modules)
			},
		},
	)

	return &command
}
