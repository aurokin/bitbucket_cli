package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type pipelineRunPayload struct {
	Host      string             `json:"host"`
	Workspace string             `json:"workspace"`
	Repo      string             `json:"repo"`
	Warnings  []string           `json:"warnings,omitempty"`
	Pipeline  bitbucket.Pipeline `json:"pipeline"`
}

type pipelineTestReportsPayload struct {
	Host      string                              `json:"host"`
	Workspace string                              `json:"workspace"`
	Repo      string                              `json:"repo"`
	Warnings  []string                            `json:"warnings,omitempty"`
	Pipeline  bitbucket.Pipeline                  `json:"pipeline"`
	Step      bitbucket.PipelineStep              `json:"step"`
	Summary   bitbucket.PipelineTestReportSummary `json:"summary"`
	TestCases []bitbucket.PipelineTestCase        `json:"test_cases,omitempty"`
}

func newPipelineRunCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var ref string
	var refType string

	cmd := &cobra.Command{
		Use:   "run [ref]",
		Short: "Trigger a pipeline run",
		Long:  "Trigger a Bitbucket pipeline run for a branch or tag. If no ref is provided, bb uses the current local branch when the repository target matches the current checkout.",
		Example: "  bb pipeline run main --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline run --repo workspace-slug/pipelines-repo-slug --ref main --json '*'\n" +
			"  bb pipeline run v1.2.3 --ref-type tag --repo workspace-slug/pipelines-repo-slug",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			target := resolved.Target

			refName, err := resolvePipelineRunRef(ref, args, target)
			if err != nil {
				return err
			}

			pipeline, err := resolved.Client.TriggerPipeline(context.Background(), target.Workspace, target.Repo, bitbucket.TriggerPipelineOptions{
				RefType: strings.TrimSpace(refType),
				RefName: refName,
			})
			if err != nil {
				return err
			}

			payload := pipelineRunPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Warnings:  append([]string(nil), target.Warnings...),
				Pipeline:  pipeline,
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
				if err := writeLabelValue(w, "Ref", pipelineRefLabel(pipeline.Target)); err != nil {
					return err
				}
				if err := writeLabelValue(w, "State", pipelineStateLabel(pipeline.State)); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", pipeline.BuildNumber, target.Workspace, target.Repo))
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch or tag name to build")
	cmd.Flags().StringVar(&refType, "ref-type", "branch", "Reference type to build: branch or tag")

	return cmd
}

func newPipelineTestReportsCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var stepRef string
	var includeCases bool
	var limit int

	cmd := &cobra.Command{
		Use:   "test-reports <number-or-uuid>",
		Short: "View pipeline test reports",
		Long:  "View Bitbucket pipeline test reports for one pipeline step. If the pipeline has exactly one step, bb selects it automatically. Otherwise pass --step with a step UUID or step name.",
		Example: "  bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug --cases --limit 50 --json '*'\n" +
			"  bb pipeline test-reports '{uuid}' --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'",
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
			target := resolved.Target

			pipeline, err := resolvePipelineReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			step, err := resolvePipelineStepReference(context.Background(), resolved.Client, target.Workspace, target.Repo, pipeline.UUID, stepRef)
			if err != nil {
				return err
			}

			summary, err := resolved.Client.GetPipelineTestReports(context.Background(), target.Workspace, target.Repo, pipeline.UUID, step.UUID)
			if err != nil {
				if apiErr, ok := bitbucket.AsAPIError(err); ok && apiErr.StatusCode == 404 {
					return fmt.Errorf("bitbucket did not expose test reports for pipeline #%d step %s", pipeline.BuildNumber, pipelineStepLabel(step))
				}
				return err
			}

			var cases []bitbucket.PipelineTestCase
			if includeCases {
				cases, err = resolved.Client.ListPipelineTestCases(context.Background(), target.Workspace, target.Repo, pipeline.UUID, step.UUID, limit)
				if err != nil {
					if apiErr, ok := bitbucket.AsAPIError(err); ok && apiErr.StatusCode == 404 {
						return fmt.Errorf("bitbucket did not expose test cases for pipeline #%d step %s", pipeline.BuildNumber, pipelineStepLabel(step))
					}
					return err
				}
			}

			payload := pipelineTestReportsPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Warnings:  append([]string(nil), target.Warnings...),
				Pipeline:  pipeline,
				Step:      step,
				Summary:   summary,
				TestCases: cases,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineTestReportsSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&stepRef, "step", "", "Pipeline step UUID or name when a pipeline has more than one step")
	cmd.Flags().BoolVar(&includeCases, "cases", false, "Include individual test cases in the response")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of test cases to return when --cases is set")

	return cmd
}

func resolvePipelineRunRef(flagRef string, args []string, target resolvedRepoTarget) (string, error) {
	flagRef = strings.TrimSpace(flagRef)
	if len(args) > 0 {
		argRef := strings.TrimSpace(args[0])
		if flagRef != "" && flagRef != argRef {
			return "", fmt.Errorf("pipeline ref %q does not match --ref %q", argRef, flagRef)
		}
		flagRef = argRef
	}
	if flagRef != "" {
		return flagRef, nil
	}
	if target.LocalRepo != nil && target.LocalRepo.RootDir != "" {
		branch, err := gitrepo.CurrentBranch(context.Background(), target.LocalRepo.RootDir)
		if err == nil && strings.TrimSpace(branch) != "" {
			return strings.TrimSpace(branch), nil
		}
	}
	return "", fmt.Errorf("could not determine a pipeline ref for %s/%s; pass a branch or tag name, or use --ref", target.Workspace, target.Repo)
}

func resolvePipelineVariableValue(stdin io.Reader, value, valueFile string) (string, error) {
	return resolveCommandVariableValue(stdin, value, valueFile, "pipeline variable")
}

func writePipelineTestReportsSummary(w io.Writer, payload pipelineTestReportsPayload) error {
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
	if _, err := fmt.Fprintln(w, "Summary:"); err != nil {
		return err
	}
	if err := writeAnyMapSummary(w, map[string]any(payload.Summary)); err != nil {
		return err
	}
	if len(payload.TestCases) > 0 {
		if _, err := fmt.Fprintln(w, "\nTest Cases:"); err != nil {
			return err
		}
		if err := writePipelineTestCaseTable(w, payload.TestCases); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", payload.Pipeline.BuildNumber, payload.Workspace, payload.Repo))
}

func writeAnyMapSummary(w io.Writer, payload map[string]any) error {
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	tw := output.NewTableWriter(w)
	for _, key := range keys {
		rendered, ok := renderSummaryValue(payload[key])
		if !ok {
			continue
		}
		if _, err := fmt.Fprintf(tw, "%s:\t%s\n", key, output.Truncate(rendered, 80)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderSummaryValue(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case string:
		if strings.TrimSpace(typed) == "" {
			return "", false
		}
		return typed, true
	case bool, float64, int, int64:
		return fmt.Sprintf("%v", typed), true
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return "", false
		}
		if string(data) == "{}" || string(data) == "[]" || string(data) == `""` {
			return "", false
		}
		return string(data), true
	}
}

func writePipelineTestCaseTable(w io.Writer, cases []bitbucket.PipelineTestCase) error {
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\tresult\tclass\tduration"); err != nil {
		return err
	}
	for _, testCase := range cases {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			output.Truncate(firstCaseValue(testCase, "name", "test_case_name"), 32),
			output.Truncate(firstCaseValue(testCase, "result", "status"), 16),
			output.Truncate(firstCaseValue(testCase, "class_name", "classname", "class"), 28),
			output.Truncate(firstCaseValue(testCase, "duration", "duration_in_seconds", "time"), 16),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func firstCaseValue(testCase bitbucket.PipelineTestCase, keys ...string) string {
	for _, key := range keys {
		if value, ok := renderSummaryValue(testCase[key]); ok {
			return value
		}
	}
	return ""
}
