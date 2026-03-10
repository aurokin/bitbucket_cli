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
