package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteRepoHookListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoHookListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Hooks: []bitbucket.RepositoryWebhook{
			{UUID: "hook-1", Active: true, Description: "fixture hook", URL: "https://example.com/hook"},
		},
	}

	if err := writeRepoHookListSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoHookListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"hook-1",
		"fixture hook",
		"https://example.com/hook",
		"Next: bb repo hook view hook-1 --repo acme/widgets",
	)
}

func TestWriteRepoHookSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoHookPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "created",
		Hook: bitbucket.RepositoryWebhook{
			UUID:        "hook-1",
			Description: "fixture hook",
			URL:         "https://example.com/hook",
			Active:      true,
			Events:      []string{"repo:push"},
		},
	}

	if err := writeRepoHookSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoHookSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Hook: hook-1",
		"Action: created",
		"State: active",
		"Events: repo:push",
		"Next: bb repo hook view hook-1 --repo acme/widgets",
	)
}

func TestWriteRepoDeployKeyListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoDeployKeyListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Keys: []bitbucket.RepositoryDeployKey{
			{ID: 7, Label: "fixture-key", LastUsed: "2026-03-13T00:00:00Z"},
		},
	}

	if err := writeRepoDeployKeyListSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoDeployKeyListSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"7",
		"fixture-key",
		"Next: bb repo deploy-key view 7 --repo acme/widgets",
	)
}

func TestWriteRepoDeployKeySummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoDeployKeyPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "created",
		Key: bitbucket.RepositoryDeployKey{
			ID:       7,
			Label:    "fixture-key",
			Comment:  "fixture",
			LastUsed: "2026-03-13T00:00:00Z",
		},
	}

	if err := writeRepoDeployKeySummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoDeployKeySummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Deploy Key: 7",
		"Action: created",
		"State: present",
		"Label: fixture-key",
		"Next: bb repo deploy-key view 7 --repo acme/widgets",
	)
}

func TestResolveDeployKeyMaterial(t *testing.T) {
	t.Parallel()

	if _, err := resolveDeployKeyMaterial("", ""); err == nil {
		t.Fatal("expected error for missing key material")
	}
	got, err := resolveDeployKeyMaterial("ssh-ed25519 AAAA fixture", "")
	if err != nil || got != "ssh-ed25519 AAAA fixture" {
		t.Fatalf("unexpected inline key resolution %q %v", got, err)
	}

	tempDir := t.TempDir()
	filePath := tempDir + "/id_ed25519.pub"
	if err := os.WriteFile(filePath, []byte("ssh-ed25519 AAAA file\n"), 0o644); err != nil {
		t.Fatalf("write temp key file: %v", err)
	}
	got, err = resolveDeployKeyMaterial("", filePath)
	if err != nil || got != "ssh-ed25519 AAAA file" {
		t.Fatalf("unexpected file key resolution %q %v", got, err)
	}
}
