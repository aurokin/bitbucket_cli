package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestPullRequestCommentState(t *testing.T) {
	t.Parallel()

	if got := pullRequestCommentState(bitbucket.PullRequestComment{}, pullRequestCommentSummaryOptions{}); got != "open" {
		t.Fatalf("expected open, got %q", got)
	}
	if got := pullRequestCommentState(bitbucket.PullRequestComment{Pending: true}, pullRequestCommentSummaryOptions{}); got != "pending" {
		t.Fatalf("expected pending, got %q", got)
	}
	if got := pullRequestCommentState(bitbucket.PullRequestComment{Resolution: &bitbucket.PullRequestCommentResolve{Type: "comment_resolution"}}, pullRequestCommentSummaryOptions{}); got != "resolved" {
		t.Fatalf("expected resolved, got %q", got)
	}
	if got := pullRequestCommentState(bitbucket.PullRequestComment{}, pullRequestCommentSummaryOptions{Deleted: true}); got != "deleted" {
		t.Fatalf("expected deleted, got %q", got)
	}
}

func TestPullRequestCommentLine(t *testing.T) {
	t.Parallel()

	if got := pullRequestCommentLine(&bitbucket.PullRequestCommentInline{StartTo: 10, To: 12}); got != "10-12" {
		t.Fatalf("expected 10-12, got %q", got)
	}
	if got := pullRequestCommentLine(&bitbucket.PullRequestCommentInline{To: 12}); got != "12" {
		t.Fatalf("expected 12, got %q", got)
	}
}

func TestWritePullRequestCommentSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prCommentPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		Action:      "resolved",
		Comment: bitbucket.PullRequestComment{
			ID: 15,
			Content: bitbucket.PullRequestCommentContent{
				Raw: "Looks good to me",
			},
			User:       bitbucket.PullRequestActor{DisplayName: "Reviewer"},
			Resolution: &bitbucket.PullRequestCommentResolve{Type: "comment_resolution"},
			Inline:     &bitbucket.PullRequestCommentInline{Path: "README.md", To: 12},
			Links:      bitbucket.PullRequestCommentLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15"}},
		},
	}

	if err := writePullRequestCommentSummary(&buf, payload, pullRequestCommentSummaryOptions{}); err != nil {
		t.Fatalf("writePullRequestCommentSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Comment:",
		"State:",
		"Path:",
		"Line:",
		"resolved",
		"README.md",
		"12",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Comment:",
		"State:",
		"Path:",
		"Line:",
		"URL:",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	)
}
