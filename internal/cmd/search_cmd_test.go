package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
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

func TestWriteSearchPRSummaryIncludesWarningsAndCounts(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	target := resolvedRepoTarget{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
	}
	prs := []bitbucket.PullRequest{
		{
			ID:           7,
			Title:        "Fixture PR",
			State:        "OPEN",
			TaskCount:    2,
			CommentCount: 4,
			Author:       bitbucket.PullRequestActor{DisplayName: "Example User"},
			Source:       bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/tasks"}},
			Destination:  bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
		},
	}

	if err := writeSearchPRSummary(&buf, target, "fixture", prs); err != nil {
		t.Fatalf("writeSearchPRSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Query: fixture",
		"tsk",
		"cmt",
		"Next: bb pr view 7 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWriteSearchPRSummaryEmptyIncludesWarnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	target := resolvedRepoTarget{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
	}

	if err := writeSearchPRSummary(&buf, target, "fixture", nil); err != nil {
		t.Fatalf("writeSearchPRSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Warning: local repository context unavailable",
		`No pull requests found for acme/widgets matching "fixture".`,
		"Next: bb pr list --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWriteSearchRepoSummaryIncludesNextStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	repos := []bitbucket.Repository{{
		Name:      "Widget Service",
		Slug:      "widgets",
		IsPrivate: true,
		Project:   bitbucket.RepositoryProject{Key: "WID"},
		UpdatedOn: "2026-03-13T00:00:00Z",
	}}

	if err := writeSearchRepoSummary(&buf, "acme", "widget", repos); err != nil {
		t.Fatalf("writeSearchRepoSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Workspace: acme",
		"Query: widget",
		"widgets",
		"Next: bb repo view --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWriteSearchIssueSummaryIncludesWarningsAndNextStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	target := resolvedRepoTarget{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
	}
	issues := []bitbucket.Issue{{
		ID:        9,
		Title:     "Fix flaky integration test",
		State:     "open",
		Reporter:  bitbucket.IssueActor{DisplayName: "Example User"},
		UpdatedOn: "2026-03-13T00:00:00Z",
	}}

	if err := writeSearchIssueSummary(&buf, target, "flaky", issues); err != nil {
		t.Fatalf("writeSearchIssueSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Query: flaky",
		"Fix flaky integration test",
		"Next: bb issue view 9 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}
