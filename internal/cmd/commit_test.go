package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteCommitViewSummary(t *testing.T) {
	t.Parallel()

	payload := commitViewPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit: bitbucket.RepositoryCommit{
			Hash:    "abc1234",
			Date:    "2026-03-13T00:00:00Z",
			Message: "Example commit\n\nbody",
			Summary: bitbucket.RepositoryCommitSummary{Raw: "Example commit"},
			Author:  bitbucket.RepositoryCommitAuthor{Raw: "A U Thor <author@example.com>"},
		},
	}

	var buf bytes.Buffer
	if err := writeCommitViewSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitViewSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Commit: abc1234",
		"Summary: Example commit",
		"Author: A U Thor <author@example.com>",
		"Next: bb commit diff abc1234 --repo acme/widgets --stat",
	)
}

func TestWriteCommitStatusesSummary(t *testing.T) {
	t.Parallel()

	payload := commitStatusesPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Statuses: []bitbucket.CommitStatus{
			{Key: "build", Name: "Build", State: "SUCCESSFUL", UpdatedOn: "2026-03-13T00:00:00Z"},
		},
	}

	var buf bytes.Buffer
	if err := writeCommitStatusesSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitStatusesSummary returned error: %v", err)
	}

	output := buf.String()
	assertOrderedSubstrings(t, output,
		"Repository: acme/widgets",
		"Commit: abc1234",
		"Summary: 1 successful",
	)
	if !strings.Contains(output, "build") {
		t.Fatalf("expected status table in output:\n%s", output)
	}
}

func TestWriteCommitDiffStatSummaryAndFallback(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := commitDiffPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Stats: []bitbucket.PullRequestDiffStat{
			{Status: "modified", Old: &bitbucket.PullRequestDiffRef{Path: "main.go"}, LinesAdded: 10, LinesRemoved: 2},
		},
	}
	if err := writeCommitDiffStatSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitDiffStatSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Commit: abc1234",
		"main.go",
	)

	buf.Reset()
	payload.Stats = nil
	if err := writeCommitDiffStatSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitDiffStatSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Commit: abc1234",
		"No commit diff stats found.",
		"Next: bb commit view abc1234 --repo acme/widgets",
	)
}

func TestWriteCommitReviewCommentAndReportSummaries(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reviewPayload := commitReviewPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Action:    "approved",
		Reviewer:  bitbucket.PullRequestParticipant{User: bitbucket.PullRequestActor{DisplayName: "Reviewer"}},
	}
	if err := writeCommitReviewSummary(&buf, reviewPayload); err != nil {
		t.Fatalf("writeCommitReviewSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Commit: abc1234",
		"Action: approved",
		"Reviewer: Reviewer",
		"Next: bb commit statuses abc1234 --repo acme/widgets",
	)

	buf.Reset()
	commentPayload := commitCommentPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Comment: bitbucket.CommitComment{
			ID:        15,
			Content:   bitbucket.CommitCommentBody{Raw: "Looks good"},
			User:      bitbucket.PullRequestActor{DisplayName: "Reviewer"},
			UpdatedOn: "2026-03-13T00:00:00Z",
			Inline:    &bitbucket.CommitCommentInline{Path: "main.go", To: 12},
		},
	}
	if err := writeCommitCommentSummary(&buf, commentPayload); err != nil {
		t.Fatalf("writeCommitCommentSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Comment: #15",
		"Author: Reviewer",
		"Inline: main.go:12",
		"Looks good",
		"Next: bb commit comment list abc1234 --repo acme/widgets",
	)

	buf.Reset()
	reportPayload := commitReportPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Report: bitbucket.CommitReport{
			ExternalID: "scanner-1",
			Title:      "Security scan",
			Result:     "PASSED",
			ReportType: "SECURITY",
			Reporter:   "scanner",
			UpdatedOn:  "2026-03-13T00:00:00Z",
			Details:    "all clear",
		},
	}
	if err := writeCommitReportSummary(&buf, reportPayload); err != nil {
		t.Fatalf("writeCommitReportSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Report: scanner-1",
		"Title: Security scan",
		"Reporter: scanner",
		"all clear",
		"Next: bb commit statuses abc1234 --repo acme/widgets",
	)
}

func TestWriteCommitDiffPatchSummary(t *testing.T) {
	t.Parallel()

	payload := commitDiffPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
		Commit:    "abc1234",
		Patch:     "diff --git a/main.go b/main.go\n",
	}

	var buf bytes.Buffer
	if err := writeCommitDiffPatchSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitDiffPatchSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Commit: abc1234",
		"diff --git a/main.go b/main.go",
	)
}

func TestWriteCommitCommentListSummary(t *testing.T) {
	t.Parallel()

	payload := commitCommentListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Comments: []bitbucket.CommitComment{
			{ID: 15, Content: bitbucket.CommitCommentBody{Raw: "Looks good"}, User: bitbucket.PullRequestActor{DisplayName: "Reviewer"}, CreatedOn: "2026-03-13T00:00:00Z"},
		},
	}

	var buf bytes.Buffer
	if err := writeCommitCommentListSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitCommentListSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Commit: abc1234",
		"Next: bb commit comment view 15 --commit abc1234 --repo acme/widgets",
	)
}

func TestWriteCommitReportListSummary(t *testing.T) {
	t.Parallel()

	payload := commitReportListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Commit:    "abc1234",
		Reports: []bitbucket.CommitReport{
			{ExternalID: "scanner-1", Title: "Security scan", ReportType: "SECURITY", Result: "PASSED", UpdatedOn: "2026-03-13T00:00:00Z"},
		},
	}

	var buf bytes.Buffer
	if err := writeCommitReportListSummary(&buf, payload); err != nil {
		t.Fatalf("writeCommitReportListSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Commit: abc1234",
		"Next: bb commit report view scanner-1 --commit abc1234 --repo acme/widgets",
	)
}
