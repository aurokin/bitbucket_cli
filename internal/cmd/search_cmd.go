package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search repositories, pull requests, and issues",
		Long:  "Search Bitbucket Cloud repositories, pull requests, and issues using Bitbucket query filters behind the scenes.",
	}

	searchCmd.AddCommand(
		newSearchReposCmd(),
		newSearchPRsCmd(),
		newSearchIssuesCmd(),
	)

	return searchCmd
}

func newSearchReposCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var limit int

	cmd := &cobra.Command{
		Use:   "repos <query>",
		Short: "Search repositories in a workspace",
		Example: "  bb search repos integration --workspace workspace-slug\n" +
			"  bb search repos bb-cli --workspace workspace-slug --json name,slug,description",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			_, client, err := resolveAuthenticatedClient(host)
			if err != nil {
				return err
			}

			resolvedWorkspace, err := resolveWorkspaceForCreate(context.Background(), client, workspace)
			if err != nil {
				return err
			}

			repos, err := client.ListRepositories(context.Background(), resolvedWorkspace, bitbucket.ListRepositoriesOptions{
				Query: buildRepositorySearchQuery(args[0]),
				Sort:  "-updated_on",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, repos, func(w io.Writer) error {
				return writeSearchRepoSummary(w, resolvedWorkspace, args[0], repos)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to search; inferred when only one workspace is available")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of repositories to return")

	return cmd
}

func newSearchPRsCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "prs <query>",
		Short: "Search pull requests in one repository",
		Example: "  bb search prs fixture --repo workspace-slug/repo-slug\n" +
			"  bb search prs feature --repo workspace-slug/repo-slug --json id,title,state",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			prs, err := client.ListPullRequests(context.Background(), target.Workspace, target.Repo, bitbucket.ListPullRequestsOptions{
				State: "ALL",
				Query: buildPullRequestSearchQuery(args[0]),
				Sort:  "-updated_on",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, prs, func(w io.Writer) error {
				return writeSearchPRSummary(w, target, args[0], prs)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pull requests to return")

	return cmd
}

func newSearchIssuesCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "issues <query>",
		Short: "Search issues in one repository",
		Example: "  bb search issues fixture --repo workspace-slug/issues-repo-slug\n" +
			"  bb search issues bug --repo workspace-slug/issues-repo-slug --json id,title,state",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			issues, err := client.ListIssues(context.Background(), target.Workspace, target.Repo, bitbucket.ListIssuesOptions{
				Query: buildIssueSearchQuery(args[0]),
				Sort:  "-updated_on",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issues, func(w io.Writer) error {
				return writeSearchIssueSummary(w, target, args[0], issues)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of issues to return")

	return cmd
}

func buildRepositorySearchQuery(query string) string {
	escaped := quoteBitbucketQueryString(query)
	return fmt.Sprintf(`name ~ "%[1]s" OR slug ~ "%[1]s" OR description ~ "%[1]s"`, escaped)
}

func buildPullRequestSearchQuery(query string) string {
	escaped := quoteBitbucketQueryString(query)
	return fmt.Sprintf(`title ~ "%[1]s" OR description ~ "%[1]s"`, escaped)
}

func buildIssueSearchQuery(query string) string {
	escaped := quoteBitbucketQueryString(query)
	return fmt.Sprintf(`title ~ "%[1]s" OR content.raw ~ "%[1]s"`, escaped)
}

func quoteBitbucketQueryString(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return replacer.Replace(strings.TrimSpace(value))
}
