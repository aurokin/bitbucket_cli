package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
)

func TestReviewPullRequestApprove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests/7/approve" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"display_name":"Reviewer","account_id":"user-1"},"role":"REVIEWER","approved":true,"state":"approved","participated_on":"2026-03-13T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	participant, err := client.ReviewPullRequest(context.Background(), "acme", "widgets", 7, PullRequestReviewApprove)
	if err != nil {
		t.Fatalf("ReviewPullRequest returned error: %v", err)
	}
	if !participant.Approved || participant.State != "approved" || participant.User.DisplayName != "Reviewer" {
		t.Fatalf("unexpected participant %+v", participant)
	}
}

func TestReviewPullRequestRequestChanges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests/7/request-changes" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"display_name":"Reviewer","account_id":"user-1"},"role":"REVIEWER","approved":false,"state":"changes_requested","participated_on":"2026-03-13T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	participant, err := client.ReviewPullRequest(context.Background(), "acme", "widgets", 7, PullRequestReviewRequestChanges)
	if err != nil {
		t.Fatalf("ReviewPullRequest returned error: %v", err)
	}
	if participant.State != "changes_requested" || participant.User.DisplayName != "Reviewer" {
		t.Fatalf("unexpected participant %+v", participant)
	}
}

func TestReviewPullRequestDeleteActionsReturnEmptyParticipant(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action PullRequestReviewAction
		path   string
	}{
		{name: "unapprove", action: PullRequestReviewUnapprove, path: "/2.0/repositories/acme/widgets/pullrequests/7/approve"},
		{name: "clear_request_changes", action: PullRequestReviewClearRequestChanges, path: "/2.0/repositories/acme/widgets/pullrequests/7/request-changes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Fatalf("unexpected method %s", r.Method)
				}
				if r.URL.Path != tc.path {
					t.Fatalf("unexpected path %q", r.URL.Path)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

			client, err := NewClient("bitbucket.org", config.HostConfig{
				Username: "auro@example.com",
				Token:    "secret",
				AuthType: config.AuthTypeAPIToken,
			})
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}

			participant, err := client.ReviewPullRequest(context.Background(), "acme", "widgets", 7, tc.action)
			if err != nil {
				t.Fatalf("ReviewPullRequest returned error: %v", err)
			}
			if participant != (PullRequestParticipant{}) {
				t.Fatalf("expected empty participant, got %+v", participant)
			}
		})
	}
}

func TestListPullRequestActivity(t *testing.T) {
	var requests int
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.RawQuery, "page=2") {
			_, _ = w.Write([]byte(`{"values":[{"approval":{"date":"2026-03-13T01:00:00Z","user":{"display_name":"Reviewer","account_id":"user-2"}}}]}`))
			return
		}
		if got := r.URL.Query().Get("pagelen"); got != "2" {
			t.Fatalf("unexpected pagelen %q", got)
		}

		_, _ = w.Write([]byte(`{"values":[{"comment":{"id":15,"content":{"raw":"Looks good"},"user":{"display_name":"Reviewer"},"created_on":"2026-03-13T00:00:00Z"}}],"next":"` + server.URL + `/2.0/repositories/acme/widgets/pullrequests/7/activity?page=2"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	activity, err := client.ListPullRequestActivity(context.Background(), "acme", "widgets", 7, ListPullRequestActivityOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ListPullRequestActivity returned error: %v", err)
	}
	if len(activity) != 2 {
		t.Fatalf("expected 2 activity entries, got %d", len(activity))
	}
	if activity[0].Comment == nil || activity[0].Comment.ID != 15 {
		t.Fatalf("unexpected comment activity %+v", activity[0])
	}
	if activity[1].Approval == nil || activity[1].Approval.User.DisplayName != "Reviewer" {
		t.Fatalf("unexpected approval activity %+v", activity[1])
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}

func TestListPullRequestCommits(t *testing.T) {
	var requests int
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.RawQuery, "page=2") {
			_, _ = w.Write([]byte(`{"values":[{"hash":"def4567","date":"2026-03-13T01:00:00Z","summary":{"raw":"Second commit"},"author":{"raw":"Reviewer <reviewer@example.com>"}}]}`))
			return
		}
		if got := r.URL.Query().Get("pagelen"); got != "2" {
			t.Fatalf("unexpected pagelen %q", got)
		}

		_, _ = w.Write([]byte(`{"values":[{"hash":"abc1234","date":"2026-03-13T00:00:00Z","message":"First commit\n\nbody","summary":{"raw":"First commit"},"author":{"raw":"Reviewer <reviewer@example.com>","user":{"display_name":"Reviewer"}}}],"next":"` + server.URL + `/2.0/repositories/acme/widgets/pullrequests/7/commits?page=2"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	commits, err := client.ListPullRequestCommits(context.Background(), "acme", "widgets", 7, ListPullRequestCommitsOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ListPullRequestCommits returned error: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" || commits[1].Hash != "def4567" {
		t.Fatalf("unexpected commits %+v", commits)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}

func TestListPullRequestStatuses(t *testing.T) {
	var requests int
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.RawQuery, "page=2") {
			_, _ = w.Write([]byte(`{"values":[{"key":"deploy","state":"FAILED","name":"Deploy","description":"deploy failed","updated_on":"2026-03-13T01:00:00Z"}]}`))
			return
		}
		if got := r.URL.Query().Get("pagelen"); got != "2" {
			t.Fatalf("unexpected pagelen %q", got)
		}
		if got := r.URL.Query().Get("q"); got != "state=\"SUCCESSFUL\"" {
			t.Fatalf("unexpected q %q", got)
		}
		if got := r.URL.Query().Get("sort"); got != "-updated_on" {
			t.Fatalf("unexpected sort %q", got)
		}

		_, _ = w.Write([]byte(`{"values":[{"key":"build","state":"SUCCESSFUL","name":"Build","description":"build passed","updated_on":"2026-03-13T00:00:00Z"}],"next":"` + server.URL + `/2.0/repositories/acme/widgets/pullrequests/7/statuses?page=2"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	statuses, err := client.ListPullRequestStatuses(context.Background(), "acme", "widgets", 7, ListPullRequestStatusesOptions{
		Limit: 2,
		Query: "state=\"SUCCESSFUL\"",
		Sort:  "-updated_on",
	})
	if err != nil {
		t.Fatalf("ListPullRequestStatuses returned error: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if statuses[0].Key != "build" || statuses[1].Key != "deploy" {
		t.Fatalf("unexpected statuses %+v", statuses)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}

func TestPullRequestActivityRequestChangesEvent(t *testing.T) {
	changes := PullRequestActivityEvent{Date: "2026-03-13T00:00:00Z"}

	if got := (PullRequestActivity{ChangesRequest: &changes}).RequestChangesEvent(); got != &changes {
		t.Fatalf("expected changes_request event, got %+v", got)
	}
	if got := (PullRequestActivity{RequestChanges: &changes}).RequestChangesEvent(); got != &changes {
		t.Fatalf("expected request_changes event, got %+v", got)
	}
}
