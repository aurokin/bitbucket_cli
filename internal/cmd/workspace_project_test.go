package cmd

import (
	"bytes"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteWorkspaceListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := workspaceListPayload{
		Workspaces: []bitbucket.Workspace{
			{Slug: "acme", Name: "Acme", IsPrivate: true},
		},
	}
	if err := writeWorkspaceListSummary(&buf, payload); err != nil {
		t.Fatalf("writeWorkspaceListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"acme",
		"Acme",
		"private",
		"Next: bb workspace view acme",
	)
}

func TestWriteWorkspaceMembershipSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := workspaceMembershipPayload{
		Workspace: "acme",
		Membership: bitbucket.WorkspaceMembership{
			Permission: "owner",
			User: bitbucket.RepositoryPermissionUser{
				AccountID:   "user-1",
				DisplayName: "Auro",
				Nickname:    "auro",
			},
		},
	}
	if err := writeWorkspaceMembershipSummary(&buf, payload, "permissions"); err != nil {
		t.Fatalf("writeWorkspaceMembershipSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Account ID: user-1",
		"User: Auro",
		"Nickname: auro",
		"Permission: owner",
		"Next: bb workspace permissions list acme",
	)
}

func TestWriteProjectMutationSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := projectMutationPayload{
		Workspace: "acme",
		Action:    "updated",
		Project: bitbucket.Project{
			Key:       "BBCLI",
			Name:      "bb cli",
			IsPrivate: false,
			Links: bitbucket.ProjectLinks{
				HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/workspace/projects/BBCLI"},
			},
		},
	}
	if err := writeProjectMutationSummary(&buf, payload); err != nil {
		t.Fatalf("writeProjectMutationSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"Name: bb cli",
		"Action: updated",
		"Visibility: public",
		"URL: https://bitbucket.org/acme/workspace/projects/BBCLI",
		"Next: bb project view BBCLI --workspace acme",
	)
}

func TestWriteProjectUserPermissionListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := projectUserPermissionListPayload{
		Workspace:  "acme",
		ProjectKey: "BBCLI",
		Permissions: []bitbucket.ProjectUserPermission{
			{Permission: "admin", User: bitbucket.RepositoryPermissionUser{DisplayName: "Auro", AccountID: "user-1"}},
		},
	}
	if err := writeProjectUserPermissionListSummary(&buf, payload); err != nil {
		t.Fatalf("writeProjectUserPermissionListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"user-1",
		"Auro",
		"admin",
		"Next: bb project permissions user view BBCLI user-1 --workspace acme",
	)
}
