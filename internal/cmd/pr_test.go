package cmd

import (
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
