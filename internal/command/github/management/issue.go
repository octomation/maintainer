package management

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v80/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"go.octolab.org/toolset/maintainer/internal/config"
	gitprovider "go.octolab.org/toolset/maintainer/internal/service/git/provider"
)

// Issue returns a command to fetch all GitHub issues.
func Issue(cmd *cobra.Command, cnf *config.Tool) *cobra.Command {
	cmd.Short = "Fetch all GitHub issues"
	cmd.Long = "Fetches all issues from the current GitHub repository."

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Get the current repository
		repo, err := gitprovider.Current()
		if err != nil {
			return fmt.Errorf("failed to open git repository: %w", err)
		}

		// Get the remote URL
		remotes, err := repo.Remotes()
		if err != nil {
			return fmt.Errorf("failed to get remotes: %w", err)
		}

		if len(remotes) == 0 {
			return fmt.Errorf("no remotes found")
		}

		// Find GitHub remote
		var owner, repoName string
		for _, remote := range remotes {
			config := remote.Config()
			if len(config.URLs) == 0 {
				continue
			}

			remoteURL := config.URLs[0]
			if !strings.Contains(remoteURL, "github.com") {
				continue
			}

			// Parse the GitHub URL to extract owner and repo
			parsedURL, err := url.Parse(remoteURL)
			if err != nil {
				continue
			}

			path := strings.TrimPrefix(parsedURL.Path, "/")
			path = strings.TrimSuffix(path, ".git")
			parts := strings.Split(path, "/")
			if len(parts) != 2 {
				continue
			}

			owner = parts[0]
			repoName = parts[1]
			break
		}

		if owner == "" || repoName == "" {
			return fmt.Errorf("could not determine GitHub owner and repository from remotes")
		}

		// Create GitHub client
		var httpClient *http.Client
		token := string(cnf.GitHub.Token)
		if token != "" {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			httpClient = oauth2.NewClient(ctx, ts)
		}
		client := github.NewClient(httpClient)

		// Fetch all issues
		opt := &github.IssueListByRepoOptions{
			State: "all",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		var allIssues []*github.Issue
		for {
			issues, resp, err := client.Issues.ListByRepo(ctx, owner, repoName, opt)
			if err != nil {
				return fmt.Errorf("failed to fetch issues: %w", err)
			}
			allIssues = append(allIssues, issues...)
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}

		// Print issues
		fmt.Printf("Found %d issues in %s/%s:\n\n", len(allIssues), owner, repoName)
		for _, issue := range allIssues {
			fmt.Printf("#%-5d %s [%s]\n", issue.GetNumber(), issue.GetTitle(), issue.GetState())
		}

		return nil
	}

	return cmd
}
