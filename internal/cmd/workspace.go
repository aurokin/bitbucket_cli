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

type workspaceListPayload struct {
	Host       string                `json:"host"`
	Workspaces []bitbucket.Workspace `json:"workspaces"`
}

type workspacePayload struct {
	Host      string              `json:"host"`
	Workspace bitbucket.Workspace `json:"workspace"`
}

type workspaceMembershipListPayload struct {
	Host      string                          `json:"host"`
	Workspace string                          `json:"workspace"`
	Query     string                          `json:"query,omitempty"`
	Members   []bitbucket.WorkspaceMembership `json:"members"`
}

type workspaceMembershipPayload struct {
	Host       string                        `json:"host"`
	Workspace  string                        `json:"workspace"`
	Membership bitbucket.WorkspaceMembership `json:"membership"`
}

type workspaceRepoPermissionListPayload struct {
	Host        string                                    `json:"host"`
	Workspace   string                                    `json:"workspace"`
	Repo        string                                    `json:"repo,omitempty"`
	Query       string                                    `json:"query,omitempty"`
	Sort        string                                    `json:"sort,omitempty"`
	Permissions []bitbucket.WorkspaceRepositoryPermission `json:"permissions"`
}

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"workspaces"},
		Short:   "Work with Bitbucket workspaces",
		Long:    "Inspect Bitbucket workspaces, members, and workspace-scoped permissions backed by the official Bitbucket Cloud workspace APIs.",
	}
	cmd.AddCommand(
		newWorkspaceListCmd(),
		newWorkspaceViewCmd(),
		newWorkspaceMemberCmd(),
		newWorkspacePermissionCmd(),
		newWorkspaceRepoPermissionCmd(),
	)
	return cmd
}

func newWorkspaceListCmd() *cobra.Command {
	var flags formatFlags
	var host string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accessible workspaces",
		Example: "  bb workspace list\n" +
			"  bb workspace list --json workspaces\n" +
			"  bb workspace list --jq '.workspaces[].slug'",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolvedHost, client, err := resolveAuthenticatedClient(host)
			if err != nil {
				return err
			}
			workspaces, err := client.ListWorkspaces(context.Background())
			if err != nil {
				return err
			}
			payload := workspaceListPayload{Host: resolvedHost, Workspaces: workspaces}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	return cmd
}

func newWorkspaceViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string

	cmd := &cobra.Command{
		Use:   "view [workspace]",
		Short: "Show workspace information",
		Long:  "Show Bitbucket workspace information. If exactly one workspace is accessible, the workspace slug can be omitted.",
		Example: "  bb workspace view workspace-slug\n" +
			"  bb workspace view --workspace workspace-slug --json workspace\n" +
			"  bb workspace view --json '*'",
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
			item, err := client.GetWorkspace(context.Background(), selectedWorkspace)
			if err != nil {
				return err
			}
			payload := workspacePayload{Host: resolvedHost, Workspace: item}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	return cmd
}

func newWorkspaceMemberCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "member",
		Aliases: []string{"members"},
		Short:   "Inspect workspace members",
	}
	cmd.AddCommand(newWorkspaceMemberListCmd(), newWorkspaceMemberViewCmd())
	return cmd
}

func newWorkspaceMemberListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, query string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [workspace]",
		Short: "List members in a workspace",
		Example: "  bb workspace member list workspace-slug\n" +
			"  bb workspace member list --workspace workspace-slug --query 'user.account_id=\"123\"'\n" +
			"  bb workspace member list workspace-slug --json members",
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
			members, err := client.ListWorkspaceMembers(context.Background(), selectedWorkspace, limit, query)
			if err != nil {
				return err
			}
			payload := workspaceMembershipListPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Query:     query,
				Members:   members,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceMembershipListSummary(w, payload, "members")
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket workspace membership query filter")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of workspace members to return")
	return cmd
}

func newWorkspaceMemberViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string

	cmd := &cobra.Command{
		Use:   "view <account-id-or-uuid>",
		Short: "View one workspace member",
		Example: "  bb workspace member view 557058:example --workspace workspace-slug\n" +
			"  bb workspace member view '{account-uuid}' --workspace workspace-slug --json membership",
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
			memberRef := strings.TrimSpace(args[0])
			membership, err := client.GetWorkspaceMember(context.Background(), selectedWorkspace, memberRef)
			if err != nil {
				return err
			}
			payload := workspaceMembershipPayload{
				Host:       resolvedHost,
				Workspace:  selectedWorkspace,
				Membership: membership,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceMembershipSummary(w, payload, "members")
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	return cmd
}

func newWorkspacePermissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "permission",
		Aliases: []string{"permissions"},
		Short:   "Inspect workspace user permissions",
		Long:    "Inspect workspace memberships and their effective workspace permission levels.",
	}
	cmd.AddCommand(newWorkspacePermissionListCmd(), newWorkspacePermissionViewCmd())
	return cmd
}

func newWorkspacePermissionListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, query string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [workspace]",
		Short: "List user permissions in a workspace",
		Example: "  bb workspace permission list workspace-slug\n" +
			"  bb workspace permission list --workspace workspace-slug --query 'permission=\"owner\"'\n" +
			"  bb workspace permission list workspace-slug --json members",
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
			memberships, err := client.ListWorkspacePermissions(context.Background(), selectedWorkspace, limit, query)
			if err != nil {
				return err
			}
			payload := workspaceMembershipListPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Query:     query,
				Members:   memberships,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceMembershipListSummary(w, payload, "permissions")
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket workspace permission query filter")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of workspace permissions to return")
	return cmd
}

func newWorkspacePermissionViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string

	cmd := &cobra.Command{
		Use:   "view <account-id-or-uuid>",
		Short: "View one workspace user permission",
		Example: "  bb workspace permission view 557058:example --workspace workspace-slug\n" +
			"  bb workspace permission view '{account-uuid}' --workspace workspace-slug --json membership",
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
			memberRef := strings.TrimSpace(args[0])
			membership, err := client.GetWorkspaceMember(context.Background(), selectedWorkspace, memberRef)
			if err != nil {
				return err
			}
			payload := workspaceMembershipPayload{
				Host:       resolvedHost,
				Workspace:  selectedWorkspace,
				Membership: membership,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceMembershipSummary(w, payload, "permissions")
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	return cmd
}

func newWorkspaceRepoPermissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repo-permission",
		Aliases: []string{"repo-permissions"},
		Short:   "Inspect effective repository permissions in a workspace",
	}
	cmd.AddCommand(newWorkspaceRepoPermissionListCmd())
	return cmd
}

func newWorkspaceRepoPermissionListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, query, sort string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [workspace]",
		Short: "List effective repository permissions in a workspace",
		Long:  "List effective repository permissions across a workspace or for one repository within a workspace. Bitbucket only exposes list endpoints for this surface.",
		Example: "  bb workspace repo-permission list workspace-slug\n" +
			"  bb workspace repo-permission list workspace-slug --repo repo-slug\n" +
			"  bb workspace repo-permission list --workspace workspace-slug --repo workspace-slug/repo-slug --json permissions",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			selectedWorkspace, selectedRepo, err := resolveWorkspaceRepoPermissionInput(workspace, repo, firstArg(args))
			if err != nil {
				return err
			}
			resolvedHost, client, err := resolveAuthenticatedClient(host)
			if err != nil {
				return err
			}
			selectedWorkspace, err = resolveWorkspaceForCreate(context.Background(), client, selectedWorkspace)
			if err != nil {
				return err
			}
			permissions, err := client.ListWorkspaceRepositoryPermissions(context.Background(), selectedWorkspace, selectedRepo, limit, query, sort)
			if err != nil {
				return err
			}
			payload := workspaceRepoPermissionListPayload{
				Host:        resolvedHost,
				Workspace:   selectedWorkspace,
				Repo:        selectedRepo,
				Query:       query,
				Sort:        sort,
				Permissions: permissions,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeWorkspaceRepoPermissionListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().StringVar(&repo, "repo", "", "Optional repository filter as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket workspace repository permission query filter")
	cmd.Flags().StringVar(&sort, "sort", "", "Bitbucket workspace repository permission sort expression")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of workspace repository permissions to return")
	return cmd
}

func resolveWorkspaceCommandTarget(host, workspaceFlag, positional string) (string, string, *bitbucket.Client, error) {
	selectedWorkspace, err := resolveWorkspaceInput(workspaceFlag, positional)
	if err != nil {
		return "", "", nil, err
	}
	resolvedHost, client, err := resolveAuthenticatedClient(host)
	if err != nil {
		return "", "", nil, err
	}
	selectedWorkspace, err = resolveWorkspaceForCreate(context.Background(), client, selectedWorkspace)
	if err != nil {
		return "", "", nil, err
	}
	return selectedWorkspace, resolvedHost, client, nil
}

func resolveWorkspaceRepoPermissionInput(workspaceFlag, repoFlag, positionalWorkspace string) (string, string, error) {
	selectedWorkspace, err := resolveWorkspaceInput(workspaceFlag, positionalWorkspace)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(repoFlag) == "" {
		return selectedWorkspace, "", nil
	}
	target, err := parseRepositoryReference(repoFlag)
	if err != nil {
		return "", "", err
	}
	if target.Workspace != "" {
		if selectedWorkspace != "" && target.Workspace != selectedWorkspace {
			return "", "", fmt.Errorf("--workspace %q does not match repository target %q", selectedWorkspace, repoFlag)
		}
		selectedWorkspace = target.Workspace
	}
	return selectedWorkspace, target.Repo, nil
}

func writeWorkspaceListSummary(w io.Writer, payload workspaceListPayload) error {
	if len(payload.Workspaces) == 0 {
		if _, err := fmt.Fprintln(w, "No accessible workspaces found."); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "slug\tname\tvis"); err != nil {
		return err
	}
	for _, workspace := range payload.Workspaces {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(workspace.Slug, 24),
			output.Truncate(workspace.Name, 28),
			repoVisibilityLabel(workspace.IsPrivate),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb workspace view %s", payload.Workspaces[0].Slug))
}

func writeWorkspaceSummary(w io.Writer, payload workspacePayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Workspace.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Host", payload.Host); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Workspace.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "UUID", payload.Workspace.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Created", payload.Workspace.CreatedOn); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Workspace.Links.HTML.Href); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb workspace member list %s", payload.Workspace.Slug))
}

func writeWorkspaceMembershipListSummary(w io.Writer, payload workspaceMembershipListPayload, noun string) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", payload.Query); err != nil {
		return err
	}
	if len(payload.Members) == 0 {
		if _, err := fmt.Fprintf(w, "No workspace %s found for %s.\n", noun, payload.Workspace); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "account-id\tuser\tnick\tpermission"); err != nil {
		return err
	}
	for _, membership := range payload.Members {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			output.Truncate(membership.User.AccountID, 28),
			output.Truncate(membership.User.DisplayName, 24),
			output.Truncate(membership.User.Nickname, 16),
			coalesce(membership.Permission, "member"),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb workspace %s view %s --workspace %s", strings.TrimSuffix(noun, "s"), payload.Members[0].User.AccountID, payload.Workspace))
}

func writeWorkspaceMembershipSummary(w io.Writer, payload workspaceMembershipPayload, noun string) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Account ID", payload.Membership.User.AccountID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "User", payload.Membership.User.DisplayName); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Nickname", payload.Membership.User.Nickname); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Permission", payload.Membership.Permission); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Added", payload.Membership.AddedOn); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Last Accessed", payload.Membership.LastAccessed); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb workspace %s list %s", noun, payload.Workspace))
}

func writeWorkspaceRepoPermissionListSummary(w io.Writer, payload workspaceRepoPermissionListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Repository", payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", payload.Query); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Sort", payload.Sort); err != nil {
		return err
	}
	if len(payload.Permissions) == 0 {
		if _, err := fmt.Fprintf(w, "No workspace repository permissions found for %s.\n", payload.Workspace); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "repo\taccount-id\tuser\tpermission"); err != nil {
		return err
	}
	for _, permission := range payload.Permissions {
		repoName := coalesce(permission.Repository.FullName, permission.Repository.Slug, permission.Repository.Name)
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			output.Truncate(repoName, 28),
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
	if payload.Repo != "" {
		return writeNextStep(w, fmt.Sprintf("bb repo permissions user list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	return writeNextStep(w, fmt.Sprintf("bb workspace permission list %s", payload.Workspace))
}
