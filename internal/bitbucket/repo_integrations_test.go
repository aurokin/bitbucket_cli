package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
)

func TestRepositoryWebhookFlow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/repositories/acme/widgets/hooks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"uuid":"hook-1","description":"fixture hook","url":"https://example.com/hook","active":true,"events":["repo:push"]}]}`))
		case http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if body["url"] != "https://example.com/hook" || body["description"] != "fixture hook" {
				t.Fatalf("unexpected webhook create body %+v", body)
			}
			_, _ = w.Write([]byte(`{"uuid":"hook-1","description":"fixture hook","url":"https://example.com/hook","active":true,"events":["repo:push"]}`))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/hooks/hook-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"uuid":"hook-1","description":"fixture hook","url":"https://example.com/hook","active":true,"events":["repo:push"]}`))
		case http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if body["description"] != "updated hook" || body["active"] != false {
				t.Fatalf("unexpected webhook update body %+v", body)
			}
			_, _ = w.Write([]byte(`{"uuid":"hook-1","description":"updated hook","url":"https://example.com/hook","active":false,"events":["repo:push"]}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := repositoryTestClient(t)

	hooks, err := client.ListRepositoryWebhooks(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListRepositoryWebhooks returned error: %v", err)
	}
	if len(hooks) != 1 || hooks[0].UUID != "hook-1" {
		t.Fatalf("unexpected hooks %+v", hooks)
	}

	created, err := client.CreateRepositoryWebhook(context.Background(), "acme", "widgets", RepositoryWebhookMutationOptions{
		URL:         "https://example.com/hook",
		Description: "fixture hook",
		Events:      []string{"repo:push"},
	})
	if err != nil {
		t.Fatalf("CreateRepositoryWebhook returned error: %v", err)
	}
	if created.UUID != "hook-1" {
		t.Fatalf("unexpected created webhook %+v", created)
	}

	viewed, err := client.GetRepositoryWebhook(context.Background(), "acme", "widgets", "hook-1")
	if err != nil {
		t.Fatalf("GetRepositoryWebhook returned error: %v", err)
	}
	if viewed.UUID != "hook-1" {
		t.Fatalf("unexpected viewed webhook %+v", viewed)
	}

	active := false
	updated, err := client.UpdateRepositoryWebhook(context.Background(), "acme", "widgets", "hook-1", RepositoryWebhookMutationOptions{
		Description: "updated hook",
		Active:      &active,
	})
	if err != nil {
		t.Fatalf("UpdateRepositoryWebhook returned error: %v", err)
	}
	if updated.Description != "updated hook" || updated.Active {
		t.Fatalf("unexpected updated webhook %+v", updated)
	}

	if err := client.DeleteRepositoryWebhook(context.Background(), "acme", "widgets", "hook-1"); err != nil {
		t.Fatalf("DeleteRepositoryWebhook returned error: %v", err)
	}
}

func TestRepositoryDeployKeyFlow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/repositories/acme/widgets/deploy-keys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"id":7,"label":"fixture-key","key":"ssh-ed25519 AAAA fixture","comment":"fixture","repository":{"full_name":"acme/widgets"},"links":{"self":{"href":"https://example.com/key/7"}}}]}`))
		case http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if body["label"] != "fixture-key" || body["comment"] != "fixture" {
				t.Fatalf("unexpected deploy key create body %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":7,"label":"fixture-key","key":"ssh-ed25519 AAAA fixture","comment":"fixture","repository":{"full_name":"acme/widgets"},"links":{"self":{"href":"https://example.com/key/7"}}}`))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/deploy-keys/7", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":7,"label":"fixture-key","key":"ssh-ed25519 AAAA fixture","comment":"fixture","repository":{"full_name":"acme/widgets"},"links":{"self":{"href":"https://example.com/key/7"}}}`))
		case http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if body["label"] != "updated-key" || body["comment"] != "updated comment" || body["key"] != "ssh-ed25519 AAAA fixture" {
				t.Fatalf("unexpected deploy key update body %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":7,"label":"updated-key","key":"ssh-ed25519 AAAA fixture","comment":"updated comment","repository":{"full_name":"acme/widgets"},"links":{"self":{"href":"https://example.com/key/7"}}}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := repositoryTestClient(t)

	keys, err := client.ListRepositoryDeployKeys(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListRepositoryDeployKeys returned error: %v", err)
	}
	if len(keys) != 1 || keys[0].ID != 7 {
		t.Fatalf("unexpected deploy keys %+v", keys)
	}

	created, err := client.CreateRepositoryDeployKey(context.Background(), "acme", "widgets", CreateRepositoryDeployKeyOptions{
		Label:   "fixture-key",
		Key:     "ssh-ed25519 AAAA fixture",
		Comment: "fixture",
	})
	if err != nil {
		t.Fatalf("CreateRepositoryDeployKey returned error: %v", err)
	}
	if created.ID != 7 {
		t.Fatalf("unexpected created deploy key %+v", created)
	}

	viewed, err := client.GetRepositoryDeployKey(context.Background(), "acme", "widgets", 7)
	if err != nil {
		t.Fatalf("GetRepositoryDeployKey returned error: %v", err)
	}
	if viewed.ID != 7 {
		t.Fatalf("unexpected viewed deploy key %+v", viewed)
	}

	if err := client.DeleteRepositoryDeployKey(context.Background(), "acme", "widgets", 7); err != nil {
		t.Fatalf("DeleteRepositoryDeployKey returned error: %v", err)
	}
}

func repositoryTestClient(t *testing.T) *Client {
	t.Helper()

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}
