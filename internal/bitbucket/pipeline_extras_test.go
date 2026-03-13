package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPipelineTestReports(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines/{uuid-1}/steps/{step-1}/test_reports" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"successful":12,"failed":1,"skipped":0}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	summary, err := client.GetPipelineTestReports(context.Background(), "acme", "widgets", "uuid-1", "step-1")
	if err != nil {
		t.Fatalf("GetPipelineTestReports returned error: %v", err)
	}
	if summary["successful"] != float64(12) || summary["failed"] != float64(1) {
		t.Fatalf("unexpected test report summary %+v", summary)
	}
}

func TestListPipelineTestCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines/{uuid-1}/steps/{step-1}/test_reports/test_cases" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("pagelen"); got != "2" {
			t.Fatalf("unexpected pagelen %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"name":"TestOne","result":"PASSED"},{"name":"TestTwo","result":"FAILED"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	cases, err := client.ListPipelineTestCases(context.Background(), "acme", "widgets", "uuid-1", "step-1", 2)
	if err != nil {
		t.Fatalf("ListPipelineTestCases returned error: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("expected 2 test cases, got %+v", cases)
	}
	if cases[1]["result"] != "FAILED" {
		t.Fatalf("unexpected second test case %+v", cases[1])
	}
}

func TestListPipelineVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/variables" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("pagelen"); got != "10" {
			t.Fatalf("unexpected pagelen %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{uuid-1}","key":"APP_ENV","value":"production","secured":false}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	variables, err := client.ListPipelineVariables(context.Background(), "acme", "widgets", ListPipelineVariablesOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListPipelineVariables returned error: %v", err)
	}
	if len(variables) != 1 || variables[0].Key != "APP_ENV" {
		t.Fatalf("unexpected pipeline variables %+v", variables)
	}
}

func TestGetPipelineVariable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/variables/{uuid-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{uuid-1}","key":"APP_ENV","value":"production","secured":false}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	variable, err := client.GetPipelineVariable(context.Background(), "acme", "widgets", "uuid-1")
	if err != nil {
		t.Fatalf("GetPipelineVariable returned error: %v", err)
	}
	if variable.UUID != "{uuid-1}" || variable.Key != "APP_ENV" {
		t.Fatalf("unexpected pipeline variable %+v", variable)
	}
}

func TestCreatePipelineVariable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/variables" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var payload PipelineVariable
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload.Key != "APP_ENV" || payload.Value != "production" || payload.Secured {
			t.Fatalf("unexpected create payload %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PipelineVariable{UUID: "{uuid-1}", Key: payload.Key, Value: payload.Value})
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	created, err := client.CreatePipelineVariable(context.Background(), "acme", "widgets", PipelineVariable{Key: "APP_ENV", Value: "production"})
	if err != nil {
		t.Fatalf("CreatePipelineVariable returned error: %v", err)
	}
	if created.UUID != "{uuid-1}" || created.Key != "APP_ENV" {
		t.Fatalf("unexpected created pipeline variable %+v", created)
	}
}

func TestUpdatePipelineVariable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/variables/{uuid-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var payload PipelineVariable
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload.Key != "APP_ENV" || payload.Value != "staging" || !payload.Secured {
			t.Fatalf("unexpected update payload %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PipelineVariable{UUID: "{uuid-1}", Key: payload.Key, Secured: payload.Secured})
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	updated, err := client.UpdatePipelineVariable(context.Background(), "acme", "widgets", "uuid-1", PipelineVariable{Key: "APP_ENV", Value: "staging", Secured: true})
	if err != nil {
		t.Fatalf("UpdatePipelineVariable returned error: %v", err)
	}
	if updated.UUID != "{uuid-1}" || !updated.Secured {
		t.Fatalf("unexpected updated pipeline variable %+v", updated)
	}
}

func TestDeletePipelineVariable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/variables/{uuid-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.DeletePipelineVariable(context.Background(), "acme", "widgets", "uuid-1"); err != nil {
		t.Fatalf("DeletePipelineVariable returned error: %v", err)
	}
}

func TestParsePipelineBuildNumber(t *testing.T) {
	if got, ok := parsePipelineBuildNumber("42"); !ok || got != 42 {
		t.Fatalf("expected build number 42, got %d %t", got, ok)
	}
	if got, ok := parsePipelineBuildNumber("0"); ok || got != 0 {
		t.Fatalf("expected invalid zero build number, got %d %t", got, ok)
	}
}
