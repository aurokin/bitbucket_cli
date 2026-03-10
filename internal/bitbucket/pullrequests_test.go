package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
