package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type issueCommentPayload struct {
	Host      string                 `json:"host"`
	Workspace string                 `json:"workspace"`
	Repo      string                 `json:"repo"`
	Issue     int                    `json:"issue"`
	Action    string                 `json:"action,omitempty"`
	Deleted   bool                   `json:"deleted,omitempty"`
	Comment   bitbucket.IssueComment `json:"comment"`
}

type issueCommentListPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Issue     int                      `json:"issue"`
	Comments  []bitbucket.IssueComment `json:"comments"`
}

func newIssueCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Work with issue comments",
		Long:  "List, view, create, edit, and delete Bitbucket issue comments.",
		Example: "  bb issue comment list 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body 'Needs follow-up'\n" +
			"  bb issue comment view 3 --issue 1 --repo workspace-slug/issues-repo-slug",
	}
	cmd.AddCommand(
		newIssueCommentListCmd(),
		newIssueCommentViewCmd(),
		newIssueCommentCreateCmd(),
		newIssueCommentEditCmd(),
		newIssueCommentDeleteCmd(),
	)
	return cmd
}

func newIssueCommentListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list <issue-id-or-url>",
		Short: "List comments on an issue",
		Example: "  bb issue comment list 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue comment list https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, issueID, err := resolveIssueReference(host, workspace, repo, args[0])
			if err != nil {
				return err
			}
			comments, err := client.ListIssueComments(context.Background(), target.Workspace, target.Repo, issueID, limit)
			if err != nil {
				return err
			}
			payload := issueCommentListPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Issue: issueID, Comments: comments}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueCommentListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of issue comments to return")
	return cmd
}

func newIssueCommentViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, issueRef string

	cmd := &cobra.Command{
		Use:   "view <comment-id>",
		Short: "View one issue comment",
		Example: "  bb issue comment view 3 --issue 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue comment view 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, issueID, commentID, err := resolveIssueCommentReference(host, workspace, repo, issueRef, args[0])
			if err != nil {
				return err
			}
			comment, err := client.GetIssueComment(context.Background(), target.Workspace, target.Repo, issueID, commentID)
			if err != nil {
				return err
			}
			payload := issueCommentPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Issue: issueID, Action: "viewed", Comment: comment}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueCommentSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	addIssueCommentTargetFlags(cmd, &host, &workspace, &repo, &issueRef)
	return cmd
}

func newIssueCommentCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var body, bodyFile string

	cmd := &cobra.Command{
		Use:   "create <issue-id-or-url>",
		Short: "Create an issue comment",
		Example: "  bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body 'Needs follow-up'\n" +
			"  bb issue comment create https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --body-file comment.md --json '*'\n" +
			"  printf 'Needs follow-up\\n' | bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body-file -",
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
			target, client, issueID, err := resolveIssueReference(host, workspace, repo, args[0])
			if err != nil {
				return err
			}
			comment, err := client.CreateIssueComment(context.Background(), target.Workspace, target.Repo, issueID, commentBody)
			if err != nil {
				return err
			}
			payload := issueCommentPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Issue: issueID, Action: "created", Comment: comment}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueCommentSummary(w, payload)
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
	return cmd
}

func newIssueCommentEditCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, issueRef string
	var body, bodyFile string

	cmd := &cobra.Command{
		Use:   "edit <comment-id>",
		Short: "Edit an issue comment",
		Example: "  bb issue comment edit 3 --issue 1 --repo workspace-slug/issues-repo-slug --body 'Updated feedback'\n" +
			"  bb issue comment edit 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --body-file comment.md --json '*'",
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
			target, client, issueID, commentID, err := resolveIssueCommentReference(host, workspace, repo, issueRef, args[0])
			if err != nil {
				return err
			}
			comment, err := client.UpdateIssueComment(context.Background(), target.Workspace, target.Repo, issueID, commentID, commentBody)
			if err != nil {
				return err
			}
			payload := issueCommentPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Issue: issueID, Action: "edited", Comment: comment}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueCommentSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	addIssueCommentTargetFlags(cmd, &host, &workspace, &repo, &issueRef)
	cmd.Flags().StringVar(&body, "body", "", "Comment body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read the comment body from a file, or '-' for stdin")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	return cmd
}

func newIssueCommentDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, issueRef string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <comment-id>",
		Short: "Delete an issue comment",
		Long:  "Delete a Bitbucket issue comment. Humans must confirm the exact repository, issue, and comment unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb issue comment delete 3 --issue 1 --repo workspace-slug/issues-repo-slug --yes\n" +
			"  bb --no-prompt issue comment delete 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --yes --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, issueID, commentID, err := resolveIssueCommentReference(host, workspace, repo, issueRef, args[0])
			if err != nil {
				return err
			}
			comment, err := client.GetIssueComment(context.Background(), target.Workspace, target.Repo, issueID, commentID)
			if err != nil {
				return err
			}
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("issue comment deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s#%d:%d", target.Workspace, target.Repo, issueID, commentID)); err != nil {
					return err
				}
			}
			if err := client.DeleteIssueComment(context.Background(), target.Workspace, target.Repo, issueID, commentID); err != nil {
				return err
			}
			payload := issueCommentPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Issue: issueID, Action: "deleted", Deleted: true, Comment: comment}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueCommentSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	addIssueCommentTargetFlags(cmd, &host, &workspace, &repo, &issueRef)
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func addIssueCommentTargetFlags(cmd *cobra.Command, host, workspace, repo, issueRef *string) {
	cmd.Flags().StringVar(host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(issueRef, "issue", "", "Issue ID or Bitbucket issue URL")
}

func resolveIssueCommentReference(host, workspace, repo, issueRef, commentRef string) (resolvedRepoTarget, *bitbucket.Client, int, int, error) {
	target, client, issueID, err := resolveIssueReference(host, workspace, repo, issueRef)
	if err != nil {
		return resolvedRepoTarget{}, nil, 0, 0, err
	}
	commentID, err := parsePositiveInt("issue comment", commentRef)
	if err != nil {
		return resolvedRepoTarget{}, nil, 0, 0, err
	}
	return target, client, issueID, commentID, nil
}

func writeIssueCommentListSummary(w io.Writer, payload issueCommentListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Issue", fmt.Sprintf("%d", payload.Issue)); err != nil {
		return err
	}
	if len(payload.Comments) == 0 {
		if _, err := fmt.Fprintf(w, "No comments found on issue %s/%s#%d.\n", payload.Workspace, payload.Repo, payload.Issue); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb issue comment create %d --repo %s/%s --body '<comment>'", payload.Issue, payload.Workspace, payload.Repo))
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tauthor\tupdated\tbody"); err != nil {
		return err
	}
	for _, comment := range payload.Comments {
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", comment.ID, output.Truncate(comment.User.DisplayName, 16), output.Truncate(comment.UpdatedOn, 20), output.Truncate(comment.Content.Raw, 40)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue comment view %d --issue %d --repo %s/%s", payload.Comments[0].ID, payload.Issue, payload.Workspace, payload.Repo))
}

func writeIssueCommentSummary(w io.Writer, payload issueCommentPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Issue", fmt.Sprintf("%d", payload.Issue)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Comment", fmt.Sprintf("%d", payload.Comment.ID)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb issue comment list %d --repo %s/%s", payload.Issue, payload.Workspace, payload.Repo))
	}
	if err := writeLabelValue(w, "Author", payload.Comment.User.DisplayName); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Updated", payload.Comment.UpdatedOn); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Body", payload.Comment.Content.Raw); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue comment list %d --repo %s/%s", payload.Issue, payload.Workspace, payload.Repo))
}
