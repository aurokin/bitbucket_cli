package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
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
				DisplayName: "Hunter Sadler With Long Name",
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
