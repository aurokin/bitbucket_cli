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

	rootDir := t.TempDir()

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
		LocalRepo: &gitrepo.RepoContext{RootDir: rootDir},
	}, "README.md", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}

	expected := "https://bitbucket.org/acme/widgets/src/main/README.md"
	if payload.Ref != "main" || payload.URL != expected {
		t.Fatalf("unexpected payload %+v", payload)
	}
	if len(payload.Warnings) == 0 || !strings.Contains(payload.Warnings[0], "falling back to the repository main branch") {
		t.Fatalf("expected fallback warning, got %+v", payload)
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

func TestBuildBrowsePayloadTreatsExistingCommitLikeFilenameAsPath(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootDir, "deadbeef"), []byte("fixture"), 0o644); err != nil {
		t.Fatalf("write commit-like file: %v", err)
	}

	previousCWD := getWorkingDirectory
	previousBranch := currentBrowseBranch
	t.Cleanup(func() {
		getWorkingDirectory = previousCWD
		currentBrowseBranch = previousBranch
	})

	getWorkingDirectory = func() (string, error) {
		return rootDir, nil
	}
	currentBrowseBranch = func(context.Context, string) (string, error) {
		return "main", nil
	}

	payload, err := buildBrowsePayload(context.Background(), &bitbucket.Client{}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
		LocalRepo: &gitrepo.RepoContext{RootDir: rootDir},
	}, "deadbeef", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}

	expected := "https://bitbucket.org/acme/widgets/src/main/deadbeef"
	if payload.Type != "path" || payload.Path != "deadbeef" || payload.URL != expected {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestBuildBrowsePayloadBranchOnlyBuildsRootPathURL(t *testing.T) {
	t.Parallel()

	payload, err := buildBrowsePayload(context.Background(), browseRepositoryClient{}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
	}, "", browseOptions{Branch: "release/1.0"})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}

	expected := "https://bitbucket.org/acme/widgets/src/release%2F1.0/"
	if payload.Type != "path" || payload.Ref != "release/1.0" || payload.Path != "" || payload.URL != expected {
		t.Fatalf("unexpected payload %+v", payload)
	}
}

func TestBuildBrowsePayloadWarnsWhenTreatingPathAsRepoRelativeWithoutLocalContext(t *testing.T) {
	t.Parallel()

	payload, err := buildBrowsePayload(context.Background(), browseRepositoryClient{
		repository: bitbucket.Repository{
			MainBranch: bitbucket.RepositoryBranch{Name: "main"},
		},
	}, resolvedRepoTarget{
		Host:      "bitbucket.org",
		Workspace: "acme",
		Repo:      "widgets",
	}, "README.md", browseOptions{})
	if err != nil {
		t.Fatalf("buildBrowsePayload returned error: %v", err)
	}

	if len(payload.Warnings) == 0 || !strings.Contains(payload.Warnings[0], "repository-relative") {
		t.Fatalf("expected repository-relative warning, got %+v", payload)
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

func TestDefaultBrowserCommandByPlatform(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		goos string
		name string
		args []string
	}{
		{goos: "darwin", name: "open", args: []string{"https://bitbucket.org/acme/widgets"}},
		{goos: "windows", name: "rundll32", args: []string{"url.dll,FileProtocolHandler", "https://bitbucket.org/acme/widgets"}},
		{goos: "linux", name: "xdg-open", args: []string{"https://bitbucket.org/acme/widgets"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.goos, func(t *testing.T) {
			t.Parallel()

			name, args := defaultBrowserCommand(tc.goos, "https://bitbucket.org/acme/widgets")
			if name != tc.name {
				t.Fatalf("expected %q, got %q", tc.name, name)
			}
			if strings.Join(args, "\n") != strings.Join(tc.args, "\n") {
				t.Fatalf("expected %v, got %v", tc.args, args)
			}
		})
	}
}

func TestResolveBrowsePathRejectsAbsolutePathOutsideRepo(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	outsideDir := t.TempDir()

	_, err := resolveBrowsePath(resolvedRepoTarget{
		LocalRepo: &gitrepo.RepoContext{RootDir: rootDir},
	}, filepath.Join(outsideDir, "README.md"))
	if err == nil || !strings.Contains(err.Error(), "outside the repository root") {
		t.Fatalf("expected outside-repo error, got %v", err)
	}
}
