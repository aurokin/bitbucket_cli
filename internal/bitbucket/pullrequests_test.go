package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestListPullRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("state"); got != "OPEN" {
			t.Fatalf("unexpected state filter %q", got)
		}
		if got := r.URL.Query().Get("pagelen"); got != "10" {
			t.Fatalf("unexpected pagelen %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":1,"title":"Example PR","state":"OPEN","author":{"display_name":"Auro"},"source":{"branch":{"name":"feature"},"commit":{"hash":"abc"},"repository":{"full_name":"acme/widgets"}},"destination":{"branch":{"name":"main"},"commit":{"hash":"def"},"repository":{"full_name":"acme/widgets"}}}]}`))
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

	prs, err := client.ListPullRequests(context.Background(), "acme", "widgets", ListPullRequestsOptions{
		State: "OPEN",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListPullRequests returned error: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 pull request, got %d", len(prs))
	}
	if prs[0].Title != "Example PR" {
		t.Fatalf("unexpected pull request %+v", prs[0])
	}
}

func TestListPullRequestsFollowsPagination(t *testing.T) {
	var requests int
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.RawQuery, "page=2") {
			_, _ = w.Write([]byte(`{"values":[{"id":2,"title":"Second","state":"OPEN","author":{},"source":{"branch":{"name":"two"},"commit":{},"repository":{}},"destination":{"branch":{"name":"main"},"commit":{},"repository":{}}}]}`))
			return
		}

		_, _ = w.Write([]byte(`{"values":[{"id":1,"title":"First","state":"OPEN","author":{},"source":{"branch":{"name":"one"},"commit":{},"repository":{}},"destination":{"branch":{"name":"main"},"commit":{},"repository":{}}}],"next":"` + server.URL + `/2.0/repositories/acme/widgets/pullrequests?page=2"}`))
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

	prs, err := client.ListPullRequests(context.Background(), "acme", "widgets", ListPullRequestsOptions{
		State: "OPEN",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("ListPullRequests returned error: %v", err)
	}
	if len(prs) != 2 {
		t.Fatalf("expected 2 pull requests, got %d", len(prs))
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}

func TestGetPullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests/7" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"title":"Example PR","state":"OPEN","author":{"display_name":"Auro"},"source":{"branch":{"name":"feature"},"commit":{"hash":"abc"},"repository":{"full_name":"acme/widgets"}},"destination":{"branch":{"name":"main"},"commit":{"hash":"def"},"repository":{"full_name":"acme/widgets"}}}`))
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

	pr, err := client.GetPullRequest(context.Background(), "acme", "widgets", 7)
	if err != nil {
		t.Fatalf("GetPullRequest returned error: %v", err)
	}
	if pr.ID != 7 || pr.Title != "Example PR" {
		t.Fatalf("unexpected pull request %+v", pr)
	}
}

func TestCreatePullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["title"] != "Add feature" {
			t.Fatalf("unexpected title %v", body["title"])
		}
		source := body["source"].(map[string]any)
		destination := body["destination"].(map[string]any)
		if source["branch"].(map[string]any)["name"] != "feature" {
			t.Fatalf("unexpected source %#v", source)
		}
		if destination["branch"].(map[string]any)["name"] != "main" {
			t.Fatalf("unexpected destination %#v", destination)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":9,"title":"Add feature","state":"OPEN","author":{"display_name":"Auro"},"source":{"branch":{"name":"feature"},"commit":{},"repository":{}},"destination":{"branch":{"name":"main"},"commit":{},"repository":{}}}`))
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

	pr, err := client.CreatePullRequest(context.Background(), "acme", "widgets", CreatePullRequestOptions{
		Title:             "Add feature",
		Description:       "desc",
		SourceBranch:      "feature",
		DestinationBranch: "main",
	})
	if err != nil {
		t.Fatalf("CreatePullRequest returned error: %v", err)
	}
	if pr.ID != 9 || pr.Title != "Add feature" {
		t.Fatalf("unexpected pull request %+v", pr)
	}
}

func TestCreatePullRequestReuseExisting(t *testing.T) {
	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet {
			t.Fatalf("expected reuse path to avoid POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":5,"title":"Add feature","state":"OPEN","author":{"display_name":"Auro"},"source":{"branch":{"name":"feature"},"commit":{},"repository":{}},"destination":{"branch":{"name":"main"},"commit":{},"repository":{}}}]}`))
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

	pr, err := client.CreatePullRequest(context.Background(), "acme", "widgets", CreatePullRequestOptions{
		Title:             "Add feature",
		SourceBranch:      "feature",
		DestinationBranch: "main",
		ReuseExisting:     true,
	})
	if err != nil {
		t.Fatalf("CreatePullRequest returned error: %v", err)
	}
	if pr.ID != 5 {
		t.Fatalf("unexpected reused pull request %+v", pr)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestMergePullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pullrequests/7/merge" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("async"); got != "true" {
			t.Fatalf("unexpected async query %q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["type"] != "pullrequest" {
			t.Fatalf("unexpected type %v", body["type"])
		}
		if body["merge_strategy"] != "merge_commit" {
			t.Fatalf("unexpected merge strategy %v", body["merge_strategy"])
		}
		if body["message"] != "Ship it" {
			t.Fatalf("unexpected merge message %v", body["message"])
		}
		if body["close_source_branch"] != true {
			t.Fatalf("expected close_source_branch to be true, got %v", body["close_source_branch"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"title":"Example PR","state":"MERGED","merge_commit":{"hash":"abc123"},"author":{"display_name":"Auro"},"source":{"branch":{"name":"feature"},"commit":{"hash":"abc"},"repository":{"full_name":"acme/widgets"}},"destination":{"branch":{"name":"main"},"commit":{"hash":"def"},"repository":{"full_name":"acme/widgets"}}}`))
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

	pr, err := client.MergePullRequest(context.Background(), "acme", "widgets", 7, MergePullRequestOptions{
		Message:           "Ship it",
		CloseSourceBranch: true,
		MergeStrategy:     "merge_commit",
	})
	if err != nil {
		t.Fatalf("MergePullRequest returned error: %v", err)
	}
	if pr.State != "MERGED" || pr.MergeCommit.Hash != "abc123" {
		t.Fatalf("unexpected merged pull request %+v", pr)
	}
}

func TestMergePullRequestWaitsForTask(t *testing.T) {
	var taskRequests int
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/2.0/repositories/acme/widgets/pullrequests/7/merge":
			w.Header().Set("Location", server.URL+"/2.0/repositories/acme/widgets/pullrequests/7/merge/task-status/abc")
			w.WriteHeader(http.StatusAccepted)
		case "/2.0/repositories/acme/widgets/pullrequests/7/merge/task-status/abc":
			taskRequests++
			w.Header().Set("Content-Type", "application/json")
			if taskRequests == 1 {
				_, _ = w.Write([]byte(`{"task_status":"PENDING"}`))
				return
			}
			_, _ = w.Write([]byte(`{"task_status":"SUCCESS","merge_result":{"id":7,"title":"Example PR","state":"MERGED","merge_commit":{"hash":"def456"},"author":{"display_name":"Auro"},"source":{"branch":{"name":"feature"},"commit":{"hash":"abc"},"repository":{"full_name":"acme/widgets"}},"destination":{"branch":{"name":"main"},"commit":{"hash":"def"},"repository":{"full_name":"acme/widgets"}}}}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
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

	pr, err := client.MergePullRequest(context.Background(), "acme", "widgets", 7, MergePullRequestOptions{
		MergeStrategy: "merge_commit",
		PollInterval:  time.Millisecond,
		PollTimeout:   time.Second,
	})
	if err != nil {
		t.Fatalf("MergePullRequest returned error: %v", err)
	}
	if pr.State != "MERGED" || pr.MergeCommit.Hash != "def456" {
		t.Fatalf("unexpected merged pull request %+v", pr)
	}
	if taskRequests != 2 {
		t.Fatalf("expected 2 task requests, got %d", taskRequests)
	}
}

func TestMergePullRequestRequiresTaskLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
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

	_, err = client.MergePullRequest(context.Background(), "acme", "widgets", 7, MergePullRequestOptions{
		MergeStrategy: "merge_commit",
	})
	if err == nil || !strings.Contains(err.Error(), "without a merge task location") {
		t.Fatalf("expected missing task location error, got %v", err)
	}
}
