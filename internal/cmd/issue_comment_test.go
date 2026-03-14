package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteIssueCommentListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := issueCommentListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Issue:     7,
		Comments: []bitbucket.IssueComment{
			{ID: 3, Content: bitbucket.IssueContent{Raw: "Needs follow-up"}, User: bitbucket.IssueActor{DisplayName: "Auro"}, UpdatedOn: "2026-03-13T00:00:00Z"},
		},
	}

	if err := writeIssueCommentListSummary(&buf, payload); err != nil {
		t.Fatalf("writeIssueCommentListSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Issue: 7",
		"3",
		"Needs follow-up",
		"Next: bb issue comment view 3 --issue 7 --repo acme/widgets",
	)
}

func TestWriteIssueCommentSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := issueCommentPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Issue:     7,
		Action:    "created",
		Comment: bitbucket.IssueComment{
			ID:        3,
			Content:   bitbucket.IssueContent{Raw: "Needs follow-up"},
			User:      bitbucket.IssueActor{DisplayName: "Auro"},
			UpdatedOn: "2026-03-13T00:00:00Z",
		},
	}

	if err := writeIssueCommentSummary(&buf, payload); err != nil {
		t.Fatalf("writeIssueCommentSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Issue: 7",
		"Comment: 3",
		"Action: created",
		"Author: Auro",
		"Body: Needs follow-up",
		"Next: bb issue comment list 7 --repo acme/widgets",
	)
}

func TestResolveIssueCommentReferenceFromIssueURL(t *testing.T) {
	configureIssueReferenceTestAuth(t)

	target, _, issueID, commentID, err := resolveIssueCommentReference("", "", "", "https://bitbucket.org/acme/widgets/issues/7", "3")
	if err != nil {
		t.Fatalf("resolveIssueCommentReference returned error: %v", err)
	}
	if issueID != 7 || commentID != 3 {
		t.Fatalf("expected issue/comment 7/3, got %d/%d", issueID, commentID)
	}
	if target.Workspace != "acme" || target.Repo != "widgets" {
		t.Fatalf("unexpected target %+v", target)
	}
}

func TestResolveIssueCommentReferenceRejectsInvalidCommentID(t *testing.T) {
	configureIssueReferenceTestAuth(t)

	_, _, _, _, err := resolveIssueCommentReference("", "", "", "https://bitbucket.org/acme/widgets/issues/7", "comment-3")
	if err == nil || !strings.Contains(err.Error(), "positive integer") {
		t.Fatalf("expected invalid comment ID error, got %v", err)
	}
}
