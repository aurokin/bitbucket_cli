package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func writePipelineListSummary(w io.Writer, target resolvedRepoTarget, pipelines []bitbucket.Pipeline) error {
	if len(pipelines) == 0 {
		if err := writeWarnings(w, target.Warnings); err != nil {
			return err
		}
		_, err := fmt.Fprintf(w, "No pipelines found for %s/%s.\n", target.Workspace, target.Repo)
		return err
	}
	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	if err := writePipelineListTable(w, pipelines); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", pipelines[0].BuildNumber, target.Workspace, target.Repo))
}

func writePipelineListTable(w io.Writer, pipelines []bitbucket.Pipeline) error {
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "#\tstate\tref\tcreator\tcreated"); err != nil {
		return err
	}
	for _, pipeline := range pipelines {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\n",
			pipeline.BuildNumber,
			output.Truncate(pipelineStateLabel(pipeline.State), 18),
			output.Truncate(pipelineRefLabel(pipeline.Target), 20),
			output.Truncate(coalesce(pipeline.Creator.DisplayName, pipeline.Creator.Nickname, pipeline.Creator.AccountID), 20),
			formatPipelineTimestamp(pipeline.CreatedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writePipelineViewSummary(w io.Writer, payload pipelineViewPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pipeline: #%d\n", payload.Pipeline.BuildNumber); err != nil {
		return err
	}
	tw := output.NewTableWriter(w)
	for _, row := range pipelineSummaryRows(payload.Pipeline) {
		if _, err := fmt.Fprintf(tw, "%s:\t%s\n", row.Label, row.Value); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if len(payload.Steps) > 0 {
		if _, err := fmt.Fprintln(w, "\nSteps:"); err != nil {
			return err
		}
		if err := writePipelineStepTable(w, payload.Steps); err != nil {
			return err
		}
	}
	return nil
}

func writePipelineStepTable(w io.Writer, steps []bitbucket.PipelineStep) error {
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\tstate\tstarted\tcompleted"); err != nil {
		return err
	}
	for _, step := range steps {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			output.Truncate(step.Name, 28),
			output.Truncate(pipelineStateLabel(step.State), 18),
			formatPipelineTimestamp(step.StartedOn),
			formatPipelineTimestamp(step.CompletedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func pipelineSummaryRows(pipeline bitbucket.Pipeline) []summaryRow {
	rows := []summaryRow{
		{Label: "State", Value: pipelineStateLabel(pipeline.State)},
		{Label: "Ref", Value: pipelineRefLabel(pipeline.Target)},
		{Label: "UUID", Value: pipeline.UUID},
	}
	rows = appendSummaryRow(rows, "Creator", pipeline.Creator.DisplayName)
	rows = appendSummaryRow(rows, "Created", pipeline.CreatedOn)
	rows = appendSummaryRow(rows, "Completed", pipeline.CompletedOn)
	rows = appendSummaryRow(rows, "Duration", pipelineDuration(pipeline))
	rows = appendSummaryRow(rows, "URL", pipeline.Links.HTML.Href)
	return rows
}

func pipelineStepChoices(steps []bitbucket.PipelineStep) string {
	choices := make([]string, 0, len(steps))
	for _, step := range steps {
		label := step.UUID
		if step.Name != "" {
			label = fmt.Sprintf("%s (%s)", step.Name, step.UUID)
		}
		choices = append(choices, label)
	}
	return strings.Join(choices, ", ")
}

func pipelineStepLabel(step bitbucket.PipelineStep) string {
	if step.Name == "" {
		return step.UUID
	}
	return fmt.Sprintf("%s (%s)", step.Name, step.UUID)
}

func pipelineStateLabel(state bitbucket.PipelineState) string {
	switch {
	case state.Result.Name != "":
		return state.Result.Name
	case state.Name != "":
		return state.Name
	default:
		return "UNKNOWN"
	}
}

func pipelineStateActive(state bitbucket.PipelineState) bool {
	label := pipelineStateLabel(state)
	return !strings.EqualFold(label, "STOPPED") &&
		!strings.EqualFold(label, "SUCCESSFUL") &&
		!strings.EqualFold(label, "FAILED") &&
		!strings.EqualFold(label, "ERROR")
}

func pipelineRefLabel(target bitbucket.PipelineTarget) string {
	if target.RefName == "" && target.Commit.Hash == "" {
		return ""
	}
	if target.RefName != "" && target.RefType != "" {
		return fmt.Sprintf("%s:%s", target.RefType, target.RefName)
	}
	if target.RefName != "" {
		return target.RefName
	}
	return target.Commit.Hash
}

func formatPipelineTimestamp(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}
	return parsed.Format("2006-01-02 15:04")
}

func pipelineDuration(pipeline bitbucket.Pipeline) string {
	if pipeline.CreatedOn == "" || pipeline.CompletedOn == "" {
		return ""
	}
	started, err := time.Parse(time.RFC3339, pipeline.CreatedOn)
	if err != nil {
		return ""
	}
	completed, err := time.Parse(time.RFC3339, pipeline.CompletedOn)
	if err != nil || completed.Before(started) {
		return ""
	}
	return completed.Sub(started).Round(time.Second).String()
}
