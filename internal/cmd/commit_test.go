package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/spf13/cobra"
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

func TestBuildCommitDiffPayloadStatAndPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/2.0/repositories/acme/widgets/diffstat/abc1234":
			if got := r.URL.Query().Get("pagelen"); got != "100" {
				t.Fatalf("expected diffstat pagelen=100, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"values":[{"status":"modified","old":{"path":"main.go"},"lines_added":2,"lines_removed":1}]}`))
		case r.URL.Path == "/2.0/repositories/acme/widgets/diff/abc1234":
			if got := r.URL.RawQuery; got != "binary=false&context=3&ignore_whitespace=true&path=main.go&renames=false" {
				t.Fatalf("unexpected diff query %q", got)
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("diff --git a/main.go b/main.go\n"))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := bitbucket.NewClient("bitbucket.org", config.HostConfig{
		Username: "agent@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resolved := resolvedCommitCommandTarget{
		Client: client,
		Target: resolvedCommitTarget{
			RepoTarget: resolvedRepoTarget{
				Host:      "bitbucket.org",
				Workspace: "acme",
				Repo:      "widgets",
				Warnings:  []string{"warning"},
			},
			Commit: "abc1234",
		},
	}

	statPayload, err := buildCommitDiffPayload(context.Background(), &cobra.Command{}, resolved, true, 0, nil, false, true, true)
	if err != nil {
		t.Fatalf("buildCommitDiffPayload stat returned error: %v", err)
	}
	if len(statPayload.Stats) != 1 || statPayload.Stats[0].Old == nil || statPayload.Stats[0].Old.Path != "main.go" {
		t.Fatalf("unexpected stat payload %+v", statPayload)
	}
	if len(statPayload.Warnings) != 1 || statPayload.Warnings[0] != "warning" {
		t.Fatalf("expected warnings to be preserved, got %+v", statPayload.Warnings)
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("binary", true, "")
	cmd.Flags().Bool("renames", true, "")
	if err := cmd.Flags().Set("binary", "false"); err != nil {
		t.Fatalf("set binary flag: %v", err)
	}
	if err := cmd.Flags().Set("renames", "false"); err != nil {
		t.Fatalf("set renames flag: %v", err)
	}

	patchPayload, err := buildCommitDiffPayload(context.Background(), cmd, resolved, false, 3, []string{"main.go"}, true, false, false)
	if err != nil {
		t.Fatalf("buildCommitDiffPayload patch returned error: %v", err)
	}
	if !strings.Contains(patchPayload.Patch, "diff --git a/main.go b/main.go") {
		t.Fatalf("unexpected patch payload %+v", patchPayload)
	}
}
