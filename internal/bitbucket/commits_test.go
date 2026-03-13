package bitbucket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
)

func TestCommitClientReadAndReviewFlows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234":
			_, _ = w.Write([]byte(`{"hash":"abc1234","date":"2026-03-13T00:00:00Z","message":"Example commit","summary":{"raw":"Example commit"},"author":{"raw":"A U Thor <author@example.com>"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/diff/abc1234":
			if got := r.URL.Query().Get("ignore_whitespace"); got != "true" {
				t.Fatalf("expected ignore_whitespace=true, got %q", got)
			}
			if got := r.URL.Query()["path"]; len(got) != 2 || got[0] != "README.md" || got[1] != "cmd/bb/main.go" {
				t.Fatalf("unexpected path filters %v", got)
			}
			_, _ = io.WriteString(w, "diff --git a/README.md b/README.md\n")
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/diffstat/abc1234":
			_, _ = w.Write([]byte(`{"values":[{"status":"modified","new":{"path":"README.md"},"lines_added":3,"lines_removed":1}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/comments":
			_, _ = w.Write([]byte(`{"values":[{"id":15,"content":{"raw":"Looks good"},"user":{"display_name":"Reviewer"},"created_on":"2026-03-13T01:00:00Z"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/comments/15":
			_, _ = w.Write([]byte(`{"id":15,"content":{"raw":"Looks good"},"user":{"display_name":"Reviewer"},"created_on":"2026-03-13T01:00:00Z"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/approve":
			_, _ = w.Write([]byte(`{"user":{"display_name":"Reviewer","account_id":"557058:reviewer"},"approved":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/approve":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/statuses":
			_, _ = w.Write([]byte(`{"values":[{"key":"build","state":"SUCCESSFUL","name":"Build","updated_on":"2026-03-13T02:00:00Z"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/reports":
			_, _ = w.Write([]byte(`{"values":[{"uuid":"{report-1}","external_id":"scanner-1","title":"Security scan","report_type":"SECURITY","result":"PASSED","updated_on":"2026-03-13T03:00:00Z"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/commit/abc1234/reports/scanner-1":
			_, _ = w.Write([]byte(`{"uuid":"{report-1}","external_id":"scanner-1","title":"Security scan","details":"All good","report_type":"SECURITY","result":"PASSED","updated_on":"2026-03-13T03:00:00Z"}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{AuthType: "api-token", Username: "user@example.com", Token: "secret"})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	commit, err := client.GetCommit(context.Background(), "acme", "widgets", "abc1234")
	if err != nil || commit.Hash != "abc1234" {
		t.Fatalf("GetCommit returned %v %+v", err, commit)
	}

	diff, err := client.GetCommitDiff(context.Background(), "acme", "widgets", "abc1234", CommitDiffOptions{
		Path:             []string{"README.md", "cmd/bb/main.go"},
		IgnoreWhitespace: true,
	})
	if err != nil || !strings.Contains(diff, "diff --git") {
		t.Fatalf("GetCommitDiff returned %v %q", err, diff)
	}

	stats, err := client.ListCommitDiffStats(context.Background(), "acme", "widgets", "abc1234")
	if err != nil || len(stats) != 1 || stats[0].New == nil || stats[0].New.Path != "README.md" {
		t.Fatalf("ListCommitDiffStats returned %v %+v", err, stats)
	}

	comments, err := client.ListCommitComments(context.Background(), "acme", "widgets", "abc1234", ListCommitCommentsOptions{Limit: 20})
	if err != nil || len(comments) != 1 || comments[0].ID != 15 {
		t.Fatalf("ListCommitComments returned %v %+v", err, comments)
	}

	comment, err := client.GetCommitComment(context.Background(), "acme", "widgets", "abc1234", 15)
	if err != nil || comment.ID != 15 {
		t.Fatalf("GetCommitComment returned %v %+v", err, comment)
	}

	reviewer, err := client.ReviewCommit(context.Background(), "acme", "widgets", "abc1234", true)
	if err != nil || reviewer.User.DisplayName != "Reviewer" {
		t.Fatalf("ReviewCommit approve returned %v %+v", err, reviewer)
	}

	if _, err := client.ReviewCommit(context.Background(), "acme", "widgets", "abc1234", false); err != nil {
		t.Fatalf("ReviewCommit unapprove returned error: %v", err)
	}

	statuses, err := client.ListCommitStatuses(context.Background(), "acme", "widgets", "abc1234", ListCommitStatusesOptions{Limit: 20})
	if err != nil || len(statuses) != 1 || statuses[0].Key != "build" {
		t.Fatalf("ListCommitStatuses returned %v %+v", err, statuses)
	}

	reports, err := client.ListCommitReports(context.Background(), "acme", "widgets", "abc1234", ListCommitReportsOptions{Limit: 20})
	if err != nil || len(reports) != 1 || reports[0].ExternalID != "scanner-1" {
		t.Fatalf("ListCommitReports returned %v %+v", err, reports)
	}

	report, err := client.GetCommitReport(context.Background(), "acme", "widgets", "abc1234", "scanner-1")
	if err != nil || report.Title != "Security scan" {
		t.Fatalf("GetCommitReport returned %v %+v", err, report)
	}
}
