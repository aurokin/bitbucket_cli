package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/output"
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
		Example: "  bb search repos integration --workspace OhBizzle\n" +
			"  bb search repos bb-cli --json name,slug,description",
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
				if len(repos) == 0 {
					_, err := fmt.Fprintf(w, "No repositories found in %s for %q.\n", resolvedWorkspace, args[0])
					return err
				}

				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintln(tw, "name\tslug\tprivate\tproject\tupdated"); err != nil {
					return err
				}
				for _, repo := range repos {
					if _, err := fmt.Fprintf(
						tw,
						"%s\t%s\t%t\t%s\t%s\n",
						output.Truncate(repo.Name, 32),
						output.Truncate(repo.Slug, 24),
						repo.IsPrivate,
						output.Truncate(repo.Project.Key, 12),
						formatPRUpdated(repo.UpdatedOn),
					); err != nil {
						return err
					}
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
		Example: "  bb search prs fixture --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb search prs feature --json id,title,state",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, true)
			if err != nil {
				return err
			}

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
				if len(prs) == 0 {
					_, err := fmt.Fprintf(w, "No pull requests found for %s/%s matching %q.\n", target.Workspace, target.Repo, args[0])
					return err
				}
				return writePRListTable(w, prs)
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
		Example: "  bb search issues fixture --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb search issues bug --json id,title,state",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, true)
			if err != nil {
				return err
			}

			issues, err := client.ListIssues(context.Background(), target.Workspace, target.Repo, bitbucket.ListIssuesOptions{
				Query: buildIssueSearchQuery(args[0]),
				Sort:  "-updated_on",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issues, func(w io.Writer) error {
				if len(issues) == 0 {
					_, err := fmt.Fprintf(w, "No issues found for %s/%s matching %q.\n", target.Workspace, target.Repo, args[0])
					return err
				}

				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintln(tw, "#\ttitle\tstate\treporter\tupdated"); err != nil {
					return err
				}
				for _, issue := range issues {
					if _, err := fmt.Fprintf(
						tw,
						"%d\t%s\t%s\t%s\t%s\n",
						issue.ID,
						output.Truncate(issue.Title, 40),
						output.Truncate(issue.State, 12),
						output.Truncate(issue.Reporter.DisplayName, 16),
						formatPRUpdated(issue.UpdatedOn),
					); err != nil {
						return err
					}
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
