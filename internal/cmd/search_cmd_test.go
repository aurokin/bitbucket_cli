package cmd

import (
	"strings"
	"testing"
)

func TestBuildRepositorySearchQuery(t *testing.T) {
	t.Parallel()

	query := buildRepositorySearchQuery(`bb "cli"`)
	if !strings.Contains(query, `name ~ "bb \"cli\""`) {
		t.Fatalf("unexpected repository query %q", query)
	}
}

func TestBuildPullRequestSearchQuery(t *testing.T) {
	t.Parallel()

	query := buildPullRequestSearchQuery("fixture")
	if !strings.Contains(query, `title ~ "fixture"`) || !strings.Contains(query, `description ~ "fixture"`) {
		t.Fatalf("unexpected pull request query %q", query)
	}
}

func TestBuildIssueSearchQuery(t *testing.T) {
	t.Parallel()

	query := buildIssueSearchQuery("bug")
	if !strings.Contains(query, `content.raw ~ "bug"`) {
		t.Fatalf("unexpected issue query %q", query)
	}
}
