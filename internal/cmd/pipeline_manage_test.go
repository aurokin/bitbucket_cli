package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/spf13/cobra"
)

func TestResolvePipelineRunRef(t *testing.T) {
	t.Parallel()

	ref, err := resolvePipelineRunRef("main", nil, resolvedRepoTarget{Workspace: "acme", Repo: "widgets"})
	if err != nil {
		t.Fatalf("resolvePipelineRunRef returned error: %v", err)
	}
	if ref != "main" {
		t.Fatalf("expected main ref, got %q", ref)
	}

	ref, err = resolvePipelineRunRef("", []string{"release/1.0"}, resolvedRepoTarget{Workspace: "acme", Repo: "widgets"})
	if err != nil {
		t.Fatalf("resolvePipelineRunRef returned error: %v", err)
	}
	if ref != "release/1.0" {
		t.Fatalf("expected release ref, got %q", ref)
	}
}

func TestResolvePipelineRunRefRequiresExplicitOrLocalRef(t *testing.T) {
	t.Parallel()

	_, err := resolvePipelineRunRef("", nil, resolvedRepoTarget{Workspace: "acme", Repo: "widgets"})
	if err == nil || !strings.Contains(err.Error(), "could not determine a pipeline ref") {
		t.Fatalf("expected missing ref error, got %v", err)
	}
}

func TestResolvePipelineVariableValue(t *testing.T) {
	t.Parallel()

	value, err := resolvePipelineVariableValue(bytes.NewBufferString(""), "production", "")
	if err != nil {
		t.Fatalf("resolvePipelineVariableValue returned error: %v", err)
	}
	if value != "production" {
		t.Fatalf("unexpected pipeline variable value %q", value)
	}

	value, err = resolvePipelineVariableValue(bytes.NewBufferString("from-stdin\n"), "", "-")
	if err != nil {
		t.Fatalf("resolvePipelineVariableValue returned error: %v", err)
	}
	if value != "from-stdin" {
		t.Fatalf("unexpected stdin pipeline variable value %q", value)
	}
}

func TestResolvePipelineVariableValueSupportsFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	valueFile := filepath.Join(tempDir, "value.txt")
	if err := os.WriteFile(valueFile, []byte("staging\n"), 0o644); err != nil {
		t.Fatalf("write value file: %v", err)
	}

	value, err := resolvePipelineVariableValue(bytes.NewBufferString(""), "", valueFile)
	if err != nil {
		t.Fatalf("resolvePipelineVariableValue returned error: %v", err)
	}
	if value != "staging" {
		t.Fatalf("unexpected file pipeline variable value %q", value)
	}
}

func TestResolvePipelineVariableReferenceByKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines_config/variables" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{uuid-1}","key":"APP_ENV","value":"production","secured":false}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := newPipelineManageTestClient(t)

	variable, err := resolvePipelineVariableReference(context.Background(), client, "acme", "widgets", "APP_ENV")
	if err != nil {
		t.Fatalf("resolvePipelineVariableReference returned error: %v", err)
	}
	if variable.UUID != "{uuid-1}" || variable.Key != "APP_ENV" {
		t.Fatalf("unexpected pipeline variable %+v", variable)
	}
}

func TestResolvePipelineVariableReferenceAmbiguous(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{uuid-1}","key":"APP_ENV"},{"uuid":"{uuid-2}","key":"APP_ENV"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := newPipelineManageTestClient(t)

	_, err := resolvePipelineVariableReference(context.Background(), client, "acme", "widgets", "APP_ENV")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous pipeline variable error, got %v", err)
	}
}

func TestWritePipelineTestReportsSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineTestReportsPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
		Pipeline:  bitbucket.Pipeline{BuildNumber: 42},
		Step:      bitbucket.PipelineStep{Name: "Tests", UUID: "{step-1}"},
		Summary: bitbucket.PipelineTestReportSummary{
			"failed":     1,
			"successful": 12,
		},
		TestCases: []bitbucket.PipelineTestCase{
			{"name": "TestOne", "result": "PASSED", "classname": "suite.A"},
		},
	}

	if err := writePipelineTestReportsSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineTestReportsSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Pipeline: #42",
		"Step: Tests ({step-1})",
		"Summary:",
		"Test Cases:",
		"Next: bb pipeline view 42 --repo acme/widgets",
	)
}

func TestWritePipelineRunSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineRunPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
		Pipeline: bitbucket.Pipeline{
			BuildNumber: 42,
			State:       bitbucket.PipelineState{Result: bitbucket.PipelineResult{Name: "PENDING"}},
			Target:      bitbucket.PipelineTarget{RefType: "branch", RefName: "main"},
		},
	}

	if err := writePipelineRunSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineRunSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Pipeline: #42",
		"Ref: branch:main",
		"State: PENDING",
		"Next: bb pipeline view 42 --repo acme/widgets",
	)
}

func TestWritePipelineVariableSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineVariablePayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "created",
		Variable:  bitbucket.PipelineVariable{UUID: "{uuid-1}", Key: "APP_ENV", Value: "production"},
	}

	if err := writePipelineVariableSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineVariableSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Variable: APP_ENV",
		"Action: created",
		"Secured: false",
		"UUID: {uuid-1}",
		"Value: production",
		"Next: bb pipeline variable list --repo acme/widgets",
	)
}

func TestWritePipelineVariableListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineVariableListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Variables: []bitbucket.PipelineVariable{{UUID: "{uuid-1}", Key: "APP_ENV", Secured: false}},
	}

	if err := writePipelineVariableListSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineVariableListSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"key",
		"APP_ENV",
		"Next: bb pipeline variable view APP_ENV --repo acme/widgets",
	)
}

func TestBuildPipelineRunPayload(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	cfg := config.Config{}
	cfg.SetHost("bitbucket.org", config.HostConfig{
		AuthType: config.AuthTypeAPIToken,
		Username: "agent@example.com",
		Token:    "secret",
	}, true)
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/pipelines" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{pipeline-1}","build_number":42,"state":{"result":{"name":"PENDING"}},"target":{"ref_type":"branch","ref_name":"main"}}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	payload, err := buildPipelineRunPayload(context.Background(), "", "acme", "widgets", "main", nil, "branch")
	if err != nil {
		t.Fatalf("buildPipelineRunPayload returned error: %v", err)
	}
	if payload.Workspace != "acme" || payload.Repo != "widgets" || payload.Pipeline.BuildNumber != 42 {
		t.Fatalf("unexpected pipeline run payload %+v", payload)
	}
}

func TestConfirmPipelineStop(t *testing.T) {
	t.Parallel()

	target := resolvedRepoTarget{Workspace: "acme", Repo: "widgets"}
	pipeline := bitbucket.Pipeline{BuildNumber: 42}

	if err := confirmPipelineStop(&cobra.Command{}, target, pipeline, true); err != nil {
		t.Fatalf("confirmPipelineStop with --yes returned error: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString(""))
	cmd.SetOut(&bytes.Buffer{})
	cmd.Flags().Bool("no-prompt", false, "")
	if err := cmd.Flags().Set("no-prompt", "true"); err != nil {
		t.Fatalf("set no-prompt flag: %v", err)
	}

	err := confirmPipelineStop(cmd, target, pipeline, false)
	if err == nil || err.Error() != "pipeline stop requires confirmation; pass --yes or run in an interactive terminal" {
		t.Fatalf("expected non-interactive confirmation error, got %v", err)
	}
}

func newPipelineManageTestClient(t *testing.T) *bitbucket.Client {
	t.Helper()

	client, err := bitbucket.NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}
