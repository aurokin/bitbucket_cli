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

func newPRTaskCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var body string
	var bodyFile string
	var commentRef string
	var pending bool

	cmd := &cobra.Command{
		Use:   "create <pr-id-or-url>",
		Short: "Create a task on a pull request",
		Long:  "Create a Bitbucket pull request task using --body, --body-file, or --body-file - for stdin. Use --comment to attach the task to a specific pull request comment when Bitbucket Cloud supports that linkage.",
		Example: "  bb pr task create 1 --repo workspace-slug/repo-slug --body 'Follow up on review feedback'\n" +
			"  bb pr task create 1 --repo workspace-slug/repo-slug --comment 15 --body-file task.md --json '*'\n" +
			"  bb pr task create https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Handle this thread'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			taskBody, err := resolveCommentBody(cmd.InOrStdin(), body, bodyFile)
			if err != nil {
				return err
			}

			resolvedPR, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			commentID := 0
			if strings.TrimSpace(commentRef) != "" {
				resolvedComment, err := resolvePullRequestCommentCommandTarget(context.Background(), host, workspace, repo, args[0], commentRef, true)
				if err != nil {
					return err
				}
				commentID = resolvedComment.Target.CommentID
				resolvedPR.Target = resolvedComment.Target.PRTarget
			}

			task, err := resolvedPR.Client.CreatePullRequestTask(context.Background(), resolvedPR.Target.RepoTarget.Workspace, resolvedPR.Target.RepoTarget.Repo, resolvedPR.Target.ID, bitbucket.CreatePullRequestTaskOptions{
				Body:      taskBody,
				CommentID: commentID,
				Pending:   pending,
			})
			if err != nil {
				return err
			}

			payload := prTaskPayload{
				Host:        resolvedPR.Target.RepoTarget.Host,
				Workspace:   resolvedPR.Target.RepoTarget.Workspace,
				Repo:        resolvedPR.Target.RepoTarget.Repo,
				PullRequest: resolvedPR.Target.ID,
				Action:      "created",
				Task:        task,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestTaskSummary(w, payload, pullRequestTaskSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&body, "body", "", "Task body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read the task body from a file, or '-' for stdin")
	cmd.Flags().StringVar(&commentRef, "comment", "", "Optional pull request comment as a numeric ID or Bitbucket pull request comment URL")
	cmd.Flags().BoolVar(&pending, "pending", false, "Mark the created task as pending")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")

	return cmd
}

func newPRTaskEditCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string
	var body string
	var bodyFile string

	cmd := &cobra.Command{
		Use:   "edit <task-id>",
		Short: "Edit a pull request task",
		Long:  "Edit the body of a Bitbucket pull request task. Tasks are addressed by numeric task ID together with --pr <id-or-url>.",
		Example: "  bb pr task edit 3 --pr 1 --repo workspace-slug/repo-slug --body 'Updated follow-up'\n" +
			"  bb pr task edit 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --body-file task.md --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			taskBody, err := resolveCommentBody(cmd.InOrStdin(), body, bodyFile)
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestTaskCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			task, err := resolved.Client.UpdatePullRequestTask(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.TaskID, bitbucket.UpdatePullRequestTaskOptions{
				Body: taskBody,
			})
			if err != nil {
				return err
			}

			payload := prTaskPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "edited",
				Task:        task,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestTaskSummary(w, payload, pullRequestTaskSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestTaskTargetFlags(cmd, &host, &workspace, &repo, &prRef)
	cmd.Flags().StringVar(&body, "body", "", "Task body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read the task body from a file, or '-' for stdin")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")

	return cmd
}

func newPRTaskDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <task-id>",
		Short: "Delete a pull request task",
		Long:  "Delete a Bitbucket pull request task. Humans must confirm the exact repository, pull request, and task unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb pr task delete 3 --pr 1 --repo workspace-slug/repo-slug --yes\n" +
			"  bb --no-prompt pr task delete 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --yes --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestTaskCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			task, err := resolved.Client.GetPullRequestTask(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.TaskID)
			if err != nil {
				return err
			}

			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("task deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, pullRequestTaskConfirmationTarget(resolved.Target)); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeletePullRequestTask(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.TaskID); err != nil {
				return err
			}

			payload := prTaskPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "deleted",
				Deleted:     true,
				Task:        task,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestTaskSummary(w, payload, pullRequestTaskSummaryOptions{Deleted: true})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestTaskTargetFlags(cmd, &host, &workspace, &repo, &prRef)
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")

	return cmd
}

func newPRTaskResolveCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string

	cmd := &cobra.Command{
		Use:   "resolve <task-id>",
		Short: "Resolve a pull request task",
		Long:  "Resolve a Bitbucket pull request task by updating its task state to RESOLVED. Tasks are addressed by numeric task ID together with --pr <id-or-url>.",
		Example: "  bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr task resolve 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestTaskCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			task, err := resolved.Client.UpdatePullRequestTask(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.TaskID, bitbucket.UpdatePullRequestTaskOptions{
				State: "RESOLVED",
			})
			if err != nil {
				return err
			}

			payload := prTaskPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "resolved",
				Task:        task,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestTaskSummary(w, payload, pullRequestTaskSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestTaskTargetFlags(cmd, &host, &workspace, &repo, &prRef)

	return cmd
}

func newPRTaskReopenCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string

	cmd := &cobra.Command{
		Use:   "reopen <task-id>",
		Short: "Reopen a pull request task",
		Long:  "Reopen a Bitbucket pull request task by updating its task state to UNRESOLVED. Tasks are addressed by numeric task ID together with --pr <id-or-url>.",
		Example: "  bb pr task reopen 3 --pr 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr task reopen 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestTaskCommandTarget(context.Background(), host, workspace, repo, prRef, args[0], true)
			if err != nil {
				return err
			}

			task, err := resolved.Client.UpdatePullRequestTask(context.Background(), resolved.Target.PRTarget.RepoTarget.Workspace, resolved.Target.PRTarget.RepoTarget.Repo, resolved.Target.PRTarget.ID, resolved.Target.TaskID, bitbucket.UpdatePullRequestTaskOptions{
				State: "UNRESOLVED",
			})
			if err != nil {
				return err
			}

			payload := prTaskPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "reopened",
				Task:        task,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestTaskSummary(w, payload, pullRequestTaskSummaryOptions{})
			})
		},
	}

	addFormatFlags(cmd, &flags)
	addPullRequestTaskTargetFlags(cmd, &host, &workspace, &repo, &prRef)

	return cmd
}
