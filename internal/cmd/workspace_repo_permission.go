package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/output"
)

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
