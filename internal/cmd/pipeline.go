package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type pipelineViewPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Warnings  []string                 `json:"warnings,omitempty"`
	Pipeline  bitbucket.Pipeline       `json:"pipeline"`
	Steps     []bitbucket.PipelineStep `json:"steps"`
}

type pipelineLogPayload struct {
	Host      string                 `json:"host"`
	Workspace string                 `json:"workspace"`
	Repo      string                 `json:"repo"`
	Warnings  []string               `json:"warnings,omitempty"`
	Pipeline  bitbucket.Pipeline     `json:"pipeline"`
	Step      bitbucket.PipelineStep `json:"step"`
	Log       string                 `json:"log"`
}

type pipelineStopPayload struct {
	Host      string             `json:"host"`
	Workspace string             `json:"workspace"`
	Repo      string             `json:"repo"`
	Warnings  []string           `json:"warnings,omitempty"`
	Pipeline  bitbucket.Pipeline `json:"pipeline"`
	Stopped   bool               `json:"stopped"`
}

func newPipelineCmd() *cobra.Command {
	pipelineCmd := &cobra.Command{
		Use:     "pipeline",
		Aliases: []string{"pipelines"},
		Short:   "Work with Bitbucket Pipelines runs",
		Long:    "List and inspect Bitbucket Pipelines runs for one repository.",
	}

	pipelineCmd.AddCommand(
		newPipelineListCmd(),
		newPipelineLogCmd(),
		newPipelineStopCmd(),
		newPipelineViewCmd(),
	)

	return pipelineCmd
}

func newPipelineListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var state string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline runs for a repository",
		Example: "  bb pipeline list --repo workspace-slug/repo-slug\n" +
			"  bb pipeline list --repo workspace-slug/repo-slug --state COMPLETED --json build_number,state,target\n" +
			"  bb pipeline list --limit 5",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			pipelines, err := client.ListPipelines(context.Background(), target.Workspace, target.Repo, bitbucket.ListPipelinesOptions{
				State: strings.TrimSpace(state),
				Sort:  "-created_on",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, pipelines, func(w io.Writer) error {
				if len(pipelines) == 0 {
					if err := writeWarnings(w, target.Warnings); err != nil {
						return err
					}
					if _, err := fmt.Fprintf(w, "No pipelines found for %s/%s.\n", target.Workspace, target.Repo); err != nil {
						return err
					}
					return nil
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
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&state, "state", "", "Filter pipelines by pipeline state name")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pipelines to return")

	return cmd
}

func newPipelineViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <number-or-uuid>",
		Short: "View one pipeline run",
		Example: "  bb pipeline view 42 --repo workspace-slug/repo-slug\n" +
			"  bb pipeline view '{uuid}' --repo workspace-slug/repo-slug --json '*'\n" +
			"  bb pipeline view 42",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			pipeline, err := resolvePipelineReference(context.Background(), client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			steps, err := client.ListPipelineSteps(context.Background(), target.Workspace, target.Repo, pipeline.UUID)
			if err != nil {
				return err
			}

			payload := pipelineViewPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Warnings:  append([]string(nil), target.Warnings...),
				Pipeline:  pipeline,
				Steps:     steps,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineViewSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newPipelineLogCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var stepRef string

	cmd := &cobra.Command{
		Use:   "log <number-or-uuid>",
		Short: "Show the log for one pipeline step",
		Long:  "Show the raw log for a pipeline step. If the pipeline has exactly one step, bb selects it automatically. Otherwise pass --step with a step UUID or name.",
		Example: "  bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'\n" +
			"  bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug --json pipeline,step,log",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			pipeline, err := resolvePipelineReference(context.Background(), client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			steps, err := client.ListPipelineSteps(context.Background(), target.Workspace, target.Repo, pipeline.UUID)
			if err != nil {
				return err
			}

			step, err := resolvePipelineStep(steps, stepRef)
			if err != nil {
				return err
			}

			logOutput, err := client.GetPipelineStepLog(context.Background(), target.Workspace, target.Repo, pipeline.UUID, step.UUID)
			if err != nil {
				if apiErr, ok := bitbucket.AsAPIError(err); ok && (apiErr.StatusCode == 404 || apiErr.StatusCode == 406) {
					return fmt.Errorf("bitbucket did not expose a log file for pipeline #%d step %s", pipeline.BuildNumber, pipelineStepLabel(step))
				}
				return err
			}

			payload := pipelineLogPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Warnings:  append([]string(nil), target.Warnings...),
				Pipeline:  pipeline,
				Step:      step,
				Log:       logOutput,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				if err := writeWarnings(w, target.Warnings); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Pipeline: #%d\n", pipeline.BuildNumber); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Step", pipelineStepLabel(step)); err != nil {
					return err
				}
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
				if _, err := io.WriteString(w, logOutput); err != nil {
					return err
				}
				if !strings.HasSuffix(logOutput, "\n") {
					if _, err := fmt.Fprintln(w); err != nil {
						return err
					}
				}
				return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", pipeline.BuildNumber, target.Workspace, target.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&stepRef, "step", "", "Pipeline step UUID or name when a pipeline has more than one step")

	return cmd
}

func newPipelineStopCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "stop <number-or-uuid>",
		Short: "Stop a running pipeline",
		Long:  "Stop a Bitbucket pipeline run. Humans must confirm the exact repository and pipeline number unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes\n" +
			"  bb pipeline stop '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'\n" +
			"  bb --no-prompt pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes --json pipeline,stopped",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			pipeline, err := resolvePipelineReference(context.Background(), client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			confirmationTarget := fmt.Sprintf("%s/%s#%d", target.Workspace, target.Repo, pipeline.BuildNumber)
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline stop requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, confirmationTarget); err != nil {
					return err
				}
			}

			if err := client.StopPipeline(context.Background(), target.Workspace, target.Repo, pipeline.UUID); err != nil {
				if apiErr, ok := bitbucket.AsAPIError(err); ok && apiErr.StatusCode == 400 {
					return fmt.Errorf("pipeline #%d is no longer stoppable; refresh it with `bb pipeline view %d --repo %s/%s`", pipeline.BuildNumber, pipeline.BuildNumber, target.Workspace, target.Repo)
				}
				return err
			}

			stoppedPipeline, err := waitForStoppedPipeline(context.Background(), client, target.Workspace, target.Repo, pipeline.UUID)
			if err != nil {
				return err
			}

			payload := pipelineStopPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Warnings:  append([]string(nil), target.Warnings...),
				Pipeline:  stoppedPipeline,
				Stopped:   true,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				if err := writeWarnings(w, target.Warnings); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Pipeline: #%d\n", stoppedPipeline.BuildNumber); err != nil {
					return err
				}
				if err := writeLabelValue(w, "State", pipelineStateLabel(stoppedPipeline.State)); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Status", "stop requested"); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", stoppedPipeline.BuildNumber, target.Workspace, target.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")

	return cmd
}

func resolvePipelineReference(ctx context.Context, client *bitbucket.Client, workspace, repo, raw string) (bitbucket.Pipeline, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return bitbucket.Pipeline{}, fmt.Errorf("pipeline reference is required")
	}

	if buildNumber, err := strconv.Atoi(raw); err == nil {
		return client.GetPipelineByBuildNumber(ctx, workspace, repo, buildNumber)
	}

	return client.GetPipeline(ctx, workspace, repo, raw)
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
	if _, err := fmt.Fprintf(tw, "State:\t%s\n", pipelineStateLabel(payload.Pipeline.State)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tw, "Ref:\t%s\n", pipelineRefLabel(payload.Pipeline.Target)); err != nil {
		return err
	}
	if payload.Pipeline.Creator.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Creator:\t%s\n", payload.Pipeline.Creator.DisplayName); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "UUID:\t%s\n", payload.Pipeline.UUID); err != nil {
		return err
	}
	if payload.Pipeline.CreatedOn != "" {
		if _, err := fmt.Fprintf(tw, "Created:\t%s\n", payload.Pipeline.CreatedOn); err != nil {
			return err
		}
	}
	if payload.Pipeline.CompletedOn != "" {
		if _, err := fmt.Fprintf(tw, "Completed:\t%s\n", payload.Pipeline.CompletedOn); err != nil {
			return err
		}
	}
	if duration := pipelineDuration(payload.Pipeline); duration != "" {
		if _, err := fmt.Fprintf(tw, "Duration:\t%s\n", duration); err != nil {
			return err
		}
	}
	if payload.Pipeline.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(tw, "URL:\t%s\n", payload.Pipeline.Links.HTML.Href); err != nil {
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

func resolvePipelineStep(steps []bitbucket.PipelineStep, raw string) (bitbucket.PipelineStep, error) {
	raw = strings.TrimSpace(raw)
	if len(steps) == 0 {
		return bitbucket.PipelineStep{}, fmt.Errorf("no pipeline steps are available for this run")
	}
	if raw == "" {
		if len(steps) == 1 {
			return steps[0], nil
		}
		return bitbucket.PipelineStep{}, fmt.Errorf("multiple pipeline steps are available; pass --step with one of: %s", pipelineStepChoices(steps))
	}

	normalized := strings.Trim(raw, "{}")
	for _, step := range steps {
		if raw == step.UUID || normalized == strings.Trim(step.UUID, "{}") || raw == step.Name {
			return step, nil
		}
	}

	return bitbucket.PipelineStep{}, fmt.Errorf("pipeline step %q was not found; available steps: %s", raw, pipelineStepChoices(steps))
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

func waitForStoppedPipeline(ctx context.Context, client *bitbucket.Client, workspace, repo, pipelineUUID string) (bitbucket.Pipeline, error) {
	var last bitbucket.Pipeline
	for attempt := 0; attempt < 12; attempt++ {
		pipeline, err := client.GetPipeline(ctx, workspace, repo, pipelineUUID)
		if err != nil {
			return bitbucket.Pipeline{}, err
		}
		last = pipeline
		label := pipelineStateLabel(pipeline.State)
		if strings.EqualFold(label, "STOPPED") || strings.EqualFold(pipeline.State.Name, "STOPPED") {
			return pipeline, nil
		}
		time.Sleep(2 * time.Second)
	}
	return last, nil
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
