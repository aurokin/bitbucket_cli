package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPipelineSchedules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/schedules" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("pagelen"); got != "10" {
			t.Fatalf("unexpected pagelen %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{schedule-1}","enabled":true,"cron_pattern":"0 0 12 * * ? *","target":{"ref_type":"branch","ref_name":"main"}}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	schedules, err := client.ListPipelineSchedules(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListPipelineSchedules returned error: %v", err)
	}
	if len(schedules) != 1 || schedules[0].UUID != "{schedule-1}" {
		t.Fatalf("unexpected pipeline schedules %+v", schedules)
	}
}

func TestGetPipelineSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/schedules/{schedule-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{schedule-1}","enabled":true,"cron_pattern":"0 0 12 * * ? *","target":{"ref_type":"branch","ref_name":"main"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	schedule, err := client.GetPipelineSchedule(context.Background(), "acme", "widgets", "schedule-1")
	if err != nil {
		t.Fatalf("GetPipelineSchedule returned error: %v", err)
	}
	if schedule.UUID != "{schedule-1}" || schedule.Target.RefName != "main" {
		t.Fatalf("unexpected pipeline schedule %+v", schedule)
	}
}

func TestCreatePipelineSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/schedules" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload["type"] != "pipeline_schedule" {
			t.Fatalf("unexpected schedule type payload %+v", payload)
		}
		target, _ := payload["target"].(map[string]any)
		selector, _ := target["selector"].(map[string]any)
		if target["type"] != "pipeline_ref_target" || target["ref_name"] != "main" || selector["type"] != "branches" || selector["pattern"] != "main" {
			t.Fatalf("unexpected schedule create payload %+v", payload)
		}
		if payload["cron_pattern"] != "0 0 12 * * ? *" {
			t.Fatalf("unexpected cron pattern %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{schedule-1}","enabled":true,"cron_pattern":"0 0 12 * * ? *","target":{"ref_type":"branch","ref_name":"main"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	schedule, err := client.CreatePipelineSchedule(context.Background(), "acme", "widgets", CreatePipelineScheduleOptions{
		RefName:     "main",
		CronPattern: "0 0 12 * * ? *",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("CreatePipelineSchedule returned error: %v", err)
	}
	if schedule.UUID != "{schedule-1}" || !schedule.Enabled {
		t.Fatalf("unexpected created schedule %+v", schedule)
	}
}

func TestUpdatePipelineScheduleEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/schedules/{schedule-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload["type"] != "pipeline_schedule" {
			t.Fatalf("unexpected schedule type payload %+v", payload)
		}
		if payload["enabled"] != false {
			t.Fatalf("unexpected schedule update payload %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{schedule-1}","enabled":false,"cron_pattern":"0 0 12 * * ? *","target":{"ref_type":"branch","ref_name":"main"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	schedule, err := client.UpdatePipelineScheduleEnabled(context.Background(), "acme", "widgets", "schedule-1", false)
	if err != nil {
		t.Fatalf("UpdatePipelineScheduleEnabled returned error: %v", err)
	}
	if schedule.Enabled {
		t.Fatalf("expected disabled schedule, got %+v", schedule)
	}
}

func TestDeletePipelineSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/schedules/{schedule-1}" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := pipelineTestClient(t)

	if err := client.DeletePipelineSchedule(context.Background(), "acme", "widgets", "schedule-1"); err != nil {
		t.Fatalf("DeletePipelineSchedule returned error: %v", err)
	}
}
