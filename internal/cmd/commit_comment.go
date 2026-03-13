package cmd

import (
	"context"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type commitCommentListPayload struct {
	Host      string                    `json:"host"`
	Workspace string                    `json:"workspace"`
	Repo      string                    `json:"repo"`
	Warnings  []string                  `json:"warnings,omitempty"`
	Commit    string                    `json:"commit"`
	Comments  []bitbucket.CommitComment `json:"comments"`
}

type commitCommentPayload struct {
	Host      string                  `json:"host"`
	Workspace string                  `json:"workspace"`
	Repo      string                  `json:"repo"`
	Warnings  []string                `json:"warnings,omitempty"`
	Commit    string                  `json:"commit"`
	Comment   bitbucket.CommitComment `json:"comment"`
}

func newCommitCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Work with commit comments",
		Long:  "List and inspect Bitbucket commit comments.",
	}
	cmd.AddCommand(newCommitCommentListCmd(), newCommitCommentViewCmd())
	return cmd
}

func newCommitCommentListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	var sort, query string

	cmd := &cobra.Command{
		Use:   "list <hash-or-url>",
		Short: "List comments on a commit",
		Long:  "List comments on a Bitbucket commit. Accepts a commit SHA or a commit URL.",
		Example: "  bb commit comment list abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit comment list https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json comments\n",
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

			comments, err := resolved.Client.ListCommitComments(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit, bitbucket.ListCommitCommentsOptions{
				Limit: limit,
				Sort:  sort,
				Query: query,
			})
			if err != nil {
				return err
			}

			payload := commitCommentListPayload{
				Host:      resolved.Target.RepoTarget.Host,
				Workspace: resolved.Target.RepoTarget.Workspace,
				Repo:      resolved.Target.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				Commit:    resolved.Target.Commit,
				Comments:  comments,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitCommentListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of commit comments to return")
	cmd.Flags().StringVar(&sort, "sort", "", "Bitbucket commit comment sort expression")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket commit comment query filter")
	return cmd
}

func newCommitCommentViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, commitRef string

	cmd := &cobra.Command{
		Use:   "view <comment-id>",
		Short: "View one commit comment",
		Long:  "View one commit comment. Pass the commit with --commit as a SHA or commit URL.",
		Example: "  bb commit comment view 15 --commit abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit comment view 15 --commit https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommentCommandTarget(context.Background(), host, workspace, repo, commitRef, args[0], true)
			if err != nil {
				return err
			}

			comment, err := resolved.Client.GetCommitComment(context.Background(), resolved.Target.CommitTarget.RepoTarget.Workspace, resolved.Target.CommitTarget.RepoTarget.Repo, resolved.Target.CommitTarget.Commit, resolved.Target.CommentID)
			if err != nil {
				return err
			}

			payload := commitCommentPayload{
				Host:      resolved.Target.CommitTarget.RepoTarget.Host,
				Workspace: resolved.Target.CommitTarget.RepoTarget.Workspace,
				Repo:      resolved.Target.CommitTarget.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.CommitTarget.RepoTarget.Warnings...),
				Commit:    resolved.Target.CommitTarget.Commit,
				Comment:   comment,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitCommentSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&commitRef, "commit", "", "Commit SHA or commit URL")
	_ = cmd.MarkFlagRequired("commit")
	return cmd
}
