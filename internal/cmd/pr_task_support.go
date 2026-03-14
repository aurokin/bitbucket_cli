package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

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
