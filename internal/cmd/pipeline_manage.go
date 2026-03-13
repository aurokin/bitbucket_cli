package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
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

type pipelineVariablePayload struct {
	Host      string                     `json:"host"`
	Workspace string                     `json:"workspace"`
	Repo      string                     `json:"repo"`
	Action    string                     `json:"action,omitempty"`
	Deleted   bool                       `json:"deleted,omitempty"`
	Variable  bitbucket.PipelineVariable `json:"variable"`
}

type pipelineVariableListPayload struct {
	Host      string                       `json:"host"`
	Workspace string                       `json:"workspace"`
	Repo      string                       `json:"repo"`
	Variables []bitbucket.PipelineVariable `json:"variables"`
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

			steps, err := resolved.Client.ListPipelineSteps(context.Background(), target.Workspace, target.Repo, pipeline.UUID)
			if err != nil {
				return err
			}

			step, err := resolvePipelineStep(steps, stepRef)
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

func newPipelineVariableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage repository pipeline variables",
		Long:  "List, inspect, create, edit, and delete Bitbucket repository pipeline variables.",
		Example: "  bb pipeline variable list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured\n" +
			"  bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes",
	}

	cmd.AddCommand(
		newPipelineVariableListCmd(),
		newPipelineVariableViewCmd(),
		newPipelineVariableCreateCmd(),
		newPipelineVariableEditCmd(),
		newPipelineVariableDeleteCmd(),
	)

	return cmd
}

func newPipelineVariableListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repository pipeline variables",
		Example: "  bb pipeline variable list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline variable list --repo workspace-slug/pipelines-repo-slug --json '*'",
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

			variables, err := resolved.Client.ListPipelineVariables(context.Background(), target.Workspace, target.Repo, bitbucket.ListPipelineVariablesOptions{Limit: limit})
			if err != nil {
				return err
			}

			payload := pipelineVariableListPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Variables: variables,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of pipeline variables to return")

	return cmd
}

func newPipelineVariableViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <key-or-uuid>",
		Short: "View one repository pipeline variable",
		Example: "  bb pipeline variable view CI_TOKEN --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline variable view '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'",
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

			variable, err := resolvePipelineVariableReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "viewed",
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newPipelineVariableCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var key string
	var value string
	var valueFile string
	var secured bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a repository pipeline variable",
		Example: "  bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured\n" +
			"  printf 'token-value\\n' | bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file - --json '*'\n" +
			"  bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value production",
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

			resolvedValue, err := resolvePipelineVariableValue(cmd.InOrStdin(), value, valueFile)
			if err != nil {
				return err
			}

			variable, err := resolved.Client.CreatePipelineVariable(context.Background(), target.Workspace, target.Repo, bitbucket.PipelineVariable{
				Key:     strings.TrimSpace(key),
				Value:   resolvedValue,
				Secured: secured,
			})
			if err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "created",
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&key, "key", "", "Pipeline variable key")
	cmd.Flags().StringVar(&value, "value", "", "Pipeline variable value")
	cmd.Flags().StringVar(&valueFile, "value-file", "", "Read the pipeline variable value from a file, or '-' for stdin")
	cmd.Flags().BoolVar(&secured, "secured", false, "Mark the variable as secured")
	cmd.MarkFlagsMutuallyExclusive("value", "value-file")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func newPipelineVariableEditCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var key string
	var value string
	var valueFile string
	var secured string

	cmd := &cobra.Command{
		Use:   "edit <key-or-uuid>",
		Short: "Edit a repository pipeline variable",
		Long:  "Edit a Bitbucket repository pipeline variable by key or UUID. By default the existing secured flag is preserved unless --secured true or --secured false is provided.",
		Example: "  bb pipeline variable edit CI_TOKEN --repo workspace-slug/pipelines-repo-slug --value-file secret.txt --secured true\n" +
			"  bb pipeline variable edit '{uuid}' --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value staging --json '*'\n" +
			"  bb pipeline variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --value production",
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

			existing, err := resolvePipelineVariableReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			resolvedValue, err := resolvePipelineVariableValue(cmd.InOrStdin(), value, valueFile)
			if err != nil {
				return err
			}
			if strings.TrimSpace(key) == "" {
				key = existing.Key
			}

			nextSecured := existing.Secured
			if strings.TrimSpace(secured) != "" {
				nextSecured, err = parseBoolString(secured)
				if err != nil {
					return fmt.Errorf("--secured must be true or false")
				}
			}

			variable, err := resolved.Client.UpdatePipelineVariable(context.Background(), target.Workspace, target.Repo, existing.UUID, bitbucket.PipelineVariable{
				Key:     strings.TrimSpace(key),
				Value:   resolvedValue,
				Secured: nextSecured,
			})
			if err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "edited",
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&key, "key", "", "Override the pipeline variable key")
	cmd.Flags().StringVar(&value, "value", "", "Pipeline variable value")
	cmd.Flags().StringVar(&valueFile, "value-file", "", "Read the pipeline variable value from a file, or '-' for stdin")
	cmd.Flags().StringVar(&secured, "secured", "", "Set secured to true or false; defaults to the existing value")
	cmd.MarkFlagsMutuallyExclusive("value", "value-file")

	return cmd
}

func newPipelineVariableDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <key-or-uuid>",
		Short: "Delete a repository pipeline variable",
		Long:  "Delete a Bitbucket repository pipeline variable by key or UUID. Humans must confirm the exact repository and variable unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes\n" +
			"  bb --no-prompt pipeline variable delete '{uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'\n" +
			"  bb pipeline variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --yes",
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

			variable, err := resolvePipelineVariableReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline variable deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", target.Workspace, target.Repo, variable.Key)); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeletePipelineVariable(context.Background(), target.Workspace, target.Repo, variable.UUID); err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "deleted",
				Deleted:   true,
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
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
	if strings.TrimSpace(valueFile) != "" {
		data, err := readRequestBody(stdin, valueFile)
		if err != nil {
			return "", err
		}
		value = string(data)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("provide a pipeline variable value with --value or --value-file")
	}
	return value, nil
}

func resolvePipelineVariableReference(ctx context.Context, client *bitbucket.Client, workspace, repo, raw string) (bitbucket.PipelineVariable, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return bitbucket.PipelineVariable{}, fmt.Errorf("pipeline variable reference is required")
	}

	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		return client.GetPipelineVariable(ctx, workspace, repo, raw)
	}

	variables, err := client.ListPipelineVariables(ctx, workspace, repo, bitbucket.ListPipelineVariablesOptions{Limit: 200})
	if err != nil {
		return bitbucket.PipelineVariable{}, err
	}

	var matches []bitbucket.PipelineVariable
	for _, variable := range variables {
		if variable.Key == raw || strings.Trim(variable.UUID, "{}") == strings.Trim(raw, "{}") || variable.UUID == raw {
			matches = append(matches, variable)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return bitbucket.PipelineVariable{}, fmt.Errorf("pipeline variable %q is ambiguous; use a UUID instead", raw)
	}

	return bitbucket.PipelineVariable{}, fmt.Errorf("pipeline variable %q was not found", raw)
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

func writePipelineVariableListSummary(w io.Writer, payload pipelineVariableListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Variables) == 0 {
		if _, err := fmt.Fprintf(w, "No pipeline variables found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pipeline variable create --repo %s/%s --key EXAMPLE --value value", payload.Workspace, payload.Repo))
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "key\tsecured\tuuid"); err != nil {
		return err
	}
	for _, variable := range payload.Variables {
		if _, err := fmt.Fprintf(tw, "%s\t%t\t%s\n", output.Truncate(variable.Key, 32), variable.Secured, output.Truncate(variable.UUID, 40)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline variable view %s --repo %s/%s", payload.Variables[0].Key, payload.Workspace, payload.Repo))
}

func writePipelineVariableSummary(w io.Writer, payload pipelineVariablePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Variable", payload.Variable.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
	} else {
		if err := writeLabelValue(w, "Secured", fmt.Sprintf("%t", payload.Variable.Secured)); err != nil {
			return err
		}
		if err := writeLabelValue(w, "UUID", payload.Variable.UUID); err != nil {
			return err
		}
		if !payload.Variable.Secured {
			if err := writeLabelValue(w, "Value", payload.Variable.Value); err != nil {
				return err
			}
		}
	}
	if payload.Deleted {
		return writeNextStep(w, fmt.Sprintf("bb pipeline variable list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline variable list --repo %s/%s", payload.Workspace, payload.Repo))
}

func parseBoolString(raw string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(raw))
}
