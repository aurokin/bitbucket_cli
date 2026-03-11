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
