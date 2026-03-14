package cmd

import (
	"context"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type commitViewPayload struct {
	Host      string                     `json:"host"`
	Workspace string                     `json:"workspace"`
	Repo      string                     `json:"repo"`
	Warnings  []string                   `json:"warnings,omitempty"`
	Commit    bitbucket.RepositoryCommit `json:"commit"`
}

type commitDiffPayload struct {
	Host      string                          `json:"host"`
	Workspace string                          `json:"workspace"`
	Repo      string                          `json:"repo"`
	Warnings  []string                        `json:"warnings,omitempty"`
	Commit    string                          `json:"commit"`
	Patch     string                          `json:"patch,omitempty"`
	Stats     []bitbucket.PullRequestDiffStat `json:"stats,omitempty"`
}

type commitStatusesPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Warnings  []string                 `json:"warnings,omitempty"`
	Commit    string                   `json:"commit"`
	Statuses  []bitbucket.CommitStatus `json:"statuses"`
}

type commitReviewPayload struct {
	Host      string                           `json:"host"`
	Workspace string                           `json:"workspace"`
	Repo      string                           `json:"repo"`
	Warnings  []string                         `json:"warnings,omitempty"`
	Commit    string                           `json:"commit"`
	Action    string                           `json:"action"`
	Reviewer  bitbucket.PullRequestParticipant `json:"reviewer,omitempty"`
}

func newCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Work with repository commits",
		Long:  "Inspect Bitbucket repository commits, diffs, comments, statuses, approvals, and code-insight reports.",
	}

	cmd.AddCommand(
		newCommitViewCmd(),
		newCommitDiffCmd(),
		newCommitStatusesCmd(),
		newCommitApproveCmd(true),
		newCommitApproveCmd(false),
		newCommitCommentCmd(),
		newCommitReportCmd(),
	)

	return cmd
}

func newCommitViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <hash-or-url>",
		Short: "View one commit",
		Long:  "View one Bitbucket commit. Accepts a commit SHA or a commit URL.",
		Example: "  bb commit view abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit view https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'\n" +
			"  bb commit view abc1234",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			commit, err := resolved.Client.GetCommit(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit)
			if err != nil {
				return err
			}

			payload := commitViewPayload{
				Host:      resolved.Target.RepoTarget.Host,
				Workspace: resolved.Target.RepoTarget.Workspace,
				Repo:      resolved.Target.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				Commit:    commit,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitViewSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newCommitDiffCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var stat bool
	var contextLines int
	var pathFilters []string
	var ignoreWhitespace bool
	var binaryValue bool
	var renamesValue bool

	cmd := &cobra.Command{
		Use:   "diff <hash-or-url>",
		Short: "Show a commit diff",
		Long:  "Show a commit patch by default or diff stats with --stat. Accepts a commit SHA or a commit URL.",
		Example: "  bb commit diff abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit diff abc1234 --repo workspace-slug/repo-slug --stat\n" +
			"  bb commit diff https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json patch,stats",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			payload, err := buildCommitDiffPayload(
				context.Background(),
				cmd,
				resolved,
				stat,
				contextLines,
				pathFilters,
				ignoreWhitespace,
				binaryValue,
				renamesValue,
			)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if stat {
					return writeCommitDiffStatSummary(w, payload)
				}
				return writeCommitDiffPatchSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&stat, "stat", false, "Show per-file diff stats instead of the raw patch")
	cmd.Flags().IntVar(&contextLines, "context", 0, "Lines of context to include in the raw diff")
	cmd.Flags().StringSliceVar(&pathFilters, "path", nil, "Limit the raw diff to one or more file paths")
	cmd.Flags().BoolVar(&ignoreWhitespace, "ignore-whitespace", false, "Ignore whitespace changes in the raw diff")
	cmd.Flags().BoolVar(&binaryValue, "binary", true, "Include binary file changes in the raw diff")
	cmd.Flags().BoolVar(&renamesValue, "renames", true, "Perform rename detection in the raw diff")
	return cmd
}

func buildCommitDiffPayload(ctx context.Context, cmd *cobra.Command, resolved resolvedCommitCommandTarget, stat bool, contextLines int, pathFilters []string, ignoreWhitespace, binaryValue, renamesValue bool) (commitDiffPayload, error) {
	payload := commitDiffPayload{
		Host:      resolved.Target.RepoTarget.Host,
		Workspace: resolved.Target.RepoTarget.Workspace,
		Repo:      resolved.Target.RepoTarget.Repo,
		Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
		Commit:    resolved.Target.Commit,
	}

	if stat {
		stats, err := resolved.Client.ListCommitDiffStats(ctx, resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit)
		if err != nil {
			return commitDiffPayload{}, err
		}
		payload.Stats = stats
		return payload, nil
	}

	diffOptions := bitbucket.CommitDiffOptions{
		Context:          contextLines,
		Path:             pathFilters,
		IgnoreWhitespace: ignoreWhitespace,
	}
	if cmd.Flags().Changed("binary") {
		diffOptions.Binary = &binaryValue
	}
	if cmd.Flags().Changed("renames") {
		diffOptions.Renames = &renamesValue
	}

	patch, err := resolved.Client.GetCommitDiff(ctx, resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit, diffOptions)
	if err != nil {
		return commitDiffPayload{}, err
	}
	payload.Patch = patch
	return payload, nil
}

func newCommitStatusesCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	var sort, query, refName string

	cmd := &cobra.Command{
		Use:     "statuses <hash-or-url>",
		Aliases: []string{"checks"},
		Short:   "List commit statuses for a commit",
		Long:    "List Bitbucket commit statuses for a commit. Accepts a commit SHA or a commit URL.",
		Example: "  bb commit statuses abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit checks https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json statuses\n" +
			"  bb commit statuses abc1234 --repo workspace-slug/repo-slug --limit 50",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			statuses, err := resolved.Client.ListCommitStatuses(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit, bitbucket.ListCommitStatusesOptions{
				Limit:   limit,
				Sort:    sort,
				Query:   query,
				RefName: refName,
			})
			if err != nil {
				return err
			}

			payload := commitStatusesPayload{
				Host:      resolved.Target.RepoTarget.Host,
				Workspace: resolved.Target.RepoTarget.Workspace,
				Repo:      resolved.Target.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				Commit:    resolved.Target.Commit,
				Statuses:  statuses,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitStatusesSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of commit statuses to return")
	cmd.Flags().StringVar(&sort, "sort", "-created_on", "Bitbucket commit status sort expression")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket commit status query filter")
	cmd.Flags().StringVar(&refName, "refname", "", "Limit statuses to the given ref name")
	return cmd
}

func newCommitApproveCmd(approve bool) *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	use := "approve"
	short := "Approve a commit"
	action := "approved"
	if !approve {
		use = "unapprove"
		short = "Withdraw your approval from a commit"
		action = "unapproved"
	}

	cmd := &cobra.Command{
		Use:   use + " <hash-or-url>",
		Short: short,
		Long:  short + ". Accepts a commit SHA or a commit URL.",
		Example: "  bb commit " + use + " abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit " + use + " https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			reviewer, err := resolved.Client.ReviewCommit(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit, approve)
			if err != nil {
				return err
			}

			payload := commitReviewPayload{
				Host:      resolved.Target.RepoTarget.Host,
				Workspace: resolved.Target.RepoTarget.Workspace,
				Repo:      resolved.Target.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				Commit:    resolved.Target.Commit,
				Action:    action,
				Reviewer:  reviewer,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitReviewSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}
