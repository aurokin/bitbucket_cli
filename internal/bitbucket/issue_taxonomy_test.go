package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIssueMilestones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/milestones" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":1,"name":"v1.0"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	items, err := client.ListIssueMilestones(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListIssueMilestones returned error: %v", err)
	}
	if len(items) != 1 || items[0].Name != "v1.0" {
		t.Fatalf("unexpected milestones %+v", items)
	}
}

func TestGetIssueMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/milestones/1" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"name":"v1.0"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	item, err := client.GetIssueMilestone(context.Background(), "acme", "widgets", 1)
	if err != nil {
		t.Fatalf("GetIssueMilestone returned error: %v", err)
	}
	if item.ID != 1 || item.Name != "v1.0" {
		t.Fatalf("unexpected milestone %+v", item)
	}
}

func TestListIssueComponents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/components" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"id":2,"name":"backend"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	items, err := client.ListIssueComponents(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListIssueComponents returned error: %v", err)
	}
	if len(items) != 1 || items[0].Name != "backend" {
		t.Fatalf("unexpected components %+v", items)
	}
}

func TestGetIssueComponent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/components/2" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":2,"name":"backend"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	item, err := client.GetIssueComponent(context.Background(), "acme", "widgets", 2)
	if err != nil {
		t.Fatalf("GetIssueComponent returned error: %v", err)
	}
	if item.ID != 2 || item.Name != "backend" {
		t.Fatalf("unexpected component %+v", item)
	}
}
