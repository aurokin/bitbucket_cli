package cmd

import (
	"strings"
	"testing"
)

func TestRepoViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "view", "--help")
	for _, fragment := range []string{
		"bb repo view --repo workspace-slug/repo-slug",
		"Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL",
		"--json string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo view help missing %q\n%s", fragment, output)
		}
	}
}

func TestRepoListHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "list", "--help")
	for _, fragment := range []string{
		"bb repo list workspace-slug",
		"bb repo list --workspace workspace-slug --limit 50",
		"--query string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo list help missing %q\n%s", fragment, output)
		}
	}
}

func TestRepoEditHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "edit", "--help")
	for _, fragment := range []string{
		"bb repo edit workspace-slug/repo-slug --description 'Updated description'",
		"bb repo edit --repo workspace-slug/repo-slug --visibility public --json '*'",
		"--visibility string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo edit help missing %q\n%s", fragment, output)
		}
	}
}

func TestRepoForkHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "fork", "--help")
	for _, fragment := range []string{
		"bb repo fork workspace-slug/repo-slug --to-workspace workspace-slug --name repo-slug-fork",
		"--to-workspace string",
		"--reuse-existing",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo fork help missing %q\n%s", fragment, output)
		}
	}
}

func TestRepoHookHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "hook", "--help")
	for _, fragment := range []string{
		"bb repo hook list --repo workspace-slug/repo-slug",
		"bb repo hook create --repo workspace-slug/repo-slug --url https://example.com/hook --event repo:push",
		"Available Commands:",
		"create      Create a repository webhook",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo hook help missing %q\n%s", fragment, output)
		}
	}
}

func TestRepoDeployKeyHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "deploy-key", "--help")
	for _, fragment := range []string{
		"bb repo deploy-key list --repo workspace-slug/repo-slug",
		"bb repo deploy-key create --repo workspace-slug/repo-slug --label ci --key-file ./id_ed25519.pub",
		"Available Commands:",
		"create      Create a repository deploy key",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo deploy-key help missing %q\n%s", fragment, output)
		}
	}
}

func TestRepoPermissionsHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "repo", "permissions", "--help")
	for _, fragment := range []string{
		"Inspect explicit Bitbucket repository user and group permissions",
		"Available Commands:",
		"user        Inspect explicit repository user permissions",
		"group       Inspect explicit repository group permissions",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("repo permissions help missing %q\n%s", fragment, output)
		}
	}
}
