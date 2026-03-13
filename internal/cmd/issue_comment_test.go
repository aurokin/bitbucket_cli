package cmd

import (
	"bytes"
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
