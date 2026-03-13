package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestParsePositiveInt(t *testing.T) {
	t.Parallel()

	value, err := parsePositiveInt("issue", "7")
	if err != nil {
		t.Fatalf("parsePositiveInt returned error: %v", err)
	}
	if value != 7 {
		t.Fatalf("expected 7, got %d", value)
	}

	if _, err := parsePositiveInt("issue", "abc"); err == nil || !strings.Contains(err.Error(), "positive integer") {
		t.Fatalf("expected positive integer error, got %v", err)
	}
}

func TestWriteIssueTable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	issues := []bitbucket.Issue{
		{
			ID:        1,
			Title:     "A very long issue title that should be compact in the table output",
			State:     "new",
			UpdatedOn: "2026-03-11T00:00:00Z",
			Reporter: bitbucket.IssueActor{
				DisplayName: "Example User With Long Name",
			},
		},
	}

	if err := writeIssueTable(&buf, issues); err != nil {
		t.Fatalf("writeIssueTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "#") || !strings.Contains(got, "title") {
		t.Fatalf("expected issue table header, got %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("expected truncated output, got %q", got)
	}
}

func TestWriteIssueMutationSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	issue := bitbucket.Issue{
		ID:    12,
		Title: "Fixture issue",
		State: "new",
		Links: bitbucket.IssueLinks{
			HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/issues/12"},
		},
	}

	if err := writeIssueMutationSummary(&buf, "Created", "acme", "widgets", issue, true); err != nil {
		t.Fatalf("writeIssueMutationSummary returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Created issue acme/widgets#12") {
		t.Fatalf("expected repo-scoped issue summary, got %q", got)
	}
	if !strings.Contains(got, "URL: https://bitbucket.org/acme/widgets/issues/12") {
		t.Fatalf("expected issue URL in summary, got %q", got)
	}
	if !strings.Contains(got, "Next: bb issue view 12 --repo acme/widgets") {
		t.Fatalf("expected next-step guidance, got %q", got)
	}
}

func TestIssueNextSteps(t *testing.T) {
	t.Parallel()

	if got := issueListEmptyNextStep("acme", "widgets"); got != "bb issue create --repo acme/widgets --title '<title>'" {
		t.Fatalf("unexpected issue list next step %q", got)
	}
	if got := issueViewNextStep("acme", "widgets", 12); got != "bb issue edit 12 --repo acme/widgets" {
		t.Fatalf("unexpected issue view next step %q", got)
	}
}

func TestWriteIssueTableWithRepositoryHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	issues := []bitbucket.Issue{
		{
			ID:        1,
			Title:     "Fixture issue",
			State:     "new",
			UpdatedOn: "2026-03-11T00:00:00Z",
			Reporter:  bitbucket.IssueActor{DisplayName: "Example User"},
		},
	}

	if err := writeTargetHeader(&buf, "Repository", "acme", "widgets"); err != nil {
		t.Fatalf("writeTargetHeader returned error: %v", err)
	}
	if err := writeIssueTable(&buf, issues); err != nil {
		t.Fatalf("writeIssueTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Repository: acme/widgets") {
		t.Fatalf("expected repository header, got %q", got)
	}
	if !strings.Contains(got, "Fixture issue") {
		t.Fatalf("expected issue row, got %q", got)
	}
}

func TestWriteIssueListSummaryIncludesWarnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	target := resolvedRepoTarget{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
	}
	issues := []bitbucket.Issue{
		{
			ID:        1,
			Title:     "Fixture issue",
			State:     "new",
			UpdatedOn: "2026-03-11T00:00:00Z",
			Reporter:  bitbucket.IssueActor{DisplayName: "Example User"},
		},
	}

	if err := writeIssueListSummary(&buf, target, issues); err != nil {
		t.Fatalf("writeIssueListSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Fixture issue",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWriteIssueViewSummaryIncludesWarnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	target := resolvedRepoTarget{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
	}
	issue := bitbucket.Issue{
		ID:        12,
		Title:     "Fixture issue",
		State:     "new",
		Kind:      "bug",
		Priority:  "major",
		UpdatedOn: "2026-03-11T00:00:00Z",
		Reporter:  bitbucket.IssueActor{DisplayName: "Example User"},
		Links: bitbucket.IssueLinks{
			HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/issues/12"},
		},
		Content: bitbucket.IssueContent{Raw: "Needs investigation"},
	}

	if err := writeIssueViewSummary(&buf, target, issue); err != nil {
		t.Fatalf("writeIssueViewSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Title:",
		"Fixture issue",
		"Next: bb issue edit 12 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}
