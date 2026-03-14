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

type projectListPayload struct {
	Host      string              `json:"host"`
	Workspace string              `json:"workspace"`
	Projects  []bitbucket.Project `json:"projects"`
}

type projectPayload struct {
	Host      string            `json:"host"`
	Workspace string            `json:"workspace"`
	Project   bitbucket.Project `json:"project"`
}

type projectMutationPayload struct {
	Host      string            `json:"host"`
	Workspace string            `json:"workspace"`
	Action    string            `json:"action"`
	Project   bitbucket.Project `json:"project"`
	Deleted   bool              `json:"deleted,omitempty"`
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "Work with Bitbucket projects",
		Long:    "Inspect and manage Bitbucket projects backed by the official Bitbucket Cloud project APIs.",
	}
	cmd.AddCommand(
		newProjectListCmd(),
		newProjectViewCmd(),
		newProjectCreateCmd(),
		newProjectEditCmd(),
		newProjectDeleteCmd(),
		newProjectDefaultReviewerCmd(),
		newProjectPermissionsCmd(),
	)
	return cmd
}

func newProjectListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [workspace]",
		Short: "List projects in a workspace",
		Long:  "List Bitbucket projects in one workspace. If you have access to exactly one workspace, the workspace can be omitted.",
		Example: "  bb project list workspace-slug\n" +
			"  bb project list --workspace workspace-slug --limit 50\n" +
			"  bb project list workspace-slug --json projects",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			selectedWorkspace, resolvedHost, client, err := resolveWorkspaceCommandTarget(host, workspace, firstArg(args))
			if err != nil {
				return err
			}
			projects, err := client.ListProjects(context.Background(), selectedWorkspace, limit)
			if err != nil {
				return err
			}
			payload := projectListPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Projects:  projects,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of projects to return")
	return cmd
}

func newProjectViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string

	cmd := &cobra.Command{
		Use:   "view <project-key>",
		Short: "Show project information",
		Example: "  bb project view BBCLI --workspace workspace-slug\n" +
			"  bb project view BBCLI --workspace workspace-slug --json project",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			selectedWorkspace, resolvedHost, client, err := resolveWorkspaceCommandTarget(host, workspace, "")
			if err != nil {
				return err
			}
			projectKey := strings.TrimSpace(args[0])
			project, err := client.GetProject(context.Background(), selectedWorkspace, projectKey)
			if err != nil {
				return err
			}
			payload := projectPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Project:   project,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	return cmd
}

func newProjectCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, name, description, visibility string

	cmd := &cobra.Command{
		Use:   "create <project-key>",
		Short: "Create a project in a workspace",
		Example: "  bb project create BBCLI --workspace workspace-slug --name 'bb cli integration'\n" +
			"  bb project create DEMO --workspace workspace-slug --name 'Demo' --visibility private --json '*'\n" +
			"  bb project create TMP --name 'Temp project'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("project creation requires --name")
			}
			selectedWorkspace, resolvedHost, client, err := resolveWorkspaceCommandTarget(host, workspace, "")
			if err != nil {
				return err
			}
			private, err := parseRepoVisibility(visibility)
			if err != nil {
				return err
			}
			project, err := client.CreateProject(context.Background(), selectedWorkspace, strings.TrimSpace(args[0]), bitbucket.CreateProjectOptions{
				Name:        strings.TrimSpace(name),
				Description: strings.TrimSpace(description),
				IsPrivate:   private,
			})
			if err != nil {
				return err
			}
			payload := projectMutationPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Action:    "created",
				Project:   project,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectMutationSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to create the project in")
	cmd.Flags().StringVar(&name, "name", "", "Project display name")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	cmd.Flags().StringVar(&visibility, "visibility", "", "Project visibility: private or public")
	return cmd
}

func newProjectEditCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, name, description, visibility string

	cmd := &cobra.Command{
		Use:   "edit <project-key>",
		Short: "Edit project metadata",
		Example: "  bb project edit BBCLI --workspace workspace-slug --description 'Updated by automation'\n" +
			"  bb project edit BBCLI --workspace workspace-slug --visibility public --json '*'\n" +
			"  bb project edit BBCLI --name 'New project name'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			private, err := parseRepoVisibility(visibility)
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" && strings.TrimSpace(description) == "" && private == nil {
				return fmt.Errorf("at least one of --name, --description, or --visibility must be provided")
			}
			selectedWorkspace, resolvedHost, client, err := resolveWorkspaceCommandTarget(host, workspace, "")
			if err != nil {
				return err
			}
			project, err := client.UpdateProject(context.Background(), selectedWorkspace, strings.TrimSpace(args[0]), bitbucket.UpdateProjectOptions{
				Name:        strings.TrimSpace(name),
				Description: strings.TrimSpace(description),
				IsPrivate:   private,
			})
			if err != nil {
				return err
			}
			payload := projectMutationPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Action:    "updated",
				Project:   project,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectMutationSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().StringVar(&name, "name", "", "New project display name")
	cmd.Flags().StringVar(&description, "description", "", "New project description")
	cmd.Flags().StringVar(&visibility, "visibility", "", "New project visibility: private or public")
	return cmd
}

func newProjectDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <project-key>",
		Short: "Delete an empty Bitbucket project",
		Long:  "Delete an empty Bitbucket project. Bitbucket requires projects to be empty before deletion.",
		Example: "  bb project delete TMP --workspace workspace-slug --yes\n" +
			"  bb --no-prompt project delete TMP --workspace workspace-slug --yes --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			selectedWorkspace, resolvedHost, client, err := resolveWorkspaceCommandTarget(host, workspace, "")
			if err != nil {
				return err
			}
			projectKey := strings.TrimSpace(args[0])
			project, err := client.GetProject(context.Background(), selectedWorkspace, projectKey)
			if err != nil {
				return err
			}
			confirmationTarget := selectedWorkspace + "/" + project.Key
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("project deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, confirmationTarget); err != nil {
					return err
				}
			}
			if err := client.DeleteProject(context.Background(), selectedWorkspace, project.Key); err != nil {
				return err
			}
			payload := projectMutationPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Action:    "deleted",
				Project:   project,
				Deleted:   true,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectMutationSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func writeProjectListSummary(w io.Writer, payload projectListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if len(payload.Projects) == 0 {
		if _, err := fmt.Fprintf(w, "No projects found in %s.\n", payload.Workspace); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "key\tname\tvis"); err != nil {
		return err
	}
	for _, project := range payload.Projects {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(project.Key, 16),
			output.Truncate(project.Name, 28),
			repoVisibilityLabel(project.IsPrivate),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project view %s --workspace %s", payload.Projects[0].Key, payload.Workspace))
}

func writeProjectSummary(w io.Writer, payload projectPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.Project.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Project.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Host", payload.Host); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Project.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "UUID", payload.Project.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Project.Links.HTML.Href); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Description", payload.Project.Description); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project default-reviewer list %s --workspace %s", payload.Project.Key, payload.Workspace))
}

func writeProjectMutationSummary(w io.Writer, payload projectMutationPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.Project.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Project.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb project create %s --workspace %s --name %q", payload.Project.Key, payload.Workspace, payload.Project.Name))
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Project.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Project.Links.HTML.Href); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project view %s --workspace %s", payload.Project.Key, payload.Workspace))
}
