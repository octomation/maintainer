package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

// EnforceUsage makes every usage error of the command tree under root a user
// input error, which exits with exit.User: a flag that does not parse, an
// argument that a command does not accept, and an unknown subcommand of a
// group, which cobra would answer with the help and success.
//
// Call it once the tree is complete and its output is set: it also adds the
// default completion command, which writes to the output of that moment.
func EnforceUsage(root *cobra.Command) {
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return exit.WithUser(err) })
	root.InitDefaultCompletionCmd()
	enforceUsage(root)
}

func enforceUsage(cmd *cobra.Command) {
	switch args := cmd.Args; {
	case !cmd.Runnable():
		// cobra answers a command that cannot run, a group, with its help
		// before it validates the arguments. Such a command now runs only to
		// have them validated: subcommand ends every run before the hooks.
		cmd.Args = subcommand
		cmd.RunE = func(cmd *cobra.Command, _ []string) error { return cmd.Help() }
		if cmd.SuggestionsMinimumDistance <= 0 {
			cmd.SuggestionsMinimumDistance = 2
		}
	case args != nil:
		cmd.Args = func(cmd *cobra.Command, in []string) error { return exit.WithUser(args(cmd, in)) }
	}
	for _, sub := range cmd.Commands() {
		enforceUsage(sub)
	}
}

// subcommand rejects an argument of a group as an unknown subcommand and
// suggests the closest ones, as cobra does for the root. Without arguments it
// asks cobra for the help of the group, so the group does not run its hooks.
func subcommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return pflag.ErrHelp
	}
	msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if !cmd.DisableSuggestions {
		if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
			msg += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
		}
	}
	return exit.WithUser(errors.New(msg))
}
