package proxy

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"go.octolab.org/unsafe"
)

func New() []*cobra.Command {
	diff := cobra.Command{
		Use:   "diff",
		Short: "compare files line by line",

		RunE: func(cmd *cobra.Command, args []string) error {
			proxy := exec.Command("diff", args...)
			proxy.Env = os.Environ()
			proxy.Stdin = cmd.InOrStdin()
			proxy.Stdout = cmd.OutOrStdout()
			proxy.Stderr = cmd.ErrOrStderr()

			unsafe.Ignore(proxy.Run())
			return nil
		},

		DisableFlagParsing: true,
	}

	return []*cobra.Command{&diff}
}
