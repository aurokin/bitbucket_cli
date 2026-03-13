package cmd

import (
	"bytes"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteRepoUserPermissionListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoUserPermissionListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Permissions: []bitbucket.RepositoryUserPermission{
			{Permission: "admin", User: bitbucket.RepositoryPermissionUser{DisplayName: "Auro", AccountID: "user-1"}},
		},
	}
	if err := writeRepoUserPermissionListSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoUserPermissionListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"user-1",
		"Auro",
		"admin",
		"Next: bb repo permissions user view user-1 --repo acme/widgets",
	)
}

func TestWriteRepoUserPermissionSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoUserPermissionPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Permission: bitbucket.RepositoryUserPermission{
			Permission: "admin",
			User:       bitbucket.RepositoryPermissionUser{DisplayName: "Auro", AccountID: "user-1"},
		},
	}
	if err := writeRepoUserPermissionSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoUserPermissionSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Account ID: user-1",
		"User: Auro",
		"Permission: admin",
		"Next: bb repo permissions user list --repo acme/widgets",
	)
}

func TestWriteRepoGroupPermissionListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoGroupPermissionListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Permissions: []bitbucket.RepositoryGroupPermission{
			{Permission: "read", Group: bitbucket.RepositoryPermissionGroup{Slug: "developers"}},
		},
	}
	if err := writeRepoGroupPermissionListSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoGroupPermissionListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"developers",
		"read",
		"Next: bb repo permissions group view developers --repo acme/widgets",
	)
}

func TestWriteRepoGroupPermissionSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoGroupPermissionPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Permission: bitbucket.RepositoryGroupPermission{
			Permission: "read",
			Group:      bitbucket.RepositoryPermissionGroup{Slug: "developers"},
		},
	}
	if err := writeRepoGroupPermissionSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoGroupPermissionSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Group: developers",
		"Permission: read",
		"Next: bb repo permissions group list --repo acme/widgets",
	)
}
