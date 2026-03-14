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

func TestWriteIssueMilestoneAndComponentSummaries(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeIssueMilestoneSummary(&buf, issueMilestonePayload{
		Workspace: "acme",
		Repo:      "widgets",
		Milestone: bitbucket.IssueMilestone{ID: 1, Name: "v1.0"},
	}); err != nil {
		t.Fatalf("writeIssueMilestoneSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Milestone: 1",
		"Name: v1.0",
		"Next: bb issue milestone list --repo acme/widgets",
	)

	buf.Reset()
	if err := writeIssueComponentSummary(&buf, issueComponentPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Component: bitbucket.IssueComponent{ID: 2, Name: "backend"},
	}); err != nil {
		t.Fatalf("writeIssueComponentSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Component: 2",
		"Name: backend",
		"Next: bb issue component list --repo acme/widgets",
	)
}
