package hub

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/google/go-github/v44/github"
	"github.com/spf13/cobra"
	"go.octolab.org/async"
	"go.octolab.org/safe"
	"golang.org/x/oauth2"
)

func New(token string) *cobra.Command {
	var (
		client = github.NewClient(
			oauth2.NewClient(
				context.TODO(),
				oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
			),
		)
	)

	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "hub",
		Short: "fetch data from Airtable, GitHub, and Trello to manage it",
		Long:  "Fetch data from Airtable, GitHub, and Trello to manage it.",

		RunE: func(cmd *cobra.Command, args []string) error {
			opt := new(github.IssueListOptions)
			opt.ListOptions.PerPage = 100

			f, err := os.Create("bin/github.issues.json")
			if err != nil {
				return err
			}
			defer safe.Close(f, func(err error) { cmd.PrintErrln(err) })

			var mx sync.Mutex
			result := make([]*github.Issue, 0, 1024)

			issues, resp, err := client.Issues.List(cmd.Context(), true, opt)
			if err != nil {
				return err
			}
			result = append(result, issues...)

			if resp.NextPage != 0 {
				job := new(async.Job)
				for i := 2; i <= resp.LastPage; i++ {
					job.Do(func() error {
						issues, _, err := client.Issues.List(cmd.Context(), true, opt)
						if err != nil {
							return err
						}
						mx.Lock()
						result = append(result, issues...)
						mx.Unlock()
						return nil
					}, func(err error) {
						cmd.PrintErrln(err)
					})
				}
				job.Wait()
			}

			return json.NewEncoder(f).Encode(result)
		},
	}

	return &command
}
