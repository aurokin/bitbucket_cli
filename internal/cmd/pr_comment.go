package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type prCommentPayload struct {
	Host        string                               `json:"host"`
	Workspace   string                               `json:"workspace"`
	Repo        string                               `json:"repo"`
	PullRequest int                                  `json:"pull_request"`
	Action      string                               `json:"action,omitempty"`
	Deleted     bool                                 `json:"deleted,omitempty"`
	Comment     bitbucket.PullRequestComment         `json:"comment"`
	Resolution  *bitbucket.PullRequestCommentResolve `json:"resolution,omitempty"`
}

type pullRequestCommentSummaryOptions struct {
	Deleted bool
}

func createPullRequestCommentCommand(ctx context.Context, stdin io.Reader, host, workspace, repo, prRef, body, bodyFile string) (resolvedPullRequestCommandTarget, bitbucket.PullRequestComment, error) {
	commentBody, err := resolveCommentBody(stdin, body, bodyFile)
	if err != nil {
		return resolvedPullRequestCommandTarget{}, bitbucket.PullRequestComment{}, err
	}

	resolved, err := resolvePullRequestCommandTarget(ctx, host, workspace, repo, prRef, true)
	if err != nil {
		return resolvedPullRequestCommandTarget{}, bitbucket.PullRequestComment{}, err
	}

	comment, err := resolved.Client.CreatePullRequestComment(ctx, resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.ID, commentBody)
	if err != nil {
		return resolvedPullRequestCommandTarget{}, bitbucket.PullRequestComment{}, err
	}

	return resolved, comment, nil
}

func writePullRequestCommentCreateSummary(w io.Writer, target resolvedPullRequestTarget, comment bitbucket.PullRequestComment) error {
	if err := writeTargetHeader(w, "Repository", target.RepoTarget.Workspace, target.RepoTarget.Repo); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", target.ID); err != nil {
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
	return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", target.ID, target.RepoTarget.Workspace, target.RepoTarget.Repo))
}

func newPRCommentViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string

	cmd := &cobra.Command{
		Use:   "view <comment-url-or-id>",
		Short: "View a pull request comment",
		Long:  "View a pull request comment. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.",
		Example: "  bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15\n" +
			"  bb pr comment view 15 --pr 1 --repo workspace-slug/repo-slug --json '*'\n" +
			"  bb pr comment view 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommentCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			comment, err := resolved.Client.GetPullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID)
			if err != nil {
				return err
			}

			payload := prCommentPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "viewed",
				Comment:     comment,
				Resolution:  comment.Resolution,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestCommentSummary(w, payload, pullRequestCommentSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestCommentTargetFlags(cmd, &host, &workspace, &repo, &prRef)

	return cmd
}

func newPRCommentEditCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string
	var body string
	var bodyFile string

	cmd := &cobra.Command{
		Use:   "edit <comment-url-or-id>",
		Short: "Edit a pull request comment",
		Long:  "Edit a pull request comment. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.",
		Example: "  bb pr comment edit https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Updated feedback'\n" +
			"  bb pr comment edit 15 --pr 1 --repo workspace-slug/repo-slug --body-file comment.md --json '*'\n" +
			"  printf 'Updated feedback\\n' | bb pr comment edit 15 --pr 1 --repo workspace-slug/repo-slug --body-file -",
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

			resolved, err := resolvePullRequestCommentCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			comment, err := resolved.Client.UpdatePullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID, commentBody)
			if err != nil {
				return err
			}

			payload := prCommentPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "edited",
				Comment:     comment,
				Resolution:  comment.Resolution,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestCommentSummary(w, payload, pullRequestCommentSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestCommentTargetFlags(cmd, &host, &workspace, &repo, &prRef)
	cmd.Flags().StringVar(&body, "body", "", "Comment body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read the comment body from a file, or '-' for stdin")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")

	return cmd
}

func newPRCommentDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <comment-url-or-id>",
		Short: "Delete a pull request comment",
		Long:  "Delete a pull request comment. Humans must confirm the exact repository, pull request, and comment unless --yes is provided. Scripts and agents should use --yes together with --no-prompt. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.",
		Example: "  bb pr comment delete https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --yes\n" +
			"  bb --no-prompt pr comment delete 15 --pr 1 --repo workspace-slug/repo-slug --yes --json '*'\n" +
			"  bb pr comment delete 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --yes",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommentCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			comment, err := resolved.Client.GetPullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID)
			if err != nil {
				return err
			}

			confirmationTarget := pullRequestCommentConfirmationTarget(resolved.Target)
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("comment deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, confirmationTarget); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeletePullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID); err != nil {
				return err
			}

			payload := prCommentPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "deleted",
				Deleted:     true,
				Comment:     comment,
				Resolution:  comment.Resolution,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestCommentSummary(w, payload, pullRequestCommentSummaryOptions{Deleted: true})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestCommentTargetFlags(cmd, &host, &workspace, &repo, &prRef)
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")

	return cmd
}

func newPRCommentResolveCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string

	cmd := &cobra.Command{
		Use:   "resolve <comment-url-or-id>",
		Short: "Resolve a pull request comment thread",
		Long:  "Resolve a pull request comment thread. Bitbucket Cloud only allows resolving top-level diff comments. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.",
		Example: "  bb pr comment resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15\n" +
			"  bb pr comment resolve 15 --pr 1 --repo workspace-slug/repo-slug --json '*'\n" +
			"  bb pr comment resolve 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommentCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			resolution, err := resolved.Client.ResolvePullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID)
			if err != nil {
				return err
			}

			comment, err := resolved.Client.GetPullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID)
			if err != nil {
				return err
			}

			payload := prCommentPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "resolved",
				Comment:     comment,
				Resolution:  &resolution,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestCommentSummary(w, payload, pullRequestCommentSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestCommentTargetFlags(cmd, &host, &workspace, &repo, &prRef)

	return cmd
}

func newPRCommentReopenCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string

	cmd := &cobra.Command{
		Use:   "reopen <comment-url-or-id>",
		Short: "Reopen a pull request comment thread",
		Long:  "Reopen a previously resolved pull request comment thread. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.",
		Example: "  bb pr comment reopen https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15\n" +
			"  bb pr comment reopen 15 --pr 1 --repo workspace-slug/repo-slug --json '*'\n" +
			"  bb pr comment reopen 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommentCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			if err := resolved.Client.ReopenPullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID); err != nil {
				return err
			}

			comment, err := resolved.Client.GetPullRequestComment(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.CommentID)
			if err != nil {
				return err
			}

			payload := prCommentPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "reopened",
				Comment:     comment,
				Resolution:  comment.Resolution,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestCommentSummary(w, payload, pullRequestCommentSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestCommentTargetFlags(cmd, &host, &workspace, &repo, &prRef)

	return cmd
}

func addPullRequestCommentTargetFlags(cmd *cobra.Command, host, workspace, repo, prRef *string) {
	cmd.Flags().StringVar(host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(prRef, "pr", "", "Parent pull request as an ID or Bitbucket pull request URL; required when the comment target is a numeric ID")
}

func writePullRequestCommentSummary(w io.Writer, payload prCommentPayload, options pullRequestCommentSummaryOptions) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", payload.PullRequest); err != nil {
		return err
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintf(tw, "Comment:\t%d\n", payload.Comment.ID); err != nil {
		return err
	}
	if payload.Action != "" {
		if _, err := fmt.Fprintf(tw, "Action:\t%s\n", payload.Action); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "State:\t%s\n", pullRequestCommentState(payload.Comment, options)); err != nil {
		return err
	}
	if payload.Comment.User.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Author:\t%s\n", payload.Comment.User.DisplayName); err != nil {
			return err
		}
	}
	if payload.Comment.Inline != nil {
		if _, err := fmt.Fprintf(tw, "Path:\t%s\n", payload.Comment.Inline.Path); err != nil {
			return err
		}
		if line := pullRequestCommentLine(payload.Comment.Inline); line != "" {
			if _, err := fmt.Fprintf(tw, "Line:\t%s\n", line); err != nil {
				return err
			}
		}
	}
	if payload.Comment.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(tw, "URL:\t%s\n", payload.Comment.Links.HTML.Href); err != nil {
			return err
		}
	}
	if payload.Comment.Content.Raw != "" {
		if _, err := fmt.Fprintf(tw, "Body:\t%s\n", payload.Comment.Content.Raw); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	return writeNextStep(w, pullRequestCommentNextStep(payload))
}

func pullRequestCommentState(comment bitbucket.PullRequestComment, options pullRequestCommentSummaryOptions) string {
	switch {
	case options.Deleted || comment.Deleted:
		return "deleted"
	case comment.Resolution != nil:
		return "resolved"
	case comment.Pending:
		return "pending"
	default:
		return "open"
	}
}

func pullRequestCommentLine(inline *bitbucket.PullRequestCommentInline) string {
	if inline == nil {
		return ""
	}
	switch {
	case inline.StartTo > 0 && inline.To > 0:
		return fmt.Sprintf("%d-%d", inline.StartTo, inline.To)
	case inline.StartFrom > 0 && inline.From > 0:
		return fmt.Sprintf("%d-%d", inline.StartFrom, inline.From)
	case inline.To > 0:
		return strconv.Itoa(inline.To)
	case inline.From > 0:
		return strconv.Itoa(inline.From)
	default:
		return ""
	}
}

func pullRequestCommentNextStep(payload prCommentPayload) string {
	switch payload.Action {
	case "viewed", "deleted":
		return fmt.Sprintf("bb pr view %d --repo %s/%s", payload.PullRequest, payload.Workspace, payload.Repo)
	default:
		return fmt.Sprintf("bb pr comment view %d --pr %d --repo %s/%s", payload.Comment.ID, payload.PullRequest, payload.Workspace, payload.Repo)
	}
}

func pullRequestCommentConfirmationTarget(target resolvedPullRequestCommentTarget) string {
	return fmt.Sprintf("%s/%s#pr-%d/comment-%d", target.PRTarget.RepoTarget.Workspace, target.PRTarget.RepoTarget.Repo, target.PRTarget.ID, target.CommentID)
}
