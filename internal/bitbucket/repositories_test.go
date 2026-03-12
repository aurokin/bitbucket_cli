package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestListWorkspaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/workspaces" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"values":[{"slug":"acme","name":"Example User"}]}`))
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

	workspaces, err := client.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatalf("ListWorkspaces returned error: %v", err)
	}
	if len(workspaces) != 1 || workspaces[0].Slug != "acme" {
		t.Fatalf("unexpected workspaces %+v", workspaces)
	}
}

func TestCreateRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["scm"] != "git" {
			t.Fatalf("unexpected scm %v", body["scm"])
		}
		if body["is_private"] != true {
			t.Fatalf("unexpected is_private %v", body["is_private"])
		}

		project, ok := body["project"].(map[string]any)
		if !ok || project["key"] != "BBCLI" {
			t.Fatalf("unexpected project payload %#v", body["project"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slug":"widgets","name":"Widgets","is_private":true,"project":{"key":"BBCLI"}}`))
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

	repo, err := client.CreateRepository(context.Background(), "acme", "widgets", CreateRepositoryOptions{
		Name:        "Widgets",
		Description: "test repo",
		ProjectKey:  "BBCLI",
		IsPrivate:   true,
	})
	if err != nil {
		t.Fatalf("CreateRepository returned error: %v", err)
	}
	if repo.Slug != "widgets" || repo.Project.Key != "BBCLI" {
		t.Fatalf("unexpected repository %+v", repo)
	}
}

func TestListRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got == "" {
			t.Fatal("expected q filter")
		}
		if got := r.URL.Query().Get("sort"); got != "-updated_on" {
			t.Fatalf("unexpected sort %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"widgets","name":"Widgets","updated_on":"2026-03-11T00:00:00Z","project":{"key":"BBCLI"}}]}`))
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

	repos, err := client.ListRepositories(context.Background(), "acme", ListRepositoriesOptions{
		Query: `name ~ "widgets"`,
		Sort:  "-updated_on",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListRepositories returned error: %v", err)
	}
	if len(repos) != 1 || repos[0].Slug != "widgets" {
		t.Fatalf("unexpected repositories %+v", repos)
	}
}

func TestDeleteRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets" {
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

	if err := client.DeleteRepository(context.Background(), "acme", "widgets"); err != nil {
		t.Fatalf("DeleteRepository returned error: %v", err)
	}
}
