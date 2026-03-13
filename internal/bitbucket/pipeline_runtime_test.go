package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPipelineRunners(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines-config/runners" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{runner-1}","name":"linux-runner","labels":["linux"],"state":{"status":"ONLINE"}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	runners, err := client.ListPipelineRunners(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListPipelineRunners returned error: %v", err)
	}
	if len(runners) != 1 || runners[0].UUID != "{runner-1}" {
		t.Fatalf("unexpected pipeline runners %+v", runners)
	}
}

func TestGetPipelineRunner(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines-config/runners/{runner-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{runner-1}","name":"linux-runner","labels":["linux"],"state":{"status":"ONLINE"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	runner, err := client.GetPipelineRunner(context.Background(), "acme", "widgets", "runner-1")
	if err != nil {
		t.Fatalf("GetPipelineRunner returned error: %v", err)
	}
	if runner.UUID != "{runner-1}" || runner.State.Status != "ONLINE" {
		t.Fatalf("unexpected pipeline runner %+v", runner)
	}
}

func TestDeletePipelineRunner(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines-config/runners/{runner-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.DeletePipelineRunner(context.Background(), "acme", "widgets", "runner-1"); err != nil {
		t.Fatalf("DeletePipelineRunner returned error: %v", err)
	}
}

func TestListPipelineCaches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines-config/caches" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{cache-1}","name":"gomod","file_size_bytes":1234}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	caches, err := client.ListPipelineCaches(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListPipelineCaches returned error: %v", err)
	}
	if len(caches) != 1 || caches[0].UUID != "{cache-1}" {
		t.Fatalf("unexpected pipeline caches %+v", caches)
	}
}

func TestDeletePipelineCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines-config/caches/{cache-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.DeletePipelineCache(context.Background(), "acme", "widgets", "cache-1"); err != nil {
		t.Fatalf("DeletePipelineCache returned error: %v", err)
	}
}

func TestDeletePipelineCachesByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines-config/caches" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "gomod" {
			t.Fatalf("unexpected cache name query %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.DeletePipelineCachesByName(context.Background(), "acme", "widgets", "gomod"); err != nil {
		t.Fatalf("DeletePipelineCachesByName returned error: %v", err)
	}
}
