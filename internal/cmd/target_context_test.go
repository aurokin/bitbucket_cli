package cmd

import (
	"context"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
)

func configureTargetContextAuth(t *testing.T) {
	t.Helper()

	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	cfg := config.Config{}
	cfg.SetHost("bitbucket.org", config.HostConfig{
		AuthType: config.AuthTypeAPIToken,
		Username: "agent@example.com",
		Token:    "secret",
	}, true)
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}
}

func TestTargetContextResolversWithExplicitURLs(t *testing.T) {
	configureTargetContextAuth(t)

	repoResolved, err := resolveRepoCommandTargetInput(context.Background(), "", "", "", "acme/widgets", false)
	if err != nil {
		t.Fatalf("resolveRepoCommandTargetInput returned error: %v", err)
	}
	if repoResolved.Target.Workspace != "acme" || repoResolved.Target.Repo != "widgets" {
		t.Fatalf("unexpected repo target %+v", repoResolved.Target)
	}

	prResolved, err := resolvePullRequestCommandTarget(context.Background(), "", "", "", "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15", false)
	if err != nil {
		t.Fatalf("resolvePullRequestCommandTarget returned error: %v", err)
	}
	if prResolved.Target.RepoTarget.Workspace != "acme" || prResolved.Target.RepoTarget.Repo != "widgets" || prResolved.Target.ID != 7 {
		t.Fatalf("unexpected PR target %+v", prResolved.Target)
	}

	commentResolved, err := resolvePullRequestCommentCommandTarget(context.Background(), "", "", "", "", "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15", false)
	if err != nil {
		t.Fatalf("resolvePullRequestCommentCommandTarget returned error: %v", err)
	}
	if commentResolved.Target.PRTarget.ID != 7 || commentResolved.Target.CommentID != 15 {
		t.Fatalf("unexpected PR comment target %+v", commentResolved.Target)
	}

	commitResolved, err := resolveCommitCommandTarget(context.Background(), "", "", "", "https://bitbucket.org/acme/widgets/commits/abc1234", false)
	if err != nil {
		t.Fatalf("resolveCommitCommandTarget returned error: %v", err)
	}
	if commitResolved.Target.RepoTarget.Workspace != "acme" || commitResolved.Target.RepoTarget.Repo != "widgets" || commitResolved.Target.Commit != "abc1234" {
		t.Fatalf("unexpected commit target %+v", commitResolved.Target)
	}

	commitCommentResolved, err := resolveCommitCommentCommandTarget(context.Background(), "", "", "", "https://bitbucket.org/acme/widgets/commits/abc1234", "5", false)
	if err != nil {
		t.Fatalf("resolveCommitCommentCommandTarget returned error: %v", err)
	}
	if commitCommentResolved.Target.CommitTarget.Commit != "abc1234" || commitCommentResolved.Target.CommentID != 5 {
		t.Fatalf("unexpected commit comment target %+v", commitCommentResolved.Target)
	}

	taskResolved, err := resolvePullRequestTaskCommandTarget(context.Background(), "", "", "", "https://bitbucket.org/acme/widgets/pull-requests/7", "3", false)
	if err != nil {
		t.Fatalf("resolvePullRequestTaskCommandTarget returned error: %v", err)
	}
	if taskResolved.Target.PRTarget.ID != 7 || taskResolved.Target.TaskID != 3 {
		t.Fatalf("unexpected PR task target %+v", taskResolved.Target)
	}

	issueTarget, _, issueID, err := resolveIssueTargetAndID("", "", "", "https://bitbucket.org/acme/widgets/issues/12")
	if err != nil {
		t.Fatalf("resolveIssueTargetAndID returned error: %v", err)
	}
	if issueTarget.Workspace != "acme" || issueTarget.Repo != "widgets" || issueID != 12 {
		t.Fatalf("unexpected issue target %+v id=%d", issueTarget, issueID)
	}
}
