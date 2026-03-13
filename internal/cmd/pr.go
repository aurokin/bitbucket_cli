package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newPRCmd() *cobra.Command {
	prCmd := &cobra.Command{
		Use:     "pr",
		Aliases: []string{"pull-request", "pullrequest"},
		Short:   "Work with pull requests",
		Long:    "List, inspect, create, check out, merge, and summarize Bitbucket pull requests.",
	}

	prCmd.AddCommand(
		newPRListCmd(),
		newPRStatusCmd(),
		newPRDiffCmd(),
		newPRCommentCmd(),
		newPRCloseCmd(),
		newPRViewCmd(),
		newPRCreateCmd(),
		newPRCheckoutCmd(),
		newPRMergeCmd(),
	)

	return prCmd
}

func newPRCloseCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "close <id-or-url>",
		Short: "Close a pull request without merging it",
		Long:  "Close a pull request without merging it. In Bitbucket Cloud this maps to declining the pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.",
		Example: "  bb pr close 1\n" +
			"  bb pr close 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json\n" +
			"  bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}
			prTarget := resolved.Target
			client := resolved.Client

			pr, err := client.GetPullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}
			if pr.State != "OPEN" {
				return fmt.Errorf("pull request #%d is %s; only OPEN pull requests can be closed", pr.ID, pr.State)
			}

			closedPR, err := client.DeclinePullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, closedPR, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo); err != nil {
					return err
				}
				if err := writePullRequestSummaryTable(w, closedPR, pullRequestSummaryOptions{}); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", closedPR.ID, prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newPRCommentCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var body string
	var bodyFile string

	cmd := &cobra.Command{
		Use:   "comment <id-or-url>",
		Short: "Add a comment to a pull request",
		Long:  "Add a comment to a pull request using --body, --body-file, or --body-file - for stdin. This first pass is intentionally deterministic for agent and script usage. Use the comment subcommands to view, edit, delete, resolve, or reopen specific pull request comments.",
		Example: "  bb pr comment 1 --body 'Looks good'\n" +
			"  bb pr comment 1 --repo workspace-slug/repo-slug --body-file comment.md\n" +
			"  printf 'Ship it\\n' | bb pr comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --body-file - --json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			commentBody, err := resolveCommentBody(cmd.InOrStdin(), body, bodyFile)
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}
			prTarget := resolved.Target
			client := resolved.Client

			comment, err := client.CreatePullRequestComment(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID, commentBody)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, comment, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", prTarget.ID); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "Comment:\t%d\n", comment.ID); err != nil {
					return err
				}
				if comment.User.DisplayName != "" {
					if _, err := fmt.Fprintf(tw, "Author:\t%s\n", comment.User.DisplayName); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintf(tw, "Body:\t%s\n", comment.Content.Raw); err != nil {
					return err
				}
				if comment.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", comment.Links.HTML.Href); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", prTarget.ID, prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&body, "body", "", "Comment body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read the comment body from a file, or '-' for stdin")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	cmd.AddCommand(
		newPRCommentViewCmd(),
		newPRCommentEditCmd(),
		newPRCommentDeleteCmd(),
		newPRCommentResolveCmd(),
		newPRCommentReopenCmd(),
	)

	return cmd
}

func newPRDiffCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var stat bool

	cmd := &cobra.Command{
		Use:   "diff <id-or-url>",
		Short: "View a pull request diff",
		Long:  "Show the patch for a pull request by default. Use --stat for a concise per-file summary, or --json for structured output that includes both the patch and diff stats. Accepts a numeric ID, pull request URL, or pull request comment URL.",
		Example: "  bb pr diff 1\n" +
			"  bb pr diff 1 --repo workspace-slug/repo-slug --stat\n" +
			"  bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json patch,stats\n" +
			"  bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --stat",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}
			prTarget := resolved.Target
			client := resolved.Client

			patch, err := client.GetPullRequestPatch(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			stats, err := client.ListPullRequestDiffStats(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			pr, err := client.GetPullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			payload := prDiffPayload{
				Host:      prTarget.RepoTarget.Host,
				Workspace: prTarget.RepoTarget.Workspace,
				Repo:      prTarget.RepoTarget.Repo,
				ID:        pr.ID,
				Title:     pr.Title,
				Patch:     patch,
				Stats:     stats,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if stat {
					if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
						return err
					}
					if _, err := fmt.Fprintf(w, "Pull Request: #%d %s\n\n", payload.ID, payload.Title); err != nil {
						return err
					}
					return writePRDiffStatTable(w, stats)
				}
				_, err := io.WriteString(w, patch)
				return err
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&stat, "stat", false, "Show a concise per-file diff summary instead of the full patch")

	return cmd
}

func newPRStatusCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show pull request status for a repository",
		Long:  "Show pull request status for one repository, including the current branch pull request when available, open pull requests created by you, and open pull requests requesting your review.",
		Example: "  bb pr status\n" +
			"  bb pr status --repo workspace-slug/repo-slug\n" +
			"  bb pr status --json current_branch,created,review_requested",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			target := resolved.Target
			client := resolved.Client

			currentUser, err := client.CurrentUser(context.Background())
			if err != nil {
				return err
			}

			prs, err := client.ListPullRequests(context.Background(), target.Workspace, target.Repo, bitbucket.ListPullRequestsOptions{
				State: "OPEN",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			currentBranch := ""
			currentBranchError := ""
			if target.LocalRepo != nil {
				currentBranch, err = gitrepo.CurrentBranch(context.Background(), target.LocalRepo.RootDir)
				if err != nil {
					currentBranchError = err.Error()
				}
			}

			payload := buildPRStatusPayload(target, currentUser, currentBranch, currentBranchError, prs)

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePRStatusSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of open pull requests to inspect for status")

	return cmd
}

func newPRListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var state string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests for a repository",
		Example: "  bb pr list\n" +
			"  bb pr list --repo workspace-slug/repo-slug\n" +
			"  bb pr list --repo https://bitbucket.org/workspace-slug/repo-slug\n" +
			"  bb pr list --state ALL --json id,title,state",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			target := resolved.Target
			client := resolved.Client

			prs, err := client.ListPullRequests(context.Background(), target.Workspace, target.Repo, bitbucket.ListPullRequestsOptions{
				State: state,
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, prs, func(w io.Writer) error {
				if len(prs) == 0 {
					if _, err := fmt.Fprintf(w, "No pull requests found for %s/%s.\n", target.Workspace, target.Repo); err != nil {
						return err
					}
					return writeNextStep(w, fmt.Sprintf("bb pr create --repo %s/%s --title '<title>'", target.Workspace, target.Repo))
				}

				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				return writePRListTable(w, prs)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&state, "state", "OPEN", "Filter pull requests by state: OPEN, MERGED, DECLINED, SUPERSEDED, or ALL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pull requests to return")

	return cmd
}
