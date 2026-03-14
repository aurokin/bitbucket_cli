package cmd

import (
	"strings"
	"testing"
)

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

func TestPRCloseHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "close", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr close help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRCheckoutHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "checkout", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"bb pr checkout https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr checkout help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRMergeHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "merge", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"bb pr merge https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr merge help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRCommentHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "comment", "--help")
	for _, fragment := range []string{
		"view, edit, delete, resolve, or reopen",
		"bb pr comment 1 --body 'Looks good'",
		"Available Commands:",
		"view        View a pull request comment",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr comment help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRReviewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "review", "--help")
	for _, fragment := range []string{
		"bb pr review approve 7 --repo workspace-slug/repo-slug",
		"Available Commands:",
		"request-changes       Request changes on a pull request",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr review help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRReviewApproveHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "review", "approve", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"--repo string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr review approve help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRActivityHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "activity", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"--limit int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr activity help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRCommitsHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "commits", "--help")
	for _, fragment := range []string{
		"pull request comment URL",
		"#comment-15",
		"--limit int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr commits help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRChecksHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "checks", "--help")
	for _, fragment := range []string{
		"Bitbucket Cloud equivalent of PR checks backed by commit statuses",
		"bb pr statuses https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15",
		"--limit int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr checks help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRCommentViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "comment", "view", "--help")
	for _, fragment := range []string{
		"numeric comment ID together with --pr <id-or-url>",
		"#comment-15",
		"--pr string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr comment view help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRCommentResolveHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "comment", "resolve", "--help")
	for _, fragment := range []string{
		"top-level diff comments",
		"#comment-15",
		"--pr string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr comment resolve help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRCommentDeleteHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "comment", "delete", "--help")
	for _, fragment := range []string{
		"Scripts and agents should use --yes together with --no-prompt",
		"--yes",
		"--pr string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr comment delete help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRTaskHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "task", "--help")
	for _, fragment := range []string{
		"bb pr task create 1 --repo workspace-slug/repo-slug --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Handle this thread'",
		"Available Commands:",
		"resolve     Resolve a pull request task",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr task help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRTaskViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "task", "view", "--help")
	for _, fragment := range []string{
		"numeric task ID together with --pr <id-or-url>",
		"--pr string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr task view help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRTaskCreateHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "task", "create", "--help")
	for _, fragment := range []string{
		"--comment string",
		"--body-file string",
		"attach the task to a specific pull request comment",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr task create help missing %q\n%s", fragment, output)
		}
	}
}

func TestPRTaskDeleteHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pr", "task", "delete", "--help")
	for _, fragment := range []string{
		"Scripts and agents should use --yes together with --no-prompt",
		"--yes",
		"--pr string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pr task delete help missing %q\n%s", fragment, output)
		}
	}
}
