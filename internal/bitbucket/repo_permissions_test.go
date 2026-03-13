package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepositoryPermissionListsAndViews(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/repositories/acme/widgets/permissions-config/users", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"admin","user":{"display_name":"Auro","account_id":"user-1"},"links":{"self":{"href":"https://example.com/u/user-1"}}}]}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/permissions-config/users/user-1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"permission":"admin","user":{"display_name":"Auro","account_id":"user-1"},"links":{"self":{"href":"https://example.com/u/user-1"}}}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/permissions-config/groups", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"permission":"read","group":{"name":"Developers","slug":"developers"},"links":{"self":{"href":"https://example.com/g/developers"}}}]}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/permissions-config/groups/developers", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"permission":"read","group":{"name":"Developers","slug":"developers"},"links":{"self":{"href":"https://example.com/g/developers"}}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := repositoryTestClient(t)

	users, err := client.ListRepositoryUserPermissions(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListRepositoryUserPermissions returned error: %v", err)
	}
	if len(users) != 1 || users[0].User.AccountID != "user-1" {
		t.Fatalf("unexpected user permissions %+v", users)
	}

	userPermission, err := client.GetRepositoryUserPermission(context.Background(), "acme", "widgets", "user-1")
	if err != nil {
		t.Fatalf("GetRepositoryUserPermission returned error: %v", err)
	}
	if userPermission.User.AccountID != "user-1" || userPermission.Permission != "admin" {
		t.Fatalf("unexpected user permission %+v", userPermission)
	}

	groups, err := client.ListRepositoryGroupPermissions(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListRepositoryGroupPermissions returned error: %v", err)
	}
	if len(groups) != 1 || groups[0].Group.Slug != "developers" {
		t.Fatalf("unexpected group permissions %+v", groups)
	}

	groupPermission, err := client.GetRepositoryGroupPermission(context.Background(), "acme", "widgets", "developers")
	if err != nil {
		t.Fatalf("GetRepositoryGroupPermission returned error: %v", err)
	}
	if groupPermission.Group.Slug != "developers" || groupPermission.Permission != "read" {
		t.Fatalf("unexpected group permission %+v", groupPermission)
	}
}
