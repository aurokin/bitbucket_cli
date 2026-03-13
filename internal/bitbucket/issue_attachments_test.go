package bitbucket

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestListIssueAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/attachments" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("pagelen"); got != "10" {
			t.Fatalf("unexpected pagelen %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"name":"trace.txt","links":{"self":{"href":"https://example.invalid/trace.txt"}}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	attachments, err := client.ListIssueAttachments(context.Background(), "acme", "widgets", 7, 10)
	if err != nil {
		t.Fatalf("ListIssueAttachments returned error: %v", err)
	}
	if len(attachments) != 1 || attachments[0].Name != "trace.txt" {
		t.Fatalf("unexpected issue attachments %+v", attachments)
	}
	if attachments[0].Links.Self.Href != "https://example.invalid/trace.txt" {
		t.Fatalf("unexpected issue attachment self link %+v", attachments[0].Links)
	}
}

func TestListIssueAttachmentsCapsPagelen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("pagelen"); got != "100" {
			t.Fatalf("unexpected capped pagelen %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if _, err := client.ListIssueAttachments(context.Background(), "acme", "widgets", 7, 200); err != nil {
		t.Fatalf("ListIssueAttachments returned error: %v", err)
	}
}

func TestListIssueAttachmentsAcceptsArraySelfHref(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"name":"trace.txt","links":{"self":{"href":["https://example.invalid/trace.txt"]}}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	attachments, err := client.ListIssueAttachments(context.Background(), "acme", "widgets", 7, 10)
	if err != nil {
		t.Fatalf("ListIssueAttachments returned error: %v", err)
	}
	if len(attachments) != 1 || attachments[0].Links.Self.Href != "https://example.invalid/trace.txt" {
		t.Fatalf("unexpected issue attachments %+v", attachments)
	}
}

func TestUploadIssueAttachments(t *testing.T) {
	tempDir := t.TempDir()
	filePath := tempDir + "/trace.txt"
	if err := os.WriteFile(filePath, []byte("trace output"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/issues/7/attachments" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("parse content-type: %v", err)
		}
		if !strings.HasPrefix(mediaType, "multipart/form-data") {
			t.Fatalf("unexpected content-type %q", mediaType)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read multipart body: %v", err)
		}
		if !strings.Contains(string(body), "trace.txt") || !strings.Contains(string(body), "trace output") {
			t.Fatalf("unexpected multipart body %q", string(body))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.UploadIssueAttachments(context.Background(), "acme", "widgets", 7, []string{filePath}); err != nil {
		t.Fatalf("UploadIssueAttachments returned error: %v", err)
	}
}
