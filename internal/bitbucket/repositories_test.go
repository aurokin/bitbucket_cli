package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
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

func TestUpdateRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["name"] != "Widgets Renamed" || body["description"] != "updated" || body["is_private"] != false {
			t.Fatalf("unexpected update body %+v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slug":"widgets-renamed","name":"Widgets Renamed","description":"updated","is_private":false,"full_name":"acme/widgets-renamed"}`))
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

	private := false
	repo, err := client.UpdateRepository(context.Background(), "acme", "widgets", UpdateRepositoryOptions{
		Name:        "Widgets Renamed",
		Description: "updated",
		IsPrivate:   &private,
	})
	if err != nil {
		t.Fatalf("UpdateRepository returned error: %v", err)
	}
	if repo.Slug != "widgets-renamed" || repo.Name != "Widgets Renamed" {
		t.Fatalf("unexpected updated repository %+v", repo)
	}
}

func TestListRepositoryForks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/forks" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("sort"); got != "-updated_on" {
			t.Fatalf("unexpected sort %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"widgets-fork","name":"Widgets Fork","full_name":"acme/widgets-fork","parent":{"full_name":"acme/widgets"}}]}`))
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

	repos, err := client.ListRepositoryForks(context.Background(), "acme", "widgets", ListRepositoriesOptions{
		Sort:  "-updated_on",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListRepositoryForks returned error: %v", err)
	}
	if len(repos) != 1 || repos[0].Slug != "widgets-fork" {
		t.Fatalf("unexpected fork repositories %+v", repos)
	}
}

func TestForkRepositoryReuseExisting(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"slug":"widgets","name":"Widgets","full_name":"acme/widgets"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/forks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"values":[{"slug":"widgets-fork","name":"Widgets Fork","full_name":"acme/widgets-fork","parent":{"full_name":"acme/widgets"}}]}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
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

	repo, err := client.ForkRepository(context.Background(), "acme", "widgets", ForkRepositoryOptions{
		Workspace:     "acme",
		Name:          "Widgets Fork",
		ReuseExisting: true,
	})
	if err != nil {
		t.Fatalf("ForkRepository returned error: %v", err)
	}
	if repo.Slug != "widgets-fork" {
		t.Fatalf("unexpected reused fork %+v", repo)
	}
	if len(requests) != 2 {
		t.Fatalf("expected only GET source and GET forks requests, got %+v", requests)
	}
}

func TestForkRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"slug":"widgets","name":"Widgets","full_name":"acme/widgets"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/repositories/acme/widgets/forks":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			workspace, ok := body["workspace"].(map[string]any)
			if !ok || workspace["slug"] != "forkspace" {
				t.Fatalf("unexpected fork workspace %#v", body["workspace"])
			}
			if body["name"] != "Widgets Fork" || body["description"] != "forked repo" || body["is_private"] != true {
				t.Fatalf("unexpected fork body %+v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"slug":"widgets-fork","name":"Widgets Fork","full_name":"forkspace/widgets-fork","is_private":true,"parent":{"full_name":"acme/widgets"}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
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

	private := true
	repo, err := client.ForkRepository(context.Background(), "acme", "widgets", ForkRepositoryOptions{
		Workspace:   "forkspace",
		Name:        "Widgets Fork",
		Description: "forked repo",
		IsPrivate:   &private,
	})
	if err != nil {
		t.Fatalf("ForkRepository returned error: %v", err)
	}
	if repo.FullName != "forkspace/widgets-fork" || repo.Parent == nil || repo.Parent.FullName != "acme/widgets" {
		t.Fatalf("unexpected forked repository %+v", repo)
	}
}
