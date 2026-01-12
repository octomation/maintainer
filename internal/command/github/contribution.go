package github

import (
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
)

// Contribution returns a command set to work with the contribution calendar,
// each command describes its arguments and examples in its --help.
func Contribution(cnf *config.Tool) *cobra.Command {
	cmd := cobra.Command{
		Use:   "contribution",
		Short: "Work with the contribution calendar",
		Long: `Allows to look up the contribution calendar of the token owner,
to find a moment to contribute, and to keep and compare snapshots of it.

The calendar is read through the GitHub GraphQL API, which requires a token
even for a public profile: set GITHUB_TOKEN or pass --token. Only diff of two
snapshot files works without it.

A date of lookup and suggest that is empty or git is the author date of HEAD
in the current repository, or in the one GIT_DIR points to; outside
a repository, or in one without commits, it is now.`,
		Example: `  maintainer github contribution lookup 2022/+10
  maintainer github contribution suggest --short
  maintainer github contribution snapshot 2013 > snap.json
  maintainer github contribution diff snap.json 2013`,
	}

	cmd.AddCommand(contribution.Diff(&cobra.Command{Use: "diff"}, cnf))
	cmd.AddCommand(contribution.Histogram(&cobra.Command{Use: "histogram"}, cnf))
	cmd.AddCommand(contribution.Lookup(&cobra.Command{Use: "lookup"}, cnf))
	cmd.AddCommand(contribution.Snapshot(&cobra.Command{Use: "snapshot"}, cnf))
	cmd.AddCommand(contribution.Suggest(&cobra.Command{Use: "suggest"}, cnf))

	return &cmd
}
