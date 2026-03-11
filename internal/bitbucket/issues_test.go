package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got == "" {
			t.Fatal("expected q filter")
		}
		if got := r.URL.Query().Get("sort"); got != "-updated_on" {
			t.Fatalf("unexpected sort %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":1,"title":"Example issue","state":"new","reporter":{"display_name":"Auro"}}]}`))
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

	issues, err := client.ListIssues(context.Background(), "acme", "widgets", ListIssuesOptions{
		Query: `title ~ "issue"`,
		Sort:  "-updated_on",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListIssues returned error: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != 1 {
		t.Fatalf("unexpected issues %+v", issues)
	}
}

func TestCreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["title"] != "Broken flow" {
			t.Fatalf("unexpected title %#v", body["title"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":2,"title":"Broken flow","state":"new"}`))
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

	issue, err := client.CreateIssue(context.Background(), "acme", "widgets", CreateIssueOptions{
		Title: "Broken flow",
		Body:  "Needs investigation",
	})
	if err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}
	if issue.ID != 2 || issue.Title != "Broken flow" {
		t.Fatalf("unexpected issue %+v", issue)
	}
}
