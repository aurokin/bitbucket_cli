package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/workspaces/acme/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"key":"BBCLI","name":"bb cli","is_private":false}]}`))
		case http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create project body: %v", err)
			}
			if body["key"] != "TMP" || body["name"] != "Temp Project" {
				t.Fatalf("unexpected create project body %+v", body)
			}
			_, _ = w.Write([]byte(`{"key":"TMP","name":"Temp Project","is_private":true}`))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/2.0/workspaces/acme/projects/BBCLI", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"key":"BBCLI","name":"bb cli","description":"fixtures","is_private":false}`))
		case http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update project body: %v", err)
			}
			if body["name"] != "bb cli" || body["description"] != "updated" || body["is_private"] != true {
				t.Fatalf("unexpected update project body %+v", body)
			}
			_, _ = w.Write([]byte(`{"key":"BBCLI","name":"bb cli","description":"updated","is_private":true}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/2.0/workspaces/acme/projects/BBCLI/default-reviewers", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"reviewer_type":"project","user":{"display_name":"Reviewer","account_id":"user-1"}}]}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/projects/BBCLI/permissions-config/users", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"admin","user":{"display_name":"Auro","account_id":"user-1"}}]}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/projects/BBCLI/permissions-config/users/user-1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"permission":"admin","user":{"display_name":"Auro","account_id":"user-1"}}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/projects/BBCLI/permissions-config/groups", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"write","group":{"slug":"developers","name":"Developers"}}]}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/projects/BBCLI/permissions-config/groups/developers", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"permission":"write","group":{"slug":"developers","name":"Developers"}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := repositoryTestClient(t)

	projects, err := client.ListProjects(context.Background(), "acme", 10)
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}
	if len(projects) != 1 || projects[0].Key != "BBCLI" {
		t.Fatalf("unexpected projects %+v", projects)
	}

	project, err := client.GetProject(context.Background(), "acme", "BBCLI")
	if err != nil {
		t.Fatalf("GetProject returned error: %v", err)
	}
	if project.Key != "BBCLI" || project.Description != "fixtures" {
		t.Fatalf("unexpected project %+v", project)
	}

	created, err := client.CreateProject(context.Background(), "acme", "TMP", CreateProjectOptions{
		Name:      "Temp Project",
		IsPrivate: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}
	if created.Key != "TMP" || !created.IsPrivate {
		t.Fatalf("unexpected created project %+v", created)
	}

	updated, err := client.UpdateProject(context.Background(), "acme", "BBCLI", UpdateProjectOptions{
		Description: "updated",
		IsPrivate:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("UpdateProject returned error: %v", err)
	}
	if updated.Description != "updated" || !updated.IsPrivate {
		t.Fatalf("unexpected updated project %+v", updated)
	}

	if err := client.DeleteProject(context.Background(), "acme", "BBCLI"); err != nil {
		t.Fatalf("DeleteProject returned error: %v", err)
	}

	reviewers, err := client.ListProjectDefaultReviewers(context.Background(), "acme", "BBCLI", 10)
	if err != nil {
		t.Fatalf("ListProjectDefaultReviewers returned error: %v", err)
	}
	if len(reviewers) != 1 || reviewers[0].User.AccountID != "user-1" {
		t.Fatalf("unexpected reviewers %+v", reviewers)
	}

	userPermissions, err := client.ListProjectUserPermissions(context.Background(), "acme", "BBCLI", 10)
	if err != nil {
		t.Fatalf("ListProjectUserPermissions returned error: %v", err)
	}
	if len(userPermissions) != 1 || userPermissions[0].User.AccountID != "user-1" {
		t.Fatalf("unexpected user permissions %+v", userPermissions)
	}

	userPermission, err := client.GetProjectUserPermission(context.Background(), "acme", "BBCLI", "user-1")
	if err != nil {
		t.Fatalf("GetProjectUserPermission returned error: %v", err)
	}
	if userPermission.Permission != "admin" {
		t.Fatalf("unexpected user permission %+v", userPermission)
	}

	groupPermissions, err := client.ListProjectGroupPermissions(context.Background(), "acme", "BBCLI", 10)
	if err != nil {
		t.Fatalf("ListProjectGroupPermissions returned error: %v", err)
	}
	if len(groupPermissions) != 1 || groupPermissions[0].Group.Slug != "developers" {
		t.Fatalf("unexpected group permissions %+v", groupPermissions)
	}

	groupPermission, err := client.GetProjectGroupPermission(context.Background(), "acme", "BBCLI", "developers")
	if err != nil {
		t.Fatalf("GetProjectGroupPermission returned error: %v", err)
	}
	if groupPermission.Permission != "write" {
		t.Fatalf("unexpected group permission %+v", groupPermission)
	}
}

func boolPtr(value bool) *bool {
	return &value
}
