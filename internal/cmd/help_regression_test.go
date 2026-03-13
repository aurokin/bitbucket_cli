package cmd

import (
	"strings"
	"testing"
)

func TestAuthLoginHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "auth", "login", "--help")
	for _, fragment := range []string{
		"BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login",
		"--username string",
		"--with-token",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("auth login help missing %q\n%s", fragment, output)
		}
	}
}

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

func TestPRCreateHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "create", "--help")
	for _, fragment := range []string{
		"bb pr create --reuse-existing --json",
		"--source string",
		"--destination string",
		"--no-prompt",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr create help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "view", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"bb pr view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr view help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRDiffHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "diff", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --stat",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr diff help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineListHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "list", "--help")
	for _, fragment := range []string{
		"bb pipeline list --repo workspace-slug/repo-slug --state COMPLETED --json build_number,state,target",
		"--state string",
		"--limit int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline list help missing %q\n%s", fragment, output)
		}
	}
}

func TestBrowseHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "browse", "--help")
	for _, fragment := range []string{
		"bb browse README.md:12 --repo workspace-slug/repo-slug --no-browser",
		"--pr int",
		"--no-browser",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("browse help missing %q\n%s", fragment, output)
		}
	}
}

func TestStatusHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "status", "--help")
	for _, fragment := range []string{
		"bb status --workspace workspace-slug --limit 10",
		"--json string",
		"--repo-limit int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("status help missing %q\n%s", fragment, output)
		}
	}
}
