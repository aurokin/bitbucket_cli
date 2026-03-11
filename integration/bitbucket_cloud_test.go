//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
)

const (
	fixtureProjectKey        = "BBCLI"
	fixtureProjectName       = "bb-cli integration"
	fixturePrimaryRepoSlug   = "bb-cli-integration-primary"
	fixtureSecondaryRepoSlug = "bb-cli-integration-secondary"
	fixtureIssuesRepoSlug    = "bb-cli-integration-issues"
	fixtureCreateRepoSlug    = "bb-cli-created-via-command"
	fixtureDeleteRepoSlug    = "bb-cli-delete-command-target"
	fixtureFeatureBranch     = "bb-cli-int-feature"
	fixturePRTitle           = "bb cli integration fixture pull request"
	fixtureCreatePRBranch    = "bb-cli-create-command-branch"
	fixtureCreatePRTitle     = "bb cli create command pull request"
	fixtureClosePRBranch     = "bb-cli-close-command-branch"
	fixtureClosePRTitle      = "bb cli close command pull request"
	fixtureMergePRBranch     = "bb-cli-merge-command-branch"
	fixtureMergePRTitle      = "bb cli merge command pull request"
)

type integrationFixture struct {
	Workspace      string
	PrimaryRepoDir string
	PrimaryRepo    repository
	SecondaryRepo  repository
	PrimaryPRID    int
}

type workspaceListResponse struct {
	Values []workspace `json:"values"`
}

type workspace struct {
	Slug string `json:"slug"`
}

type projectListResponse struct {
	Values []project `json:"values"`
}

type project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type repository struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	FullName   string `json:"full_name"`
	HasIssues  bool   `json:"has_issues"`
	MainBranch branch `json:"mainbranch"`
}

type branch struct {
	Name string `json:"name"`
}

func TestBitbucketCloudPRList(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	cfg, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "pr", "list", "--json", "*")
	cmd.Dir = fixture.PrimaryRepoDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr list failed: %v\n%s", err, output)
	}

	var prs []bitbucket.PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		t.Fatalf("parse pr list JSON: %v\n%s", err, output)
	}

	if len(prs) == 0 {
		t.Fatalf("expected at least one pull request in fixture repo")
	}

	var found bool
	for _, pr := range prs {
		if pr.Title == fixturePRTitle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected fixture pull request %q in pr list output", fixturePRTitle)
	}

	if cfg.DefaultHost == "" {
		t.Fatalf("expected configured default host")
	}
}

func TestBitbucketCloudPRStatus(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	runGitAllowFailure(t, fixture.PrimaryRepoDir, "switch", fixtureFeatureBranch)

	cmd := exec.Command(binary, "pr", "status", "--json", "*")
	cmd.Dir = fixture.PrimaryRepoDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr status failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspace         string                  `json:"workspace"`
		Repo              string                  `json:"repo"`
		CurrentBranchName string                  `json:"current_branch_name"`
		CurrentBranch     *bitbucket.PullRequest  `json:"current_branch"`
		Created           []bitbucket.PullRequest `json:"created"`
		ReviewRequested   []bitbucket.PullRequest `json:"review_requested"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr status JSON: %v\n%s", err, output)
	}

	if payload.Workspace != workspace || payload.Repo != fixture.PrimaryRepo.Slug {
		t.Fatalf("unexpected pr status payload %+v", payload)
	}
	if payload.CurrentBranchName != fixtureFeatureBranch {
		t.Fatalf("expected current branch %q, got %+v", fixtureFeatureBranch, payload)
	}
	if payload.CurrentBranch == nil || payload.CurrentBranch.ID != fixture.PrimaryPRID {
		t.Fatalf("expected current branch PR #%d, got %+v", fixture.PrimaryPRID, payload.CurrentBranch)
	}
	if len(payload.Created) == 0 {
		t.Fatalf("expected authored pull requests in status payload %+v", payload)
	}
}

func TestBitbucketCloudPRDiff(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	cmd := exec.Command(binary, "pr", "diff", prURL, "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr diff failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspace string                          `json:"workspace"`
		Repo      string                          `json:"repo"`
		ID        int                             `json:"id"`
		Patch     string                          `json:"patch"`
		Stats     []bitbucket.PullRequestDiffStat `json:"stats"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr diff JSON: %v\n%s", err, output)
	}

	if payload.Workspace != workspace || payload.Repo != fixture.PrimaryRepo.Slug || payload.ID != fixture.PrimaryPRID {
		t.Fatalf("unexpected pr diff payload %+v", payload)
	}
	if !strings.Contains(payload.Patch, "fixture.txt") {
		t.Fatalf("expected patch to contain fixture file, got %q", payload.Patch)
	}
	if len(payload.Stats) == 0 {
		t.Fatalf("expected diff stats in payload %+v", payload)
	}
}

func TestBitbucketCloudPRComment(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	commentBody := fmt.Sprintf("integration comment %d", time.Now().UTC().UnixNano())
	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	cmd := exec.Command(binary, "pr", "comment", prURL, "--body", commentBody, "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr comment failed: %v\n%s", err, output)
	}

	var payload struct {
		ID      int `json:"id"`
		Content struct {
			Raw string `json:"raw"`
		} `json:"content"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr comment JSON: %v\n%s", err, output)
	}

	if payload.ID == 0 || payload.Content.Raw != commentBody {
		t.Fatalf("unexpected pr comment payload %+v", payload)
	}
}

func TestBitbucketCloudPRClose(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	prID := ensureClosePullRequest(t, client, fixture.PrimaryRepoDir, workspace, fixture.PrimaryRepo.Slug)
	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", workspace, fixture.PrimaryRepo.Slug, prID)
	cmd := exec.Command(binary, "pr", "close", prURL, "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr close failed: %v\n%s", err, output)
	}

	var payload bitbucket.PullRequest
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr close JSON: %v\n%s", err, output)
	}

	if payload.ID != prID || payload.State != "DECLINED" {
		t.Fatalf("unexpected pr close payload %+v", payload)
	}

	updated := getPullRequest(t, client, workspace, fixture.PrimaryRepo.Slug, prID)
	if updated.State != "DECLINED" {
		t.Fatalf("expected pull request %d to be declined, got %+v", prID, updated)
	}
}

func TestBitbucketCloudIssueFlow(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	ensureFixture(t, client, hostConfig, workspace)
	issueRepo := ensureIssueRepository(t, client, workspace)
	binary := buildBinary(t)

	createCmd := exec.Command(
		binary,
		"issue", "create",
		"--repo", workspace+"/"+issueRepo.Slug,
		"--title", "bb cli issue integration flow",
		"--body", "created by the issue integration flow",
		"--json", "*",
	)
	createCmd.Env = os.Environ()

	createOutput, err := createCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue create failed: %v\n%s", err, createOutput)
	}

	var created bitbucket.Issue
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse issue create JSON: %v\n%s", err, createOutput)
	}
	if created.ID == 0 || created.Title == "" {
		t.Fatalf("unexpected issue create payload %+v", created)
	}

	listCmd := exec.Command(binary, "issue", "list", "--repo", workspace+"/"+issueRepo.Slug, "--json", "*")
	listCmd.Env = os.Environ()
	listOutput, err := listCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue list failed: %v\n%s", err, listOutput)
	}

	var issues []bitbucket.Issue
	if err := json.Unmarshal(listOutput, &issues); err != nil {
		t.Fatalf("parse issue list JSON: %v\n%s", err, listOutput)
	}
	var found bool
	for _, issue := range issues {
		if issue.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected created issue %d in issue list %+v", created.ID, issues)
	}

	closeCmd := exec.Command(binary, "issue", "close", fmt.Sprintf("%d", created.ID), "--repo", workspace+"/"+issueRepo.Slug, "--json", "*")
	closeCmd.Env = os.Environ()
	closeOutput, err := closeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue close failed: %v\n%s", err, closeOutput)
	}

	var closed bitbucket.Issue
	if err := json.Unmarshal(closeOutput, &closed); err != nil {
		t.Fatalf("parse issue close JSON: %v\n%s", err, closeOutput)
	}
	if closed.State != "resolved" {
		t.Fatalf("expected resolved issue, got %+v", closed)
	}

	reopenCmd := exec.Command(binary, "issue", "reopen", fmt.Sprintf("%d", created.ID), "--repo", workspace+"/"+issueRepo.Slug, "--json", "*")
	reopenCmd.Env = os.Environ()
	reopenOutput, err := reopenCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue reopen failed: %v\n%s", err, reopenOutput)
	}

	var reopened bitbucket.Issue
	if err := json.Unmarshal(reopenOutput, &reopened); err != nil {
		t.Fatalf("parse issue reopen JSON: %v\n%s", err, reopenOutput)
	}
	if reopened.State != "new" {
		t.Fatalf("expected reopened issue to be new, got %+v", reopened)
	}
}

func TestBitbucketCloudStatus(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	issueRepo := ensureIssueRepository(t, client, workspace)
	ensureOpenIssue(t, client, workspace, issueRepo.Slug)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "status", "--workspace", workspace, "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb status failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspaces   []string `json:"workspaces"`
		Repositories int      `json:"repositories_scanned"`
		AuthoredPRs  []struct {
			Workspace   string                `json:"workspace"`
			Repo        string                `json:"repo"`
			PullRequest bitbucket.PullRequest `json:"pull_request"`
		} `json:"authored_prs"`
		ReviewRequestedPRs []struct {
			Workspace   string                `json:"workspace"`
			Repo        string                `json:"repo"`
			PullRequest bitbucket.PullRequest `json:"pull_request"`
		} `json:"review_requested_prs"`
		YourIssues []struct {
			Workspace string          `json:"workspace"`
			Repo      string          `json:"repo"`
			Issue     bitbucket.Issue `json:"issue"`
		} `json:"your_issues"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse status JSON: %v\n%s", err, output)
	}

	if len(payload.Workspaces) != 1 || payload.Workspaces[0] != workspace {
		t.Fatalf("unexpected status workspaces %+v", payload.Workspaces)
	}
	if payload.Repositories == 0 {
		t.Fatalf("expected scanned repositories in status payload %+v", payload)
	}

	var foundPR bool
	for _, pr := range payload.AuthoredPRs {
		if pr.Workspace == workspace && pr.Repo == fixture.PrimaryRepo.Slug {
			foundPR = true
			break
		}
	}
	if !foundPR {
		t.Fatalf("expected authored fixture pull requests in status payload %+v", payload.AuthoredPRs)
	}

	var foundIssue bool
	for _, issue := range payload.YourIssues {
		if issue.Workspace == workspace && issue.Repo == issueRepo.Slug {
			foundIssue = true
			break
		}
	}
	if !foundIssue {
		t.Fatalf("expected issue fixture in status payload %+v", payload.YourIssues)
	}
}

func TestBitbucketCloudRepoCreate(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	_ = ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "repo", "create", fixtureCreateRepoSlug, "--workspace", workspace, "--project-key", fixtureProjectKey, "--reuse-existing", "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo create failed: %v\n%s", err, output)
	}

	var repo bitbucket.Repository
	if err := json.Unmarshal(output, &repo); err != nil {
		t.Fatalf("parse repo create JSON: %v\n%s", err, output)
	}

	if repo.Slug != fixtureCreateRepoSlug {
		t.Fatalf("unexpected repository %+v", repo)
	}
}

func TestBitbucketCloudRepoView(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "repo", "view", "--json", "*")
	cmd.Dir = fixture.PrimaryRepoDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo view failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspace  string `json:"workspace"`
		Repo       string `json:"repo"`
		Name       string `json:"name"`
		ProjectKey string `json:"project_key"`
		MainBranch string `json:"main_branch"`
		Remote     string `json:"remote"`
		Root       string `json:"root"`
		Private    bool   `json:"private"`
		HTMLURL    string `json:"html_url"`
		HTTPSClone string `json:"https_clone"`
		LocalClone string `json:"local_clone_url"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse repo view JSON: %v\n%s", err, output)
	}

	if payload.Workspace != workspace || payload.Repo != fixture.PrimaryRepo.Slug {
		t.Fatalf("unexpected repo identity %+v", payload)
	}
	if payload.ProjectKey != fixtureProjectKey || payload.MainBranch != "main" || payload.Remote != "origin" {
		t.Fatalf("unexpected repo metadata %+v", payload)
	}
	if payload.Root == "" || payload.HTMLURL == "" || payload.HTTPSClone == "" || payload.LocalClone == "" {
		t.Fatalf("expected repo view to include local and remote URLs %+v", payload)
	}
}

func TestBitbucketCloudRepoClone(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	cloneDir := filepath.Join(t.TempDir(), fixture.PrimaryRepo.Slug+"-clone")
	cmd := exec.Command(binary, "repo", "clone", workspace+"/"+fixture.PrimaryRepo.Slug, cloneDir, "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo clone failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Directory string `json:"directory"`
		CloneURL  string `json:"clone_url"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse repo clone JSON: %v\n%s", err, output)
	}

	if payload.Workspace != workspace || payload.Repo != fixture.PrimaryRepo.Slug {
		t.Fatalf("unexpected repo clone payload %+v", payload)
	}

	expectedDir, err := filepath.Abs(cloneDir)
	if err != nil {
		t.Fatalf("resolve clone directory: %v", err)
	}
	if payload.Directory != expectedDir {
		t.Fatalf("expected clone directory %q, got %q", expectedDir, payload.Directory)
	}

	if _, err := os.Stat(filepath.Join(cloneDir, ".git")); err != nil {
		t.Fatalf("expected cloned git repository: %v", err)
	}

	originURL := gitOutput(t, cloneDir, "remote", "get-url", "origin")
	if strings.Contains(originURL, hostConfig.Token) {
		t.Fatalf("origin remote should not contain the API token: %s", originURL)
	}
	if !strings.Contains(originURL, "x-bitbucket-api-token-auth@") {
		t.Fatalf("expected sanitized origin remote to keep Bitbucket API token username, got %s", originURL)
	}
}

func TestBitbucketCloudRepoDelete(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	ensureFixture(t, client, hostConfig, workspace)
	_ = ensureRepository(t, client, workspace, fixtureDeleteRepoSlug)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "repo", "delete", workspace+"/"+fixtureDeleteRepoSlug, "--yes", "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo delete failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Deleted   bool   `json:"deleted"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse repo delete JSON: %v\n%s", err, output)
	}

	if payload.Workspace != workspace || payload.Repo != fixtureDeleteRepoSlug || !payload.Deleted {
		t.Fatalf("unexpected repo delete payload %+v", payload)
	}

	if repositoryExists(t, client, workspace, fixtureDeleteRepoSlug) {
		t.Fatalf("expected repository %s/%s to be deleted", workspace, fixtureDeleteRepoSlug)
	}
}

func TestBitbucketCloudHumanOutputSmoke(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	issueRepo := ensureIssueRepository(t, client, workspace)
	issueID := ensureOpenIssue(t, client, workspace, issueRepo.Slug)
	binary := buildBinary(t)

	repoViewCmd := exec.Command(binary, "repo", "view")
	repoViewCmd.Dir = fixture.PrimaryRepoDir
	repoViewCmd.Env = os.Environ()
	repoViewOutput, err := repoViewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo view human output failed: %v\n%s", err, repoViewOutput)
	}
	if !strings.Contains(string(repoViewOutput), "Repository: "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in repo view output:\n%s", repoViewOutput)
	}
	if !strings.Contains(string(repoViewOutput), "Next: bb pr list --repo "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo view next step:\n%s", repoViewOutput)
	}

	repoCreateCmd := exec.Command(binary, "repo", "create", fixtureCreateRepoSlug, "--workspace", workspace, "--project-key", fixtureProjectKey, "--reuse-existing")
	repoCreateCmd.Env = os.Environ()
	repoCreateOutput, err := repoCreateCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo create human output failed: %v\n%s", err, repoCreateOutput)
	}
	if !strings.Contains(string(repoCreateOutput), "Repository: "+workspace+"/"+fixtureCreateRepoSlug) {
		t.Fatalf("expected repo header in repo create output:\n%s", repoCreateOutput)
	}
	if !strings.Contains(string(repoCreateOutput), "Next: bb repo clone "+workspace+"/"+fixtureCreateRepoSlug) {
		t.Fatalf("expected repo create next step:\n%s", repoCreateOutput)
	}

	cloneDir := filepath.Join(t.TempDir(), fixture.PrimaryRepo.Slug+"-human-clone")
	repoCloneCmd := exec.Command(binary, "repo", "clone", workspace+"/"+fixture.PrimaryRepo.Slug, cloneDir)
	repoCloneCmd.Env = os.Environ()
	repoCloneOutput, err := repoCloneCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo clone human output failed: %v\n%s", err, repoCloneOutput)
	}
	if !strings.Contains(string(repoCloneOutput), "Repository: "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in repo clone output:\n%s", repoCloneOutput)
	}
	if !strings.Contains(string(repoCloneOutput), "Next: bb repo view --repo "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo clone next step:\n%s", repoCloneOutput)
	}

	_ = ensureRepository(t, client, workspace, fixtureDeleteRepoSlug)
	repoDeleteCmd := exec.Command(binary, "repo", "delete", workspace+"/"+fixtureDeleteRepoSlug, "--yes")
	repoDeleteCmd.Env = os.Environ()
	repoDeleteOutput, err := repoDeleteCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo delete human output failed: %v\n%s", err, repoDeleteOutput)
	}
	if !strings.Contains(string(repoDeleteOutput), "Repository: "+workspace+"/"+fixtureDeleteRepoSlug) {
		t.Fatalf("expected repo header in repo delete output:\n%s", repoDeleteOutput)
	}
	if !strings.Contains(string(repoDeleteOutput), "Status: deleted") {
		t.Fatalf("expected deleted status in repo delete output:\n%s", repoDeleteOutput)
	}
	if !strings.Contains(string(repoDeleteOutput), "Next: bb repo create "+workspace+"/"+fixtureDeleteRepoSlug) {
		t.Fatalf("expected repo delete next step:\n%s", repoDeleteOutput)
	}

	prListCmd := exec.Command(binary, "pr", "list", "--repo", workspace+"/"+fixture.PrimaryRepo.Slug)
	prListCmd.Env = os.Environ()
	prListOutput, err := prListCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr list human output failed: %v\n%s", err, prListOutput)
	}
	if !strings.Contains(string(prListOutput), "Repository: "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in pr list output:\n%s", prListOutput)
	}

	prViewCmd := exec.Command(binary, "pr", "view", fmt.Sprintf("%d", fixture.PrimaryPRID), "--repo", workspace+"/"+fixture.PrimaryRepo.Slug)
	prViewCmd.Env = os.Environ()
	prViewOutput, err := prViewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr view human output failed: %v\n%s", err, prViewOutput)
	}
	if !strings.Contains(string(prViewOutput), "Repository: "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in pr view output:\n%s", prViewOutput)
	}
	if !strings.Contains(string(prViewOutput), "Next: bb pr diff "+fmt.Sprintf("%d", fixture.PrimaryPRID)+" --repo "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected pr view next step:\n%s", prViewOutput)
	}

	issueViewCmd := exec.Command(binary, "issue", "view", fmt.Sprintf("%d", issueID), "--repo", workspace+"/"+issueRepo.Slug)
	issueViewCmd.Env = os.Environ()
	issueViewOutput, err := issueViewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue view human output failed: %v\n%s", err, issueViewOutput)
	}
	if !strings.Contains(string(issueViewOutput), "Repository: "+workspace+"/"+issueRepo.Slug) {
		t.Fatalf("expected repo header in issue view output:\n%s", issueViewOutput)
	}
	if !strings.Contains(string(issueViewOutput), "Next: bb issue edit "+fmt.Sprintf("%d", issueID)+" --repo "+workspace+"/"+issueRepo.Slug) {
		t.Fatalf("expected issue view next step:\n%s", issueViewOutput)
	}

	issueListCmd := exec.Command(binary, "issue", "list", "--repo", workspace+"/"+issueRepo.Slug)
	issueListCmd.Env = os.Environ()
	issueListOutput, err := issueListCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue list human output failed: %v\n%s", err, issueListOutput)
	}
	if !strings.Contains(string(issueListOutput), "Repository: "+workspace+"/"+issueRepo.Slug) {
		t.Fatalf("expected repo header in issue list output:\n%s", issueListOutput)
	}

	searchPRsCmd := exec.Command(binary, "search", "prs", "fixture", "--repo", workspace+"/"+fixture.PrimaryRepo.Slug)
	searchPRsCmd.Env = os.Environ()
	searchPRsOutput, err := searchPRsCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb search prs human output failed: %v\n%s", err, searchPRsOutput)
	}
	if !strings.Contains(string(searchPRsOutput), "Repository: "+workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in search prs output:\n%s", searchPRsOutput)
	}
	if !strings.Contains(string(searchPRsOutput), "Query: fixture") {
		t.Fatalf("expected query line in search prs output:\n%s", searchPRsOutput)
	}
}

func TestBitbucketCloudGeneratedDocsSmoke(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	issueRepo := ensureIssueRepository(t, client, workspace)
	issueID := ensureOpenIssue(t, client, workspace, issueRepo.Slug)
	binary := buildBinary(t)

	authStatusCmd := exec.Command(binary, "auth", "status", "--check", "--json", "default_host,hosts")
	authStatusCmd.Env = os.Environ()
	authStatusOutput, err := authStatusCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb auth status generated-docs smoke failed: %v\n%s", err, authStatusOutput)
	}
	var authStatus struct {
		DefaultHost string `json:"default_host"`
		Hosts       []struct {
			Host            string `json:"host"`
			TokenConfigured bool   `json:"token_configured"`
		} `json:"hosts"`
	}
	if err := json.Unmarshal(authStatusOutput, &authStatus); err != nil {
		t.Fatalf("parse auth status JSON: %v\n%s", err, authStatusOutput)
	}
	if authStatus.DefaultHost == "" || len(authStatus.Hosts) == 0 || !authStatus.Hosts[0].TokenConfigured {
		t.Fatalf("unexpected auth status payload %+v", authStatus)
	}

	repoViewCmd := exec.Command(binary, "repo", "view", "--repo", workspace+"/"+fixture.PrimaryRepo.Slug, "--json", "host,workspace,repo,name")
	repoViewCmd.Env = os.Environ()
	repoViewOutput, err := repoViewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb repo view generated-docs smoke failed: %v\n%s", err, repoViewOutput)
	}
	var repoView struct {
		Host      string `json:"host"`
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(repoViewOutput, &repoView); err != nil {
		t.Fatalf("parse repo view JSON: %v\n%s", err, repoViewOutput)
	}
	if repoView.Workspace != workspace || repoView.Repo != fixture.PrimaryRepo.Slug || repoView.Name == "" {
		t.Fatalf("unexpected repo view payload %+v", repoView)
	}

	prStatusCmd := exec.Command(binary, "pr", "status", "--repo", workspace+"/"+fixture.PrimaryRepo.Slug, "--json", "workspace,repo,current_branch_name,created,review_requested")
	prStatusCmd.Dir = fixture.PrimaryRepoDir
	prStatusCmd.Env = os.Environ()
	prStatusOutput, err := prStatusCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr status generated-docs smoke failed: %v\n%s", err, prStatusOutput)
	}
	var prStatus struct {
		Workspace         string                  `json:"workspace"`
		Repo              string                  `json:"repo"`
		CurrentBranchName string                  `json:"current_branch_name"`
		Created           []bitbucket.PullRequest `json:"created"`
	}
	if err := json.Unmarshal(prStatusOutput, &prStatus); err != nil {
		t.Fatalf("parse pr status JSON: %v\n%s", err, prStatusOutput)
	}
	if prStatus.Workspace != workspace || prStatus.Repo != fixture.PrimaryRepo.Slug {
		t.Fatalf("unexpected pr status payload %+v", prStatus)
	}

	issueViewCmd := exec.Command(binary, "issue", "view", fmt.Sprintf("%d", issueID), "--repo", workspace+"/"+issueRepo.Slug, "--json", "id,title,state")
	issueViewCmd.Env = os.Environ()
	issueViewOutput, err := issueViewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb issue view generated-docs smoke failed: %v\n%s", err, issueViewOutput)
	}
	var issueView struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(issueViewOutput, &issueView); err != nil {
		t.Fatalf("parse issue view JSON: %v\n%s", err, issueViewOutput)
	}
	if issueView.ID != issueID || issueView.Title == "" {
		t.Fatalf("unexpected issue view payload %+v", issueView)
	}

	statusCmd := exec.Command(binary, "status", "--workspace", workspace, "--json", "user,workspaces,warnings")
	statusCmd.Env = os.Environ()
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb status generated-docs smoke failed: %v\n%s", err, statusOutput)
	}
	var statusPayload struct {
		User       string   `json:"user"`
		Workspaces []string `json:"workspaces"`
		Warnings   []string `json:"warnings"`
	}
	if err := json.Unmarshal(statusOutput, &statusPayload); err != nil {
		t.Fatalf("parse status JSON: %v\n%s", err, statusOutput)
	}
	if statusPayload.User == "" || len(statusPayload.Workspaces) == 0 || statusPayload.Workspaces[0] != workspace {
		t.Fatalf("unexpected status payload %+v", statusPayload)
	}
}

func TestBitbucketCloudPRView(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "pr", "view", fmt.Sprintf("%d", fixture.PrimaryPRID), "--workspace", workspace, "--repo", fixture.PrimaryRepo.Slug, "--json", "*")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr view failed: %v\n%s", err, output)
	}

	var pr bitbucket.PullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		t.Fatalf("parse pr view JSON: %v\n%s", err, output)
	}

	if pr.ID != fixture.PrimaryPRID || pr.Title != fixturePRTitle {
		t.Fatalf("unexpected pull request %+v", pr)
	}
}

func TestBitbucketCloudPRCreate(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	ensureBranchCommit(t, fixture.PrimaryRepoDir, fixtureCreatePRBranch, "fixture-create-command.txt", "created by pr create integration\n")

	cmd := exec.Command(
		binary,
		"pr", "create",
		"--title", fixtureCreatePRTitle,
		"--description", "Fixture pull request created by the bb pr create command.",
		"--source", fixtureCreatePRBranch,
		"--destination", "main",
		"--reuse-existing",
		"--json", "*",
	)
	cmd.Dir = fixture.PrimaryRepoDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr create failed: %v\n%s", err, output)
	}

	var pr bitbucket.PullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		t.Fatalf("parse pr create JSON: %v\n%s", err, output)
	}

	if pr.Title != fixtureCreatePRTitle || pr.Source.Branch.Name != fixtureCreatePRBranch || pr.Destination.Branch.Name != "main" {
		t.Fatalf("unexpected pull request %+v", pr)
	}
}

func TestBitbucketCloudPRCheckout(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	cmd := exec.Command(binary, "pr", "checkout", fmt.Sprintf("%d", fixture.PrimaryPRID))
	cmd.Dir = fixture.PrimaryRepoDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr checkout failed: %v\n%s", err, output)
	}

	branch := currentGitBranch(t, fixture.PrimaryRepoDir)
	if branch != fixtureFeatureBranch {
		t.Fatalf("expected checked out branch %q, got %q", fixtureFeatureBranch, branch)
	}
}

func TestBitbucketCloudPRMerge(t *testing.T) {
	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}

	_, client, hostConfig := loadIntegrationClient(t)
	workspace := resolveWorkspace(t, client)
	fixture := ensureFixture(t, client, hostConfig, workspace)
	binary := buildBinary(t)

	prID := ensureMergePullRequest(t, client, fixture.PrimaryRepoDir, workspace, fixture.PrimaryRepo.Slug)

	cmd := exec.Command(
		binary,
		"pr", "merge", fmt.Sprintf("%d", prID),
		"--workspace", workspace,
		"--repo", fixture.PrimaryRepo.Slug,
		"--message", "Fixture merge executed by bb pr merge integration test.",
		"--json", "*",
	)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb pr merge failed: %v\n%s", err, output)
	}

	var merged bitbucket.PullRequest
	if err := json.Unmarshal(output, &merged); err != nil {
		t.Fatalf("parse pr merge JSON: %v\n%s", err, output)
	}

	if merged.ID != prID || merged.State != "MERGED" {
		t.Fatalf("unexpected merged pull request %+v", merged)
	}
	if merged.MergeCommit.Hash == "" {
		t.Fatalf("expected merged pull request to include merge commit %+v", merged)
	}

	updated := getPullRequest(t, client, workspace, fixture.PrimaryRepo.Slug, prID)
	if updated.State != "MERGED" {
		t.Fatalf("expected pull request %d to be merged, got %+v", prID, updated)
	}
}

func loadIntegrationClient(t *testing.T) (config.Config, *bitbucket.Client, config.HostConfig) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	host, err := cfg.ResolveHost("")
	if err != nil {
		t.Fatalf("resolve host: %v", err)
	}

	hostConfig, ok := cfg.Hosts[host]
	if !ok {
		t.Fatalf("expected host config for %s", host)
	}

	client, err := bitbucket.NewClient(host, hostConfig)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	return cfg, client, hostConfig
}

func resolveWorkspace(t *testing.T, client *bitbucket.Client) string {
	t.Helper()

	if workspace := strings.TrimSpace(os.Getenv("BB_TEST_WORKSPACE")); workspace != "" {
		return workspace
	}

	var response workspaceListResponse
	mustRequestJSON(t, client, http.MethodGet, "/workspaces?role=member", nil, &response)
	if len(response.Values) != 1 {
		t.Fatalf("expected exactly one workspace or BB_TEST_WORKSPACE to be set, got %d", len(response.Values))
	}

	return response.Values[0].Slug
}

func ensureFixture(t *testing.T, client *bitbucket.Client, hostConfig config.HostConfig, workspace string) integrationFixture {
	t.Helper()

	ensureProject(t, client, workspace)
	primaryRepo := ensureRepository(t, client, workspace, fixturePrimaryRepoSlug)
	secondaryRepo := ensureRepository(t, client, workspace, fixtureSecondaryRepoSlug)

	baseDir := t.TempDir()
	primaryDir := filepath.Join(baseDir, fixturePrimaryRepoSlug)
	secondaryDir := filepath.Join(baseDir, fixtureSecondaryRepoSlug)

	cloneRepository(t, hostConfig, workspace, primaryRepo.Slug, primaryDir)
	cloneRepository(t, hostConfig, workspace, secondaryRepo.Slug, secondaryDir)

	ensurePrimaryRepoContent(t, primaryDir, hostConfig)
	ensureSecondaryRepoContent(t, secondaryDir, hostConfig)
	prID := ensurePullRequest(t, client, primaryDir, workspace, primaryRepo.Slug, hostConfig)

	return integrationFixture{
		Workspace:      workspace,
		PrimaryRepoDir: primaryDir,
		PrimaryRepo:    primaryRepo,
		SecondaryRepo:  secondaryRepo,
		PrimaryPRID:    prID,
	}
}

func ensureProject(t *testing.T, client *bitbucket.Client, workspace string) {
	t.Helper()

	var response projectListResponse
	mustRequestJSON(t, client, http.MethodGet, fmt.Sprintf("/workspaces/%s/projects?pagelen=100", workspace), nil, &response)
	for _, project := range response.Values {
		if project.Key == fixtureProjectKey {
			return
		}
	}

	body := map[string]any{
		"key":  fixtureProjectKey,
		"name": fixtureProjectName,
	}
	mustRequestJSON(t, client, http.MethodPost, fmt.Sprintf("/workspaces/%s/projects", workspace), body, nil)
}

func ensureRepository(t *testing.T, client *bitbucket.Client, workspace, slug string) repository {
	t.Helper()

	var repo repository
	resp, err := request(t, client, http.MethodGet, fmt.Sprintf("/repositories/%s/%s", workspace, slug), nil)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
				t.Fatalf("decode repository: %v", err)
			}
			return repo
		}
		if resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("lookup repository %s failed: %s %s", slug, resp.Status, strings.TrimSpace(string(body)))
		}
		resp.Body.Close()
	} else {
		t.Fatalf("lookup repository %s: %v", slug, err)
	}

	body := map[string]any{
		"scm":        "git",
		"is_private": true,
		"project": map[string]string{
			"key": fixtureProjectKey,
		},
	}
	mustRequestJSON(t, client, http.MethodPost, fmt.Sprintf("/repositories/%s/%s", workspace, slug), body, &repo)
	return repo
}

func ensureIssueRepository(t *testing.T, client *bitbucket.Client, workspace string) repository {
	t.Helper()

	repo := ensureRepository(t, client, workspace, fixtureIssuesRepoSlug)
	if repo.HasIssues {
		return repo
	}

	body := map[string]any{
		"has_issues": true,
	}
	mustRequestJSON(t, client, http.MethodPut, fmt.Sprintf("/repositories/%s/%s", workspace, fixtureIssuesRepoSlug), body, &repo)
	return repo
}

func ensureOpenIssue(t *testing.T, client *bitbucket.Client, workspace, repoSlug string) int {
	t.Helper()

	issues, err := client.ListIssues(context.Background(), workspace, repoSlug, bitbucket.ListIssuesOptions{
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("list issue fixtures: %v", err)
	}

	for _, issue := range issues {
		if issue.Title == "bb cli integration fixture issue" {
			if issue.State != "new" {
				if err := client.ChangeIssueState(context.Background(), workspace, repoSlug, issue.ID, bitbucket.IssueChangeOptions{State: "new"}); err != nil {
					t.Fatalf("reopen issue fixture: %v", err)
				}
			}
			return issue.ID
		}
	}

	issue, err := client.CreateIssue(context.Background(), workspace, repoSlug, bitbucket.CreateIssueOptions{
		Title: "bb cli integration fixture issue",
		Body:  "Created by integration status fixture setup.",
	})
	if err != nil {
		t.Fatalf("create issue fixture: %v", err)
	}
	return issue.ID
}

func repositoryExists(t *testing.T, client *bitbucket.Client, workspace, slug string) bool {
	t.Helper()

	resp, err := request(t, client, http.MethodGet, fmt.Sprintf("/repositories/%s/%s", workspace, slug), nil)
	if err != nil {
		t.Fatalf("lookup repository %s: %v", slug, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true
	case http.StatusNotFound:
		return false
	default:
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("lookup repository %s failed: %s %s", slug, resp.Status, strings.TrimSpace(string(body)))
		return false
	}
}

func ensurePrimaryRepoContent(t *testing.T, repoDir string, hostConfig config.HostConfig) {
	t.Helper()

	configureGitIdentity(t, repoDir)

	if !hasCommit(t, repoDir) {
		writeFile(t, filepath.Join(repoDir, "README.md"), "# bb cli integration fixture\n")
		writeFile(t, filepath.Join(repoDir, "main.go"), "package main\n\nfunc main() {}\n")
		runGit(t, repoDir, "switch", "--orphan", "main")
		runGit(t, repoDir, "add", "README.md", "main.go")
		runGit(t, repoDir, "commit", "-m", "Seed primary integration repo")
		runGit(t, repoDir, "push", "-u", "origin", "main")
		return
	}

	runGitAllowFailure(t, repoDir, "switch", "main")
}

func ensureSecondaryRepoContent(t *testing.T, repoDir string, hostConfig config.HostConfig) {
	t.Helper()

	_ = hostConfig
	configureGitIdentity(t, repoDir)

	if hasCommit(t, repoDir) {
		return
	}

	writeFile(t, filepath.Join(repoDir, "README.md"), "# secondary fixture repo\n")
	runGit(t, repoDir, "switch", "--orphan", "main")
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "Seed secondary integration repo")
	runGit(t, repoDir, "push", "-u", "origin", "main")
}

func ensurePullRequest(t *testing.T, client *bitbucket.Client, repoDir, workspace, repoSlug string, hostConfig config.HostConfig) int {
	t.Helper()

	prs, err := client.ListPullRequests(context.Background(), workspace, repoSlug, bitbucket.ListPullRequestsOptions{
		State: "OPEN",
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("list fixture pull requests: %v", err)
	}

	for _, pr := range prs {
		if pr.Title == fixturePRTitle || pr.Source.Branch.Name == fixtureFeatureBranch {
			return pr.ID
		}
	}

	configureGitIdentity(t, repoDir)
	runGitAllowFailure(t, repoDir, "switch", "main")

	if remoteBranchExists(t, repoDir, fixtureFeatureBranch) {
		runGitAllowFailure(t, repoDir, "switch", fixtureFeatureBranch)
	} else {
		runGit(t, repoDir, "switch", "-c", fixtureFeatureBranch)
	}

	content := fmt.Sprintf("integration fixture updated %s\n", time.Now().UTC().Format(time.RFC3339))
	writeFile(t, filepath.Join(repoDir, "fixture.txt"), content)
	runGit(t, repoDir, "add", "fixture.txt")
	runGit(t, repoDir, "commit", "-m", "Update integration fixture branch")
	runGit(t, repoDir, "push", "-u", "origin", fixtureFeatureBranch)
	runGitAllowFailure(t, repoDir, "switch", "main")

	body := map[string]any{
		"title":       fixturePRTitle,
		"description": "Fixture pull request created by manual integration test.",
		"source": map[string]any{
			"branch": map[string]string{
				"name": fixtureFeatureBranch,
			},
		},
		"destination": map[string]any{
			"branch": map[string]string{
				"name": "main",
			},
		},
	}

	var created bitbucket.PullRequest
	mustRequestJSON(t, client, http.MethodPost, fmt.Sprintf("/repositories/%s/%s/pullrequests", workspace, repoSlug), body, &created)
	_ = hostConfig
	return created.ID
}

func ensureMergePullRequest(t *testing.T, client *bitbucket.Client, repoDir, workspace, repoSlug string) int {
	t.Helper()

	prs, err := client.ListPullRequests(context.Background(), workspace, repoSlug, bitbucket.ListPullRequestsOptions{
		State: "OPEN",
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("list merge fixture pull requests: %v", err)
	}

	for _, pr := range prs {
		if pr.Title == fixtureMergePRTitle || pr.Source.Branch.Name == fixtureMergePRBranch {
			return pr.ID
		}
	}

	configureGitIdentity(t, repoDir)
	runGitAllowFailure(t, repoDir, "switch", "main")
	runGit(t, repoDir, "pull", "--ff-only", "origin", "main")
	runGit(t, repoDir, "switch", "-C", fixtureMergePRBranch, "main")

	content := fmt.Sprintf("merge fixture updated %s\n", time.Now().UTC().Format(time.RFC3339))
	writeFile(t, filepath.Join(repoDir, "merge-fixture.txt"), content)
	runGit(t, repoDir, "add", "merge-fixture.txt")
	runGit(t, repoDir, "commit", "-m", "Update merge integration fixture")
	runGit(t, repoDir, "push", "-u", "-f", "origin", fixtureMergePRBranch)
	runGitAllowFailure(t, repoDir, "switch", "main")

	created, err := client.CreatePullRequest(context.Background(), workspace, repoSlug, bitbucket.CreatePullRequestOptions{
		Title:             fixtureMergePRTitle,
		Description:       "Fixture pull request created by manual merge integration test.",
		SourceBranch:      fixtureMergePRBranch,
		DestinationBranch: "main",
	})
	if err != nil {
		t.Fatalf("create merge fixture pull request: %v", err)
	}

	return created.ID
}

func ensureClosePullRequest(t *testing.T, client *bitbucket.Client, repoDir, workspace, repoSlug string) int {
	t.Helper()

	prs, err := client.ListPullRequests(context.Background(), workspace, repoSlug, bitbucket.ListPullRequestsOptions{
		State: "OPEN",
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("list close fixture pull requests: %v", err)
	}

	for _, pr := range prs {
		if pr.Title == fixtureClosePRTitle || pr.Source.Branch.Name == fixtureClosePRBranch {
			return pr.ID
		}
	}

	configureGitIdentity(t, repoDir)
	runGitAllowFailure(t, repoDir, "switch", "main")
	runGit(t, repoDir, "pull", "--ff-only", "origin", "main")
	runGit(t, repoDir, "switch", "-C", fixtureClosePRBranch, "main")

	content := fmt.Sprintf("close fixture updated %s\n", time.Now().UTC().Format(time.RFC3339))
	writeFile(t, filepath.Join(repoDir, "close-fixture.txt"), content)
	runGit(t, repoDir, "add", "close-fixture.txt")
	runGit(t, repoDir, "commit", "-m", "Update close integration fixture")
	runGit(t, repoDir, "push", "-u", "-f", "origin", fixtureClosePRBranch)
	runGitAllowFailure(t, repoDir, "switch", "main")

	created, err := client.CreatePullRequest(context.Background(), workspace, repoSlug, bitbucket.CreatePullRequestOptions{
		Title:             fixtureClosePRTitle,
		Description:       "Fixture pull request created by manual close integration test.",
		SourceBranch:      fixtureClosePRBranch,
		DestinationBranch: "main",
	})
	if err != nil {
		t.Fatalf("create close fixture pull request: %v", err)
	}

	return created.ID
}

func ensureBranchCommit(t *testing.T, repoDir, branchName, fileName, content string) {
	t.Helper()

	configureGitIdentity(t, repoDir)
	runGitAllowFailure(t, repoDir, "switch", "main")

	if remoteBranchExists(t, repoDir, branchName) {
		runGitAllowFailure(t, repoDir, "switch", branchName)
		runGit(t, repoDir, "pull", "--ff-only", "origin", branchName)
	} else {
		runGit(t, repoDir, "switch", "-c", branchName)
	}

	writeFile(t, filepath.Join(repoDir, fileName), content)
	runGit(t, repoDir, "add", fileName)

	if workingTreeClean(t, repoDir) {
		runGit(t, repoDir, "push", "-u", "origin", branchName)
		runGitAllowFailure(t, repoDir, "switch", "main")
		return
	}

	runGit(t, repoDir, "commit", "-m", "Update "+branchName)
	runGit(t, repoDir, "push", "-u", "origin", branchName)
	runGitAllowFailure(t, repoDir, "switch", "main")
}

func buildBinary(t *testing.T) string {
	t.Helper()

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	binary := filepath.Join(t.TempDir(), "bb")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/bb")
	cmd.Dir = repoRoot
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build bb binary: %v\n%s", err, output)
	}

	return binary
}

func cloneRepository(t *testing.T, hostConfig config.HostConfig, workspace, repoSlug, dir string) {
	t.Helper()

	cloneURL := authenticatedCloneURL(hostConfig, workspace, repoSlug)
	cmd := exec.Command("git", "clone", cloneURL, dir)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git clone %s failed: %v\n%s", repoSlug, err, scrub(output))
	}
}

func authenticatedCloneURL(hostConfig config.HostConfig, workspace, repoSlug string) string {
	user := strings.TrimSpace(hostConfig.Username)
	if user == "" || strings.Contains(user, "@") {
		user = "x-bitbucket-api-token-auth"
	}

	return (&url.URL{
		Scheme: "https",
		Host:   "bitbucket.org",
		Path:   fmt.Sprintf("/%s/%s.git", workspace, repoSlug),
		User:   url.UserPassword(user, hostConfig.Token),
	}).String()
}

func configureGitIdentity(t *testing.T, repoDir string) {
	t.Helper()

	runGit(t, repoDir, "config", "user.name", "bb cli integration")
	runGit(t, repoDir, "config", "user.email", "hsadlersemail@gmail.com")
}

func hasCommit(t *testing.T, repoDir string) bool {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func remoteBranchExists(t *testing.T, repoDir, branch string) bool {
	t.Helper()

	cmd := exec.Command("git", "ls-remote", "--exit-code", "--heads", "origin", branch)
	cmd.Dir = repoDir
	return cmd.Run() == nil
}

func currentGitBranch(t *testing.T, repoDir string) string {
	t.Helper()

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current failed: %v\n%s", err, scrub(output))
	}

	return strings.TrimSpace(string(output))
}

func gitOutput(t *testing.T, repoDir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, scrub(output))
	}

	return strings.TrimSpace(string(output))
}

func getPullRequest(t *testing.T, client *bitbucket.Client, workspace, repoSlug string, id int) bitbucket.PullRequest {
	t.Helper()

	pr, err := client.GetPullRequest(context.Background(), workspace, repoSlug, id)
	if err != nil {
		t.Fatalf("get pull request %d: %v", id, err)
	}

	return pr
}

func workingTreeClean(t *testing.T, repoDir string) bool {
	t.Helper()

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git status failed: %v\n%s", err, scrub(output))
	}

	return strings.TrimSpace(string(output)) == ""
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func mustRequestJSON(t *testing.T, client *bitbucket.Client, method, path string, body any, target any) {
	t.Helper()

	resp, err := request(t, client, method, path, body)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s returned %s: %s", method, path, resp.Status, strings.TrimSpace(string(payload)))
	}

	if target == nil {
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode %s %s: %v", method, path, err)
	}
}

func request(t *testing.T, client *bitbucket.Client, method, path string, body any) (*http.Response, error) {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	return client.Do(context.Background(), method, path, payload, nil)
}

func runGit(t *testing.T, repoDir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, scrub(output))
	}
}

func runGitAllowFailure(t *testing.T, repoDir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	_, _ = cmd.CombinedOutput()
}

func scrub(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte(configToken()), []byte("[REDACTED]"))
}

func configToken() string {
	cfg, err := config.Load()
	if err != nil {
		return ""
	}
	host, err := cfg.ResolveHost("")
	if err != nil {
		return ""
	}
	return cfg.Hosts[host].Token
}
