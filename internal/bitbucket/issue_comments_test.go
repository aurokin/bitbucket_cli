package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIssueComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/comments" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":3,"content":{"raw":"Looks good"},"user":{"display_name":"Auro"}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	comments, err := client.ListIssueComments(context.Background(), "acme", "widgets", 7, 10)
	if err != nil {
		t.Fatalf("ListIssueComments returned error: %v", err)
	}
	if len(comments) != 1 || comments[0].ID != 3 {
		t.Fatalf("unexpected issue comments %+v", comments)
	}
}

func TestGetIssueComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/comments/3" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"content":{"raw":"Looks good"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	comment, err := client.GetIssueComment(context.Background(), "acme", "widgets", 7, 3)
	if err != nil {
		t.Fatalf("GetIssueComment returned error: %v", err)
	}
	if comment.ID != 3 || comment.Content.Raw != "Looks good" {
		t.Fatalf("unexpected issue comment %+v", comment)
	}
}

func TestCreateIssueComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/comments" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		content, _ := payload["content"].(map[string]any)
		if content["raw"] != "Looks good" {
			t.Fatalf("unexpected create comment payload %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"content":{"raw":"Looks good"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	comment, err := client.CreateIssueComment(context.Background(), "acme", "widgets", 7, "Looks good")
	if err != nil {
		t.Fatalf("CreateIssueComment returned error: %v", err)
	}
	if comment.ID != 3 {
		t.Fatalf("unexpected created issue comment %+v", comment)
	}
}

func TestUpdateIssueComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/comments/3" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		content, _ := payload["content"].(map[string]any)
		if content["raw"] != "Updated feedback" {
			t.Fatalf("unexpected update comment payload %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"content":{"raw":"Updated feedback"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	comment, err := client.UpdateIssueComment(context.Background(), "acme", "widgets", 7, 3, "Updated feedback")
	if err != nil {
		t.Fatalf("UpdateIssueComment returned error: %v", err)
	}
	if comment.Content.Raw != "Updated feedback" {
		t.Fatalf("unexpected updated issue comment %+v", comment)
	}
}

func TestDeleteIssueComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/comments/3" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.DeleteIssueComment(context.Background(), "acme", "widgets", 7, 3); err != nil {
		t.Fatalf("DeleteIssueComment returned error: %v", err)
	}
}
