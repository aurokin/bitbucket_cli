package cmd

import (
	"fmt"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
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
