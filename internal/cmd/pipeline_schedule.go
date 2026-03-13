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

type pipelineSchedulePayload struct {
	Host      string                     `json:"host"`
	Workspace string                     `json:"workspace"`
	Repo      string                     `json:"repo"`
	Action    string                     `json:"action,omitempty"`
	Deleted   bool                       `json:"deleted,omitempty"`
	Schedule  bitbucket.PipelineSchedule `json:"schedule"`
}

type pipelineScheduleListPayload struct {
	Host      string                       `json:"host"`
	Workspace string                       `json:"workspace"`
	Repo      string                       `json:"repo"`
	Schedules []bitbucket.PipelineSchedule `json:"schedules"`
}

func newPipelineScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage pipeline schedules",
		Long:  "List, inspect, create, enable, disable, and delete Bitbucket pipeline schedules.",
		Example: "  bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'\n" +
			"  bb pipeline schedule disable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug",
	}

	cmd.AddCommand(
		newPipelineScheduleListCmd(),
		newPipelineScheduleViewCmd(),
		newPipelineScheduleCreateCmd(),
		newPipelineScheduleEnableCmd(),
		newPipelineScheduleDisableCmd(),
		newPipelineScheduleDeleteCmd(),
	)

	return cmd
}

func newPipelineScheduleListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline schedules",
		Example: "  bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			schedules, err := resolved.Client.ListPipelineSchedules(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}

			payload := pipelineScheduleListPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Schedules: schedules,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineScheduleListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of pipeline schedules to return")
	return cmd
}

func newPipelineScheduleViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <uuid>",
		Short: "View one pipeline schedule",
		Example: "  bb pipeline schedule view '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline schedule view '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'",
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

			schedule, err := resolved.Client.GetPipelineSchedule(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}

			payload := pipelineSchedulePayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Action:    "viewed",
				Schedule:  schedule,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineScheduleSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newPipelineScheduleCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var ref, cronPattern, selectorType, selectorPattern string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pipeline schedule",
		Long:  "Create a Bitbucket pipeline schedule for a branch. By default bb uses the branch name as the selector pattern and the branches selector type.",
		Example: "  bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'\n" +
			"  bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref release --cron '0 30 9 * * ? *' --enabled=false --json '*'\n" +
			"  bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --selector-type custom --selector-pattern nightly --cron '0 0 1 * * ? *'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			schedule, err := resolved.Client.CreatePipelineSchedule(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.CreatePipelineScheduleOptions{
				RefName:         strings.TrimSpace(ref),
				CronPattern:     strings.TrimSpace(cronPattern),
				Enabled:         enabled,
				SelectorType:    strings.TrimSpace(selectorType),
				SelectorPattern: strings.TrimSpace(selectorPattern),
			})
			if err != nil {
				return err
			}

			payload := pipelineSchedulePayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Action:    "created",
				Schedule:  schedule,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineScheduleSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch name to run on the schedule")
	cmd.Flags().StringVar(&cronPattern, "cron", "", "Seven-field cron pattern in UTC")
	cmd.Flags().StringVar(&selectorType, "selector-type", "", "Pipeline selector type, for example branches or custom")
	cmd.Flags().StringVar(&selectorPattern, "selector-pattern", "", "Pipeline selector pattern; defaults to the ref name")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Create the schedule enabled")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("cron")
	return cmd
}

func newPipelineScheduleEnableCmd() *cobra.Command {
	return newPipelineScheduleStateCmd("enable", true)
}

func newPipelineScheduleDisableCmd() *cobra.Command {
	return newPipelineScheduleStateCmd("disable", false)
}

func newPipelineScheduleStateCmd(name string, enabled bool) *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	action := "enabled"
	shortAction := "Enable"
	if !enabled {
		action = "disabled"
		shortAction = "Disable"
	}

	cmd := &cobra.Command{
		Use:   name + " <uuid>",
		Short: shortAction + " a pipeline schedule",
		Example: fmt.Sprintf("  bb pipeline schedule %s '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug\n", name) +
			fmt.Sprintf("  bb pipeline schedule %s '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'", name),
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

			schedule, err := resolved.Client.UpdatePipelineScheduleEnabled(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0], enabled)
			if err != nil {
				return err
			}

			payload := pipelineSchedulePayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Action:    action,
				Schedule:  schedule,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineScheduleSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newPipelineScheduleDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a pipeline schedule",
		Long:  "Delete a Bitbucket pipeline schedule. Humans must confirm the exact repository and schedule UUID unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb pipeline schedule delete '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --yes\n" +
			"  bb --no-prompt pipeline schedule delete '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'",
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

			schedule, err := resolved.Client.GetPipelineSchedule(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}

			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline schedule deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, schedule.UUID)); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeletePipelineSchedule(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, schedule.UUID); err != nil {
				return err
			}

			payload := pipelineSchedulePayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Action:    "deleted",
				Deleted:   true,
				Schedule:  schedule,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineScheduleSummary(w, payload)
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

func writePipelineScheduleListSummary(w io.Writer, payload pipelineScheduleListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Schedules) == 0 {
		if _, err := fmt.Fprintf(w, "No pipeline schedules found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pipeline schedule create --repo %s/%s --ref main --cron '0 0 12 * * ? *'", payload.Workspace, payload.Repo))
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "uuid\tenabled\tref\tcron"); err != nil {
		return err
	}
	for _, schedule := range payload.Schedules {
		if _, err := fmt.Fprintf(tw, "%s\t%t\t%s\t%s\n",
			output.Truncate(schedule.UUID, 38),
			schedule.Enabled,
			output.Truncate(schedule.Target.RefName, 16),
			output.Truncate(schedule.CronPattern, 28),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline schedule view %s --repo %s/%s", payload.Schedules[0].UUID, payload.Workspace, payload.Repo))
}

func writePipelineScheduleSummary(w io.Writer, payload pipelineSchedulePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Schedule", payload.Schedule.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pipeline schedule list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	if err := writeLabelValue(w, "Enabled", fmt.Sprintf("%t", payload.Schedule.Enabled)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Ref", pipelineRefLabel(payload.Schedule.Target)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Cron", payload.Schedule.CronPattern); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Created", payload.Schedule.CreatedOn); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Updated", payload.Schedule.UpdatedOn); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline schedule list --repo %s/%s", payload.Workspace, payload.Repo))
}
