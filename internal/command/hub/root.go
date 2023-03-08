package hub

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/go-github/v84/github"
	"github.com/spf13/cobra"
	"go.octolab.org/async"
	"go.octolab.org/safe"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/http"
)

func New(cnf *config.Tool) *cobra.Command {
	command := cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "hub",
		Short: "fetch data from GitHub and Trello to manage it",

		RunE: func(cmd *cobra.Command, args []string) error {
			client := github.NewClient(http.TokenSourcedClient(cmd.Context(), cnf.Token))

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
