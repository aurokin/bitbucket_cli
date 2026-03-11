package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
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
		Long:  "Close a pull request without merging it. In Bitbucket Cloud this maps to declining the pull request.",
		Example: "  bb pr close 1\n" +
			"  bb pr close 1 --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pr close https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1 --json",
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
			prTarget, err := resolvePullRequestTarget(context.Background(), selector, client, args[0], true)
			if err != nil {
				return err
			}

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
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", closedPR.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", closedPR.Title); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", closedPR.State); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Source:\t%s\n", closedPR.Source.Branch.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", closedPR.Destination.Branch.Name); err != nil {
					return err
				}
				if closedPR.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", closedPR.Links.HTML.Href); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", closedPR.ID, prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo))
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
		Long:  "Add a comment to a pull request using --body, --body-file, or --body-file - for stdin. This first pass is intentionally deterministic for agent and script usage.",
		Example: "  bb pr comment 1 --body 'Looks good'\n" +
			"  bb pr comment 1 --repo OhBizzle/bb-cli-integration-primary --body-file comment.md\n" +
			"  printf 'Ship it\\n' | bb pr comment https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1 --body-file - --json",
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

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			prTarget, err := resolvePullRequestTarget(context.Background(), selector, client, args[0], true)
			if err != nil {
				return err
			}

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

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&body, "body", "", "Comment body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read the comment body from a file, or '-' for stdin")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")

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
		Long:  "Show the patch for a pull request by default. Use --stat for a concise per-file summary, or --json for structured output that includes both the patch and diff stats.",
		Example: "  bb pr diff 1\n" +
			"  bb pr diff 1 --repo OhBizzle/bb-cli-integration-primary --stat\n" +
			"  bb pr diff https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1 --json patch,stats",
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
			prTarget, err := resolvePullRequestTarget(context.Background(), selector, client, args[0], true)
			if err != nil {
				return err
			}

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

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
			"  bb pr status --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pr status --json current_branch,created,review_requested",
		Args: cobra.NoArgs,
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
			if target.LocalRepo != nil {
				currentBranch, _ = gitrepo.CurrentBranch(context.Background(), target.LocalRepo.RootDir)
			}

			payload := buildPRStatusPayload(target, currentUser, currentBranch, prs)

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if _, err := fmt.Fprintf(w, "Repository: %s/%s\n", payload.Workspace, payload.Repo); err != nil {
					return err
				}

				if payload.CurrentBranchName != "" {
					if _, err := fmt.Fprintf(w, "Current Branch: %s\n", payload.CurrentBranchName); err != nil {
						return err
					}
				} else if _, err := fmt.Fprintln(w, "Current Branch: unavailable"); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(w, ""); err != nil {
					return err
				}
				if _, err := fmt.Fprintln(w, "Current Branch Pull Request"); err != nil {
					return err
				}
				currentBranchPRs := make([]bitbucket.PullRequest, 0, 1)
				if payload.CurrentBranch != nil {
					currentBranchPRs = append(currentBranchPRs, *payload.CurrentBranch)
				}
				if err := writePRStatusSection(w, currentBranchPRs...); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(w, ""); err != nil {
					return err
				}
				if _, err := fmt.Fprintln(w, "Created By You"); err != nil {
					return err
				}
				if err := writePRStatusSection(w, payload.Created...); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(w, ""); err != nil {
					return err
				}
				if _, err := fmt.Fprintln(w, "Review Requested"); err != nil {
					return err
				}
				if err := writePRStatusSection(w, payload.ReviewRequested...); err != nil {
					return err
				}
				if len(payload.Created) == 0 && len(payload.ReviewRequested) == 0 && payload.CurrentBranch == nil {
					return writeNextStep(w, fmt.Sprintf("bb pr list --repo %s/%s", payload.Workspace, payload.Repo))
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
			"  bb pr list --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pr list --repo https://bitbucket.org/OhBizzle/bb-cli-integration-primary\n" +
			"  bb pr list --state ALL --json id,title,state",
		Args: cobra.NoArgs,
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

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&state, "state", "OPEN", "Filter pull requests by state: OPEN, MERGED, DECLINED, SUPERSEDED, or ALL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pull requests to return")

	return cmd
}

func newPRCheckoutCmd() *cobra.Command {
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "checkout <id-or-url>",
		Short: "Check out a pull request locally",
		Long:  "Fetch the pull request source branch from the current repository's remote and switch to it locally.",
		Example: "  bb pr checkout 1\n" +
			"  bb pr checkout 1 --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pr checkout https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoContext, err := resolveLocalRepoContext(context.Background())
			if err != nil {
				return fmt.Errorf("pr checkout must be run inside a local git checkout of the target repository")
			}

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			selector, err = mergeRepoSelectors(selector, repoSelector{
				Host:      repoContext.Host,
				Workspace: repoContext.Workspace,
				Repo:      repoContext.RepoSlug,
			})
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			prTarget, err := resolvePullRequestTarget(context.Background(), selector, client, args[0], true)
			if err != nil {
				return err
			}
			if prTarget.RepoTarget.Workspace != repoContext.Workspace || prTarget.RepoTarget.Repo != repoContext.RepoSlug {
				return fmt.Errorf("pr checkout must be run inside a local git checkout of the target repository")
			}

			pr, err := client.GetPullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			if err := gitrepo.CheckoutRemoteBranch(context.Background(), repoContext.RootDir, repoContext.RemoteName, pr.Source.Branch.Name); err != nil {
				return err
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Checked out %s/%s PR #%d onto %s\n", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, pr.ID, pr.Source.Branch.Name); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Root: %s\n", repoContext.RootDir); err != nil {
				return err
			}
			return writeNextStep(cmd.OutOrStdout(), fmt.Sprintf("bb pr view %d --repo %s/%s", pr.ID, prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo))
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newPRMergeCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var message string
	var strategy string
	var closeSourceBranch bool

	cmd := &cobra.Command{
		Use:   "merge <id-or-url>",
		Short: "Merge a pull request",
		Long:  "Merge an open pull request in Bitbucket Cloud. bb uses the destination branch default merge strategy when Bitbucket exposes one, or falls back to the repository default when Bitbucket does not include strategy metadata on the pull request.",
		Example: "  bb pr merge 7\n" +
			"  bb pr merge 7 --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pr merge 7 --strategy merge_commit\n" +
			"  bb pr merge 7 --message 'Ship feature' --close-source-branch --json",
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
			prTarget, err := resolvePullRequestTarget(context.Background(), selector, client, args[0], true)
			if err != nil {
				return err
			}

			pr, err := client.GetPullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}
			if pr.State != "OPEN" {
				return fmt.Errorf("pull request #%d is %s; only OPEN pull requests can be merged", pr.ID, pr.State)
			}

			mergeStrategy, err := resolveMergeStrategy(pr, strategy)
			if err != nil {
				return err
			}

			mergedPR, err := client.MergePullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID, bitbucket.MergePullRequestOptions{
				Message:           message,
				CloseSourceBranch: closeSourceBranch,
				MergeStrategy:     mergeStrategy,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, mergedPR, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", mergedPR.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", mergedPR.Title); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", mergedPR.State); err != nil {
					return err
				}
				if mergeStrategy != "" {
					if _, err := fmt.Fprintf(tw, "Strategy:\t%s\n", mergeStrategy); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintf(tw, "Source:\t%s\n", mergedPR.Source.Branch.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", mergedPR.Destination.Branch.Name); err != nil {
					return err
				}
				if mergedPR.MergeCommit.Hash != "" {
					if _, err := fmt.Fprintf(tw, "Merge Commit:\t%s\n", mergedPR.MergeCommit.Hash); err != nil {
						return err
					}
				}
				if mergedPR.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", mergedPR.Links.HTML.Href); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", mergedPR.ID, prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo))
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&message, "message", "", "Merge commit message")
	cmd.Flags().StringVar(&strategy, "strategy", "", "Merge strategy to use; required when Bitbucket does not expose a default")
	cmd.Flags().BoolVar(&closeSourceBranch, "close-source-branch", false, "Close the source branch when the pull request is merged")

	return cmd
}

func newPRCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var title string
	var description string
	var source string
	var destination string
	var closeSourceBranch bool
	var draft bool
	var reuseExisting bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		Long:  "Create a pull request in Bitbucket Cloud. When run interactively, bb prompts for missing fields. The source branch defaults to the current branch and the destination defaults to the repository main branch.",
		Example: "  bb pr create --title 'Add feature'\n" +
			"  bb pr create --source feature --destination main --description 'Ready for review'\n" +
			"  bb pr create --reuse-existing --json",
		Args: cobra.NoArgs,
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
			repoTarget, err := resolveRepoTarget(context.Background(), selector, client, true)
			if err != nil {
				return err
			}

			interactive := promptsEnabled(cmd)

			sourceBranch, err := resolveSourceBranchInput(cmd, source, interactive, repoTarget.Explicit, repoTarget.Workspace, repoTarget.Repo)
			if err != nil {
				return err
			}

			destinationBranch, err := resolveDestinationBranchInput(cmd, client, repoTarget.Workspace, repoTarget.Repo, destination, interactive)
			if err != nil {
				return err
			}

			if title == "" {
				title = defaultPRTitle(sourceBranch)
			}
			if interactive {
				title, err = promptRequiredString(cmd, "Title", title)
				if err != nil {
					return err
				}
			}

			pr, err := client.CreatePullRequest(context.Background(), repoTarget.Workspace, repoTarget.Repo, bitbucket.CreatePullRequestOptions{
				Title:             title,
				Description:       description,
				SourceBranch:      sourceBranch,
				DestinationBranch: destinationBranch,
				CloseSourceBranch: closeSourceBranch,
				Draft:             draft,
				ReuseExisting:     reuseExisting,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, pr, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", repoTarget.Workspace, repoTarget.Repo); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", pr.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", pr.Title); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", pr.State); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Source:\t%s\n", pr.Source.Branch.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", pr.Destination.Branch.Name); err != nil {
					return err
				}
				if pr.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", pr.Links.HTML.Href); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", pr.ID, repoTarget.Workspace, repoTarget.Repo))
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&title, "title", "", "Pull request title; defaults to the source branch name")
	cmd.Flags().StringVar(&description, "description", "", "Pull request description")
	cmd.Flags().StringVar(&source, "source", "", "Source branch; defaults to the current git branch")
	cmd.Flags().StringVar(&destination, "destination", "", "Destination branch; defaults to the repository main branch")
	cmd.Flags().BoolVar(&closeSourceBranch, "close-source-branch", false, "Close the source branch when the pull request is merged")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create the pull request as a draft")
	cmd.Flags().BoolVar(&reuseExisting, "reuse-existing", false, "Return an existing matching open pull request instead of creating a new one")

	return cmd
}

func newPRViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <id-or-url>",
		Short: "View a pull request",
		Example: "  bb pr view 1\n" +
			"  bb pr view 1 --json title,state,source,destination\n" +
			"  bb pr view https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1",
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
			prTarget, err := resolvePullRequestTarget(context.Background(), selector, client, args[0], true)
			if err != nil {
				return err
			}

			pr, err := client.GetPullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, pr, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", pr.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", pr.Title); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", pr.State); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Author:\t%s\n", pr.Author.DisplayName); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Source:\t%s\n", pr.Source.Branch.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", pr.Destination.Branch.Name); err != nil {
					return err
				}
				if pr.UpdatedOn != "" {
					if _, err := fmt.Fprintf(tw, "Updated:\t%s\n", pr.UpdatedOn); err != nil {
						return err
					}
				}
				if pr.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", pr.Links.HTML.Href); err != nil {
						return err
					}
				}
				if pr.Description != "" {
					if _, err := fmt.Fprintf(tw, "Description:\t%s\n", pr.Description); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, prViewNextStep(prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, pr.ID))
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func resolveSourceBranch(source string) (string, error) {
	if source != "" {
		return source, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	branch, err := gitrepo.CurrentBranch(context.Background(), currentDir)
	if err != nil {
		return "", fmt.Errorf("resolve source branch: %w", err)
	}
	if branch == "" {
		return "", fmt.Errorf("could not determine current branch; pass --source")
	}

	return branch, nil
}

func prViewNextStep(workspace, repo string, id int) string {
	return fmt.Sprintf("bb pr diff %d --repo %s/%s", id, workspace, repo)
}

func resolveSourceBranchInput(cmd *cobra.Command, source string, interactive bool, explicitRepoSelector bool, workspace, repo string) (string, error) {
	if source != "" {
		return source, nil
	}

	if explicitRepoSelector {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}

		localRepo, err := gitrepo.ResolveRepoContext(context.Background(), currentDir)
		if err != nil || localRepo.Workspace != workspace || localRepo.RepoSlug != repo {
			if interactive {
				return promptRequiredString(cmd, "Source branch", "")
			}
			return "", fmt.Errorf("could not determine the source branch for %s/%s from the current directory; pass --source or run in an interactive terminal", workspace, repo)
		}
	}

	defaultSource, err := resolveSourceBranch(source)
	if err == nil {
		if interactive {
			return promptRequiredString(cmd, "Source branch", defaultSource)
		}
		return defaultSource, nil
	}

	if interactive {
		return promptRequiredString(cmd, "Source branch", "")
	}

	return "", fmt.Errorf("could not determine the source branch; pass --source or run in an interactive terminal")
}

func resolveDestinationBranch(ctx context.Context, client *bitbucket.Client, workspace, repo, destination string) (string, error) {
	if destination != "" {
		return destination, nil
	}

	repository, err := client.GetRepository(ctx, workspace, repo)
	if err != nil {
		return "", err
	}
	if repository.MainBranch.Name == "" {
		return "", fmt.Errorf("repository main branch is unknown; pass --destination")
	}

	return repository.MainBranch.Name, nil
}

func resolveDestinationBranchInput(cmd *cobra.Command, client *bitbucket.Client, workspace, repo, destination string, interactive bool) (string, error) {
	if destination != "" {
		return destination, nil
	}

	defaultDestination, err := resolveDestinationBranch(context.Background(), client, workspace, repo, "")
	if err == nil {
		if interactive {
			return promptRequiredString(cmd, "Destination branch", defaultDestination)
		}
		return defaultDestination, nil
	}

	if interactive {
		return promptRequiredString(cmd, "Destination branch", "")
	}

	return "", fmt.Errorf("could not determine the destination branch; pass --destination or run in an interactive terminal")
}

func resolveMergeStrategy(pr bitbucket.PullRequest, requested string) (string, error) {
	available := uniqueNonEmptyStrings(pr.Destination.Branch.MergeStrategies)

	requested = strings.TrimSpace(requested)
	if requested != "" {
		if len(available) > 0 && !stringSliceContains(available, requested) {
			return "", fmt.Errorf("merge strategy %q is not allowed for destination branch %s; available: %s", requested, pr.Destination.Branch.Name, strings.Join(available, ", "))
		}
		return requested, nil
	}

	if defaultStrategy := strings.TrimSpace(pr.Destination.Branch.DefaultMergeStrategy); defaultStrategy != "" {
		return defaultStrategy, nil
	}

	if len(available) == 1 {
		return available[0], nil
	}
	if len(available) > 1 {
		return "", fmt.Errorf("multiple merge strategies are available for destination branch %s; pass --strategy (%s)", pr.Destination.Branch.Name, strings.Join(available, ", "))
	}

	return "", nil
}

type prStatusPayload struct {
	Host              string                  `json:"host"`
	Workspace         string                  `json:"workspace"`
	Repo              string                  `json:"repo"`
	CurrentUser       bitbucket.CurrentUser   `json:"current_user"`
	CurrentBranchName string                  `json:"current_branch_name,omitempty"`
	CurrentBranch     *bitbucket.PullRequest  `json:"current_branch,omitempty"`
	Created           []bitbucket.PullRequest `json:"created"`
	ReviewRequested   []bitbucket.PullRequest `json:"review_requested"`
}

type prDiffPayload struct {
	Host      string                          `json:"host"`
	Workspace string                          `json:"workspace"`
	Repo      string                          `json:"repo"`
	ID        int                             `json:"id"`
	Title     string                          `json:"title"`
	Patch     string                          `json:"patch"`
	Stats     []bitbucket.PullRequestDiffStat `json:"stats"`
}

func buildPRStatusPayload(target resolvedRepoTarget, currentUser bitbucket.CurrentUser, currentBranch string, prs []bitbucket.PullRequest) prStatusPayload {
	payload := prStatusPayload{
		Host:              target.Host,
		Workspace:         target.Workspace,
		Repo:              target.Repo,
		CurrentUser:       currentUser,
		CurrentBranchName: currentBranch,
		Created:           make([]bitbucket.PullRequest, 0),
		ReviewRequested:   make([]bitbucket.PullRequest, 0),
	}

	currentBranchID := 0
	for i := range prs {
		pr := prs[i]
		if payload.CurrentBranch == nil && currentBranch != "" && pr.Source.Branch.Name == currentBranch {
			prCopy := pr
			payload.CurrentBranch = &prCopy
			currentBranchID = pr.ID
			continue
		}
	}

	for _, pr := range prs {
		if currentBranchID != 0 && pr.ID == currentBranchID {
			continue
		}
		if sameActor(currentUser, pr.Author) {
			payload.Created = append(payload.Created, pr)
			continue
		}
		if reviewRequestedFromUser(currentUser, pr) {
			payload.ReviewRequested = append(payload.ReviewRequested, pr)
		}
	}

	return payload
}

func writePRStatusSection(w io.Writer, prs ...bitbucket.PullRequest) error {
	if len(prs) == 0 {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}

	for _, pr := range prs {
		line := fmt.Sprintf("  #%d  %s [%s] %s -> %s", pr.ID, pr.Title, pr.State, pr.Source.Branch.Name, pr.Destination.Branch.Name)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

func writePRDiffStatTable(w io.Writer, stats []bitbucket.PullRequestDiffStat) error {
	if len(stats) == 0 {
		_, err := fmt.Fprintln(w, "No changed files.")
		return err
	}

	pathWidth := diffStatPathWidth(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "status\tfile\t+add\t-rem"); err != nil {
		return err
	}

	totalAdded := 0
	totalRemoved := 0
	for _, stat := range stats {
		totalAdded += stat.LinesAdded
		totalRemoved += stat.LinesRemoved
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%d\n", output.Truncate(diffStatus(stat), 10), output.TruncateMiddle(diffPath(stat), pathWidth), stat.LinesAdded, stat.LinesRemoved); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "total\t%d files\t%d\t%d\n", len(stats), totalAdded, totalRemoved); err != nil {
		return err
	}

	return tw.Flush()
}

func writePRListTable(w io.Writer, prs []bitbucket.PullRequest) error {
	titleWidth, authorWidth, branchWidth := prListColumnWidths(output.TerminalWidth(w))

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "#\ttitle\tstate\tauthor\tsrc\tdst\tupdated"); err != nil {
		return err
	}

	for _, pr := range prs {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			pr.ID,
			output.Truncate(pr.Title, titleWidth),
			output.Truncate(pr.State, 10),
			output.Truncate(pr.Author.DisplayName, authorWidth),
			output.TruncateMiddle(pr.Source.Branch.Name, branchWidth),
			output.TruncateMiddle(pr.Destination.Branch.Name, branchWidth),
			formatPRUpdated(pr.UpdatedOn),
		); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func formatPRUpdated(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}

	return parsed.Local().Format("2006-01-02 15:04")
}

func prListColumnWidths(termWidth int) (title, author, branch int) {
	switch {
	case termWidth >= 160:
		return 52, 18, 24
	case termWidth >= 132:
		return 40, 16, 18
	case termWidth >= 110:
		return 32, 14, 14
	default:
		return 24, 12, 12
	}
}

func diffStatPathWidth(termWidth int) int {
	switch {
	case termWidth >= 160:
		return 72
	case termWidth >= 132:
		return 56
	case termWidth >= 110:
		return 44
	default:
		return 32
	}
}

func diffPath(stat bitbucket.PullRequestDiffStat) string {
	switch {
	case stat.New != nil && stat.Old != nil && stat.New.Path != "" && stat.Old.Path != "" && stat.New.Path != stat.Old.Path:
		return stat.Old.Path + " -> " + stat.New.Path
	case stat.New != nil && stat.New.Path != "":
		return stat.New.Path
	case stat.Old != nil && stat.Old.Path != "":
		return stat.Old.Path
	default:
		return "(unknown)"
	}
}

func diffStatus(stat bitbucket.PullRequestDiffStat) string {
	status := strings.TrimSpace(stat.Status)
	if status == "" {
		return "changed"
	}
	return status
}

func resolveCommentBody(stdin io.Reader, body, bodyFile string) (string, error) {
	if trimmed := strings.TrimSpace(body); trimmed != "" {
		return trimmed, nil
	}

	if strings.TrimSpace(bodyFile) != "" {
		data, err := readRequestBody(stdin, bodyFile)
		if err != nil {
			return "", err
		}
		trimmed := strings.TrimSpace(string(data))
		if trimmed == "" {
			return "", fmt.Errorf("comment body is empty")
		}
		return trimmed, nil
	}

	return "", fmt.Errorf("provide a comment body with --body or --body-file")
}

func sameActor(user bitbucket.CurrentUser, actor bitbucket.PullRequestActor) bool {
	switch {
	case user.AccountID != "" && actor.AccountID != "":
		return user.AccountID == actor.AccountID
	case user.Username != "" && actor.Nickname != "":
		return user.Username == actor.Nickname
	case user.DisplayName != "" && actor.DisplayName != "":
		return user.DisplayName == actor.DisplayName
	default:
		return false
	}
}

func reviewRequestedFromUser(user bitbucket.CurrentUser, pr bitbucket.PullRequest) bool {
	for _, reviewer := range pr.Reviewers {
		if sameActor(user, reviewer) {
			return true
		}
	}
	return false
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func defaultPRTitle(sourceBranch string) string {
	return sourceBranch
}
