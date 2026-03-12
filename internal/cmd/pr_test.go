package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
)

func TestResolveMergeStrategyUsesExplicitValue(t *testing.T) {
	t.Parallel()

	pr := bitbucket.PullRequest{
		Destination: bitbucket.PullRequestRef{
			Branch: bitbucket.PullRequestBranch{
				Name:            "main",
				MergeStrategies: []string{"merge_commit", "squash"},
			},
		},
	}

	strategy, err := resolveMergeStrategy(pr, "squash")
	if err != nil {
		t.Fatalf("resolveMergeStrategy returned error: %v", err)
	}
	if strategy != "squash" {
		t.Fatalf("expected squash, got %q", strategy)
	}
}

func TestResolveMergeStrategyRejectsUnsupportedValue(t *testing.T) {
	t.Parallel()

	pr := bitbucket.PullRequest{
		Destination: bitbucket.PullRequestRef{
			Branch: bitbucket.PullRequestBranch{
				Name:            "main",
				MergeStrategies: []string{"merge_commit", "squash"},
			},
		},
	}

	_, err := resolveMergeStrategy(pr, "fast-forward")
	if err == nil || !strings.Contains(err.Error(), "available: merge_commit, squash") {
		t.Fatalf("expected unsupported merge strategy error, got %v", err)
	}
}

func TestResolveMergeStrategyUsesDefault(t *testing.T) {
	t.Parallel()

	pr := bitbucket.PullRequest{
		Destination: bitbucket.PullRequestRef{
			Branch: bitbucket.PullRequestBranch{
				Name:                 "main",
				DefaultMergeStrategy: "merge_commit",
				MergeStrategies:      []string{"merge_commit", "squash"},
			},
		},
	}

	strategy, err := resolveMergeStrategy(pr, "")
	if err != nil {
		t.Fatalf("resolveMergeStrategy returned error: %v", err)
	}
	if strategy != "merge_commit" {
		t.Fatalf("expected merge_commit, got %q", strategy)
	}
}

func TestResolveMergeStrategyRequiresExplicitChoiceWhenAmbiguous(t *testing.T) {
	t.Parallel()

	pr := bitbucket.PullRequest{
		Destination: bitbucket.PullRequestRef{
			Branch: bitbucket.PullRequestBranch{
				Name:            "main",
				MergeStrategies: []string{"merge_commit", "squash"},
			},
		},
	}

	_, err := resolveMergeStrategy(pr, "")
	if err == nil || !strings.Contains(err.Error(), "pass --strategy") {
		t.Fatalf("expected ambiguous merge strategy error, got %v", err)
	}
}

func TestResolveMergeStrategyAllowsServerDefaultWhenUnavailable(t *testing.T) {
	t.Parallel()

	pr := bitbucket.PullRequest{
		Destination: bitbucket.PullRequestRef{
			Branch: bitbucket.PullRequestBranch{
				Name: "main",
			},
		},
	}

	strategy, err := resolveMergeStrategy(pr, "")
	if err != nil {
		t.Fatalf("resolveMergeStrategy returned error: %v", err)
	}
	if strategy != "" {
		t.Fatalf("expected empty strategy to defer to server default, got %q", strategy)
	}
}

func TestBuildPRStatusPayload(t *testing.T) {
	t.Parallel()

	target := resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "OhBizzle",
		Repo:      "widgets",
	}
	user := bitbucket.CurrentUser{
		AccountID:   "user-1",
		DisplayName: "Hunter Sadler",
	}

	prs := []bitbucket.PullRequest{
		{
			ID:    1,
			Title: "Current branch PR",
			State: "OPEN",
			Author: bitbucket.PullRequestActor{
				AccountID: "user-1",
			},
			Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/current"}},
			Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
		},
		{
			ID:    2,
			Title: "Created by you",
			State: "OPEN",
			Author: bitbucket.PullRequestActor{
				AccountID: "user-1",
			},
			Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/authored"}},
			Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
		},
		{
			ID:    3,
			Title: "Needs your review",
			State: "OPEN",
			Author: bitbucket.PullRequestActor{
				AccountID: "other-user",
			},
			Reviewers: []bitbucket.PullRequestActor{
				{AccountID: "user-1"},
			},
			Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/review"}},
			Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
		},
	}

	payload := buildPRStatusPayload(target, user, "feature/current", prs)
	if payload.CurrentBranch == nil || payload.CurrentBranch.ID != 1 {
		t.Fatalf("expected current branch PR #1, got %+v", payload.CurrentBranch)
	}
	if len(payload.Created) != 1 || payload.Created[0].ID != 2 {
		t.Fatalf("expected authored PR #2, got %+v", payload.Created)
	}
	if len(payload.ReviewRequested) != 1 || payload.ReviewRequested[0].ID != 3 {
		t.Fatalf("expected review requested PR #3, got %+v", payload.ReviewRequested)
	}
}

func TestReviewRequestedFromUser(t *testing.T) {
	t.Parallel()

	user := bitbucket.CurrentUser{AccountID: "user-1"}
	pr := bitbucket.PullRequest{
		Reviewers: []bitbucket.PullRequestActor{
			{AccountID: "user-1"},
		},
	}

	if !reviewRequestedFromUser(user, pr) {
		t.Fatal("expected reviewRequestedFromUser to match reviewer")
	}
}

func TestDiffPath(t *testing.T) {
	t.Parallel()

	renamed := bitbucket.PullRequestDiffStat{
		Old: &bitbucket.PullRequestDiffRef{Path: "old.txt"},
		New: &bitbucket.PullRequestDiffRef{Path: "new.txt"},
	}
	if got := diffPath(renamed); got != "old.txt -> new.txt" {
		t.Fatalf("expected rename path, got %q", got)
	}

	added := bitbucket.PullRequestDiffStat{
		New: &bitbucket.PullRequestDiffRef{Path: "file.txt"},
	}
	if got := diffPath(added); got != "file.txt" {
		t.Fatalf("expected new path, got %q", got)
	}
}

func TestResolveCommentBody(t *testing.T) {
	t.Parallel()

	body, err := resolveCommentBody(bytes.NewBufferString(""), "Looks good", "")
	if err != nil {
		t.Fatalf("resolveCommentBody returned error: %v", err)
	}
	if body != "Looks good" {
		t.Fatalf("expected inline body, got %q", body)
	}

	body, err = resolveCommentBody(bytes.NewBufferString("From stdin\n"), "", "-")
	if err != nil {
		t.Fatalf("resolveCommentBody returned error: %v", err)
	}
	if body != "From stdin" {
		t.Fatalf("expected stdin body, got %q", body)
	}
}

func TestWritePRListTableCompactsWideFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	prs := []bitbucket.PullRequest{
		{
			ID:        7,
			Title:     "Add a very long pull request title that should be truncated for human output",
			State:     "OPEN",
			UpdatedOn: "2026-03-10T12:34:56Z",
			Author: bitbucket.PullRequestActor{
				DisplayName: "Hunter Sadler With A Long Name",
			},
			Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/some-super-long-branch-name-for-testing"}},
			Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
		},
	}

	if err := writePRListTable(&buf, prs); err != nil {
		t.Fatalf("writePRListTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "#\ttitle\tstate\tauthor\tsrc\tdst\tupdated") && !strings.Contains(got, "# title") {
		t.Fatalf("expected compact header, got %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("expected truncation marker in output, got %q", got)
	}
	if !strings.Contains(got, "2026-03-10") {
		t.Fatalf("expected compact timestamp, got %q", got)
	}
}

func TestWritePRDiffStatTableCompactsPaths(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	stats := []bitbucket.PullRequestDiffStat{
		{
			Status:       "modified",
			LinesAdded:   10,
			LinesRemoved: 4,
			Old:          &bitbucket.PullRequestDiffRef{Path: "src/components/a/very/long/path/original-file-name.go"},
			New:          &bitbucket.PullRequestDiffRef{Path: "src/components/a/very/long/path/renamed-file-name.go"},
		},
	}

	if err := writePRDiffStatTable(&buf, stats); err != nil {
		t.Fatalf("writePRDiffStatTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "status") || !strings.Contains(got, "file") {
		t.Fatalf("expected compact header, got %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("expected truncated path in output, got %q", got)
	}
	if !strings.Contains(got, "total") {
		t.Fatalf("expected total row, got %q", got)
	}
}

func TestPRViewNextStep(t *testing.T) {
	t.Parallel()

	if got := prViewNextStep("acme", "widgets", 7); got != "bb pr diff 7 --repo acme/widgets" {
		t.Fatalf("unexpected PR view next step %q", got)
	}
}

func TestWritePRListTableWithRepositoryHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	prs := []bitbucket.PullRequest{
		{
			ID:        7,
			Title:     "Fixture PR",
			State:     "OPEN",
			UpdatedOn: "2026-03-10T12:34:56Z",
			Author:    bitbucket.PullRequestActor{DisplayName: "Hunter Sadler"},
			Source:    bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/test"}},
			Destination: bitbucket.PullRequestRef{
				Branch: bitbucket.PullRequestBranch{Name: "main"},
			},
		},
	}

	if err := writeTargetHeader(&buf, "Repository", "acme", "widgets"); err != nil {
		t.Fatalf("writeTargetHeader returned error: %v", err)
	}
	if err := writePRListTable(&buf, prs); err != nil {
		t.Fatalf("writePRListTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Repository: acme/widgets") {
		t.Fatalf("expected repository header, got %q", got)
	}
	if !strings.Contains(got, "Fixture PR") {
		t.Fatalf("expected PR row, got %q", got)
	}
}

func TestWritePullRequestSummaryTableIncludesOptionalFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	pr := bitbucket.PullRequest{
		ID:          9,
		Title:       "Ship feature",
		State:       "MERGED",
		UpdatedOn:   "2026-03-10T12:34:56Z",
		Description: "Ready to land",
		Author:      bitbucket.PullRequestActor{DisplayName: "Hunter Sadler"},
		Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/ship"}},
		Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
		Links:       bitbucket.PullRequestLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/pull-requests/9"}},
	}

	if err := writePullRequestSummaryTable(&buf, pr, pullRequestSummaryOptions{
		IncludeAuthor:      true,
		IncludeUpdated:     true,
		IncludeDescription: true,
		Strategy:           "merge_commit",
		MergeCommit:        "abc1234",
	}); err != nil {
		t.Fatalf("writePullRequestSummaryTable returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"ID:",
		"Title:",
		"Author:",
		"Strategy:",
		"Merge Commit:",
		"Updated:",
		"Description:",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}
