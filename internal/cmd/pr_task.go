package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type prTaskPayload struct {
	Host        string                    `json:"host"`
	Workspace   string                    `json:"workspace"`
	Repo        string                    `json:"repo"`
	PullRequest int                       `json:"pull_request"`
	Action      string                    `json:"action,omitempty"`
	Deleted     bool                      `json:"deleted,omitempty"`
	Task        bitbucket.PullRequestTask `json:"task"`
}

type prTaskListPayload struct {
	Host        string                      `json:"host"`
	Workspace   string                      `json:"workspace"`
	Repo        string                      `json:"repo"`
	PullRequest int                         `json:"pull_request"`
	State       string                      `json:"state"`
	Tasks       []bitbucket.PullRequestTask `json:"tasks"`
}

type pullRequestTaskSummaryOptions struct {
	Deleted bool
}

func newPRTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Work with pull request tasks",
		Long:  "List, inspect, create, edit, delete, resolve, and reopen Bitbucket pull request tasks. Tasks can be attached to specific pull request comments when the Bitbucket Cloud REST API supports it.",
		Example: "  bb pr task list 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr task create 1 --repo workspace-slug/repo-slug --body 'Follow up on reviewer feedback'\n" +
			"  bb pr task create 1 --repo workspace-slug/repo-slug --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Handle this thread'\n" +
			"  bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug",
	}

	taskCmd.AddCommand(
		newPRTaskListCmd(),
		newPRTaskViewCmd(),
		newPRTaskCreateCmd(),
		newPRTaskEditCmd(),
		newPRTaskDeleteCmd(),
		newPRTaskResolveCmd(),
		newPRTaskReopenCmd(),
	)

	return taskCmd
}

func newPRTaskListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var state string
	var limit int

	cmd := &cobra.Command{
		Use:   "list <pr-id-or-url>",
		Short: "List tasks on a pull request",
		Long:  "List tasks on a Bitbucket pull request. Accepts a numeric pull request ID or Bitbucket pull request URL. Defaults to unresolved tasks; pass --state all to see everything.",
		Example: "  bb pr task list 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr task list https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --state all --json '*'\n" +
			"  bb pr task list 1 --repo workspace-slug/repo-slug --state resolved --limit 50",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolvedState, err := normalizePullRequestTaskListState(state)
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			tasks, err := resolved.Client.ListPullRequestTasks(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.ID, bitbucket.ListPullRequestTasksOptions{
				State: resolvedState,
				Limit: limit,
			})
			if err != nil {
				return err
			}

			payload := prTaskListPayload{
				Host:        resolved.Target.RepoTarget.Host,
				Workspace:   resolved.Target.RepoTarget.Workspace,
				Repo:        resolved.Target.RepoTarget.Repo,
				PullRequest: resolved.Target.ID,
				State:       resolvedState,
				Tasks:       tasks,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePullRequestTaskListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&state, "state", "unresolved", "Task state filter: unresolved, resolved, or all")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of tasks to list")

	return cmd
}

func newPRTaskViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var prRef string

	cmd := &cobra.Command{
		Use:   "view <task-id>",
		Short: "View a pull request task",
		Long:  "View a specific Bitbucket pull request task. Tasks are addressed by numeric task ID together with --pr <id-or-url>.",
		Example: "  bb pr task view 3 --pr 1 --repo workspace-slug/repo-slug\n" +
			"  bb pr task view 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'",
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

			payload := prTaskPayload{
				Host:        resolved.Target.PRTarget.RepoTarget.Host,
				Workspace:   resolved.Target.PRTarget.RepoTarget.Workspace,
				Repo:        resolved.Target.PRTarget.RepoTarget.Repo,
				PullRequest: resolved.Target.PRTarget.ID,
				Action:      "viewed",
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

func addPullRequestTaskTargetFlags(cmd *cobra.Command, host, workspace, repo, prRef *string) {
	cmd.Flags().StringVar(host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(prRef, "pr", "", "Parent pull request as an ID or Bitbucket pull request URL")
	_ = cmd.MarkFlagRequired("pr")
}

func normalizePullRequestTaskListState(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "", "OPEN", "UNRESOLVED":
		return "UNRESOLVED", nil
	case "RESOLVED":
		return "RESOLVED", nil
	case "ALL":
		return "ALL", nil
	default:
		return "", fmt.Errorf("invalid task state %q; expected unresolved, resolved, or all", raw)
	}
}

func pullRequestTaskState(task bitbucket.PullRequestTask, options pullRequestTaskSummaryOptions) string {
	if options.Deleted {
		return "deleted"
	}
	switch strings.ToUpper(strings.TrimSpace(task.State)) {
	case "RESOLVED":
		return "resolved"
	case "UNRESOLVED":
		if task.Pending {
			return "pending"
		}
		return "open"
	default:
		if task.Pending {
			return "pending"
		}
		return "open"
	}
}

func writePullRequestTaskListSummary(w io.Writer, payload prTaskListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", payload.PullRequest); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Filter", strings.ToLower(payload.State)); err != nil {
		return err
	}
	if len(payload.Tasks) == 0 {
		if _, err := fmt.Fprintln(w, "Tasks: None."); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pr task create %d --repo %s/%s --body '<task body>'", payload.PullRequest, payload.Workspace, payload.Repo))
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "ID\tSTATE\tCOMMENT\tBODY"); err != nil {
		return err
	}
	for _, task := range payload.Tasks {
		comment := ""
		if task.Comment != nil && task.Comment.ID > 0 {
			comment = "#" + strconv.Itoa(task.Comment.ID)
		}
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", task.ID, pullRequestTaskState(task, pullRequestTaskSummaryOptions{}), comment, task.Content.Raw); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	return writeNextStep(w, fmt.Sprintf("bb pr task view %d --pr %d --repo %s/%s", payload.Tasks[0].ID, payload.PullRequest, payload.Workspace, payload.Repo))
}

func writePullRequestTaskSummary(w io.Writer, payload prTaskPayload, options pullRequestTaskSummaryOptions) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", payload.PullRequest); err != nil {
		return err
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintf(tw, "Task:\t%d\n", payload.Task.ID); err != nil {
		return err
	}
	if payload.Action != "" {
		if _, err := fmt.Fprintf(tw, "Action:\t%s\n", payload.Action); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "State:\t%s\n", pullRequestTaskState(payload.Task, options)); err != nil {
		return err
	}
	if payload.Task.Creator.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Author:\t%s\n", payload.Task.Creator.DisplayName); err != nil {
			return err
		}
	}
	if payload.Task.Content.Raw != "" {
		if _, err := fmt.Fprintf(tw, "Body:\t%s\n", payload.Task.Content.Raw); err != nil {
			return err
		}
	}
	if payload.Task.Comment != nil && payload.Task.Comment.ID > 0 {
		if _, err := fmt.Fprintf(tw, "Comment:\t%d\n", payload.Task.Comment.ID); err != nil {
			return err
		}
		if payload.Task.Comment.Links.HTML.Href != "" {
			if _, err := fmt.Fprintf(tw, "Comment URL:\t%s\n", payload.Task.Comment.Links.HTML.Href); err != nil {
				return err
			}
		}
	}
	if payload.Task.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(tw, "URL:\t%s\n", payload.Task.Links.HTML.Href); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	return writeNextStep(w, pullRequestTaskNextStep(payload, options))
}

func pullRequestTaskNextStep(payload prTaskPayload, options pullRequestTaskSummaryOptions) string {
	repoTarget := payload.Workspace + "/" + payload.Repo
	switch {
	case options.Deleted:
		return fmt.Sprintf("bb pr task list %d --repo %s", payload.PullRequest, repoTarget)
	case strings.EqualFold(payload.Task.State, "RESOLVED"):
		return fmt.Sprintf("bb pr task reopen %d --pr %d --repo %s", payload.Task.ID, payload.PullRequest, repoTarget)
	default:
		return fmt.Sprintf("bb pr task resolve %d --pr %d --repo %s", payload.Task.ID, payload.PullRequest, repoTarget)
	}
}

func pullRequestTaskConfirmationTarget(target resolvedPullRequestTaskTarget) string {
	return fmt.Sprintf("%s/%s#pr-%d/task-%d", target.PRTarget.RepoTarget.Workspace, target.PRTarget.RepoTarget.Repo, target.PRTarget.ID, target.TaskID)
}
