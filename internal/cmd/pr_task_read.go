package cmd

import (
	"context"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

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
