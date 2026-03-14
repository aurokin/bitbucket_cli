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
