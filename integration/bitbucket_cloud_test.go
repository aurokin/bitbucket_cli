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
	fixtureProjectKey         = "BBCLI"
	fixtureProjectName        = "bb-cli integration"
	fixturePrimaryRepoSlug    = "bb-cli-integration-primary"
	fixtureSecondaryRepoSlug  = "bb-cli-integration-secondary"
	fixtureIssuesRepoSlug     = "bb-cli-integration-issues"
	fixturePipelinesRepoSlug  = "bb-cli-integration-pipelines"
	fixtureCreateRepoSlug     = "bb-cli-created-via-command"
	fixtureDeleteRepoSlug     = "bb-cli-delete-command-target"
	fixturePipelineStopBranch = "bb-cli-pipeline-stop"
	fixtureFeatureBranch      = "bb-cli-int-feature"
	fixturePRTitle            = "bb cli integration fixture pull request"
	fixtureCreatePRBranch     = "bb-cli-create-command-branch"
	fixtureCreatePRTitle      = "bb cli create command pull request"
	fixtureClosePRBranch      = "bb-cli-close-command-branch"
	fixtureClosePRTitle       = "bb cli close command pull request"
	fixtureMergePRBranch      = "bb-cli-merge-command-branch"
	fixtureMergePRTitle       = "bb cli merge command pull request"
)

type integrationFixture struct {
	Workspace      string
	PrimaryRepoDir string
	PrimaryRepo    repository
	SecondaryRepo  repository
	PrimaryPRID    int
}

type pipelineFixture struct {
	RepoDir       string
	Repo          repository
	Pipeline      bitbucket.Pipeline
	PipelineSteps []bitbucket.PipelineStep
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
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	output := session.Run(t, fixture.PrimaryRepoDir, "pr", "list", "--json", "*")

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

	if session.Config.DefaultHost == "" {
		t.Fatalf("expected configured default host")
	}
}

func TestBitbucketCloudPRStatus(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	runGitAllowFailure(t, fixture.PrimaryRepoDir, "switch", fixtureFeatureBranch)
	output := session.Run(t, fixture.PrimaryRepoDir, "pr", "status", "--json", "*")

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

	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug {
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
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	output := session.Run(t, "", "pr", "diff", prURL, "--json", "*")

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

	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug || payload.ID != fixture.PrimaryPRID {
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
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	commentBody := fmt.Sprintf("integration comment %d", time.Now().UTC().UnixNano())
	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	output := session.Run(t, "", "pr", "comment", prURL, "--body", commentBody, "--json", "*")

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
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	prID := ensureClosePullRequest(t, session.Client, fixture.PrimaryRepoDir, session.Workspace, fixture.PrimaryRepo.Slug)
	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, prID)
	output := session.Run(t, "", "pr", "close", prURL, "--json", "*")

	var payload bitbucket.PullRequest
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr close JSON: %v\n%s", err, output)
	}

	if payload.ID != prID || payload.State != "DECLINED" {
		t.Fatalf("unexpected pr close payload %+v", payload)
	}

	updated := getPullRequest(t, session.Client, session.Workspace, fixture.PrimaryRepo.Slug, prID)
	if updated.State != "DECLINED" {
		t.Fatalf("expected pull request %d to be declined, got %+v", prID, updated)
	}
}

func TestBitbucketCloudIssueFlow(t *testing.T) {
	session := newIntegrationSession(t)
	ensureFixture(t, session.Client, session.HostConfig, session.Workspace)
	issueRepo := ensureIssueRepository(t, session.Client, session.Workspace)

	createOutput := session.Run(
		t,
		"",
		"issue", "create",
		"--repo", session.Workspace+"/"+issueRepo.Slug,
		"--title", "bb cli issue integration flow",
		"--body", "created by the issue integration flow",
		"--json", "*",
	)

	var created bitbucket.Issue
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse issue create JSON: %v\n%s", err, createOutput)
	}
	if created.ID == 0 || created.Title == "" {
		t.Fatalf("unexpected issue create payload %+v", created)
	}

	listOutput := session.Run(t, "", "issue", "list", "--repo", session.Workspace+"/"+issueRepo.Slug, "--json", "*")

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

	closeOutput := session.Run(t, "", "issue", "close", fmt.Sprintf("%d", created.ID), "--repo", session.Workspace+"/"+issueRepo.Slug, "--json", "*")

	var closed bitbucket.Issue
	if err := json.Unmarshal(closeOutput, &closed); err != nil {
		t.Fatalf("parse issue close JSON: %v\n%s", err, closeOutput)
	}
	if closed.State != "resolved" {
		t.Fatalf("expected resolved issue, got %+v", closed)
	}

	reopenOutput := session.Run(t, "", "issue", "reopen", fmt.Sprintf("%d", created.ID), "--repo", session.Workspace+"/"+issueRepo.Slug, "--json", "*")

	var reopened bitbucket.Issue
	if err := json.Unmarshal(reopenOutput, &reopened); err != nil {
		t.Fatalf("parse issue reopen JSON: %v\n%s", err, reopenOutput)
	}
	if reopened.State != "new" {
		t.Fatalf("expected reopened issue to be new, got %+v", reopened)
	}
}

func TestBitbucketCloudPipelineList(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	output := session.Run(t, "", "pipeline", "list", "--repo", session.Workspace+"/"+pipelines.Repo.Slug, "--json", "*")

	var payload []bitbucket.Pipeline
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline list JSON: %v\n%s", err, output)
	}
	if len(payload) == 0 {
		t.Fatalf("expected at least one pipeline in fixture repo")
	}

	var found bool
	for _, pipeline := range payload {
		if pipeline.UUID == pipelines.Pipeline.UUID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected fixture pipeline %s in pipeline list %+v", pipelines.Pipeline.UUID, payload)
	}
}

func TestBitbucketCloudPipelineView(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	output := session.Run(t, "", "pipeline", "view", fmt.Sprintf("%d", pipelines.Pipeline.BuildNumber), "--repo", session.Workspace+"/"+pipelines.Repo.Slug, "--json", "*")

	var payload struct {
		Host      string                   `json:"host"`
		Workspace string                   `json:"workspace"`
		Repo      string                   `json:"repo"`
		Pipeline  bitbucket.Pipeline       `json:"pipeline"`
		Steps     []bitbucket.PipelineStep `json:"steps"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline view JSON: %v\n%s", err, output)
	}

	if payload.Workspace != session.Workspace || payload.Repo != pipelines.Repo.Slug {
		t.Fatalf("unexpected pipeline view identity %+v", payload)
	}
	if payload.Pipeline.UUID != pipelines.Pipeline.UUID || payload.Pipeline.BuildNumber != pipelines.Pipeline.BuildNumber {
		t.Fatalf("unexpected pipeline payload %+v", payload.Pipeline)
	}
	if len(payload.Steps) == 0 {
		t.Fatalf("expected pipeline steps in payload %+v", payload)
	}
}

func TestBitbucketCloudPipelineLog(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	if !pipelineLogAvailable(t, session.Client, session.Workspace, pipelines.Repo.Slug, pipelines.Pipeline.UUID, pipelines.PipelineSteps[0].UUID) {
		t.Skip("Bitbucket did not expose a raw step log for the pipeline fixture")
	}
	output := session.Run(
		t,
		"",
		"pipeline", "log", fmt.Sprintf("%d", pipelines.Pipeline.BuildNumber),
		"--repo", session.Workspace+"/"+pipelines.Repo.Slug,
		"--step", pipelines.PipelineSteps[0].UUID,
		"--json", "*",
	)

	var payload struct {
		Workspace string                 `json:"workspace"`
		Repo      string                 `json:"repo"`
		Pipeline  bitbucket.Pipeline     `json:"pipeline"`
		Step      bitbucket.PipelineStep `json:"step"`
		Log       string                 `json:"log"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline log JSON: %v\n%s", err, output)
	}

	if payload.Workspace != session.Workspace || payload.Repo != pipelines.Repo.Slug {
		t.Fatalf("unexpected pipeline log identity %+v", payload)
	}
	if payload.Pipeline.UUID != pipelines.Pipeline.UUID || payload.Step.UUID != pipelines.PipelineSteps[0].UUID {
		t.Fatalf("unexpected pipeline log payload %+v", payload)
	}
	if strings.TrimSpace(payload.Log) == "" {
		t.Fatalf("expected non-empty pipeline log payload %+v", payload)
	}
}

func TestBitbucketCloudPipelineStop(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	running := ensureRunningPipeline(t, session.Client, pipelines.RepoDir, session.Workspace, pipelines.Repo.Slug)

	output, err := session.RunAllowFailure(t, "", "pipeline", "stop", fmt.Sprintf("%d", running.BuildNumber), "--repo", session.Workspace+"/"+pipelines.Repo.Slug, "--yes", "--json", "*")
	if err != nil {
		if bytes.Contains(output, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("pipeline stop requires broader pipeline write scopes:\n%s", output)
		}
		if bytes.Contains(output, []byte("is no longer stoppable")) {
			t.Skipf("pipeline stop fixture completed before the stop signal was accepted:\n%s", output)
		}
		t.Fatalf("bb pipeline stop failed: %v\n%s", err, output)
	}

	var payload struct {
		Workspace string             `json:"workspace"`
		Repo      string             `json:"repo"`
		Pipeline  bitbucket.Pipeline `json:"pipeline"`
		Stopped   bool               `json:"stopped"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline stop JSON: %v\n%s", err, output)
	}

	if payload.Workspace != session.Workspace || payload.Repo != pipelines.Repo.Slug || !payload.Stopped {
		t.Fatalf("unexpected pipeline stop payload %+v", payload)
	}

	stopped := waitForPipelineState(t, session.Client, session.Workspace, pipelines.Repo.Slug, running.UUID, 36, 5*time.Second)
	if !strings.EqualFold(pipelineStateName(stopped), "STOPPED") {
		t.Fatalf("expected stopped pipeline state, got %+v", stopped)
	}
}

func TestBitbucketCloudStatus(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	issueRepo := ensureIssueRepository(t, session.Client, session.Workspace)
	ensureOpenIssue(t, session.Client, session.Workspace, issueRepo.Slug)
	output := session.Run(t, "", "status", "--workspace", session.Workspace, "--json", "*")

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

	if len(payload.Workspaces) != 1 || payload.Workspaces[0] != session.Workspace {
		t.Fatalf("unexpected status workspaces %+v", payload.Workspaces)
	}
	if payload.Repositories == 0 {
		t.Fatalf("expected scanned repositories in status payload %+v", payload)
	}

	var foundPR bool
	for _, pr := range payload.AuthoredPRs {
		if pr.Workspace == session.Workspace && pr.Repo == fixture.PrimaryRepo.Slug {
			foundPR = true
			break
		}
	}
	if !foundPR {
		t.Fatalf("expected authored fixture pull requests in status payload %+v", payload.AuthoredPRs)
	}

	var foundIssue bool
	for _, issue := range payload.YourIssues {
		if issue.Workspace == session.Workspace && issue.Repo == issueRepo.Slug {
			foundIssue = true
			break
		}
	}
	if !foundIssue {
		t.Fatalf("expected issue fixture in status payload %+v", payload.YourIssues)
	}
}

func TestBitbucketCloudRepoCreate(t *testing.T) {
	session := newIntegrationSession(t)
	_ = session.Fixture(t)
	output := session.Run(t, "", "repo", "create", fixtureCreateRepoSlug, "--workspace", session.Workspace, "--project-key", fixtureProjectKey, "--reuse-existing", "--json", "*")

	var repo bitbucket.Repository
	if err := json.Unmarshal(output, &repo); err != nil {
		t.Fatalf("parse repo create JSON: %v\n%s", err, output)
	}

	if repo.Slug != fixtureCreateRepoSlug {
		t.Fatalf("unexpected repository %+v", repo)
	}
}

func TestBitbucketCloudRepoView(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	output := session.Run(t, fixture.PrimaryRepoDir, "repo", "view", "--json", "*")

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

	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug {
		t.Fatalf("unexpected repo identity %+v", payload)
	}
	if payload.ProjectKey != fixtureProjectKey || payload.MainBranch != "main" || payload.Remote != "origin" {
		t.Fatalf("unexpected repo metadata %+v", payload)
	}
	if payload.Root == "" || payload.HTMLURL == "" || payload.HTTPSClone == "" || payload.LocalClone == "" {
		t.Fatalf("expected repo view to include local and remote URLs %+v", payload)
	}
}

func TestBitbucketCloudBrowse(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	output := session.Run(t, "", "browse", "--pr", fmt.Sprintf("%d", fixture.PrimaryPRID), "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug, "--no-browser", "--json", "*")

	var payload struct {
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Type      string `json:"type"`
		URL       string `json:"url"`
		PR        int    `json:"pr"`
		Opened    bool   `json:"opened"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse browse JSON: %v\n%s", err, output)
	}

	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug || payload.Type != "pull-request" || payload.PR != fixture.PrimaryPRID || payload.Opened {
		t.Fatalf("unexpected browse payload %+v", payload)
	}
	if !strings.Contains(payload.URL, fmt.Sprintf("/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)) {
		t.Fatalf("unexpected browse URL %q", payload.URL)
	}
}

func TestBitbucketCloudRepoClone(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	cloneDir := filepath.Join(t.TempDir(), fixture.PrimaryRepo.Slug+"-clone")
	output := session.Run(t, "", "repo", "clone", session.Workspace+"/"+fixture.PrimaryRepo.Slug, cloneDir, "--json", "*")

	var payload struct {
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Directory string `json:"directory"`
		CloneURL  string `json:"clone_url"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse repo clone JSON: %v\n%s", err, output)
	}

	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug {
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
	if strings.Contains(originURL, session.HostConfig.Token) {
		t.Fatalf("origin remote should not contain the API token: %s", originURL)
	}
	if !strings.Contains(originURL, "x-bitbucket-api-token-auth@") {
		t.Fatalf("expected sanitized origin remote to keep Bitbucket API token username, got %s", originURL)
	}
}

func TestBitbucketCloudRepoDelete(t *testing.T) {
	session := newIntegrationSession(t)
	ensureFixture(t, session.Client, session.HostConfig, session.Workspace)
	_ = ensureRepository(t, session.Client, session.Workspace, fixtureDeleteRepoSlug)
	output := session.Run(t, "", "repo", "delete", session.Workspace+"/"+fixtureDeleteRepoSlug, "--yes", "--json", "*")

	var payload struct {
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Deleted   bool   `json:"deleted"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse repo delete JSON: %v\n%s", err, output)
	}

	if payload.Workspace != session.Workspace || payload.Repo != fixtureDeleteRepoSlug || !payload.Deleted {
		t.Fatalf("unexpected repo delete payload %+v", payload)
	}

	if repositoryExists(t, session.Client, session.Workspace, fixtureDeleteRepoSlug) {
		t.Fatalf("expected repository %s/%s to be deleted", session.Workspace, fixtureDeleteRepoSlug)
	}
}

func TestBitbucketCloudHumanOutputSmoke(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	issueRepo := ensureIssueRepository(t, session.Client, session.Workspace)
	pipelineRepo := session.PipelineFixture(t)
	canReadPipelineLog := pipelineLogAvailable(t, session.Client, session.Workspace, pipelineRepo.Repo.Slug, pipelineRepo.Pipeline.UUID, pipelineRepo.PipelineSteps[0].UUID)
	issueID := ensureOpenIssue(t, session.Client, session.Workspace, issueRepo.Slug)

	repoViewOutput := session.Run(t, fixture.PrimaryRepoDir, "repo", "view")
	if !strings.Contains(string(repoViewOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in repo view output:\n%s", repoViewOutput)
	}
	if !strings.Contains(string(repoViewOutput), "Next: bb pr list --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo view next step:\n%s", repoViewOutput)
	}

	browseOutput := session.Run(t, "", "browse", "--pr", fmt.Sprintf("%d", fixture.PrimaryPRID), "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug, "--no-browser")
	if !strings.Contains(string(browseOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in browse output:\n%s", browseOutput)
	}
	if !strings.Contains(string(browseOutput), "Type: pull-request") {
		t.Fatalf("expected browse type in output:\n%s", browseOutput)
	}
	if !strings.Contains(string(browseOutput), "Status: printed") {
		t.Fatalf("expected browse printed status in output:\n%s", browseOutput)
	}

	repoCreateOutput := session.Run(t, "", "repo", "create", fixtureCreateRepoSlug, "--workspace", session.Workspace, "--project-key", fixtureProjectKey, "--reuse-existing")
	if !strings.Contains(string(repoCreateOutput), "Repository: "+session.Workspace+"/"+fixtureCreateRepoSlug) {
		t.Fatalf("expected repo header in repo create output:\n%s", repoCreateOutput)
	}
	if !strings.Contains(string(repoCreateOutput), "Next: bb repo clone "+session.Workspace+"/"+fixtureCreateRepoSlug) {
		t.Fatalf("expected repo create next step:\n%s", repoCreateOutput)
	}

	cloneDir := filepath.Join(t.TempDir(), fixture.PrimaryRepo.Slug+"-human-clone")
	repoCloneOutput := session.Run(t, "", "repo", "clone", session.Workspace+"/"+fixture.PrimaryRepo.Slug, cloneDir)
	if !strings.Contains(string(repoCloneOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in repo clone output:\n%s", repoCloneOutput)
	}
	if !strings.Contains(string(repoCloneOutput), "Next: bb repo view --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo clone next step:\n%s", repoCloneOutput)
	}

	_ = ensureRepository(t, session.Client, session.Workspace, fixtureDeleteRepoSlug)
	repoDeleteOutput := session.Run(t, "", "repo", "delete", session.Workspace+"/"+fixtureDeleteRepoSlug, "--yes")
	if !strings.Contains(string(repoDeleteOutput), "Repository: "+session.Workspace+"/"+fixtureDeleteRepoSlug) {
		t.Fatalf("expected repo header in repo delete output:\n%s", repoDeleteOutput)
	}
	if !strings.Contains(string(repoDeleteOutput), "Status: deleted") {
		t.Fatalf("expected deleted status in repo delete output:\n%s", repoDeleteOutput)
	}
	if !strings.Contains(string(repoDeleteOutput), "Next: bb repo create "+session.Workspace+"/"+fixtureDeleteRepoSlug) {
		t.Fatalf("expected repo delete next step:\n%s", repoDeleteOutput)
	}

	pipelineListOutput := session.Run(t, "", "pipeline", "list", "--repo", session.Workspace+"/"+pipelineRepo.Repo.Slug)
	if !strings.Contains(string(pipelineListOutput), "Repository: "+session.Workspace+"/"+pipelineRepo.Repo.Slug) {
		t.Fatalf("expected repo header in pipeline list output:\n%s", pipelineListOutput)
	}
	if !strings.Contains(string(pipelineListOutput), fmt.Sprintf("Next: bb pipeline view %d --repo %s/%s", pipelineRepo.Pipeline.BuildNumber, session.Workspace, pipelineRepo.Repo.Slug)) {
		t.Fatalf("expected pipeline list next step:\n%s", pipelineListOutput)
	}

	pipelineViewOutput := session.Run(t, "", "pipeline", "view", fmt.Sprintf("%d", pipelineRepo.Pipeline.BuildNumber), "--repo", session.Workspace+"/"+pipelineRepo.Repo.Slug)
	if !strings.Contains(string(pipelineViewOutput), "Repository: "+session.Workspace+"/"+pipelineRepo.Repo.Slug) {
		t.Fatalf("expected repo header in pipeline view output:\n%s", pipelineViewOutput)
	}
	if !strings.Contains(string(pipelineViewOutput), fmt.Sprintf("Pipeline: #%d", pipelineRepo.Pipeline.BuildNumber)) {
		t.Fatalf("expected pipeline number in pipeline view output:\n%s", pipelineViewOutput)
	}
	if len(pipelineRepo.PipelineSteps) > 0 && !strings.Contains(string(pipelineViewOutput), "Steps:") {
		t.Fatalf("expected steps section in pipeline view output:\n%s", pipelineViewOutput)
	}

	if canReadPipelineLog {
		pipelineLogOutput := session.Run(t, "", "pipeline", "log", fmt.Sprintf("%d", pipelineRepo.Pipeline.BuildNumber), "--repo", session.Workspace+"/"+pipelineRepo.Repo.Slug, "--step", pipelineRepo.PipelineSteps[0].UUID)
		if !strings.Contains(string(pipelineLogOutput), "Repository: "+session.Workspace+"/"+pipelineRepo.Repo.Slug) {
			t.Fatalf("expected repo header in pipeline log output:\n%s", pipelineLogOutput)
		}
		if !strings.Contains(string(pipelineLogOutput), "Step:") {
			t.Fatalf("expected step label in pipeline log output:\n%s", pipelineLogOutput)
		}
	}

	prListOutput := session.Run(t, "", "pr", "list", "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug)
	if !strings.Contains(string(prListOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in pr list output:\n%s", prListOutput)
	}

	prViewOutput := session.Run(t, "", "pr", "view", fmt.Sprintf("%d", fixture.PrimaryPRID), "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug)
	if !strings.Contains(string(prViewOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in pr view output:\n%s", prViewOutput)
	}
	if !strings.Contains(string(prViewOutput), "Next: bb pr diff "+fmt.Sprintf("%d", fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected pr view next step:\n%s", prViewOutput)
	}

	issueViewOutput := session.Run(t, "", "issue", "view", fmt.Sprintf("%d", issueID), "--repo", session.Workspace+"/"+issueRepo.Slug)
	if !strings.Contains(string(issueViewOutput), "Repository: "+session.Workspace+"/"+issueRepo.Slug) {
		t.Fatalf("expected repo header in issue view output:\n%s", issueViewOutput)
	}
	if !strings.Contains(string(issueViewOutput), "Next: bb issue edit "+fmt.Sprintf("%d", issueID)+" --repo "+session.Workspace+"/"+issueRepo.Slug) {
		t.Fatalf("expected issue view next step:\n%s", issueViewOutput)
	}

	issueListOutput := session.Run(t, "", "issue", "list", "--repo", session.Workspace+"/"+issueRepo.Slug)
	if !strings.Contains(string(issueListOutput), "Repository: "+session.Workspace+"/"+issueRepo.Slug) {
		t.Fatalf("expected repo header in issue list output:\n%s", issueListOutput)
	}

	searchPRsOutput := session.Run(t, "", "search", "prs", "fixture", "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug)
	if !strings.Contains(string(searchPRsOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in search prs output:\n%s", searchPRsOutput)
	}
	if !strings.Contains(string(searchPRsOutput), "Query: fixture") {
		t.Fatalf("expected query line in search prs output:\n%s", searchPRsOutput)
	}
}

func TestBitbucketCloudGeneratedDocsSmoke(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	issueRepo := ensureIssueRepository(t, session.Client, session.Workspace)
	pipelineRepo := session.PipelineFixture(t)
	canReadPipelineLog := pipelineLogAvailable(t, session.Client, session.Workspace, pipelineRepo.Repo.Slug, pipelineRepo.Pipeline.UUID, pipelineRepo.PipelineSteps[0].UUID)
	issueID := ensureOpenIssue(t, session.Client, session.Workspace, issueRepo.Slug)

	authStatusOutput := session.Run(t, "", "auth", "status", "--check", "--json", "default_host,hosts")
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

	repoViewOutput := session.Run(t, "", "repo", "view", "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug, "--json", "host,workspace,repo,name")
	var repoView struct {
		Host      string `json:"host"`
		Workspace string `json:"workspace"`
		Repo      string `json:"repo"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(repoViewOutput, &repoView); err != nil {
		t.Fatalf("parse repo view JSON: %v\n%s", err, repoViewOutput)
	}
	if repoView.Workspace != session.Workspace || repoView.Repo != fixture.PrimaryRepo.Slug || repoView.Name == "" {
		t.Fatalf("unexpected repo view payload %+v", repoView)
	}

	browseOutput := session.Run(t, "", "browse", "--pr", fmt.Sprintf("%d", fixture.PrimaryPRID), "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug, "--no-browser", "--json", "url,type,pr")
	var browse struct {
		URL  string `json:"url"`
		Type string `json:"type"`
		PR   int    `json:"pr"`
	}
	if err := json.Unmarshal(browseOutput, &browse); err != nil {
		t.Fatalf("parse browse JSON: %v\n%s", err, browseOutput)
	}
	if browse.Type != "pull-request" || browse.PR != fixture.PrimaryPRID || browse.URL == "" {
		t.Fatalf("unexpected browse payload %+v", browse)
	}

	pipelineViewOutput := session.Run(t, "", "pipeline", "view", fmt.Sprintf("%d", pipelineRepo.Pipeline.BuildNumber), "--repo", session.Workspace+"/"+pipelineRepo.Repo.Slug, "--json", "host,workspace,repo,pipeline,steps")
	var pipelineView struct {
		Host      string                   `json:"host"`
		Workspace string                   `json:"workspace"`
		Repo      string                   `json:"repo"`
		Pipeline  bitbucket.Pipeline       `json:"pipeline"`
		Steps     []bitbucket.PipelineStep `json:"steps"`
	}
	if err := json.Unmarshal(pipelineViewOutput, &pipelineView); err != nil {
		t.Fatalf("parse pipeline view JSON: %v\n%s", err, pipelineViewOutput)
	}
	if pipelineView.Workspace != session.Workspace || pipelineView.Repo != pipelineRepo.Repo.Slug || pipelineView.Pipeline.BuildNumber == 0 {
		t.Fatalf("unexpected pipeline view payload %+v", pipelineView)
	}

	if canReadPipelineLog {
		pipelineLogOutput := session.Run(t, "", "pipeline", "log", fmt.Sprintf("%d", pipelineRepo.Pipeline.BuildNumber), "--repo", session.Workspace+"/"+pipelineRepo.Repo.Slug, "--step", pipelineRepo.PipelineSteps[0].UUID, "--json", "pipeline,step,log")
		var pipelineLog struct {
			Pipeline bitbucket.Pipeline     `json:"pipeline"`
			Step     bitbucket.PipelineStep `json:"step"`
			Log      string                 `json:"log"`
		}
		if err := json.Unmarshal(pipelineLogOutput, &pipelineLog); err != nil {
			t.Fatalf("parse pipeline log JSON: %v\n%s", err, pipelineLogOutput)
		}
		if pipelineLog.Pipeline.BuildNumber == 0 || pipelineLog.Step.UUID == "" || strings.TrimSpace(pipelineLog.Log) == "" {
			t.Fatalf("unexpected pipeline log payload %+v", pipelineLog)
		}
	}

	prStatusOutput := session.Run(t, fixture.PrimaryRepoDir, "pr", "status", "--repo", session.Workspace+"/"+fixture.PrimaryRepo.Slug, "--json", "workspace,repo,current_branch_name,created,review_requested")
	var prStatus struct {
		Workspace         string                  `json:"workspace"`
		Repo              string                  `json:"repo"`
		CurrentBranchName string                  `json:"current_branch_name"`
		Created           []bitbucket.PullRequest `json:"created"`
	}
	if err := json.Unmarshal(prStatusOutput, &prStatus); err != nil {
		t.Fatalf("parse pr status JSON: %v\n%s", err, prStatusOutput)
	}
	if prStatus.Workspace != session.Workspace || prStatus.Repo != fixture.PrimaryRepo.Slug {
		t.Fatalf("unexpected pr status payload %+v", prStatus)
	}

	issueViewOutput := session.Run(t, "", "issue", "view", fmt.Sprintf("%d", issueID), "--repo", session.Workspace+"/"+issueRepo.Slug, "--json", "id,title,state")
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

	statusOutput := session.Run(t, "", "status", "--workspace", session.Workspace, "--json", "user,workspaces,warnings")
	var statusPayload struct {
		User       string   `json:"user"`
		Workspaces []string `json:"workspaces"`
		Warnings   []string `json:"warnings"`
	}
	if err := json.Unmarshal(statusOutput, &statusPayload); err != nil {
		t.Fatalf("parse status JSON: %v\n%s", err, statusOutput)
	}
	if statusPayload.User == "" || len(statusPayload.Workspaces) == 0 || statusPayload.Workspaces[0] != session.Workspace {
		t.Fatalf("unexpected status payload %+v", statusPayload)
	}
}

func TestBitbucketCloudPRView(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	output := session.Run(t, "", "pr", "view", fmt.Sprintf("%d", fixture.PrimaryPRID), "--workspace", session.Workspace, "--repo", fixture.PrimaryRepo.Slug, "--json", "*")

	var pr bitbucket.PullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		t.Fatalf("parse pr view JSON: %v\n%s", err, output)
	}

	if pr.ID != fixture.PrimaryPRID || pr.Title != fixturePRTitle {
		t.Fatalf("unexpected pull request %+v", pr)
	}
}

func TestBitbucketCloudPRCreate(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	ensureBranchCommit(t, fixture.PrimaryRepoDir, fixtureCreatePRBranch, "fixture-create-command.txt", "created by pr create integration\n")

	output := session.Run(
		t,
		fixture.PrimaryRepoDir,
		"pr", "create",
		"--title", fixtureCreatePRTitle,
		"--description", "Fixture pull request created by the bb pr create command.",
		"--source", fixtureCreatePRBranch,
		"--destination", "main",
		"--reuse-existing",
		"--json", "*",
	)

	var pr bitbucket.PullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		t.Fatalf("parse pr create JSON: %v\n%s", err, output)
	}

	if pr.Title != fixtureCreatePRTitle || pr.Source.Branch.Name != fixtureCreatePRBranch || pr.Destination.Branch.Name != "main" {
		t.Fatalf("unexpected pull request %+v", pr)
	}
}

func TestBitbucketCloudPRCheckout(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	_ = session.Run(t, fixture.PrimaryRepoDir, "pr", "checkout", fmt.Sprintf("%d", fixture.PrimaryPRID))

	branch := currentGitBranch(t, fixture.PrimaryRepoDir)
	if branch != fixtureFeatureBranch {
		t.Fatalf("expected checked out branch %q, got %q", fixtureFeatureBranch, branch)
	}
}

func TestBitbucketCloudPRMerge(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)
	prID := ensureMergePullRequest(t, session.Client, fixture.PrimaryRepoDir, session.Workspace, fixture.PrimaryRepo.Slug)

	output := session.Run(
		t,
		"",
		"pr", "merge", fmt.Sprintf("%d", prID),
		"--workspace", session.Workspace,
		"--repo", fixture.PrimaryRepo.Slug,
		"--message", "Fixture merge executed by bb pr merge integration test.",
		"--json", "*",
	)

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

	updated := getPullRequest(t, session.Client, session.Workspace, fixture.PrimaryRepo.Slug, prID)
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

func ensurePipelineFixture(t *testing.T, client *bitbucket.Client, hostConfig config.HostConfig, workspace string) pipelineFixture {
	t.Helper()

	ensureProject(t, client, workspace)
	repo := ensureRepository(t, client, workspace, fixturePipelinesRepoSlug)

	baseDir := t.TempDir()
	repoDir := filepath.Join(baseDir, fixturePipelinesRepoSlug)
	cloneRepository(t, hostConfig, workspace, repo.Slug, repoDir)
	ensurePipelineRepoContent(t, repoDir)
	ensurePipelinesEnabled(t, client, workspace, repo.Slug)
	pipeline := ensurePipelineRun(t, client, repoDir, workspace, repo.Slug)
	steps := waitForPipelineSteps(t, client, workspace, repo.Slug, pipeline.UUID)

	return pipelineFixture{
		RepoDir:       repoDir,
		Repo:          repo,
		Pipeline:      pipeline,
		PipelineSteps: steps,
	}
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

func ensurePipelineRepoContent(t *testing.T, repoDir string) {
	t.Helper()

	configureGitIdentity(t, repoDir)

	if hasCommit(t, repoDir) {
		runGitAllowFailure(t, repoDir, "switch", "main")
		runGitAllowFailure(t, repoDir, "pull", "--ff-only", "origin", "main")
	} else {
		runGit(t, repoDir, "switch", "--orphan", "main")
	}

	writeFile(t, filepath.Join(repoDir, "README.md"), "# pipelines fixture repo\n")
	writeFile(t, filepath.Join(repoDir, "bitbucket-pipelines.yml"), "pipelines:\n  default:\n    - step:\n        name: fixture-step\n        script:\n          - echo 'bb pipeline fixture'\n  branches:\n    bb-cli-pipeline-stop:\n      - step:\n          name: stop-fixture-step\n          script:\n            - echo 'starting stop fixture'\n            - sleep 180\n")
	runGit(t, repoDir, "add", "README.md", "bitbucket-pipelines.yml")

	if workingTreeClean(t, repoDir) {
		if !remoteBranchExists(t, repoDir, "main") {
			runGit(t, repoDir, "push", "-u", "origin", "main")
		}
		return
	}

	runGit(t, repoDir, "commit", "-m", "Seed pipelines integration repo")
	runGit(t, repoDir, "push", "-u", "origin", "main")
}

func ensurePipelinesEnabled(t *testing.T, client *bitbucket.Client, workspace, repoSlug string) {
	t.Helper()

	cfg, err := client.GetPipelineConfig(context.Background(), workspace, repoSlug)
	if err == nil {
		if cfg.Enabled {
			return
		}
		cfg.Enabled = true
		if _, err := client.UpdatePipelineConfig(context.Background(), workspace, repoSlug, cfg); err != nil {
			t.Fatalf("enable pipelines: %v", err)
		}
		return
	}

	if apiErr, ok := bitbucket.AsAPIError(err); ok && apiErr.StatusCode == http.StatusNotFound {
		if _, err := client.UpdatePipelineConfig(context.Background(), workspace, repoSlug, bitbucket.PipelineConfig{Enabled: true}); err != nil {
			t.Fatalf("create pipeline config: %v", err)
		}
		return
	}

	t.Fatalf("get pipeline config: %v", err)
}

func ensurePipelineRun(t *testing.T, client *bitbucket.Client, repoDir, workspace, repoSlug string) bitbucket.Pipeline {
	t.Helper()

	pipelines, err := client.ListPipelines(context.Background(), workspace, repoSlug, bitbucket.ListPipelinesOptions{
		Sort:  "-created_on",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("list pipeline fixtures: %v", err)
	}
	if len(pipelines) > 0 {
		return pipelines[0]
	}

	configureGitIdentity(t, repoDir)
	runGitAllowFailure(t, repoDir, "switch", "main")
	writeFile(t, filepath.Join(repoDir, "pipeline-trigger.txt"), fmt.Sprintf("triggered at %s\n", time.Now().UTC().Format(time.RFC3339Nano)))
	runGit(t, repoDir, "add", "pipeline-trigger.txt")
	if !workingTreeClean(t, repoDir) {
		runGit(t, repoDir, "commit", "-m", "Trigger pipelines integration fixture")
		runGit(t, repoDir, "push", "-u", "origin", "main")
	}

	for attempt := 0; attempt < 24; attempt++ {
		pipelines, err = client.ListPipelines(context.Background(), workspace, repoSlug, bitbucket.ListPipelinesOptions{
			Sort:  "-created_on",
			Limit: 5,
		})
		if err == nil && len(pipelines) > 0 {
			return pipelines[0]
		}
		time.Sleep(5 * time.Second)
	}

	triggered, err := client.TriggerPipeline(context.Background(), workspace, repoSlug, bitbucket.TriggerPipelineOptions{
		RefType: "branch",
		RefName: "main",
	})
	if err != nil {
		t.Fatalf("trigger pipeline fixture: %v", err)
	}
	return triggered
}

func waitForPipelineSteps(t *testing.T, client *bitbucket.Client, workspace, repoSlug, pipelineUUID string) []bitbucket.PipelineStep {
	t.Helper()

	for attempt := 0; attempt < 24; attempt++ {
		steps, err := client.ListPipelineSteps(context.Background(), workspace, repoSlug, pipelineUUID)
		if err == nil && len(steps) > 0 {
			return steps
		}
		time.Sleep(5 * time.Second)
	}

	steps, err := client.ListPipelineSteps(context.Background(), workspace, repoSlug, pipelineUUID)
	if err != nil {
		t.Fatalf("list pipeline steps: %v", err)
	}
	return steps
}

func pipelineLogAvailable(t *testing.T, client *bitbucket.Client, workspace, repoSlug, pipelineUUID, stepUUID string) bool {
	t.Helper()

	_, err := client.GetPipelineStepLog(context.Background(), workspace, repoSlug, pipelineUUID, stepUUID)
	if err == nil {
		return true
	}
	if apiErr, ok := bitbucket.AsAPIError(err); ok && (apiErr.StatusCode == http.StatusNotFound || apiErr.StatusCode == http.StatusNotAcceptable) {
		return false
	}
	t.Fatalf("probe pipeline log availability: %v", err)
	return false
}

func ensureRunningPipeline(t *testing.T, client *bitbucket.Client, repoDir, workspace, repoSlug string) bitbucket.Pipeline {
	t.Helper()

	pipelines, err := client.ListPipelines(context.Background(), workspace, repoSlug, bitbucket.ListPipelinesOptions{
		Sort:  "-created_on",
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("list running pipeline fixtures: %v", err)
	}
	for _, pipeline := range pipelines {
		if pipeline.Target.RefName == fixturePipelineStopBranch && isActivePipeline(pipeline) {
			return pipeline
		}
	}

	ensureBranchCommit(t, repoDir, fixturePipelineStopBranch, "stop-trigger.txt", fmt.Sprintf("stop fixture updated %s\n", time.Now().UTC().Format(time.RFC3339Nano)))
	for attempt := 0; attempt < 24; attempt++ {
		pipelines, err = client.ListPipelines(context.Background(), workspace, repoSlug, bitbucket.ListPipelinesOptions{
			Sort:  "-created_on",
			Limit: 20,
		})
		if err == nil {
			for _, pipeline := range pipelines {
				if pipeline.Target.RefName == fixturePipelineStopBranch && isActivePipeline(pipeline) {
					return pipeline
				}
			}
		}
		time.Sleep(5 * time.Second)
	}

	triggered, err := client.TriggerPipeline(context.Background(), workspace, repoSlug, bitbucket.TriggerPipelineOptions{
		RefType: "branch",
		RefName: fixturePipelineStopBranch,
	})
	if err != nil {
		t.Fatalf("trigger stop pipeline fixture: %v", err)
	}
	return triggered
}

func waitForPipelineState(t *testing.T, client *bitbucket.Client, workspace, repoSlug, pipelineUUID string, attempts int, delay time.Duration) bitbucket.Pipeline {
	t.Helper()

	var pipeline bitbucket.Pipeline
	for attempt := 0; attempt < attempts; attempt++ {
		var err error
		pipeline, err = client.GetPipeline(context.Background(), workspace, repoSlug, pipelineUUID)
		if err != nil {
			t.Fatalf("get pipeline %s: %v", pipelineUUID, err)
		}
		if !isActivePipeline(pipeline) {
			return pipeline
		}
		time.Sleep(delay)
	}

	return pipeline
}

func isActivePipeline(pipeline bitbucket.Pipeline) bool {
	return !strings.EqualFold(pipelineStateName(pipeline), "STOPPED") &&
		!strings.EqualFold(pipelineStateName(pipeline), "SUCCESSFUL") &&
		!strings.EqualFold(pipelineStateName(pipeline), "FAILED") &&
		!strings.EqualFold(pipelineStateName(pipeline), "ERROR")
}

func pipelineStateName(pipeline bitbucket.Pipeline) string {
	if pipeline.State.Result.Name != "" {
		return pipeline.State.Result.Name
	}
	return pipeline.State.Name
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
