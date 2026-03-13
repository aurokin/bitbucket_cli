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

type repoUserPermissionListPayload struct {
	Host        string                               `json:"host"`
	Workspace   string                               `json:"workspace"`
	Repo        string                               `json:"repo"`
	Permissions []bitbucket.RepositoryUserPermission `json:"permissions"`
}

type repoUserPermissionPayload struct {
	Host       string                             `json:"host"`
	Workspace  string                             `json:"workspace"`
	Repo       string                             `json:"repo"`
	Permission bitbucket.RepositoryUserPermission `json:"permission"`
}

type repoGroupPermissionListPayload struct {
	Host        string                                `json:"host"`
	Workspace   string                                `json:"workspace"`
	Repo        string                                `json:"repo"`
	Permissions []bitbucket.RepositoryGroupPermission `json:"permissions"`
}

type repoGroupPermissionPayload struct {
	Host       string                              `json:"host"`
	Workspace  string                              `json:"workspace"`
	Repo       string                              `json:"repo"`
	Permission bitbucket.RepositoryGroupPermission `json:"permission"`
}

func newRepoPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "permissions",
		Aliases: []string{"permission"},
		Short:   "Inspect explicit repository permissions",
		Long:    "Inspect explicit Bitbucket repository user and group permissions. Bitbucket's write/delete permission APIs remain out of scope until the API-token path is verified live.",
	}
	cmd.AddCommand(newRepoUserPermissionsCmd(), newRepoGroupPermissionsCmd())
	return cmd
}

func newRepoUserPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Inspect explicit repository user permissions",
	}
	cmd.AddCommand(newRepoUserPermissionListCmd(), newRepoUserPermissionViewCmd())
	return cmd
}

func newRepoGroupPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Inspect explicit repository group permissions",
	}
	cmd.AddCommand(newRepoGroupPermissionListCmd(), newRepoGroupPermissionViewCmd())
	return cmd
}

func newRepoUserPermissionListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List explicit repository user permissions",
		Example: "  bb repo permissions user list --repo workspace-slug/repo-slug\n" +
			"  bb repo permissions user list --repo workspace-slug/repo-slug --json permissions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			permissions, err := resolved.Client.ListRepositoryUserPermissions(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := repoUserPermissionListPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Permissions: permissions}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoUserPermissionListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of explicit repository user permissions to return")
	return cmd
}

func newRepoUserPermissionViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "view <account-id>",
		Short: "View one explicit repository user permission",
		Example: "  bb repo permissions user view 557058:example --repo workspace-slug/repo-slug\n" +
			"  bb repo permissions user view 557058:example --repo workspace-slug/repo-slug --json permission",
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
			accountID := strings.TrimSpace(args[0])
			permission, err := resolved.Client.GetRepositoryUserPermission(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, accountID)
			if err != nil {
				return err
			}
			payload := repoUserPermissionPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Permission: permission}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoUserPermissionSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newRepoGroupPermissionListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List explicit repository group permissions",
		Example: "  bb repo permissions group list --repo workspace-slug/repo-slug\n" +
			"  bb repo permissions group list --repo workspace-slug/repo-slug --json permissions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			permissions, err := resolved.Client.ListRepositoryGroupPermissions(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := repoGroupPermissionListPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Permissions: permissions}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoGroupPermissionListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of explicit repository group permissions to return")
	return cmd
}

func newRepoGroupPermissionViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "view <group-slug>",
		Short: "View one explicit repository group permission",
		Example: "  bb repo permissions group view developers --repo workspace-slug/repo-slug\n" +
			"  bb repo permissions group view developers --repo workspace-slug/repo-slug --json permission",
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
			groupSlug := strings.TrimSpace(args[0])
			permission, err := resolved.Client.GetRepositoryGroupPermission(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, groupSlug)
			if err != nil {
				return err
			}
			payload := repoGroupPermissionPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Permission: permission}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoGroupPermissionSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func writeRepoUserPermissionListSummary(w io.Writer, payload repoUserPermissionListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Permissions) == 0 {
		_, err := fmt.Fprintf(w, "No explicit repository user permissions found for %s/%s.\n", payload.Workspace, payload.Repo)
		return err
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
	return writeNextStep(w, fmt.Sprintf("bb repo permissions user view %s --repo %s/%s", payload.Permissions[0].User.AccountID, payload.Workspace, payload.Repo))
}

func writeRepoUserPermissionSummary(w io.Writer, payload repoUserPermissionPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
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
	return writeNextStep(w, fmt.Sprintf("bb repo permissions user list --repo %s/%s", payload.Workspace, payload.Repo))
}

func writeRepoGroupPermissionListSummary(w io.Writer, payload repoGroupPermissionListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Permissions) == 0 {
		_, err := fmt.Fprintf(w, "No explicit repository group permissions found for %s/%s.\n", payload.Workspace, payload.Repo)
		return err
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
	return writeNextStep(w, fmt.Sprintf("bb repo permissions group view %s --repo %s/%s", payload.Permissions[0].Group.Slug, payload.Workspace, payload.Repo))
}

func writeRepoGroupPermissionSummary(w io.Writer, payload repoGroupPermissionPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Group", payload.Permission.Group.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Permission", payload.Permission.Permission); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo permissions group list --repo %s/%s", payload.Workspace, payload.Repo))
}
