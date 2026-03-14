package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestParseRepoVisibility(t *testing.T) {
	t.Parallel()

	private, err := parseRepoVisibility("private")
	if err != nil || private == nil || !*private {
		t.Fatalf("expected private visibility, got %v %v", private, err)
	}
	public, err := parseRepoVisibility("public")
	if err != nil || public == nil || *public {
		t.Fatalf("expected public visibility, got %v %v", public, err)
	}
	none, err := parseRepoVisibility("")
	if err != nil || none != nil {
		t.Fatalf("expected nil visibility, got %v %v", none, err)
	}
	if _, err := parseRepoVisibility("internal"); err == nil {
		t.Fatal("expected invalid visibility error")
	}
}

func TestResolveWorkspaceInput(t *testing.T) {
	t.Parallel()

	got, err := resolveWorkspaceInput("acme", "")
	if err != nil || got != "acme" {
		t.Fatalf("unexpected workspace resolution %q %v", got, err)
	}
	got, err = resolveWorkspaceInput("", "acme")
	if err != nil || got != "acme" {
		t.Fatalf("unexpected workspace resolution %q %v", got, err)
	}
	if _, err := resolveWorkspaceInput("acme", "other"); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestResolveWorkspaceForCreate(t *testing.T) {
	t.Parallel()

	got, err := resolveWorkspaceForCreate(context.Background(), stubWorkspaceResolver{
		workspaces: []bitbucket.Workspace{{Slug: "acme"}},
	}, "")
	if err != nil || got != "acme" {
		t.Fatalf("unexpected workspace resolution %q %v", got, err)
	}

	if _, err := resolveWorkspaceForCreate(context.Background(), stubWorkspaceResolver{
		workspaces: []bitbucket.Workspace{{Slug: "acme"}, {Slug: "other"}},
	}, ""); err == nil {
		t.Fatal("expected multiple workspace error")
	}

	if _, err := resolveWorkspaceForCreate(context.Background(), stubWorkspaceResolver{}, ""); err == nil {
		t.Fatal("expected no workspace error")
	}
}

func TestResolveWorkspaceRepoPermissionInput(t *testing.T) {
	t.Parallel()

	workspace, repo, err := resolveWorkspaceRepoPermissionInput("", "acme/widgets", "")
	if err != nil {
		t.Fatalf("resolveWorkspaceRepoPermissionInput returned error: %v", err)
	}
	if workspace != "acme" || repo != "widgets" {
		t.Fatalf("unexpected workspace/repo %q %q", workspace, repo)
	}

	workspace, repo, err = resolveWorkspaceRepoPermissionInput("acme", "widgets", "")
	if err != nil {
		t.Fatalf("resolveWorkspaceRepoPermissionInput returned error: %v", err)
	}
	if workspace != "acme" || repo != "widgets" {
		t.Fatalf("unexpected workspace/repo %q %q", workspace, repo)
	}

	if _, _, err := resolveWorkspaceRepoPermissionInput("acme", "other/widgets", ""); err == nil {
		t.Fatal("expected workspace mismatch error")
	}

	workspace, repo, err = resolveWorkspaceRepoPermissionInput("", "https://bitbucket.org/acme/widgets", "")
	if err != nil {
		t.Fatalf("resolveWorkspaceRepoPermissionInput returned error for repo URL: %v", err)
	}
	if workspace != "acme" || repo != "widgets" {
		t.Fatalf("unexpected repo URL workspace/repo %q %q", workspace, repo)
	}
}

func TestWriteRepoListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoListPayload{
		Workspace: "acme",
		Query:     `name ~ "widget"`,
		Repos: []bitbucket.Repository{
			{Slug: "widgets", Name: "Widgets", IsPrivate: true, UpdatedOn: "2026-03-13T00:00:00Z", Project: bitbucket.RepositoryProject{Key: "BBCLI"}},
		},
	}

	if err := writeRepoListSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Workspace: acme",
		"Query: name ~ \"widget\"",
		"widgets",
		"Widgets",
		"private",
		"Next: bb repo view --repo acme/widgets",
	)
}

func TestWriteRepoEditSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoEditPayload{
		Workspace:    "acme",
		PreviousRepo: "widgets-old",
		Action:       "updated",
		Repository: bitbucket.Repository{
			Slug:        "widgets",
			Name:        "Widgets",
			IsPrivate:   false,
			Description: "updated description",
			Links:       bitbucket.RepositoryLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets"}},
		},
	}

	if err := writeRepoEditSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoEditSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Action: updated",
		"Previous Repository: widgets-old",
		"Visibility: public",
		"URL: https://bitbucket.org/acme/widgets",
		"Next: bb repo view --repo acme/widgets",
	)
}

func TestWriteRepoForkSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoForkPayload{
		SourceWorkspace: "acme",
		SourceRepo:      "widgets",
		Action:          "forked",
		Repository: bitbucket.Repository{
			Slug:      "widgets-fork",
			Name:      "widgets-fork",
			FullName:  "acme/widgets-fork",
			IsPrivate: true,
			Links:     bitbucket.RepositoryLinks{HTML: bitbucket.Link{Href: "https://bitbucket.org/acme/widgets-fork"}},
		},
	}

	if err := writeRepoForkSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoForkSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets-fork",
		"Source: acme/widgets",
		"Action: forked",
		"Visibility: private",
		"Next: bb repo clone acme/widgets-fork",
	)
}
