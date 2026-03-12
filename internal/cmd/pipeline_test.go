package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
)

func TestPipelineStateLabel(t *testing.T) {
	t.Parallel()

	if got := pipelineStateLabel(bitbucket.PipelineState{Result: bitbucket.PipelineResult{Name: "SUCCESSFUL"}}); got != "SUCCESSFUL" {
		t.Fatalf("unexpected pipeline state label %q", got)
	}
	if got := pipelineStateLabel(bitbucket.PipelineState{Name: "IN_PROGRESS"}); got != "IN_PROGRESS" {
		t.Fatalf("unexpected pipeline state label %q", got)
	}
}

func TestPipelineRefLabel(t *testing.T) {
	t.Parallel()

	if got := pipelineRefLabel(bitbucket.PipelineTarget{RefType: "branch", RefName: "main"}); got != "branch:main" {
		t.Fatalf("unexpected ref label %q", got)
	}
	if got := pipelineRefLabel(bitbucket.PipelineTarget{Commit: bitbucket.PipelineCommit{Hash: "abc123"}}); got != "abc123" {
		t.Fatalf("unexpected commit label %q", got)
	}
}

func TestPipelineDuration(t *testing.T) {
	t.Parallel()

	pipeline := bitbucket.Pipeline{
		CreatedOn:   "2026-03-11T00:00:00Z",
		CompletedOn: "2026-03-11T00:01:05Z",
	}
	if got := pipelineDuration(pipeline); got != "1m5s" {
		t.Fatalf("unexpected duration %q", got)
	}
}

func TestWritePipelineListTable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	pipelines := []bitbucket.Pipeline{
		{
			BuildNumber: 12,
			State:       bitbucket.PipelineState{Result: bitbucket.PipelineResult{Name: "FAILED"}},
			Target:      bitbucket.PipelineTarget{RefType: "branch", RefName: "feature/some-super-long-branch-name"},
			Creator:     bitbucket.PipelineActor{DisplayName: "Hunter Sadler With A Long Name"},
			CreatedOn:   "2026-03-11T12:34:56Z",
		},
	}

	if err := writePipelineListTable(&buf, pipelines); err != nil {
		t.Fatalf("writePipelineListTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "FAILED") {
		t.Fatalf("expected pipeline state in output, got %q", got)
	}
	if !strings.Contains(got, "2026-03-11") {
		t.Fatalf("expected timestamp in output, got %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("expected truncation marker in output, got %q", got)
	}
}
