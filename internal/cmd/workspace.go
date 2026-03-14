package cmd

import (
	"context"
	"fmt"
	"io"

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
