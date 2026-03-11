package cmd

import (
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
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

func TestSearchNextSteps(t *testing.T) {
	t.Parallel()

	if got := searchReposNextStep("acme", nil); got != "bb repo create acme/<repo>" {
		t.Fatalf("unexpected repo empty next step %q", got)
	}
	if got := searchReposNextStep("acme", []bitbucket.Repository{{Slug: "widgets"}}); got != "bb repo view --repo acme/widgets" {
		t.Fatalf("unexpected repo single next step %q", got)
	}
	if got := searchPRsNextStep("acme", "widgets", nil); got != "bb pr list --repo acme/widgets" {
		t.Fatalf("unexpected PR empty next step %q", got)
	}
	if got := searchPRsNextStep("acme", "widgets", []bitbucket.PullRequest{{ID: 7}}); got != "bb pr view 7 --repo acme/widgets" {
		t.Fatalf("unexpected PR single next step %q", got)
	}
	if got := searchIssuesNextStep("acme", "widgets", nil); got != "bb issue list --repo acme/widgets" {
		t.Fatalf("unexpected issue empty next step %q", got)
	}
	if got := searchIssuesNextStep("acme", "widgets", []bitbucket.Issue{{ID: 9}}); got != "bb issue view 9 --repo acme/widgets" {
		t.Fatalf("unexpected issue single next step %q", got)
	}
}
