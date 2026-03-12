package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestListPipelines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("status"); got != "COMPLETED" {
			t.Fatalf("unexpected status %q", got)
		}
		if got := r.URL.Query().Get("sort"); got != "-created_on" {
			t.Fatalf("unexpected sort %q", got)
		}
		if got := r.URL.Query().Get("pagelen"); got != "10" {
			t.Fatalf("unexpected pagelen %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{uuid-1}","build_number":7,"state":{"name":"COMPLETED","result":{"name":"SUCCESSFUL"}},"target":{"ref_name":"main"},"creator":{"display_name":"Auro"}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	pipelines, err := client.ListPipelines(context.Background(), "acme", "widgets", ListPipelinesOptions{
		State: "COMPLETED",
		Sort:  "-created_on",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListPipelines returned error: %v", err)
	}
	if len(pipelines) != 1 || pipelines[0].BuildNumber != 7 {
		t.Fatalf("unexpected pipelines %+v", pipelines)
	}
}

func TestGetPipeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines/{uuid-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{uuid-1}","build_number":7,"state":{"name":"IN_PROGRESS"},"target":{"ref_name":"main"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	pipeline, err := client.GetPipeline(context.Background(), "acme", "widgets", "uuid-1")
	if err != nil {
		t.Fatalf("GetPipeline returned error: %v", err)
	}
	if pipeline.UUID != "{uuid-1}" || pipeline.BuildNumber != 7 {
		t.Fatalf("unexpected pipeline %+v", pipeline)
	}
}

func TestGetPipelineByBuildNumber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "build_number=12" {
			t.Fatalf("unexpected q %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{uuid-12}","build_number":12,"state":{"name":"COMPLETED"}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	pipeline, err := client.GetPipelineByBuildNumber(context.Background(), "acme", "widgets", 12)
	if err != nil {
		t.Fatalf("GetPipelineByBuildNumber returned error: %v", err)
	}
	if pipeline.UUID != "{uuid-12}" || pipeline.BuildNumber != 12 {
		t.Fatalf("unexpected pipeline %+v", pipeline)
	}
}

func TestListPipelineSteps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines/{uuid-1}/steps" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{step-1}","name":"Build","state":{"name":"COMPLETED","result":{"name":"SUCCESSFUL"}}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	steps, err := client.ListPipelineSteps(context.Background(), "acme", "widgets", "{uuid-1}")
	if err != nil {
		t.Fatalf("ListPipelineSteps returned error: %v", err)
	}
	if len(steps) != 1 || steps[0].Name != "Build" {
		t.Fatalf("unexpected steps %+v", steps)
	}
}

func TestGetPipelineConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"repository_pipelines_configuration","enabled":false,"repository":{"type":"repository"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	config, err := client.GetPipelineConfig(context.Background(), "acme", "widgets")
	if err != nil {
		t.Fatalf("GetPipelineConfig returned error: %v", err)
	}
	if config.Type != "repository_pipelines_configuration" || config.Enabled {
		t.Fatalf("unexpected pipeline config %+v", config)
	}
}

func TestUpdatePipelineConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var payload PipelineConfig
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload.Type != "repository_pipelines_configuration" || !payload.Enabled || payload.Repository.Type != "repository" {
			t.Fatalf("unexpected update payload %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	updated, err := client.UpdatePipelineConfig(context.Background(), "acme", "widgets", PipelineConfig{Enabled: true})
	if err != nil {
		t.Fatalf("UpdatePipelineConfig returned error: %v", err)
	}
	if !updated.Enabled {
		t.Fatalf("expected enabled pipeline config, got %+v", updated)
	}
}

func TestTriggerPipeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		target, _ := payload["target"].(map[string]any)
		if target["type"] != "pipeline_ref_target" || target["ref_type"] != "branch" || target["ref_name"] != "main" {
			t.Fatalf("unexpected trigger payload %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{uuid-2}","build_number":18,"state":{"name":"PENDING"},"target":{"ref_type":"branch","ref_name":"main"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	pipeline, err := client.TriggerPipeline(context.Background(), "acme", "widgets", TriggerPipelineOptions{RefName: "main"})
	if err != nil {
		t.Fatalf("TriggerPipeline returned error: %v", err)
	}
	if pipeline.UUID != "{uuid-2}" || pipeline.BuildNumber != 18 {
		t.Fatalf("unexpected triggered pipeline %+v", pipeline)
	}
}

func TestNormalizePipelineUUID(t *testing.T) {
	if got := normalizePipelineUUID("uuid-1"); got != "{uuid-1}" {
		t.Fatalf("unexpected normalized uuid %q", got)
	}
	if got := normalizePipelineUUID("{uuid-1}"); got != "{uuid-1}" {
		t.Fatalf("unexpected preserved uuid %q", got)
	}
	if got := normalizePipelineUUID("  "); got != "" {
		t.Fatalf("unexpected empty uuid %q", got)
	}
}

func pipelineTestClient(t *testing.T) *Client {
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

func TestGetPipelineByBuildNumberNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	_, err := client.GetPipelineByBuildNumber(context.Background(), "acme", "widgets", 12)
	if err == nil || !strings.Contains(err.Error(), "pipeline #12 was not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}
