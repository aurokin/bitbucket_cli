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
