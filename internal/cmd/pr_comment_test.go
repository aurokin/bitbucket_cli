package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
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
		"Action:",
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
		"Action:",
		"State:",
		"Path:",
		"Line:",
		"URL:",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	)
}

func TestWritePullRequestCommentCreateSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	target := resolvedPullRequestTarget{
		RepoTarget: resolvedRepoTarget{
			Workspace: "acme",
			Repo:      "widgets",
		},
		ID: 7,
	}
	comment := bitbucket.PullRequestComment{
		ID: 15,
		Content: bitbucket.PullRequestCommentContent{
			Raw: "Looks good to me",
		},
		User:  bitbucket.PullRequestActor{DisplayName: "Reviewer"},
		Links: bitbucket.PullRequestCommentLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15"}},
	}

	if err := writePullRequestCommentCreateSummary(&buf, target, comment); err != nil {
		t.Fatalf("writePullRequestCommentCreateSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Comment:",
		"Author:",
		"Body:",
		"URL:",
		"Next: bb pr view 7 --repo acme/widgets",
	)
}

func TestPullRequestCommentConfirmationTarget(t *testing.T) {
	t.Parallel()

	target := resolvedPullRequestCommentTarget{
		PRTarget: resolvedPullRequestTarget{
			RepoTarget: resolvedRepoTarget{
				Workspace: "acme",
				Repo:      "widgets",
			},
			ID: 7,
		},
		CommentID: 15,
	}

	if got := pullRequestCommentConfirmationTarget(target); got != "acme/widgets#pr-7/comment-15" {
		t.Fatalf("unexpected confirmation target %q", got)
	}
}

func TestCreatePullRequestCommentCommand(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	cfg := config.Config{}
	cfg.SetHost("bitbucket.org", config.HostConfig{
		AuthType: config.AuthTypeAPIToken,
		Username: "agent@example.com",
		Token:    "secret",
	}, true)
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests/7/comments" {
			t.Fatalf("unexpected %s %q", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":15,"content":{"raw":"Looks good"},"links":{"html":{"href":"https://bitbucket.org/acme/widgets/pull-requests/7#comment-15"}}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	resolved, comment, err := createPullRequestCommentCommand(context.Background(), bytes.NewBufferString(""), "", "acme", "widgets", "7", "Looks good", "")
	if err != nil {
		t.Fatalf("createPullRequestCommentCommand returned error: %v", err)
	}
	if resolved.Target.ID != 7 || comment.ID != 15 || comment.Content.Raw != "Looks good" {
		t.Fatalf("unexpected PR comment result target=%+v comment=%+v", resolved.Target, comment)
	}
}
