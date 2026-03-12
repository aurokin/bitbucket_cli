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

func TestResolvePipelineStep(t *testing.T) {
	t.Parallel()

	steps := []bitbucket.PipelineStep{
		{UUID: "{step-1}", Name: "Build"},
		{UUID: "{step-2}", Name: "Test"},
	}

	step, err := resolvePipelineStep(steps, "Test")
	if err != nil {
		t.Fatalf("resolvePipelineStep returned error: %v", err)
	}
	if step.UUID != "{step-2}" {
		t.Fatalf("unexpected step %+v", step)
	}
}

func TestResolvePipelineStepRequiresExplicitSelectionWhenMultiple(t *testing.T) {
	t.Parallel()

	_, err := resolvePipelineStep([]bitbucket.PipelineStep{
		{UUID: "{step-1}", Name: "Build"},
		{UUID: "{step-2}", Name: "Test"},
	}, "")
	if err == nil || !strings.Contains(err.Error(), "multiple pipeline steps are available") {
		t.Fatalf("expected multiple-step error, got %v", err)
	}
}

func TestResolvePipelineStepFallsBackToSingleStep(t *testing.T) {
	t.Parallel()

	step, err := resolvePipelineStep([]bitbucket.PipelineStep{{UUID: "{step-1}", Name: "Build"}}, "")
	if err != nil {
		t.Fatalf("resolvePipelineStep returned error: %v", err)
	}
	if step.UUID != "{step-1}" {
		t.Fatalf("unexpected step %+v", step)
	}
}

func TestPipelineStepLabel(t *testing.T) {
	t.Parallel()

	label := pipelineStepLabel(bitbucket.PipelineStep{UUID: "{step-1}", Name: "Build"})
	if label != "Build ({step-1})" {
		t.Fatalf("unexpected step label %q", label)
	}
}

func TestWritePipelineViewSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineViewPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Pipeline: bitbucket.Pipeline{
			BuildNumber: 42,
			UUID:        "{pipeline-42}",
			State:       bitbucket.PipelineState{Result: bitbucket.PipelineResult{Name: "SUCCESSFUL"}},
			Target:      bitbucket.PipelineTarget{RefType: "branch", RefName: "main"},
			Creator:     bitbucket.PipelineActor{DisplayName: "Hunter Sadler"},
			CreatedOn:   "2026-03-11T00:00:00Z",
			CompletedOn: "2026-03-11T00:01:05Z",
			Links:       bitbucket.PipelineLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets/pipelines/results/42"}},
		},
		Steps: []bitbucket.PipelineStep{
			{Name: "Build", UUID: "{step-1}", State: bitbucket.PipelineState{Name: "COMPLETED"}},
		},
	}

	if err := writePipelineViewSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineViewSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Pipeline: #42",
		"State:",
		"https://bitbucket.org/acme/widgets/pipelines/results/42",
		"Steps:",
		"Build",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}
