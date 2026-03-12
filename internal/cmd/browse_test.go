package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
)

type browseRepositoryClient struct {
	repository bitbucket.Repository
	err        error
}

func (c browseRepositoryClient) GetRepository(context.Context, string, string) (bitbucket.Repository, error) {
	return c.repository, c.err
}

func TestBuildBrowsePayloadRepository(t *testing.T) {
	t.Parallel()

	payload, err := buildBrowsePayload(context.Background(), &bitbucket.Client{}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
	}, "", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}
	if payload.Type != "repository" || payload.URL != "https://bitbucket.org/acme/widgets" {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestBuildBrowsePayloadPullRequest(t *testing.T) {
	t.Parallel()

	payload, err := buildBrowsePayload(context.Background(), &bitbucket.Client{}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
	}, "", browseOptions{PR: 7})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}
	if payload.Type != "pull-request" || payload.PR != 7 || payload.URL != "https://bitbucket.org/acme/widgets/pull-requests/7" {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestBuildBrowsePayloadPathUsesLocalBranchAndLine(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootDir, "internal", "cmd"), 0o755); err != nil {
		t.Fatalf("create repo subdir: %v", err)
	}

	previousCWD := getWorkingDirectory
	previousBranch := currentBrowseBranch
	t.Cleanup(func() {
		getWorkingDirectory = previousCWD
		currentBrowseBranch = previousBranch
	})

	getWorkingDirectory = func() (string, error) {
		return filepath.Join(rootDir, "internal", "cmd"), nil
	}
	currentBrowseBranch = func(context.Context, string) (string, error) {
		return "feature/refactor", nil
	}

	payload, err := buildBrowsePayload(context.Background(), &bitbucket.Client{}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
		LocalRepo: &gitrepo.RepoContext{RootDir: rootDir},
	}, "browse.go:42", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}

	expected := "https://bitbucket.org/acme/widgets/src/feature%2Frefactor/internal/cmd/browse.go#lines-42"
	if payload.Type != "path" || payload.Ref != "feature/refactor" || payload.Path != "internal/cmd/browse.go" || payload.Line != 42 || payload.URL != expected {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestBuildBrowsePayloadPathFallsBackToRepositoryMainBranch(t *testing.T) {
	t.Parallel()

	previousBranch := currentBrowseBranch
	t.Cleanup(func() { currentBrowseBranch = previousBranch })
	currentBrowseBranch = func(context.Context, string) (string, error) {
		return "", fmt.Errorf("no branch")
	}

	client := browseRepositoryClient{
		repository: bitbucket.Repository{
			MainBranch: bitbucket.RepositoryBranch{Name: "main"},
		},
	}

	payload, err := buildBrowsePayload(context.Background(), client, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
	}, "README.md", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}

	expected := "https://bitbucket.org/acme/widgets/src/main/README.md"
	if payload.Ref != "main" || payload.URL != expected {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestBuildBrowsePayloadCommit(t *testing.T) {
	t.Parallel()

	payload, err := buildBrowsePayload(context.Background(), &bitbucket.Client{}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
	}, "deadbeef", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}
	if payload.Type != "commit" || payload.Commit != "deadbeef" || payload.URL != "https://bitbucket.org/acme/widgets/commits/deadbeef" {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestValidateBrowseOptionsRejectsConflicts(t *testing.T) {
	t.Parallel()

	err := validateBrowseOptions("README.md", browseOptions{PR: 1})
	if err == nil || !strings.Contains(err.Error(), "positional browse target") {
		t.Fatalf("expected positional conflict error, got %v", err)
	}

	err = validateBrowseOptions("", browseOptions{PR: 1, Issue: 2})
	if err == nil || !strings.Contains(err.Error(), "choose only one") {
		t.Fatalf("expected mutually exclusive error, got %v", err)
	}
}

func TestConfiguredBrowserCommandPrefersEnvThenConfig(t *testing.T) {
	previousLoad := loadBrowseConfig
	t.Cleanup(func() { loadBrowseConfig = previousLoad })
	loadBrowseConfig = func() (config.Config, error) {
		return config.Config{Settings: config.Settings{Browser: "firefox"}}, nil
	}

	t.Setenv("BROWSER", "chromium --incognito")
	command, err := configuredBrowserCommand()
	if err != nil {
		t.Fatalf("configuredBrowserCommand returned error: %v", err)
	}
	if command != "chromium --incognito" {
		t.Fatalf("expected env browser command, got %q", command)
	}

	t.Setenv("BROWSER", "")
	command, err = configuredBrowserCommand()
	if err != nil {
		t.Fatalf("configuredBrowserCommand returned error: %v", err)
	}
	if command != "firefox" {
		t.Fatalf("expected config browser command, got %q", command)
	}
}
