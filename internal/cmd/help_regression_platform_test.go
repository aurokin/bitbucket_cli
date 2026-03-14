package cmd

import (
	"strings"
	"testing"
)

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

func TestBranchDeleteHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "branch", "delete", "--help")
	for _, fragment := range []string{
		"Scripts and agents should use --yes together with --no-prompt",
		"--yes",
		"--repo string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("branch delete help missing %q\n%s", fragment, output)
		}
	}
}

func TestTagCreateHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "tag", "create", "--help")
	for _, fragment := range []string{
		"defaults the target to the current branch",
		"--target string",
		"--message string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("tag create help missing %q\n%s", fragment, output)
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
