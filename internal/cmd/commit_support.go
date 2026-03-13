package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func writeCommitViewSummary(w io.Writer, payload commitViewPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit.Hash); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Summary", commitSummary(payload.Commit)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Author", commitAuthor(payload.Commit)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Date", formatPRUpdated(payload.Commit.Date)); err != nil {
		return err
	}
	if message := strings.TrimSpace(payload.Commit.Message); message != "" {
		if err := writeLabelValue(w, "Message", firstLine(message)); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb commit diff %s --repo %s/%s --stat", payload.Commit.Hash, payload.Workspace, payload.Repo))
}

func writeCommitDiffStatSummary(w io.Writer, payload commitDiffPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	if len(payload.Stats) == 0 {
		if _, err := fmt.Fprintln(w, "No commit diff stats found."); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb commit view %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
	}
	return writePRDiffStatTable(w, payload.Stats)
}

func writeCommitStatusesSummary(w io.Writer, payload commitStatusesPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Summary", summarizeCommitStatuses(payload.Statuses)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if len(payload.Statuses) == 0 {
		return writeNextStep(w, fmt.Sprintf("bb commit view %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
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

func writeCommitReviewSummary(w io.Writer, payload commitReviewPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if reviewer := strings.TrimSpace(payload.Reviewer.User.DisplayName); reviewer != "" {
		if err := writeLabelValue(w, "Reviewer", reviewer); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb commit statuses %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
}

func writeCommitCommentListSummary(w io.Writer, payload commitCommentListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	if len(payload.Comments) == 0 {
		if _, err := fmt.Fprintln(w, "No commit comments found."); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb commit view %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
	}

	bodyWidth := output.TerminalWidth(w) - 48
	if bodyWidth < 24 {
		bodyWidth = 24
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tauthor\tupdated\tbody"); err != nil {
		return err
	}
	for _, comment := range payload.Comments {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\n",
			comment.ID,
			output.Truncate(comment.User.DisplayName, 20),
			formatPRUpdated(commitCommentUpdatedOnOrCreated(comment)),
			output.Truncate(strings.TrimSpace(comment.Content.Raw), bodyWidth),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb commit comment view %d --commit %s --repo %s/%s", payload.Comments[0].ID, payload.Commit, payload.Workspace, payload.Repo))
}

func writeCommitCommentSummary(w io.Writer, payload commitCommentPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Comment", fmt.Sprintf("#%d", payload.Comment.ID)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Author", payload.Comment.User.DisplayName); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Updated", formatPRUpdated(commitCommentUpdatedOnOrCreated(payload.Comment))); err != nil {
		return err
	}
	if payload.Comment.Inline != nil && strings.TrimSpace(payload.Comment.Inline.Path) != "" {
		location := payload.Comment.Inline.Path
		if payload.Comment.Inline.To > 0 {
			location += fmt.Sprintf(":%d", payload.Comment.Inline.To)
		} else if payload.Comment.Inline.From > 0 {
			location += fmt.Sprintf(":%d", payload.Comment.Inline.From)
		}
		if err := writeLabelValue(w, "Inline", location); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "\n%s\n", strings.TrimSpace(payload.Comment.Content.Raw)); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb commit comment list %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
}

func writeCommitReportListSummary(w io.Writer, payload commitReportListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	if len(payload.Reports) == 0 {
		if _, err := fmt.Fprintln(w, "No commit reports found."); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb commit statuses %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tresult\ttype\ttitle\tupdated"); err != nil {
		return err
	}
	for _, report := range payload.Reports {
		identifier := report.ExternalID
		if identifier == "" {
			identifier = report.UUID
		}
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			output.Truncate(identifier, 20),
			output.Truncate(report.Result, 8),
			output.Truncate(report.ReportType, 8),
			output.Truncate(report.Title, 40),
			formatPRUpdated(reportUpdatedOnOrCreated(report)),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	firstID := payload.Reports[0].ExternalID
	if firstID == "" {
		firstID = payload.Reports[0].UUID
	}
	return writeNextStep(w, fmt.Sprintf("bb commit report view %s --commit %s --repo %s/%s", firstID, payload.Commit, payload.Workspace, payload.Repo))
}

func writeCommitReportSummary(w io.Writer, payload commitReportPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
		return err
	}
	reportID := payload.Report.ExternalID
	if reportID == "" {
		reportID = payload.Report.UUID
	}
	if err := writeLabelValue(w, "Report", reportID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Title", payload.Report.Title); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Result", payload.Report.Result); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Type", payload.Report.ReportType); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Reporter", payload.Report.Reporter); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Updated", formatPRUpdated(reportUpdatedOnOrCreated(payload.Report))); err != nil {
		return err
	}
	if details := strings.TrimSpace(payload.Report.Details); details != "" {
		if _, err := fmt.Fprintf(w, "\n%s\n", details); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb commit statuses %s --repo %s/%s", payload.Commit, payload.Workspace, payload.Repo))
}

func commitCommentUpdatedOnOrCreated(comment bitbucket.CommitComment) string {
	if strings.TrimSpace(comment.UpdatedOn) != "" {
		return comment.UpdatedOn
	}
	return comment.CreatedOn
}

func reportUpdatedOnOrCreated(report bitbucket.CommitReport) string {
	if strings.TrimSpace(report.UpdatedOn) != "" {
		return report.UpdatedOn
	}
	return report.CreatedOn
}
