package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWritePipelineRunnerListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineRunnerListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Runners: []bitbucket.PipelineRunner{
			{UUID: "{runner-1}", Name: "linux-runner", Labels: []string{"linux"}, State: bitbucket.PipelineRunnerState{Status: "ONLINE"}},
		},
	}

	if err := writePipelineRunnerListSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineRunnerListSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"linux-runner",
		"ONLINE",
		"Next: bb pipeline runner view {runner-1} --repo acme/widgets",
	)
}

func TestWritePipelineCacheListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineCacheListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Caches:    []bitbucket.PipelineCache{{UUID: "{cache-1}", Name: "gomod", FileSizeBytes: 1234}},
	}

	if err := writePipelineCacheListSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineCacheListSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"gomod",
		"{cache-1}",
		"Next: bb pipeline cache delete {cache-1} --repo acme/widgets --yes",
	)
}

func TestWritePipelineRunnerSummaryDelete(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineRunnerPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "deleted",
		Deleted:   true,
		Runner:    bitbucket.PipelineRunner{UUID: "{runner-1}"},
	}

	if err := writePipelineRunnerSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineRunnerSummary returned error: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "Status: deleted") {
		t.Fatalf("expected deleted status in output, got %q", got)
	}
}

func TestWritePipelineCacheSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineCachePayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "deleted",
		Deleted:   true,
		Name:      "gomod",
		Cache:     bitbucket.PipelineCache{UUID: "{cache-1}"},
	}

	if err := writePipelineCacheSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineCacheSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Cache: gomod",
		"Action: deleted",
		"Status: deleted",
		"Next: bb pipeline cache list --repo acme/widgets",
	)
}
