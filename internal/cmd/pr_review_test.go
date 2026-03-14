package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWritePRReviewSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prReviewPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		Action:      string(bitbucket.PullRequestReviewRequestChanges),
		Reviewer:    bitbucket.PullRequestActor{DisplayName: "Reviewer"},
		ReviewState: "changes_requested",
		Participant: &bitbucket.PullRequestParticipant{
			Role:           "REVIEWER",
			ParticipatedOn: "2026-03-13T00:00:00Z",
		},
	}

	if err := writePRReviewSummary(&buf, payload); err != nil {
		t.Fatalf("writePRReviewSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Action:",
		"requested changes",
		"Reviewer:",
		"Reviewer",
		"State:",
		"changes_requested",
		"Next: bb pr view 7 --repo acme/widgets",
	)
}

func TestWritePRActivitySummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prActivityPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		Warnings:    []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
		PullRequest: 7,
		Activity: []bitbucket.PullRequestActivity{
			{
				Comment: &bitbucket.PullRequestComment{
					Content:   bitbucket.PullRequestCommentContent{Raw: "Needs a regression test\n\nPlease add one."},
					User:      bitbucket.PullRequestActor{DisplayName: "Reviewer"},
					CreatedOn: "2026-03-13T00:00:00Z",
				},
			},
			{
				Approval: &bitbucket.PullRequestActivityEvent{
					Date: "2026-03-13T01:00:00Z",
					User: bitbucket.PullRequestActor{DisplayName: "Maintainer"},
				},
			},
		},
	}

	if err := writePRActivitySummary(&buf, payload); err != nil {
		t.Fatalf("writePRActivitySummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Pull Request: #7",
		"comment",
		"approval",
		"Needs a regression test",
		"approved pull request",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWritePRCommitsSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prCommitsPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		Commits: []bitbucket.RepositoryCommit{
			{
				Hash:    "abc123456789",
				Date:    "2026-03-13T00:00:00Z",
				Summary: bitbucket.RepositoryCommitSummary{Raw: "Add regression coverage"},
				Author:  bitbucket.RepositoryCommitAuthor{User: bitbucket.PullRequestActor{DisplayName: "Reviewer"}},
			},
		},
	}

	if err := writePRCommitsSummary(&buf, payload); err != nil {
		t.Fatalf("writePRCommitsSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Pull Request: #7",
		"abc123",
		"Add regression coverage",
		"Reviewer",
		"2026-03-",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWritePRChecksSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prChecksPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		Statuses: []bitbucket.CommitStatus{
			{State: "FAILED", Name: "Deploy", Key: "deploy", UpdatedOn: "2026-03-13T01:00:00Z"},
			{State: "SUCCESSFUL", Name: "Build", Key: "build", UpdatedOn: "2026-03-13T00:00:00Z"},
		},
	}

	if err := writePRChecksSummary(&buf, payload); err != nil {
		t.Fatalf("writePRChecksSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Summary: 1 failed, 1 successful",
		"FAILED",
		"Deploy",
		"SUCCESSFUL",
		"Build",
	)
}

func TestSummarizeCommitStatuses(t *testing.T) {
	t.Parallel()

	got := summarizeCommitStatuses([]bitbucket.CommitStatus{
		{State: "FAILED"},
		{State: "SUCCESSFUL"},
		{State: "SUCCESSFUL"},
	})
	if got != "1 failed, 2 successful" {
		t.Fatalf("unexpected status summary %q", got)
	}
}

func TestReviewStateForAction(t *testing.T) {
	t.Parallel()

	if got := reviewStateForAction(bitbucket.PullRequestReviewApprove, bitbucket.PullRequestParticipant{State: "APPROVED"}); got != "APPROVED" {
		t.Fatalf("expected explicit participant state, got %q", got)
	}
	if got := reviewStateForAction(bitbucket.PullRequestReviewApprove, bitbucket.PullRequestParticipant{Approved: true}); got != "approved" {
		t.Fatalf("expected approved fallback, got %q", got)
	}
	if got := reviewStateForAction(bitbucket.PullRequestReviewRequestChanges, bitbucket.PullRequestParticipant{}); got != "changes_requested" {
		t.Fatalf("expected changes_requested, got %q", got)
	}
	if got := reviewStateForAction(bitbucket.PullRequestReviewClearRequestChanges, bitbucket.PullRequestParticipant{}); got != "changes_request_cleared" {
		t.Fatalf("expected changes_request_cleared, got %q", got)
	}
}
