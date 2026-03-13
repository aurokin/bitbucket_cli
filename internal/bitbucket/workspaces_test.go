package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkspaceReadEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/workspaces/acme", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"slug":"acme","name":"Acme","is_private":true,"links":{"html":{"href":"https://bitbucket.org/acme/"}}}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/members", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"member","user":{"display_name":"Auro","account_id":"user-1","nickname":"auro"}}]}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/members/user-1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"permission":"owner","user":{"display_name":"Auro","account_id":"user-1","nickname":"auro"}}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/permissions", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"owner","user":{"display_name":"Auro","account_id":"user-1"}}]}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/permissions/repositories", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"admin","user":{"display_name":"Auro","account_id":"user-1"},"repository":{"full_name":"acme/widgets","slug":"widgets"}}]}`))
	})
	mux.HandleFunc("/2.0/workspaces/acme/permissions/repositories/widgets", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"admin","user":{"display_name":"Auro","account_id":"user-1"},"repository":{"full_name":"acme/widgets","slug":"widgets"}}]}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := repositoryTestClient(t)

	workspace, err := client.GetWorkspace(context.Background(), "acme")
	if err != nil {
		t.Fatalf("GetWorkspace returned error: %v", err)
	}
	if workspace.Slug != "acme" || workspace.Name != "Acme" {
		t.Fatalf("unexpected workspace %+v", workspace)
	}

	members, err := client.ListWorkspaceMembers(context.Background(), "acme", 10, "")
	if err != nil {
		t.Fatalf("ListWorkspaceMembers returned error: %v", err)
	}
	if len(members) != 1 || members[0].User.AccountID != "user-1" {
		t.Fatalf("unexpected workspace members %+v", members)
	}

	member, err := client.GetWorkspaceMember(context.Background(), "acme", "user-1")
	if err != nil {
		t.Fatalf("GetWorkspaceMember returned error: %v", err)
	}
	if member.Permission != "owner" || member.User.AccountID != "user-1" {
		t.Fatalf("unexpected workspace member %+v", member)
	}

	permissions, err := client.ListWorkspacePermissions(context.Background(), "acme", 10, "")
	if err != nil {
		t.Fatalf("ListWorkspacePermissions returned error: %v", err)
	}
	if len(permissions) != 1 || permissions[0].Permission != "owner" {
		t.Fatalf("unexpected workspace permissions %+v", permissions)
	}

	repoPermissions, err := client.ListWorkspaceRepositoryPermissions(context.Background(), "acme", "", 10, "", "")
	if err != nil {
		t.Fatalf("ListWorkspaceRepositoryPermissions returned error: %v", err)
	}
	if len(repoPermissions) != 1 || repoPermissions[0].Repository.Slug != "widgets" {
		t.Fatalf("unexpected workspace repository permissions %+v", repoPermissions)
	}

	filteredRepoPermissions, err := client.ListWorkspaceRepositoryPermissions(context.Background(), "acme", "widgets", 10, "", "")
	if err != nil {
		t.Fatalf("ListWorkspaceRepositoryPermissions with repo returned error: %v", err)
	}
	if len(filteredRepoPermissions) != 1 || filteredRepoPermissions[0].Repository.Slug != "widgets" {
		t.Fatalf("unexpected filtered workspace repository permissions %+v", filteredRepoPermissions)
	}
}
