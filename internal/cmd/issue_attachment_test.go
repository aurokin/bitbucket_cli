package cmd

import (
	"bytes"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestFilterIssueAttachmentsByName(t *testing.T) {
	t.Parallel()

	attachments := []bitbucket.IssueAttachment{
		{Name: "trace.txt"},
		{Name: "screenshot.png"},
		{Name: "ignored.log"},
	}

	filtered := filterIssueAttachmentsByName(attachments, []string{"./trace.txt", "/tmp/screenshot.png"})
	if len(filtered) != 2 {
		t.Fatalf("expected 2 attachments, got %+v", filtered)
	}
	if filtered[0].Name != "trace.txt" || filtered[1].Name != "screenshot.png" {
		t.Fatalf("unexpected filtered attachments %+v", filtered)
	}
}

func TestWriteIssueAttachmentListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := issueAttachmentListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Issue:     7,
		Attachments: []bitbucket.IssueAttachment{
			{Name: "trace.txt", Links: bitbucket.IssueAttachmentLinks{Self: bitbucket.IssueAttachmentLink{Href: "https://example.invalid/trace.txt"}}},
		},
	}

	if err := writeIssueAttachmentListSummary(&buf, payload); err != nil {
		t.Fatalf("writeIssueAttachmentListSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Issue: 7",
		"trace.txt",
		"https://example.invalid/trace.txt",
		"Next: bb issue view 7 --repo acme/widgets",
	)
}

func TestWriteIssueAttachmentUploadSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := issueAttachmentUploadPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Issue:     7,
		Action:    "uploaded",
		Files:     []string{"trace.txt", "screenshot.png"},
	}

	if err := writeIssueAttachmentUploadSummary(&buf, payload); err != nil {
		t.Fatalf("writeIssueAttachmentUploadSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Issue: 7",
		"Action: uploaded",
		"Files: trace.txt, screenshot.png",
		"Next: bb issue attachment list 7 --repo acme/widgets",
	)
}
