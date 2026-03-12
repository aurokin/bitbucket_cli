package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type pipelineViewPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Pipeline  bitbucket.Pipeline       `json:"pipeline"`
	Steps     []bitbucket.PipelineStep `json:"steps"`
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
		Example: "  bb pipeline list --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pipeline list --repo OhBizzle/bb-cli-integration-primary --state COMPLETED --json build_number,state,target\n" +
			"  bb pipeline list --limit 5",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, true)
			if err != nil {
				return err
			}

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
					if _, err := fmt.Fprintf(w, "No pipelines found for %s/%s.\n", target.Workspace, target.Repo); err != nil {
						return err
					}
					return nil
				}
				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				if err := writePipelineListTable(w, pipelines); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb pipeline view %d --repo %s/%s", pipelines[0].BuildNumber, target.Workspace, target.Repo))
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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
		Example: "  bb pipeline view 42 --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb pipeline view '{uuid}' --repo OhBizzle/bb-cli-integration-primary --json '*'\n" +
			"  bb pipeline view 42",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, true)
			if err != nil {
				return err
			}

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
				Pipeline:  pipeline,
				Steps:     steps,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Pipeline: #%d\n", pipeline.BuildNumber); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", pipelineStateLabel(pipeline.State)); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Ref:\t%s\n", pipelineRefLabel(pipeline.Target)); err != nil {
					return err
				}
				if pipeline.Creator.DisplayName != "" {
					if _, err := fmt.Fprintf(tw, "Creator:\t%s\n", pipeline.Creator.DisplayName); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintf(tw, "UUID:\t%s\n", pipeline.UUID); err != nil {
					return err
				}
				if pipeline.CreatedOn != "" {
					if _, err := fmt.Fprintf(tw, "Created:\t%s\n", pipeline.CreatedOn); err != nil {
						return err
					}
				}
				if pipeline.CompletedOn != "" {
					if _, err := fmt.Fprintf(tw, "Completed:\t%s\n", pipeline.CompletedOn); err != nil {
						return err
					}
				}
				if duration := pipelineDuration(pipeline); duration != "" {
					if _, err := fmt.Fprintf(tw, "Duration:\t%s\n", duration); err != nil {
						return err
					}
				}
				if pipeline.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", pipeline.Links.HTML.Href); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				if len(steps) > 0 {
					if _, err := fmt.Fprintln(w, "\nSteps:"); err != nil {
						return err
					}
					if err := writePipelineStepTable(w, steps); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

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
