package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		host      string
		workspace string
		repo      string
	}{
		{
			name:      "https",
			raw:       "https://bitbucket.org/acme/widgets.git",
			host:      "bitbucket.org",
			workspace: "acme",
			repo:      "widgets",
		},
		{
			name:      "ssh scp",
			raw:       "git@bitbucket.org:acme/widgets.git",
			host:      "bitbucket.org",
			workspace: "acme",
			repo:      "widgets",
		},
		{
			name:      "ssh url",
			raw:       "ssh://git@bitbucket.org/acme/widgets.git",
			host:      "bitbucket.org",
			workspace: "acme",
			repo:      "widgets",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseRemoteURL(test.raw)
			if err != nil {
				t.Fatalf("ParseRemoteURL returned error: %v", err)
			}
			if parsed.Host != test.host || parsed.Workspace != test.workspace || parsed.RepoSlug != test.repo {
				t.Fatalf("unexpected parse result: %+v", parsed)
			}
		})
	}
}

func TestResolveRepoContext(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "remote", "add", "origin", "git@bitbucket.org:acme/widgets.git")

	filePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(filePath, []byte("test\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")

	repo, err := ResolveRepoContext(context.Background(), dir)
	if err != nil {
		t.Fatalf("ResolveRepoContext returned error: %v", err)
	}

	if repo.Workspace != "acme" || repo.RepoSlug != "widgets" {
		t.Fatalf("unexpected repo context: %+v", repo)
	}
	if repo.RemoteName != "origin" {
		t.Fatalf("expected origin remote, got %q", repo.RemoteName)
	}
}

func TestCurrentBranch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("test\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")
	runGit(t, dir, "switch", "-c", "feature/test")

	branch, err := CurrentBranch(context.Background(), dir)
	if err != nil {
		t.Fatalf("CurrentBranch returned error: %v", err)
	}
	if branch != "feature/test" {
		t.Fatalf("expected feature/test, got %q", branch)
	}
}

func TestCheckoutRemoteBranchRequiresInputs(t *testing.T) {
	t.Parallel()

	if err := CheckoutRemoteBranch(context.Background(), t.TempDir(), "", "feature"); err == nil {
		t.Fatalf("expected remote name validation error")
	}
	if err := CheckoutRemoteBranch(context.Background(), t.TempDir(), "origin", ""); err == nil {
		t.Fatalf("expected branch validation error")
	}
}

func TestAuthenticatedHTTPSURL(t *testing.T) {
	t.Parallel()

	got, err := authenticatedHTTPSURL("https://bitbucket.org/acme/widgets.git", "x-bitbucket-api-token-auth", "secret-token")
	if err != nil {
		t.Fatalf("authenticatedHTTPSURL returned error: %v", err)
	}

	want := "https://x-bitbucket-api-token-auth:secret-token@bitbucket.org/acme/widgets.git"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSanitizedHTTPSURL(t *testing.T) {
	t.Parallel()

	got, err := sanitizedHTTPSURL("https://bitbucket.org/acme/widgets.git", "x-bitbucket-api-token-auth")
	if err != nil {
		t.Fatalf("sanitizedHTTPSURL returned error: %v", err)
	}

	want := "https://x-bitbucket-api-token-auth@bitbucket.org/acme/widgets.git"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
