package command

import (
	"errors"
	"strings"

	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

// ErrTokenFlag refuses the retired --token flag. The flag stays known to the
// parser, hidden, so that the help stays clean and no path reports it as an
// unknown flag; this scan refuses it first, with the replacement.
var ErrTokenFlag = errors.New("--token is removed: other accounts of this machine " +
	"can read command lines; set GITHUB_TOKEN in the environment instead")

// RejectToken returns a usage error if any argument is --token or starts
// with --token=, with any value, also an empty one. Run it on the raw
// arguments before the root command executes: cobra answers --help and
// a group without a subcommand before any hook. No argument is exempt, not
// the value of another flag and not one after --, since -- itself can be
// the value of a flag. The value is neither used nor printed.
func RejectToken(args []string) error {
	for _, arg := range args {
		if arg == "--token" || strings.HasPrefix(arg, "--token=") {
			return exit.WithUser(ErrTokenFlag)
		}
	}
	return nil
}
