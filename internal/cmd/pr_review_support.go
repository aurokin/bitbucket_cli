package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func writePRReviewSummary(w io.Writer, payload prReviewPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", payload.PullRequest); err != nil {
		return err
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintf(tw, "Action:\t%s\n", reviewActionSummaryLabel(payload.Action)); err != nil {
		return err
	}
	if payload.Reviewer.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Reviewer:\t%s\n", payload.Reviewer.DisplayName); err != nil {
			return err
		}
	}
	if payload.ReviewState != "" {
		if _, err := fmt.Fprintf(tw, "State:\t%s\n", payload.ReviewState); err != nil {
			return err
		}
	}
	if payload.Participant != nil && payload.Participant.Role != "" {
		if _, err := fmt.Fprintf(tw, "Role:\t%s\n", payload.Participant.Role); err != nil {
			return err
		}
	}
	if payload.Participant != nil && payload.Participant.ParticipatedOn != "" {
		if _, err := fmt.Fprintf(tw, "Updated:\t%s\n", formatPRUpdated(payload.Participant.ParticipatedOn)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", payload.PullRequest, payload.Workspace, payload.Repo))
}

func writePRActivitySummary(w io.Writer, payload prActivityPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n\n", payload.PullRequest); err != nil {
		return err
	}
	if len(payload.Activity) == 0 {
		if _, err := fmt.Fprintln(w, "No pull request activity found."); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", payload.PullRequest, payload.Workspace, payload.Repo))
	}

	typeWidth, actorWidth, summaryWidth := prActivityColumnWidths(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "date\ttype\tactor\tsummary"); err != nil {
		return err
	}
	for _, item := range payload.Activity {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			output.Truncate(activityTimestamp(item), 16),
			output.Truncate(activityType(item), typeWidth),
			output.Truncate(activityActor(item), actorWidth),
			output.Truncate(activitySummary(item), summaryWidth),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writePRCommitsSummary(w io.Writer, payload prCommitsPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n\n", payload.PullRequest); err != nil {
		return err
	}
	if len(payload.Commits) == 0 {
		if _, err := fmt.Fprintln(w, "No pull request commits found."); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", payload.PullRequest, payload.Workspace, payload.Repo))
	}

	hashWidth, summaryWidth, authorWidth := prCommitColumnWidths(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "hash\tsummary\tauthor\tdate"); err != nil {
		return err
	}
	for _, commit := range payload.Commits {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			output.Truncate(commit.Hash, hashWidth),
			output.Truncate(commitSummary(commit), summaryWidth),
			output.Truncate(commitAuthor(commit), authorWidth),
			formatPRUpdated(commit.Date),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writePRChecksSummary(w io.Writer, payload prChecksPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d\n", payload.PullRequest); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Summary: %s\n\n", summarizeCommitStatuses(payload.Statuses)); err != nil {
		return err
	}
	if len(payload.Statuses) == 0 {
		return writeNextStep(w, fmt.Sprintf("bb pr view %d --repo %s/%s", payload.PullRequest, payload.Workspace, payload.Repo))
	}

	nameWidth, keyWidth := prChecksColumnWidths(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "state\tname\tkey\tupdated"); err != nil {
		return err
	}
	for _, status := range payload.Statuses {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			output.Truncate(status.State, 12),
			output.Truncate(commitStatusName(status), nameWidth),
			output.Truncate(status.Key, keyWidth),
			formatPRUpdated(commitStatusUpdated(status)),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func reviewActionSummaryLabel(action string) string {
	switch action {
	case string(bitbucket.PullRequestReviewApprove):
		return "approved"
	case string(bitbucket.PullRequestReviewUnapprove):
		return "unapproved"
	case string(bitbucket.PullRequestReviewRequestChanges):
		return "requested changes"
	case string(bitbucket.PullRequestReviewClearRequestChanges):
		return "cleared requested changes"
	default:
		return action
	}
}

func activityType(item bitbucket.PullRequestActivity) string {
	switch {
	case item.Comment != nil:
		return "comment"
	case item.Update != nil:
		return "update"
	case item.Approval != nil:
		return "approval"
	case item.RequestChangesEvent() != nil:
		return "request changes"
	default:
		return "activity"
	}
}

func activityActor(item bitbucket.PullRequestActivity) string {
	switch {
	case item.Comment != nil:
		return item.Comment.User.DisplayName
	case item.Update != nil:
		return item.Update.Author.DisplayName
	case item.Approval != nil:
		return item.Approval.User.DisplayName
	case item.RequestChangesEvent() != nil:
		return item.RequestChangesEvent().User.DisplayName
	default:
		return ""
	}
}

func activityTimestamp(item bitbucket.PullRequestActivity) string {
	switch {
	case item.Comment != nil:
		return formatPRUpdated(commentUpdatedOnOrCreated(*item.Comment))
	case item.Update != nil:
		return formatPRUpdated(item.Update.Date)
	case item.Approval != nil:
		return formatPRUpdated(item.Approval.Date)
	case item.RequestChangesEvent() != nil:
		return formatPRUpdated(item.RequestChangesEvent().Date)
	default:
		return ""
	}
}

func activitySummary(item bitbucket.PullRequestActivity) string {
	switch {
	case item.Comment != nil:
		return firstLine(item.Comment.Content.Raw)
	case item.Update != nil:
		if item.Update.Title != "" && item.Update.State != "" {
			return fmt.Sprintf("%s [%s]", item.Update.Title, item.Update.State)
		}
		if item.Update.Title != "" {
			return item.Update.Title
		}
		if item.Update.State != "" {
			return item.Update.State
		}
		return "pull request updated"
	case item.Approval != nil:
		return "approved pull request"
	case item.RequestChangesEvent() != nil:
		return "requested changes"
	default:
		return "pull request activity"
	}
}

func commitSummary(commit bitbucket.RepositoryCommit) string {
	switch {
	case strings.TrimSpace(commit.Summary.Raw) != "":
		return strings.TrimSpace(commit.Summary.Raw)
	case strings.TrimSpace(commit.Message) != "":
		return firstLine(commit.Message)
	default:
		return "(no summary)"
	}
}

func commitAuthor(commit bitbucket.RepositoryCommit) string {
	switch {
	case strings.TrimSpace(commit.Author.User.DisplayName) != "":
		return strings.TrimSpace(commit.Author.User.DisplayName)
	case strings.TrimSpace(commit.Author.Raw) != "":
		return strings.TrimSpace(commit.Author.Raw)
	default:
		return ""
	}
}

func summarizeCommitStatuses(statuses []bitbucket.CommitStatus) string {
	if len(statuses) == 0 {
		return "no commit statuses"
	}

	counts := map[string]int{}
	for _, status := range statuses {
		counts[status.State]++
	}

	parts := make([]string, 0, 4)
	for _, state := range []string{"FAILED", "INPROGRESS", "SUCCESSFUL", "STOPPED"} {
		if counts[state] == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s", counts[state], strings.ToLower(state)))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%d statuses", len(statuses))
	}
	return strings.Join(parts, ", ")
}

func commitStatusName(status bitbucket.CommitStatus) string {
	if strings.TrimSpace(status.Name) != "" {
		return strings.TrimSpace(status.Name)
	}
	if strings.TrimSpace(status.Description) != "" {
		return strings.TrimSpace(status.Description)
	}
	return status.Key
}

func commitStatusUpdated(status bitbucket.CommitStatus) string {
	if strings.TrimSpace(status.UpdatedOn) != "" {
		return status.UpdatedOn
	}
	return status.CreatedOn
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		return strings.TrimSpace(value[:idx])
	}
	return value
}

func prActivityColumnWidths(termWidth int) (typ, actor, summary int) {
	switch {
	case termWidth >= 160:
		return 18, 18, 72
	case termWidth >= 132:
		return 16, 16, 52
	case termWidth >= 110:
		return 14, 14, 40
	default:
		return 12, 12, 28
	}
}

func prCommitColumnWidths(termWidth int) (hash, summary, author int) {
	switch {
	case termWidth >= 160:
		return 12, 72, 24
	case termWidth >= 132:
		return 10, 52, 18
	case termWidth >= 110:
		return 10, 40, 16
	default:
		return 8, 28, 14
	}
}

func prChecksColumnWidths(termWidth int) (name, key int) {
	switch {
	case termWidth >= 160:
		return 52, 28
	case termWidth >= 132:
		return 40, 24
	case termWidth >= 110:
		return 30, 18
	default:
		return 24, 14
	}
}

func commentUpdatedOnOrCreated(comment bitbucket.PullRequestComment) string {
	if strings.TrimSpace(comment.UpdatedOn) != "" {
		return comment.UpdatedOn
	}
	return comment.CreatedOn
}
