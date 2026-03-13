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

type pipelineRunnerPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Action    string                   `json:"action,omitempty"`
	Deleted   bool                     `json:"deleted,omitempty"`
	Runner    bitbucket.PipelineRunner `json:"runner"`
}

type pipelineRunnerListPayload struct {
	Host      string                     `json:"host"`
	Workspace string                     `json:"workspace"`
	Repo      string                     `json:"repo"`
	Runners   []bitbucket.PipelineRunner `json:"runners"`
}

type pipelineCachePayload struct {
	Host      string                  `json:"host"`
	Workspace string                  `json:"workspace"`
	Repo      string                  `json:"repo"`
	Action    string                  `json:"action,omitempty"`
	Deleted   bool                    `json:"deleted,omitempty"`
	Cache     bitbucket.PipelineCache `json:"cache,omitempty"`
	Name      string                  `json:"name,omitempty"`
}

type pipelineCacheListPayload struct {
	Host      string                    `json:"host"`
	Workspace string                    `json:"workspace"`
	Repo      string                    `json:"repo"`
	Caches    []bitbucket.PipelineCache `json:"caches"`
}

func newPipelineRunnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner",
		Short: "Inspect repository pipeline runners",
		Long:  "List, inspect, and delete Bitbucket repository pipeline runners. Runner creation and update remain out of scope until the official API request shape is clearer.",
	}
	cmd.AddCommand(
		newPipelineRunnerListCmd(),
		newPipelineRunnerViewCmd(),
		newPipelineRunnerDeleteCmd(),
	)
	return cmd
}

func newPipelineRunnerListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline runners",
		Example: "  bb pipeline runner list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline runner list --repo workspace-slug/pipelines-repo-slug --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			runners, err := resolved.Client.ListPipelineRunners(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := pipelineRunnerListPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Runners: runners}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineRunnerListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of pipeline runners to return")
	return cmd
}

func newPipelineRunnerViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <uuid>",
		Short: "View one pipeline runner",
		Example: "  bb pipeline runner view '{runner-uuid}' --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline runner view '{runner-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'",
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
			runner, err := resolved.Client.GetPipelineRunner(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}
			payload := pipelineRunnerPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "viewed", Runner: runner}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineRunnerSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newPipelineRunnerDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a pipeline runner",
		Long:  "Delete a Bitbucket repository pipeline runner. Humans must confirm the exact repository and runner UUID unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			runner, err := resolved.Client.GetPipelineRunner(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline runner deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, runner.UUID)); err != nil {
					return err
				}
			}
			if err := resolved.Client.DeletePipelineRunner(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, runner.UUID); err != nil {
				return err
			}
			payload := pipelineRunnerPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "deleted", Deleted: true, Runner: runner}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineRunnerSummary(w, payload)
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

func newPipelineCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Inspect and clear pipeline caches",
		Long:  "List Bitbucket repository pipeline caches, delete one cache by UUID, or clear caches by name when the official API supports it.",
	}
	cmd.AddCommand(
		newPipelineCacheListCmd(),
		newPipelineCacheDeleteCmd(),
		newPipelineCacheClearCmd(),
	)
	return cmd
}

func newPipelineCacheListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline caches",
		Example: "  bb pipeline cache list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline cache list --repo workspace-slug/pipelines-repo-slug --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			caches, err := resolved.Client.ListPipelineCaches(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := pipelineCacheListPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Caches: caches}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineCacheListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of pipeline caches to return")
	return cmd
}

func newPipelineCacheDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete one pipeline cache by UUID",
		Long:  "Delete a Bitbucket pipeline cache by UUID. Humans must confirm the exact repository and cache UUID unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline cache deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, args[0])); err != nil {
					return err
				}
			}
			if err := resolved.Client.DeletePipelineCache(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0]); err != nil {
				return err
			}
			payload := pipelineCachePayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "deleted", Deleted: true, Cache: bitbucket.PipelineCache{UUID: args[0]}}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineCacheSummary(w, payload)
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

func newPipelineCacheClearCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "clear <name>",
		Short: "Clear pipeline caches by name",
		Long:  "Clear Bitbucket pipeline caches by cache name. Humans must confirm the exact repository and cache name unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline cache clearing requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, args[0])); err != nil {
					return err
				}
			}
			if err := resolved.Client.DeletePipelineCachesByName(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0]); err != nil {
				return err
			}
			payload := pipelineCachePayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "cleared", Deleted: true, Name: args[0]}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineCacheSummary(w, payload)
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

func writePipelineRunnerListSummary(w io.Writer, payload pipelineRunnerListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Runners) == 0 {
		if _, err := fmt.Fprintf(w, "No pipeline runners found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\tstatus\tlabels\tuuid"); err != nil {
		return err
	}
	for _, runner := range payload.Runners {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			output.Truncate(runner.Name, 24),
			output.Truncate(runner.State.Status, 16),
			output.Truncate(strings.Join(runner.Labels, ","), 24),
			output.Truncate(runner.UUID, 38),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline runner view %s --repo %s/%s", payload.Runners[0].UUID, payload.Workspace, payload.Repo))
}

func writePipelineRunnerSummary(w io.Writer, payload pipelineRunnerPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Runner", payload.Runner.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pipeline runner list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	if err := writeLabelValue(w, "Name", payload.Runner.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Status", payload.Runner.State.Status); err != nil {
		return err
	}
	if len(payload.Runner.Labels) > 0 {
		if err := writeLabelValue(w, "Labels", strings.Join(payload.Runner.Labels, ", ")); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline runner list --repo %s/%s", payload.Workspace, payload.Repo))
}

func writePipelineCacheListSummary(w io.Writer, payload pipelineCacheListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Caches) == 0 {
		if _, err := fmt.Fprintf(w, "No pipeline caches found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\tsize\tuuid"); err != nil {
		return err
	}
	for _, cache := range payload.Caches {
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\n",
			output.Truncate(cache.Name, 24),
			cache.FileSizeBytes,
			output.Truncate(cache.UUID, 38),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline cache delete %s --repo %s/%s --yes", payload.Caches[0].UUID, payload.Workspace, payload.Repo))
}

func writePipelineCacheSummary(w io.Writer, payload pipelineCachePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if payload.Name != "" {
		if err := writeLabelValue(w, "Cache", payload.Name); err != nil {
			return err
		}
	} else {
		if err := writeLabelValue(w, "Cache", payload.Cache.UUID); err != nil {
			return err
		}
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline cache list --repo %s/%s", payload.Workspace, payload.Repo))
}
