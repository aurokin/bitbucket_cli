package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestNormalizePullRequestTaskListState(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":           "UNRESOLVED",
		"unresolved": "UNRESOLVED",
		"resolved":   "RESOLVED",
		"all":        "ALL",
	}

	for input, want := range cases {
		got, err := normalizePullRequestTaskListState(input)
		if err != nil {
			t.Fatalf("normalizePullRequestTaskListState(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("normalizePullRequestTaskListState(%q) = %q, want %q", input, got, want)
		}
	}

	if _, err := normalizePullRequestTaskListState("pending"); err == nil {
		t.Fatalf("expected invalid task state error")
	}
}

func TestPullRequestTaskState(t *testing.T) {
	t.Parallel()

	if got := pullRequestTaskState(bitbucket.PullRequestTask{}, pullRequestTaskSummaryOptions{}); got != "open" {
		t.Fatalf("expected open, got %q", got)
	}
	if got := pullRequestTaskState(bitbucket.PullRequestTask{Pending: true}, pullRequestTaskSummaryOptions{}); got != "pending" {
		t.Fatalf("expected pending, got %q", got)
	}
	if got := pullRequestTaskState(bitbucket.PullRequestTask{State: "RESOLVED"}, pullRequestTaskSummaryOptions{}); got != "resolved" {
		t.Fatalf("expected resolved, got %q", got)
	}
	if got := pullRequestTaskState(bitbucket.PullRequestTask{}, pullRequestTaskSummaryOptions{Deleted: true}); got != "deleted" {
		t.Fatalf("expected deleted, got %q", got)
	}
}

func TestWritePullRequestTaskSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prTaskPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		Action:      "resolved",
		Task: bitbucket.PullRequestTask{
			ID:      3,
			State:   "RESOLVED",
			Content: bitbucket.PullRequestTaskContent{Raw: "Handle reviewer feedback"},
			Creator: bitbucket.PullRequestActor{DisplayName: "Auro"},
			Comment: &bitbucket.PullRequestComment{
				ID:    15,
				Links: bitbucket.PullRequestCommentLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15"}},
			},
			Links: bitbucket.PullRequestTaskLinks{
				HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/pull-requests/7?_task=3"},
			},
		},
	}

	if err := writePullRequestTaskSummary(&buf, payload, pullRequestTaskSummaryOptions{}); err != nil {
		t.Fatalf("writePullRequestTaskSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Task:",
		"Action:",
		"State:",
		"Comment:",
		"Comment URL:",
		"URL:",
		"Next: bb pr task reopen 3 --pr 7 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Task:",
		"Action:",
		"State:",
		"Comment:",
		"URL:",
		"Next: bb pr task reopen 3 --pr 7 --repo acme/widgets",
	)
}

func TestWritePullRequestTaskListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prTaskListPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		State:       "UNRESOLVED",
		Tasks: []bitbucket.PullRequestTask{
			{
				ID:      3,
				State:   "UNRESOLVED",
				Content: bitbucket.PullRequestTaskContent{Raw: "Handle reviewer feedback"},
				Comment: &bitbucket.PullRequestComment{ID: 15},
			},
		},
	}

	if err := writePullRequestTaskListSummary(&buf, payload); err != nil {
		t.Fatalf("writePullRequestTaskListSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Pull Request: #7",
		"Filter: unresolved",
		"ID STATE COMMENT BODY",
		"Next: bb pr task view 3 --pr 7 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWritePullRequestTaskListSummaryEmpty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := prTaskListPayload{
		Workspace:   "acme",
		Repo:        "widgets",
		PullRequest: 7,
		State:       "ALL",
	}

	if err := writePullRequestTaskListSummary(&buf, payload); err != nil {
		t.Fatalf("writePullRequestTaskListSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Tasks: None.",
		"Next: bb pr task create 7 --repo acme/widgets --body '<task body>'",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestParsePullRequestTaskID(t *testing.T) {
	t.Parallel()

	got, err := parsePullRequestTaskID("3")
	if err != nil {
		t.Fatalf("parsePullRequestTaskID returned error: %v", err)
	}
	if got != 3 {
		t.Fatalf("expected task id 3, got %d", got)
	}

	if _, err := parsePullRequestTaskID("https://bitbucket.org/acme/widgets/pull-requests/7"); err == nil {
		t.Fatalf("expected non-numeric task id error")
	}

	if _, err := parsePullRequestTaskID("0"); err == nil || !strings.Contains(err.Error(), "numeric task ID") {
		t.Fatalf("expected positive task id error, got %v", err)
	}

	if _, err := parsePullRequestTaskID(""); err == nil || !strings.Contains(err.Error(), "reference is required") {
		t.Fatalf("expected missing task reference error, got %v", err)
	}
}
