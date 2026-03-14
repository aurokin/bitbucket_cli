package cmd

import (
	"context"
	"io"
	"strings"

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

func newPipelineCmd() *cobra.Command {
	pipelineCmd := &cobra.Command{
		Use:     "pipeline",
		Aliases: []string{"pipelines"},
		Short:   "Work with Bitbucket Pipelines runs",
		Long:    "List and inspect Bitbucket Pipelines runs for one repository.",
	}

	pipelineCmd.AddCommand(
		newPipelineCacheCmd(),
		newPipelineListCmd(),
		newPipelineRunCmd(),
		newPipelineRunnerCmd(),
		newPipelineScheduleCmd(),
		newPipelineTestReportsCmd(),
		newPipelineVariableCmd(),
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
				return writePipelineListSummary(w, target, pipelines)
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
