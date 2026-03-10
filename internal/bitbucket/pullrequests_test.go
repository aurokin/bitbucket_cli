package bitbucket

import (
	"context"
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

	client, err := NewClient("bitbucket.org", config.HostConfig{Token: "secret"})
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

	client, err := NewClient("bitbucket.org", config.HostConfig{Token: "secret"})
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
