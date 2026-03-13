package cmd

import (
	"bytes"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteIssueMilestoneListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := issueMilestoneListPayload{
		Workspace:  "acme",
		Repo:       "widgets",
		Milestones: []bitbucket.IssueMilestone{{ID: 1, Name: "v1.0"}},
	}
	if err := writeIssueMilestoneListSummary(&buf, payload); err != nil {
		t.Fatalf("writeIssueMilestoneListSummary returned error: %v", err)
	}
	got := buf.String()
	assertOrderedSubstrings(t, got, "Repository: acme/widgets", "1", "v1.0", "Next: bb issue milestone view 1 --repo acme/widgets")
}

func TestWriteIssueComponentListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := issueComponentListPayload{
		Workspace:  "acme",
		Repo:       "widgets",
		Components: []bitbucket.IssueComponent{{ID: 2, Name: "backend"}},
	}
	if err := writeIssueComponentListSummary(&buf, payload); err != nil {
		t.Fatalf("writeIssueComponentListSummary returned error: %v", err)
	}
	got := buf.String()
	assertOrderedSubstrings(t, got, "Repository: acme/widgets", "2", "backend", "Next: bb issue component view 2 --repo acme/widgets")
}
