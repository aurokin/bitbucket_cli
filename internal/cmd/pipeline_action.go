package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

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

			payload, err := buildPipelineLogPayload(context.Background(), host, workspace, repo, args[0], stepRef)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineLogSummary(w, payload)
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

			payload, err := stopPipelineCommand(context.Background(), cmd, host, workspace, repo, args[0], yes)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineStopSummary(w, payload)
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

func buildPipelineLogPayload(ctx context.Context, host, workspace, repo, pipelineRef, stepRef string) (pipelineLogPayload, error) {
	resolved, pipeline, step, err := resolvePipelineStepCommandTarget(ctx, host, workspace, repo, pipelineRef, stepRef)
	if err != nil {
		return pipelineLogPayload{}, err
	}

	logOutput, err := resolved.Client.GetPipelineStepLog(ctx, resolved.Target.Workspace, resolved.Target.Repo, pipeline.UUID, step.UUID)
	if err != nil {
		if apiErr, ok := bitbucket.AsAPIError(err); ok && (apiErr.StatusCode == 404 || apiErr.StatusCode == 406) {
			return pipelineLogPayload{}, fmt.Errorf("bitbucket did not expose a log file for pipeline #%d step %s", pipeline.BuildNumber, pipelineStepLabel(step))
		}
		return pipelineLogPayload{}, err
	}

	return pipelineLogPayload{
		Host:      resolved.Target.Host,
		Workspace: resolved.Target.Workspace,
		Repo:      resolved.Target.Repo,
		Warnings:  append([]string(nil), resolved.Target.Warnings...),
		Pipeline:  pipeline,
		Step:      step,
		Log:       logOutput,
	}, nil
}

func stopPipelineCommand(ctx context.Context, cmd *cobra.Command, host, workspace, repo, pipelineRef string, yes bool) (pipelineStopPayload, error) {
	resolved, pipeline, err := resolvePipelineCommandTarget(ctx, host, workspace, repo, pipelineRef)
	if err != nil {
		return pipelineStopPayload{}, err
	}

	if err := confirmPipelineStop(cmd, resolved.Target, pipeline, yes); err != nil {
		return pipelineStopPayload{}, err
	}

	if err := resolved.Client.StopPipeline(ctx, resolved.Target.Workspace, resolved.Target.Repo, pipeline.UUID); err != nil {
		if apiErr, ok := bitbucket.AsAPIError(err); ok && apiErr.StatusCode == 400 {
			return pipelineStopPayload{}, fmt.Errorf("pipeline #%d is no longer stoppable; refresh it with `bb pipeline view %d --repo %s/%s`", pipeline.BuildNumber, pipeline.BuildNumber, resolved.Target.Workspace, resolved.Target.Repo)
		}
		return pipelineStopPayload{}, err
	}

	stoppedPipeline, err := waitForStoppedPipeline(ctx, resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, pipeline.UUID)
	if err != nil {
		return pipelineStopPayload{}, err
	}
	stopped := strings.EqualFold(pipelineStateLabel(stoppedPipeline.State), "STOPPED")

	return pipelineStopPayload{
		Host:      resolved.Target.Host,
		Workspace: resolved.Target.Workspace,
		Repo:      resolved.Target.Repo,
		Warnings:  append([]string(nil), resolved.Target.Warnings...),
		Pipeline:  stoppedPipeline,
		Stopped:   stopped,
	}, nil
}

func confirmPipelineStop(cmd *cobra.Command, target resolvedRepoTarget, pipeline bitbucket.Pipeline, yes bool) error {
	if yes {
		return nil
	}
	if !promptsEnabled(cmd) {
		return fmt.Errorf("pipeline stop requires confirmation; pass --yes or run in an interactive terminal")
	}
	return confirmExactMatch(cmd, fmt.Sprintf("%s/%s#%d", target.Workspace, target.Repo, pipeline.BuildNumber))
}

func writePipelineLogSummary(w io.Writer, payload pipelineLogPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pipeline: #%d\n", payload.Pipeline.BuildNumber); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Step", pipelineStepLabel(payload.Step)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, payload.Log); err != nil {
		return err
	}
	if !strings.HasSuffix(payload.Log, "\n") {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", payload.Pipeline.BuildNumber, payload.Workspace, payload.Repo))
}

func writePipelineStopSummary(w io.Writer, payload pipelineStopPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Pipeline: #%d\n", payload.Pipeline.BuildNumber); err != nil {
		return err
	}
	if err := writeLabelValue(w, "State", pipelineStateLabel(payload.Pipeline.State)); err != nil {
		return err
	}
	status := "finished before stop completed"
	if payload.Stopped {
		status = "stopped"
	}
	if err := writeLabelValue(w, "Status", status); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", payload.Pipeline.BuildNumber, payload.Workspace, payload.Repo))
}
