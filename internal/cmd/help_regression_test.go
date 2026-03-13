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

func TestAuthStatusHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "auth", "status", "--help")
	for _, fragment := range []string{
		"bb auth status --check --json",
		"--check",
		"--host string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("auth status help missing %q\n%s", fragment, output)
		}
	}
}

func TestConfigSetHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "config", "set", "--help")
	for _, fragment := range []string{
		"bb config set browser 'firefox --new-window'",
		"bb config set output.format json",
		"bb config get output.format",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("config set help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueAttachmentHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "attachment", "--help")
	for _, fragment := range []string{
		"bb issue attachment list 1 --repo workspace-slug/issues-repo-slug",
		"bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug",
		"Available Commands:",
		"upload      Upload attachments to an issue",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue attachment help missing %q\n%s", fragment, output)
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

func TestCommitViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "commit", "view", "--help")
	for _, fragment := range []string{
		"commit SHA or a commit URL",
		"/commits/abc1234",
		"--repo string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("commit view help missing %q\n%s", fragment, output)
		}
	}
}

func TestCommitCommentViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "commit", "comment", "view", "--help")
	for _, fragment := range []string{
		"Pass the commit with --commit as a SHA or commit URL",
		"--commit string",
		"/commits/abc1234",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("commit comment view help missing %q\n%s", fragment, output)
		}
	}
}

func TestCommitReportViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "commit", "report", "view", "--help")
	for _, fragment := range []string{
		"code-insight report",
		"--commit string",
		"/commits/abc1234",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("commit report view help missing %q\n%s", fragment, output)
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

func TestPipelineRunHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "run", "--help")
	for _, fragment := range []string{
		"bb pipeline run --repo workspace-slug/pipelines-repo-slug --ref main --json '*'",
		"--ref-type string",
		"current local branch",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline run help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineTestReportsHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "test-reports", "--help")
	for _, fragment := range []string{
		"--cases",
		"--step string",
		"bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug --cases --limit 50 --json '*'",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline test-reports help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineVariableHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "variable", "--help")
	for _, fragment := range []string{
		"bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured",
		"Available Commands:",
		"delete      Delete a repository pipeline variable",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline variable help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineVariableDeleteHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "variable", "delete", "--help")
	for _, fragment := range []string{
		"Scripts and agents should use --yes together with --no-prompt",
		"--yes",
		"key or UUID",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline variable delete help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineScheduleHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "schedule", "--help")
	for _, fragment := range []string{
		"bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'",
		"Available Commands:",
		"disable     Disable a pipeline schedule",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline schedule help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineScheduleCreateHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "schedule", "create", "--help")
	for _, fragment := range []string{
		"--cron string",
		"--selector-type string",
		"Seven-field cron pattern in UTC",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline schedule create help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineScheduleDeleteHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "schedule", "delete", "--help")
	for _, fragment := range []string{
		"Scripts and agents should use --yes together with --no-prompt",
		"--yes",
		"{schedule-uuid}",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline schedule delete help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineRunnerHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "runner", "--help")
	for _, fragment := range []string{
		"List, inspect, and delete Bitbucket repository pipeline runners",
		"Available Commands:",
		"view        View one pipeline runner",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline runner help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineCacheHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "cache", "--help")
	for _, fragment := range []string{
		"List Bitbucket repository pipeline caches, delete one cache by UUID, or clear caches by name",
		"Available Commands:",
		"clear       Clear pipeline caches by name",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline cache help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueCommentHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "comment", "--help")
	for _, fragment := range []string{
		"bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body 'Needs follow-up'",
		"Available Commands:",
		"delete      Delete an issue comment",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue comment help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueCommentViewHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "comment", "view", "--help")
	for _, fragment := range []string{
		"--issue string",
		"Bitbucket issue URL",
		"bb issue comment view 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue comment view help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueCommentDeleteHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "comment", "delete", "--help")
	for _, fragment := range []string{
		"Scripts and agents should use --yes together with --no-prompt",
		"--yes",
		"--issue string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue comment delete help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueMilestoneHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "milestone", "--help")
	for _, fragment := range []string{
		"List and view Bitbucket issue tracker milestones",
		"Available Commands:",
		"view        View one issue milestone",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue milestone help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueComponentHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "component", "--help")
	for _, fragment := range []string{
		"List and view Bitbucket issue tracker components",
		"Available Commands:",
		"view        View one issue component",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue component help missing %q\n%s", fragment, output)
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
