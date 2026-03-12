package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newPRCheckoutCmd() *cobra.Command {
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "checkout <id-or-url>",
		Short: "Check out a pull request locally",
		Long:  "Fetch the pull request source branch from the current repository's remote and switch to it locally.",
		Example: "  bb pr checkout 1\n" +
			"  bb pr checkout 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr checkout https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1",
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

			resolved, err := resolvePullRequestCommandTarget(context.Background(), selector.Host, selector.Workspace, selector.Repo, args[0], true)
			if err != nil {
				return err
			}
			prTarget := resolved.Target
			client := resolved.Client
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

			if err := writeTargetHeader(cmd.OutOrStdout(), "Repository", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Pull Request: #%d\n", pr.ID); err != nil {
				return err
			}
			if err := writeLabelValue(cmd.OutOrStdout(), "Branch", pr.Source.Branch.Name); err != nil {
				return err
			}
			if err := writeLabelValue(cmd.OutOrStdout(), "Local Root", repoContext.RootDir); err != nil {
				return err
			}
			if err := writeLabelValue(cmd.OutOrStdout(), "Status", "checked out"); err != nil {
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
			"  bb pr merge 7 --repo workspace-slug/repo-slug\n" +
			"  bb pr merge 7 --strategy merge_commit\n" +
			"  bb pr merge 7 --message 'Ship feature' --close-source-branch --json",
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
				if err := writePullRequestSummaryTable(w, mergedPR, pullRequestSummaryOptions{
					Strategy:    mergeStrategy,
					MergeCommit: mergedPR.MergeCommit.Hash,
				}); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", mergedPR.ID, prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
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

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			repoTarget := resolved.Target
			client := resolved.Client

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
				if err := writePullRequestSummaryTable(w, pr, pullRequestSummaryOptions{}); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", pr.ID, repoTarget.Workspace, repoTarget.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
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
