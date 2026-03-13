package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWritePipelineScheduleListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineScheduleListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Schedules: []bitbucket.PipelineSchedule{
			{
				UUID:        "{schedule-1}",
				Enabled:     true,
				CronPattern: "0 0 12 * * ? *",
				Target:      bitbucket.PipelineTarget{RefType: "branch", RefName: "main"},
			},
		},
	}

	if err := writePipelineScheduleListSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineScheduleListSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"uuid",
		"{schedule-1}",
		"Next: bb pipeline schedule view {schedule-1} --repo acme/widgets",
	)
}

func TestWritePipelineScheduleSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineSchedulePayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "created",
		Schedule: bitbucket.PipelineSchedule{
			UUID:        "{schedule-1}",
			Enabled:     true,
			CronPattern: "0 0 12 * * ? *",
			Target:      bitbucket.PipelineTarget{RefType: "branch", RefName: "main"},
			CreatedOn:   "2026-03-13T00:00:00Z",
		},
	}

	if err := writePipelineScheduleSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineScheduleSummary returned error: %v", err)
	}

	got := buf.String()
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Schedule: {schedule-1}",
		"Action: created",
		"Enabled: true",
		"Ref: branch:main",
		"Cron: 0 0 12 * * ? *",
		"Next: bb pipeline schedule list --repo acme/widgets",
	)
}

func TestWritePipelineScheduleDeleteSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := pipelineSchedulePayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "deleted",
		Deleted:   true,
		Schedule:  bitbucket.PipelineSchedule{UUID: "{schedule-1}"},
	}

	if err := writePipelineScheduleSummary(&buf, payload); err != nil {
		t.Fatalf("writePipelineScheduleSummary returned error: %v", err)
	}

	if got := buf.String(); !strings.Contains(got, "Status: deleted") {
		t.Fatalf("expected deleted status in output, got %q", got)
	}
}
