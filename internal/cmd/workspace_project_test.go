package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
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

func TestResolveWorkspaceCommandTargetAutoSelectsSingleWorkspace(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	cfg := config.Config{}
	cfg.SetHost("bitbucket.org", config.HostConfig{
		AuthType: config.AuthTypeAPIToken,
		Username: "agent@example.com",
		Token:    "secret",
	}, true)
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/workspaces" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"acme","name":"Acme"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	workspace, host, client, err := resolveWorkspaceCommandTarget("", "", "")
	if err != nil {
		t.Fatalf("resolveWorkspaceCommandTarget returned error: %v", err)
	}
	if workspace != "acme" || host != "bitbucket.org" || client == nil {
		t.Fatalf("unexpected resolution %q %q %v", workspace, host, client)
	}
}

func TestResolveWorkspaceCommandTargetRejectsMultipleWorkspacesWithoutSelection(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	cfg := config.Config{}
	cfg.SetHost("bitbucket.org", config.HostConfig{
		AuthType: config.AuthTypeAPIToken,
		Username: "agent@example.com",
		Token:    "secret",
	}, true)
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"acme","name":"Acme"},{"slug":"other","name":"Other"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")

	_, _, _, err := resolveWorkspaceCommandTarget("", "", "")
	if err == nil || err.Error() != "multiple workspaces available; specify --workspace" {
		t.Fatalf("expected multiple workspace error, got %v", err)
	}
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

func TestWriteProjectListAndSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listPayload := projectListPayload{
		Workspace: "acme",
		Projects: []bitbucket.Project{
			{Key: "BBCLI", Name: "bb cli", IsPrivate: true},
		},
	}
	if err := writeProjectListSummary(&buf, listPayload); err != nil {
		t.Fatalf("writeProjectListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"BBCLI",
		"bb cli",
		"private",
		"Next: bb project view BBCLI --workspace acme",
	)

	buf.Reset()
	summaryPayload := projectPayload{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Project: bitbucket.Project{
			Key:         "BBCLI",
			Name:        "bb cli",
			IsPrivate:   false,
			UUID:        "{project-1}",
			Description: "project description",
			Links: bitbucket.ProjectLinks{
				HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/workspace/projects/BBCLI"},
			},
		},
	}
	if err := writeProjectSummary(&buf, summaryPayload); err != nil {
		t.Fatalf("writeProjectSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"Name: bb cli",
		"Host: bitbucket.org",
		"UUID: {project-1}",
		"Description: project description",
		"Next: bb project default-reviewer list BBCLI --workspace acme",
	)
}

func TestWriteProjectDefaultReviewerListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := projectDefaultReviewerListPayload{
		Workspace:  "acme",
		ProjectKey: "BBCLI",
		DefaultReviewers: []bitbucket.DefaultReviewer{
			{ReviewerType: "project", User: bitbucket.RepositoryPermissionUser{AccountID: "user-1", DisplayName: "Auro"}},
		},
	}
	if err := writeProjectDefaultReviewerListSummary(&buf, payload); err != nil {
		t.Fatalf("writeProjectDefaultReviewerListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"user-1",
		"Auro",
		"project",
		"Next: bb project permissions user list BBCLI --workspace acme",
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

func TestWriteProjectPermissionSummaries(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	userPayload := projectUserPermissionPayload{
		Workspace:  "acme",
		ProjectKey: "BBCLI",
		Permission: bitbucket.ProjectUserPermission{
			Permission: "admin",
			User:       bitbucket.RepositoryPermissionUser{AccountID: "user-1", DisplayName: "Auro"},
		},
	}
	if err := writeProjectUserPermissionSummary(&buf, userPayload); err != nil {
		t.Fatalf("writeProjectUserPermissionSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"Account ID: user-1",
		"User: Auro",
		"Permission: admin",
		"Next: bb project permissions user list BBCLI --workspace acme",
	)

	buf.Reset()
	groupListPayload := projectGroupPermissionListPayload{
		Workspace:  "acme",
		ProjectKey: "BBCLI",
		Permissions: []bitbucket.ProjectGroupPermission{
			{Permission: "write", Group: bitbucket.RepositoryPermissionGroup{Slug: "eng"}},
		},
	}
	if err := writeProjectGroupPermissionListSummary(&buf, groupListPayload); err != nil {
		t.Fatalf("writeProjectGroupPermissionListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"eng",
		"write",
		"Next: bb project permissions group view BBCLI eng --workspace acme",
	)

	buf.Reset()
	groupPayload := projectGroupPermissionPayload{
		Workspace:  "acme",
		ProjectKey: "BBCLI",
		Permission: bitbucket.ProjectGroupPermission{
			Permission: "write",
			Group:      bitbucket.RepositoryPermissionGroup{Slug: "eng"},
		},
	}
	if err := writeProjectGroupPermissionSummary(&buf, groupPayload); err != nil {
		t.Fatalf("writeProjectGroupPermissionSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Project: BBCLI",
		"Group: eng",
		"Permission: write",
		"Next: bb project permissions group list BBCLI --workspace acme",
	)
}

func TestWriteWorkspaceAndMembershipListSummaries(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	workspacePayload := workspacePayload{
		Host: "bitbucket.org",
		Workspace: bitbucket.Workspace{
			Slug:      "acme",
			Name:      "Acme",
			IsPrivate: true,
			UUID:      "{workspace-1}",
			CreatedOn: "2026-03-13T00:00:00Z",
			Links:     bitbucket.WorkspaceLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/workspace/overview"}},
		},
	}
	if err := writeWorkspaceSummary(&buf, workspacePayload); err != nil {
		t.Fatalf("writeWorkspaceSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Name: Acme",
		"Host: bitbucket.org",
		"UUID: {workspace-1}",
		"Created: 2026-03-13T00:00:00Z",
		"Next: bb workspace member list acme",
	)

	buf.Reset()
	memberListPayload := workspaceMembershipListPayload{
		Workspace: "acme",
		Query:     `display_name ~ "Auro"`,
		Members: []bitbucket.WorkspaceMembership{
			{Permission: "owner", User: bitbucket.RepositoryPermissionUser{AccountID: "user-1", DisplayName: "Auro", Nickname: "auro"}},
		},
	}
	if err := writeWorkspaceMembershipListSummary(&buf, memberListPayload, "members"); err != nil {
		t.Fatalf("writeWorkspaceMembershipListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Query: display_name ~ \"Auro\"",
		"user-1",
		"Auro",
		"owner",
		"Next: bb workspace member view user-1 --workspace acme",
	)

	buf.Reset()
	repoPermPayload := workspaceRepoPermissionListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Query:     "permission = \"write\"",
		Sort:      "repository.slug",
		Permissions: []bitbucket.WorkspaceRepositoryPermission{
			{
				Permission: "write",
				Repository: bitbucket.WorkspacePermissionRepository{FullName: "acme/widgets"},
				User:       bitbucket.RepositoryPermissionUser{AccountID: "user-1", DisplayName: "Auro"},
			},
		},
	}
	if err := writeWorkspaceRepoPermissionListSummary(&buf, repoPermPayload); err != nil {
		t.Fatalf("writeWorkspaceRepoPermissionListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Repository: widgets",
		"Query: permission = \"write\"",
		"Sort: repository.slug",
		"acme/widgets",
		"Next: bb repo permissions user list --repo acme/widgets",
	)
}
