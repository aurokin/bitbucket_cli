package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
)

func TestWriteBranchListSummary(t *testing.T) {
	t.Parallel()

	payload := branchListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Branches: []bitbucket.RepositoryBranch{
			{Name: "main", Target: bitbucket.RepositoryCommit{Hash: "abc1234"}, DefaultMergeStrategy: "squash"},
		},
	}

	var buf bytes.Buffer
	if err := writeBranchListSummary(&buf, payload); err != nil {
		t.Fatalf("writeBranchListSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Next: bb branch view main --repo acme/widgets",
	)
}

func TestWriteTagListSummary(t *testing.T) {
	t.Parallel()

	payload := tagListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Tags: []bitbucket.RepositoryTag{
			{Name: "v1.0.0", Target: bitbucket.RepositoryCommit{Hash: "abc1234"}, Date: "2026-03-13T00:00:00Z"},
		},
	}

	var buf bytes.Buffer
	if err := writeTagListSummary(&buf, payload); err != nil {
		t.Fatalf("writeTagListSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Next: bb tag view v1.0.0 --repo acme/widgets",
	)
}

func TestWriteBranchSummaryLifecycle(t *testing.T) {
	t.Parallel()

	payload := branchPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "created",
		Branch:    bitbucket.RepositoryBranch{Name: "feature/demo", Target: bitbucket.RepositoryCommit{Hash: "abc1234"}, DefaultMergeStrategy: "squash"},
	}

	var buf bytes.Buffer
	if err := writeBranchSummary(&buf, payload); err != nil {
		t.Fatalf("writeBranchSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Branch: feature/demo",
		"Target: abc1234",
		"Next: bb browse --repo acme/widgets --branch feature/demo --no-browser",
	)

	buf.Reset()
	if err := writeBranchCreateSummary(&buf, payload); err != nil {
		t.Fatalf("writeBranchCreateSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Branch: feature/demo",
		"Action: created",
		"Next: bb branch view feature/demo --repo acme/widgets",
	)

	buf.Reset()
	if err := writeBranchDeleteSummary(&buf, branchDeletePayload{Workspace: "acme", Repo: "widgets", Branch: "feature/demo", Deleted: true}); err != nil {
		t.Fatalf("writeBranchDeleteSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Branch: feature/demo",
		"Action: deleted",
		"Next: bb branch list --repo acme/widgets",
	)
}

func TestWriteTagSummaryLifecycle(t *testing.T) {
	t.Parallel()

	payload := tagPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "created",
		Tag:       bitbucket.RepositoryTag{Name: "v1.0.0", Target: bitbucket.RepositoryCommit{Hash: "abc1234"}, Date: "2026-03-13T00:00:00Z", Message: "release\nnotes"},
	}

	var buf bytes.Buffer
	if err := writeTagSummary(&buf, payload); err != nil {
		t.Fatalf("writeTagSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Tag: v1.0.0",
		"Target: abc1234",
		"Message: release",
		"Next: bb browse --repo acme/widgets --commit abc1234 --no-browser",
	)

	buf.Reset()
	if err := writeTagCreateSummary(&buf, payload); err != nil {
		t.Fatalf("writeTagCreateSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Tag: v1.0.0",
		"Action: created",
		"Next: bb tag view v1.0.0 --repo acme/widgets",
	)

	buf.Reset()
	if err := writeTagDeleteSummary(&buf, tagDeletePayload{Workspace: "acme", Repo: "widgets", Tag: "v1.0.0", Deleted: true}); err != nil {
		t.Fatalf("writeTagDeleteSummary returned error: %v", err)
	}
	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Tag: v1.0.0",
		"Action: deleted",
		"Next: bb tag list --repo acme/widgets",
	)
}

func TestResolveRefCreateTarget(t *testing.T) {
	rootDir := t.TempDir()

	if got, err := resolveRefCreateTarget(resolvedRepoTarget{Workspace: "acme", Repo: "widgets"}, "main"); err != nil || got != "main" {
		t.Fatalf("expected explicit target, got %q %v", got, err)
	}

	runGitCmd(t, rootDir, "init", "-b", "main")
	runGitCmd(t, rootDir, "config", "user.name", "Test User")
	runGitCmd(t, rootDir, "config", "user.email", "test@example.com")
	if err := os.WriteFile(rootDir+"/README.md", []byte("test\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	runGitCmd(t, rootDir, "add", "README.md")
	runGitCmd(t, rootDir, "commit", "-m", "initial")
	runGitCmd(t, rootDir, "switch", "-c", "feature/demo")

	got, err := resolveRefCreateTarget(resolvedRepoTarget{
		Workspace: "acme",
		Repo:      "widgets",
		LocalRepo: &gitrepo.RepoContext{RootDir: rootDir},
	}, "")
	if err != nil || got != "feature/demo" {
		t.Fatalf("expected local branch fallback, got %q %v", got, err)
	}

	if _, err := resolveRefCreateTarget(resolvedRepoTarget{Workspace: "acme", Repo: "widgets"}, ""); err == nil {
		t.Fatal("expected missing ref error")
	}
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
