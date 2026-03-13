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

type projectUserPermissionListPayload struct {
	Host        string                            `json:"host"`
	Workspace   string                            `json:"workspace"`
	ProjectKey  string                            `json:"project_key"`
	Permissions []bitbucket.ProjectUserPermission `json:"permissions"`
}

type projectUserPermissionPayload struct {
	Host       string                          `json:"host"`
	Workspace  string                          `json:"workspace"`
	ProjectKey string                          `json:"project_key"`
	Permission bitbucket.ProjectUserPermission `json:"permission"`
}

type projectGroupPermissionListPayload struct {
	Host        string                             `json:"host"`
	Workspace   string                             `json:"workspace"`
	ProjectKey  string                             `json:"project_key"`
	Permissions []bitbucket.ProjectGroupPermission `json:"permissions"`
}

type projectGroupPermissionPayload struct {
	Host       string                           `json:"host"`
	Workspace  string                           `json:"workspace"`
	ProjectKey string                           `json:"project_key"`
	Permission bitbucket.ProjectGroupPermission `json:"permission"`
}

func newProjectPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "permissions",
		Aliases: []string{"permission"},
		Short:   "Inspect explicit project permissions",
		Long:    "Inspect explicit Bitbucket project user and group permissions. Project permission writes remain out of scope until the API-token path is verified live.",
	}
	cmd.AddCommand(newProjectUserPermissionsCmd(), newProjectGroupPermissionsCmd())
	return cmd
}

func newProjectUserPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Inspect explicit project user permissions",
	}
	cmd.AddCommand(newProjectUserPermissionListCmd(), newProjectUserPermissionViewCmd())
	return cmd
}

func newProjectGroupPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Inspect explicit project group permissions",
	}
	cmd.AddCommand(newProjectGroupPermissionListCmd(), newProjectGroupPermissionViewCmd())
	return cmd
}

func newProjectUserPermissionListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string
	var limit int

	cmd := &cobra.Command{
		Use:   "list <project-key>",
		Short: "List explicit project user permissions",
		Example: "  bb project permissions user list BBCLI --workspace workspace-slug\n" +
			"  bb project permissions user list BBCLI --workspace workspace-slug --json permissions",
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
			permissions, err := client.ListProjectUserPermissions(context.Background(), selectedWorkspace, projectKey, limit)
			if err != nil {
				return err
			}
			payload := projectUserPermissionListPayload{
				Host:        resolvedHost,
				Workspace:   selectedWorkspace,
				ProjectKey:  projectKey,
				Permissions: permissions,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectUserPermissionListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of explicit project user permissions to return")
	return cmd
}

func newProjectUserPermissionViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string

	cmd := &cobra.Command{
		Use:   "view <project-key> <account-id>",
		Short: "View one explicit project user permission",
		Example: "  bb project permissions user view BBCLI 557058:example --workspace workspace-slug\n" +
			"  bb project permissions user view BBCLI 557058:example --workspace workspace-slug --json permission",
		Args: cobra.ExactArgs(2),
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
			accountID := strings.TrimSpace(args[1])
			permission, err := client.GetProjectUserPermission(context.Background(), selectedWorkspace, projectKey, accountID)
			if err != nil {
				return err
			}
			payload := projectUserPermissionPayload{
				Host:       resolvedHost,
				Workspace:  selectedWorkspace,
				ProjectKey: projectKey,
				Permission: permission,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectUserPermissionSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	return cmd
}

func newProjectGroupPermissionListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string
	var limit int

	cmd := &cobra.Command{
		Use:   "list <project-key>",
		Short: "List explicit project group permissions",
		Example: "  bb project permissions group list BBCLI --workspace workspace-slug\n" +
			"  bb project permissions group list BBCLI --workspace workspace-slug --json permissions",
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
			permissions, err := client.ListProjectGroupPermissions(context.Background(), selectedWorkspace, projectKey, limit)
			if err != nil {
				return err
			}
			payload := projectGroupPermissionListPayload{
				Host:        resolvedHost,
				Workspace:   selectedWorkspace,
				ProjectKey:  projectKey,
				Permissions: permissions,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectGroupPermissionListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of explicit project group permissions to return")
	return cmd
}

func newProjectGroupPermissionViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string

	cmd := &cobra.Command{
		Use:   "view <project-key> <group-slug>",
		Short: "View one explicit project group permission",
		Example: "  bb project permissions group view BBCLI developers --workspace workspace-slug\n" +
			"  bb project permissions group view BBCLI developers --workspace workspace-slug --json permission",
		Args: cobra.ExactArgs(2),
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
			groupSlug := strings.TrimSpace(args[1])
			permission, err := client.GetProjectGroupPermission(context.Background(), selectedWorkspace, projectKey, groupSlug)
			if err != nil {
				return err
			}
			payload := projectGroupPermissionPayload{
				Host:       resolvedHost,
				Workspace:  selectedWorkspace,
				ProjectKey: projectKey,
				Permission: permission,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectGroupPermissionSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	return cmd
}

func writeProjectUserPermissionListSummary(w io.Writer, payload projectUserPermissionListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.ProjectKey); err != nil {
		return err
	}
	if len(payload.Permissions) == 0 {
		if _, err := fmt.Fprintf(w, "No explicit project user permissions found for %s/%s.\n", payload.Workspace, payload.ProjectKey); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "account-id\tuser\tpermission"); err != nil {
		return err
	}
	for _, permission := range payload.Permissions {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(permission.User.AccountID, 28),
			output.Truncate(permission.User.DisplayName, 24),
			permission.Permission,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project permissions user view %s %s --workspace %s", payload.ProjectKey, payload.Permissions[0].User.AccountID, payload.Workspace))
}

func writeProjectUserPermissionSummary(w io.Writer, payload projectUserPermissionPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.ProjectKey); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Account ID", payload.Permission.User.AccountID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "User", payload.Permission.User.DisplayName); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Permission", payload.Permission.Permission); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project permissions user list %s --workspace %s", payload.ProjectKey, payload.Workspace))
}

func writeProjectGroupPermissionListSummary(w io.Writer, payload projectGroupPermissionListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.ProjectKey); err != nil {
		return err
	}
	if len(payload.Permissions) == 0 {
		if _, err := fmt.Fprintf(w, "No explicit project group permissions found for %s/%s.\n", payload.Workspace, payload.ProjectKey); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "group\tpermission"); err != nil {
		return err
	}
	for _, permission := range payload.Permissions {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n",
			output.Truncate(permission.Group.Slug, 24),
			permission.Permission,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project permissions group view %s %s --workspace %s", payload.ProjectKey, payload.Permissions[0].Group.Slug, payload.Workspace))
}

func writeProjectGroupPermissionSummary(w io.Writer, payload projectGroupPermissionPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.ProjectKey); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Group", payload.Permission.Group.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Permission", payload.Permission.Permission); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project permissions group list %s --workspace %s", payload.ProjectKey, payload.Workspace))
}
