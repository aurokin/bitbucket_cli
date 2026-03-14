//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestBitbucketCloudWorkspaceRead(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	repoTarget := session.Workspace + "/" + fixture.PrimaryRepo.Slug

	listOutput := session.Run(t, "", "workspace", "list", "--json", "*")
	var listed struct {
		Workspaces []bitbucket.Workspace `json:"workspaces"`
	}
	if err := json.Unmarshal(listOutput, &listed); err != nil {
		t.Fatalf("parse workspace list JSON: %v\n%s", err, listOutput)
	}
	if len(listed.Workspaces) == 0 {
		t.Fatalf("expected at least one workspace")
	}

	viewOutput := session.Run(t, "", "workspace", "view", session.Workspace, "--json", "*")
	var viewed struct {
		Workspace bitbucket.Workspace `json:"workspace"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse workspace view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Workspace.Slug != session.Workspace {
		t.Fatalf("unexpected workspace view payload %+v", viewed)
	}

	memberListOutput := session.Run(t, "", "workspace", "member", "list", session.Workspace, "--json", "*")
	var members struct {
		Members []bitbucket.WorkspaceMembership `json:"members"`
	}
	if err := json.Unmarshal(memberListOutput, &members); err != nil {
		t.Fatalf("parse workspace member list JSON: %v\n%s", err, memberListOutput)
	}
	if len(members.Members) == 0 {
		t.Fatalf("expected at least one workspace member")
	}

	currentUser, err := session.Client.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("resolve current user: %v", err)
	}

	memberViewOutput := session.Run(t, "", "workspace", "member", "view", currentUser.AccountID, "--workspace", session.Workspace, "--json", "*")
	var memberView struct {
		Membership bitbucket.WorkspaceMembership `json:"membership"`
	}
	if err := json.Unmarshal(memberViewOutput, &memberView); err != nil {
		t.Fatalf("parse workspace member view JSON: %v\n%s", err, memberViewOutput)
	}
	if memberView.Membership.User.AccountID != currentUser.AccountID {
		t.Fatalf("unexpected workspace member view payload %+v", memberView)
	}

	permissionListOutput := session.Run(t, "", "workspace", "permission", "list", session.Workspace, "--json", "*")
	var permissions struct {
		Members []bitbucket.WorkspaceMembership `json:"members"`
	}
	if err := json.Unmarshal(permissionListOutput, &permissions); err != nil {
		t.Fatalf("parse workspace permission list JSON: %v\n%s", err, permissionListOutput)
	}
	if len(permissions.Members) == 0 {
		t.Fatalf("expected at least one workspace permission")
	}

	repoPermissionOutput, err := session.RunAllowFailure(t, "", "workspace", "repo-permission", "list", session.Workspace, "--repo", repoTarget, "--json", "*")
	if err != nil {
		if bytes.Contains(repoPermissionOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("workspace repository permission inspection requires broader repository admin scopes:\n%s", repoPermissionOutput)
		}
		t.Fatalf("bb workspace repo-permission list failed: %v\n%s", err, repoPermissionOutput)
	}

	var repoPermissions struct {
		Permissions []bitbucket.WorkspaceRepositoryPermission `json:"permissions"`
	}
	if err := json.Unmarshal(repoPermissionOutput, &repoPermissions); err != nil {
		t.Fatalf("parse workspace repo permission list JSON: %v\n%s", err, repoPermissionOutput)
	}

	humanOutput := session.Run(t, "", "workspace", "view", session.Workspace)
	assertContainsOrdered(t, string(humanOutput),
		"Workspace: "+session.Workspace,
		"Next: bb workspace member list "+session.Workspace,
	)
}

func TestBitbucketCloudProjectRead(t *testing.T) {
	session := newIntegrationSession(t)
	session.Fixture(t)

	listOutput := session.Run(t, "", "project", "list", session.Workspace, "--json", "*")
	var listed struct {
		Projects []bitbucket.Project `json:"projects"`
	}
	if err := json.Unmarshal(listOutput, &listed); err != nil {
		t.Fatalf("parse project list JSON: %v\n%s", err, listOutput)
	}
	if len(listed.Projects) == 0 {
		t.Fatalf("expected at least one project")
	}

	viewOutput := session.Run(t, "", "project", "view", fixtureProjectKey, "--workspace", session.Workspace, "--json", "*")
	var viewed struct {
		Project bitbucket.Project `json:"project"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse project view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Project.Key != fixtureProjectKey {
		t.Fatalf("unexpected project view payload %+v", viewed)
	}

	reviewerOutput, err := session.RunAllowFailure(t, "", "project", "default-reviewer", "list", fixtureProjectKey, "--workspace", session.Workspace, "--json", "*")
	if err != nil {
		t.Fatalf("bb project default-reviewer list failed: %v\n%s", err, reviewerOutput)
	}

	var reviewers struct {
		DefaultReviewers []bitbucket.DefaultReviewer `json:"default_reviewers"`
	}
	if err := json.Unmarshal(reviewerOutput, &reviewers); err != nil {
		t.Fatalf("parse project default reviewer list JSON: %v\n%s", err, reviewerOutput)
	}

	userPermissionOutput, err := session.RunAllowFailure(t, "", "project", "permissions", "user", "list", fixtureProjectKey, "--workspace", session.Workspace, "--json", "*")
	if err != nil {
		if bytes.Contains(userPermissionOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("project permission inspection requires broader project admin scopes:\n%s", userPermissionOutput)
		}
		t.Fatalf("bb project permissions user list failed: %v\n%s", err, userPermissionOutput)
	}

	var userPermissions struct {
		Permissions []bitbucket.ProjectUserPermission `json:"permissions"`
	}
	if err := json.Unmarshal(userPermissionOutput, &userPermissions); err != nil {
		t.Fatalf("parse project user permission list JSON: %v\n%s", err, userPermissionOutput)
	}

	currentUser, err := session.Client.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("resolve current user: %v", err)
	}
	var hasUserPermission bool
	for _, permission := range userPermissions.Permissions {
		if permission.User.AccountID == currentUser.AccountID {
			hasUserPermission = true
			break
		}
	}
	if hasUserPermission {
		userViewOutput := session.Run(t, "", "project", "permissions", "user", "view", fixtureProjectKey, currentUser.AccountID, "--workspace", session.Workspace, "--json", "*")
		var userView struct {
			Permission bitbucket.ProjectUserPermission `json:"permission"`
		}
		if err := json.Unmarshal(userViewOutput, &userView); err != nil {
			t.Fatalf("parse project user permission view JSON: %v\n%s", err, userViewOutput)
		}
		if userView.Permission.User.AccountID != currentUser.AccountID {
			t.Fatalf("unexpected project user permission view payload %+v", userView)
		}
	}

	groupPermissionOutput, err := session.RunAllowFailure(t, "", "project", "permissions", "group", "list", fixtureProjectKey, "--workspace", session.Workspace, "--json", "*")
	if err != nil {
		if bytes.Contains(groupPermissionOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("project group permission inspection requires broader project admin scopes:\n%s", groupPermissionOutput)
		}
		t.Fatalf("bb project permissions group list failed: %v\n%s", err, groupPermissionOutput)
	}

	var groupPermissions struct {
		Permissions []bitbucket.ProjectGroupPermission `json:"permissions"`
	}
	if err := json.Unmarshal(groupPermissionOutput, &groupPermissions); err != nil {
		t.Fatalf("parse project group permission list JSON: %v\n%s", err, groupPermissionOutput)
	}
	if len(groupPermissions.Permissions) > 0 {
		groupViewOutput := session.Run(t, "", "project", "permissions", "group", "view", fixtureProjectKey, groupPermissions.Permissions[0].Group.Slug, "--workspace", session.Workspace, "--json", "*")
		var groupView struct {
			Permission bitbucket.ProjectGroupPermission `json:"permission"`
		}
		if err := json.Unmarshal(groupViewOutput, &groupView); err != nil {
			t.Fatalf("parse project group permission view JSON: %v\n%s", err, groupViewOutput)
		}
		if groupView.Permission.Group.Slug != groupPermissions.Permissions[0].Group.Slug {
			t.Fatalf("unexpected project group permission view payload %+v", groupView)
		}
	}

	humanOutput := session.Run(t, "", "project", "view", fixtureProjectKey, "--workspace", session.Workspace)
	assertContainsOrdered(t, string(humanOutput),
		"Workspace: "+session.Workspace,
		"Project: "+fixtureProjectKey,
		"Next: bb project default-reviewer list "+fixtureProjectKey+" --workspace "+session.Workspace,
	)
}

func TestBitbucketCloudProjectMutation(t *testing.T) {
	session := newIntegrationSession(t)

	deleteProjectIfExists(t, session.Client, session.Workspace, fixtureTempProjectKey)

	createOutput, err := session.RunAllowFailure(t, "", "project", "create", fixtureTempProjectKey, "--workspace", session.Workspace, "--name", fixtureTempProjectName, "--visibility", "private", "--json", "*")
	if err != nil {
		if bytes.Contains(createOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("project creation requires broader project admin scopes:\n%s", createOutput)
		}
		t.Fatalf("bb project create failed: %v\n%s", err, createOutput)
	}

	var created struct {
		Action  string            `json:"action"`
		Project bitbucket.Project `json:"project"`
	}
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse project create JSON: %v\n%s", err, createOutput)
	}
	if created.Action != "created" || created.Project.Key != fixtureTempProjectKey {
		t.Fatalf("unexpected project create payload %+v", created)
	}

	editOutput := session.Run(t, "", "project", "edit", fixtureTempProjectKey, "--workspace", session.Workspace, "--description", "updated by integration", "--visibility", "public", "--json", "*")
	var edited struct {
		Action  string            `json:"action"`
		Project bitbucket.Project `json:"project"`
	}
	if err := json.Unmarshal(editOutput, &edited); err != nil {
		t.Fatalf("parse project edit JSON: %v\n%s", err, editOutput)
	}
	if edited.Action != "updated" || edited.Project.Key != fixtureTempProjectKey || edited.Project.IsPrivate {
		t.Fatalf("unexpected project edit payload %+v", edited)
	}

	deleteHuman, err := session.RunAllowFailure(t, "", "project", "delete", fixtureTempProjectKey, "--workspace", session.Workspace, "--yes")
	if err != nil {
		t.Fatalf("bb project delete failed: %v\n%s", err, deleteHuman)
	}
	assertContainsOrdered(t, string(deleteHuman),
		"Workspace: "+session.Workspace,
		"Project: "+fixtureTempProjectKey,
		"Action: deleted",
		"Status: deleted",
	)
}
