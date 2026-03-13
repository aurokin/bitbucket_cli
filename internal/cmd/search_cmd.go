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
				if len(repos) == 0 {
					if _, err := fmt.Fprintf(w, "No repositories found in %s for %q.\n", resolvedWorkspace, args[0]); err != nil {
						return err
					}
					return writeNextStep(w, searchReposNextStep(resolvedWorkspace, repos))
				}

				if err := writeLabelValue(w, "Workspace", resolvedWorkspace); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Query", args[0]); err != nil {
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
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, searchReposNextStep(resolvedWorkspace, repos))
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
				if len(issues) == 0 {
					if err := writeWarnings(w, target.Warnings); err != nil {
						return err
					}
					if _, err := fmt.Fprintf(w, "No issues found for %s/%s matching %q.\n", target.Workspace, target.Repo, args[0]); err != nil {
						return err
					}
					return writeNextStep(w, searchIssuesNextStep(target.Workspace, target.Repo, issues))
				}

				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				if err := writeWarnings(w, target.Warnings); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Query", args[0]); err != nil {
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
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, searchIssuesNextStep(target.Workspace, target.Repo, issues))
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

func searchReposNextStep(workspace string, repos []bitbucket.Repository) string {
	if len(repos) == 1 {
		return fmt.Sprintf("bb repo view --repo %s/%s", workspace, repos[0].Slug)
	}
	if len(repos) > 1 {
		return fmt.Sprintf("bb repo view --repo %s/<repo>", workspace)
	}
	return fmt.Sprintf("bb repo create %s/<repo>", workspace)
}

func searchPRsNextStep(workspace, repo string, prs []bitbucket.PullRequest) string {
	if len(prs) == 1 {
		return fmt.Sprintf("bb pr view %d --repo %s/%s", prs[0].ID, workspace, repo)
	}
	if len(prs) > 1 {
		return fmt.Sprintf("bb pr view <id> --repo %s/%s", workspace, repo)
	}
	return fmt.Sprintf("bb pr list --repo %s/%s", workspace, repo)
}

func writeSearchPRSummary(w io.Writer, target resolvedRepoTarget, query string, prs []bitbucket.PullRequest) error {
	if len(prs) == 0 {
		if err := writeWarnings(w, target.Warnings); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "No pull requests found for %s/%s matching %q.\n", target.Workspace, target.Repo, query); err != nil {
			return err
		}
		return writeNextStep(w, searchPRsNextStep(target.Workspace, target.Repo, prs))
	}
	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", query); err != nil {
		return err
	}
	if err := writePRListTable(w, prs); err != nil {
		return err
	}
	return writeNextStep(w, searchPRsNextStep(target.Workspace, target.Repo, prs))
}

func searchIssuesNextStep(workspace, repo string, issues []bitbucket.Issue) string {
	if len(issues) == 1 {
		return fmt.Sprintf("bb issue view %d --repo %s/%s", issues[0].ID, workspace, repo)
	}
	if len(issues) > 1 {
		return fmt.Sprintf("bb issue view <id> --repo %s/%s", workspace, repo)
	}
	return fmt.Sprintf("bb issue list --repo %s/%s", workspace, repo)
}
