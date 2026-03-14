package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func buildPRStatusPayload(target resolvedRepoTarget, currentUser bitbucket.CurrentUser, currentBranch, currentBranchError string, prs []bitbucket.PullRequest) prStatusPayload {
	payload := prStatusPayload{
		Host:               target.Host,
		Workspace:          target.Workspace,
		Repo:               target.Repo,
		Warnings:           append([]string(nil), target.Warnings...),
		CurrentUser:        currentUser,
		CurrentBranchName:  currentBranch,
		CurrentBranchError: currentBranchError,
		Created:            make([]bitbucket.PullRequest, 0),
		ReviewRequested:    make([]bitbucket.PullRequest, 0),
	}

	currentBranchID := 0
	for i := range prs {
		pr := prs[i]
		if payload.CurrentBranch == nil && currentBranch != "" && pr.Source.Branch.Name == currentBranch {
			prCopy := pr
			payload.CurrentBranch = &prCopy
			currentBranchID = pr.ID
			continue
		}
	}

	for _, pr := range prs {
		if currentBranchID != 0 && pr.ID == currentBranchID {
			continue
		}
		if sameActor(currentUser, pr.Author) {
			payload.Created = append(payload.Created, pr)
			continue
		}
		if reviewRequestedFromUser(currentUser, pr) {
			payload.ReviewRequested = append(payload.ReviewRequested, pr)
		}
	}

	return payload
}

func writePRStatusSummary(w io.Writer, payload prStatusPayload) error {
	if err := writePRStatusHeader(w, payload); err != nil {
		return err
	}
	currentBranchPRs := make([]bitbucket.PullRequest, 0, 1)
	if payload.CurrentBranch != nil {
		currentBranchPRs = append(currentBranchPRs, *payload.CurrentBranch)
	}
	if err := writePRStatusSectionWithTitle(w, "Current Branch Pull Request", currentBranchPRs...); err != nil {
		return err
	}
	if err := writePRStatusSectionWithTitle(w, "Created By You", payload.Created...); err != nil {
		return err
	}
	if err := writePRStatusSectionWithTitle(w, "Review Requested", payload.ReviewRequested...); err != nil {
		return err
	}
	if len(payload.Created) == 0 && len(payload.ReviewRequested) == 0 && payload.CurrentBranch == nil {
		return writeNextStep(w, fmt.Sprintf("bb pr list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	return nil
}

func writePRStatusHeader(w io.Writer, payload prStatusPayload) error {
	if _, err := fmt.Fprintf(w, "Repository: %s/%s\n", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	currentBranchLabel := payload.CurrentBranchName
	if currentBranchLabel == "" {
		currentBranchLabel = "unavailable"
	}
	if _, err := fmt.Fprintf(w, "Current Branch: %s\n", currentBranchLabel); err != nil {
		return err
	}
	if payload.CurrentBranchError != "" {
		if err := writeLabelValue(w, "Current Branch Error", payload.CurrentBranchError); err != nil {
			return err
		}
	}
	return nil
}

func writePRStatusSectionWithTitle(w io.Writer, title string, prs ...bitbucket.PullRequest) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, title); err != nil {
		return err
	}
	return writePRStatusSection(w, prs...)
}

func writePRStatusSection(w io.Writer, prs ...bitbucket.PullRequest) error {
	if len(prs) == 0 {
		_, err := fmt.Fprintln(w, "  None.")
		return err
	}

	for _, pr := range prs {
		line := fmt.Sprintf("  #%d  %s [%s] %s -> %s", pr.ID, pr.Title, pr.State, pr.Source.Branch.Name, pr.Destination.Branch.Name)
		if pr.TaskCount > 0 {
			line += fmt.Sprintf("  tasks:%d", pr.TaskCount)
		}
		if pr.CommentCount > 0 {
			line += fmt.Sprintf("  comments:%d", pr.CommentCount)
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

func writePRDiffStatSummary(w io.Writer, payload prDiffPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pull Request: #%d %s\n\n", payload.ID, payload.Title); err != nil {
		return err
	}
	return writePRDiffStatTable(w, payload.Stats)
}

func writePRDiffStatTable(w io.Writer, stats []bitbucket.PullRequestDiffStat) error {
	if len(stats) == 0 {
		_, err := fmt.Fprintln(w, "No changed files.")
		return err
	}

	pathWidth := diffStatPathWidth(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "status\tfile\t+add\t-rem"); err != nil {
		return err
	}

	totalAdded := 0
	totalRemoved := 0
	for _, stat := range stats {
		totalAdded += stat.LinesAdded
		totalRemoved += stat.LinesRemoved
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%d\n", output.Truncate(diffStatus(stat), 10), output.TruncateMiddle(diffPath(stat), pathWidth), stat.LinesAdded, stat.LinesRemoved); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "total\t%d files\t%d\t%d\n", len(stats), totalAdded, totalRemoved); err != nil {
		return err
	}

	return tw.Flush()
}

func writePRListSummary(w io.Writer, target resolvedRepoTarget, prs []bitbucket.PullRequest) error {
	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	if len(prs) == 0 {
		if _, err := fmt.Fprintf(w, "No pull requests found for %s/%s.\n", target.Workspace, target.Repo); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pr create --repo %s/%s --title '<title>'", target.Workspace, target.Repo))
	}
	return writePRListTable(w, prs)
}

func writePRListTable(w io.Writer, prs []bitbucket.PullRequest) error {
	titleWidth, authorWidth, branchWidth := prListColumnWidths(output.TerminalWidth(w))

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "#\ttitle\tstate\tauthor\tsrc\tdst\ttsk\tcmt\tupdated"); err != nil {
		return err
	}

	for _, pr := range prs {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
			pr.ID,
			output.Truncate(pr.Title, titleWidth),
			output.Truncate(pr.State, 10),
			output.Truncate(pr.Author.DisplayName, authorWidth),
			output.TruncateMiddle(pr.Source.Branch.Name, branchWidth),
			output.TruncateMiddle(pr.Destination.Branch.Name, branchWidth),
			pr.TaskCount,
			pr.CommentCount,
			formatPRUpdated(pr.UpdatedOn),
		); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func writePRViewSummary(w io.Writer, target resolvedPullRequestTarget, pr bitbucket.PullRequest) error {
	if err := writeTargetHeader(w, "Repository", target.RepoTarget.Workspace, target.RepoTarget.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.RepoTarget.Warnings); err != nil {
		return err
	}
	if err := writePullRequestSummaryTable(w, pr, pullRequestSummaryOptions{
		IncludeAuthor:      true,
		IncludeUpdated:     true,
		IncludeDescription: true,
	}); err != nil {
		return err
	}
	return writeNextStep(w, prViewNextStep(target.RepoTarget.Workspace, target.RepoTarget.Repo, pr.ID))
}

func formatPRUpdated(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}

	return parsed.Local().Format("2006-01-02 15:04")
}

func prListColumnWidths(termWidth int) (title, author, branch int) {
	switch {
	case termWidth >= 160:
		return 52, 18, 24
	case termWidth >= 132:
		return 40, 16, 18
	case termWidth >= 110:
		return 32, 14, 14
	default:
		return 24, 12, 12
	}
}

func diffStatPathWidth(termWidth int) int {
	switch {
	case termWidth >= 160:
		return 72
	case termWidth >= 132:
		return 56
	case termWidth >= 110:
		return 44
	default:
		return 32
	}
}

func diffPath(stat bitbucket.PullRequestDiffStat) string {
	switch {
	case stat.New != nil && stat.Old != nil && stat.New.Path != "" && stat.Old.Path != "" && stat.New.Path != stat.Old.Path:
		return stat.Old.Path + " -> " + stat.New.Path
	case stat.New != nil && stat.New.Path != "":
		return stat.New.Path
	case stat.Old != nil && stat.Old.Path != "":
		return stat.Old.Path
	default:
		return "(unknown)"
	}
}

func diffStatus(stat bitbucket.PullRequestDiffStat) string {
	status := strings.TrimSpace(stat.Status)
	if status == "" {
		return "changed"
	}
	return status
}

type pullRequestSummaryOptions struct {
	IncludeAuthor      bool
	IncludeUpdated     bool
	IncludeDescription bool
	Strategy           string
	MergeCommit        string
}

type summaryRow struct {
	Label string
	Value string
}

func writePullRequestSummaryTable(w io.Writer, pr bitbucket.PullRequest, options pullRequestSummaryOptions) error {
	tw := output.NewTableWriter(w)
	for _, row := range pullRequestSummaryRows(pr, options) {
		if _, err := fmt.Fprintf(tw, "%s:\t%s\n", row.Label, row.Value); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func pullRequestSummaryRows(pr bitbucket.PullRequest, options pullRequestSummaryOptions) []summaryRow {
	rows := []summaryRow{
		{Label: "ID", Value: fmt.Sprintf("%d", pr.ID)},
		{Label: "Title", Value: pr.Title},
		{Label: "State", Value: pr.State},
		{Label: "Source", Value: pr.Source.Branch.Name},
		{Label: "Destination", Value: pr.Destination.Branch.Name},
	}
	if options.IncludeAuthor {
		rows = appendSummaryRow(rows, "Author", pr.Author.DisplayName)
	}
	rows = appendSummaryRow(rows, "Strategy", options.Strategy)
	if pr.TaskCount > 0 {
		rows = append(rows, summaryRow{Label: "Tasks", Value: fmt.Sprintf("%d", pr.TaskCount)})
	}
	if pr.CommentCount > 0 {
		rows = append(rows, summaryRow{Label: "Comments", Value: fmt.Sprintf("%d", pr.CommentCount)})
	}
	if options.IncludeUpdated {
		rows = appendSummaryRow(rows, "Updated", pr.UpdatedOn)
	}
	rows = appendSummaryRow(rows, "Merge Commit", options.MergeCommit)
	rows = appendSummaryRow(rows, "URL", pr.Links.HTML.Href)
	if options.IncludeDescription {
		rows = appendSummaryRow(rows, "Description", pr.Description)
	}
	return rows
}

func appendSummaryRow(rows []summaryRow, label, value string) []summaryRow {
	if strings.TrimSpace(value) == "" {
		return rows
	}
	return append(rows, summaryRow{Label: label, Value: value})
}
