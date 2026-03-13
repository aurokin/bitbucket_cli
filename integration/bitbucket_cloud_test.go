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
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
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

func TestBitbucketCloudPRReview(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)

	requestChangesOutput, err := session.RunAllowFailure(t, "", "pr", "review", "request-changes", prURL, "--json", "*")
	if err != nil {
		t.Skipf("request-changes not available for this fixture/account: %v\n%s", err, requestChangesOutput)
	}

	var requested struct {
		Workspace   string `json:"workspace"`
		Repo        string `json:"repo"`
		PullRequest int    `json:"pull_request"`
		Action      string `json:"action"`
		ReviewState string `json:"review_state"`
	}
	if err := json.Unmarshal(requestChangesOutput, &requested); err != nil {
		t.Fatalf("parse pr review request-changes JSON: %v\n%s", err, requestChangesOutput)
	}
	if requested.Workspace != session.Workspace || requested.Repo != fixture.PrimaryRepo.Slug || requested.PullRequest != fixture.PrimaryPRID {
		t.Fatalf("unexpected request-changes payload %+v", requested)
	}
	if requested.Action != "request-changes" || requested.ReviewState != "changes_requested" {
		t.Fatalf("unexpected request-changes state %+v", requested)
	}

	clearHuman, err := session.RunAllowFailure(t, "", "pr", "review", "clear-request-changes", prURL)
	if err != nil {
		t.Fatalf("clear-request-changes failed: %v\n%s", err, clearHuman)
	}
	assertContainsOrdered(t, string(clearHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Action:",
		"cleared requested changes",
		"Next: bb pr view "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)

	approveOutput, err := session.RunAllowFailure(t, "", "pr", "review", "approve", prURL, "--json", "*")
	if err != nil {
		t.Skipf("approve not available for this fixture/account: %v\n%s", err, approveOutput)
	}

	var approved struct {
		Action      string `json:"action"`
		ReviewState string `json:"review_state"`
	}
	if err := json.Unmarshal(approveOutput, &approved); err != nil {
		t.Fatalf("parse pr review approve JSON: %v\n%s", err, approveOutput)
	}
	if approved.Action != "approve" || approved.ReviewState != "approved" {
		t.Fatalf("unexpected approve payload %+v", approved)
	}

	unapproveHuman, err := session.RunAllowFailure(t, "", "pr", "review", "unapprove", prURL)
	if err != nil {
		t.Fatalf("unapprove failed: %v\n%s", err, unapproveHuman)
	}
	assertContainsOrdered(t, string(unapproveHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Action:",
		"unapproved",
		"Next: bb pr view "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)
}

func TestBitbucketCloudPRActivity(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	output := session.Run(t, "", "pr", "activity", prURL, "--json", "*")

	var payload struct {
		Workspace   string                          `json:"workspace"`
		Repo        string                          `json:"repo"`
		PullRequest int                             `json:"pull_request"`
		Activity    []bitbucket.PullRequestActivity `json:"activity"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr activity JSON: %v\n%s", err, output)
	}
	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug || payload.PullRequest != fixture.PrimaryPRID {
		t.Fatalf("unexpected pr activity payload %+v", payload)
	}
	if len(payload.Activity) == 0 {
		t.Fatalf("expected pull request activity entries in %+v", payload)
	}
}

func TestBitbucketCloudPRCommits(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	output := session.Run(t, "", "pr", "commits", prURL, "--json", "*")

	var payload struct {
		Workspace   string                       `json:"workspace"`
		Repo        string                       `json:"repo"`
		PullRequest int                          `json:"pull_request"`
		Commits     []bitbucket.RepositoryCommit `json:"commits"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr commits JSON: %v\n%s", err, output)
	}
	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug || payload.PullRequest != fixture.PrimaryPRID {
		t.Fatalf("unexpected pr commits payload %+v", payload)
	}
	if len(payload.Commits) == 0 {
		t.Fatalf("expected pull request commits in %+v", payload)
	}
}

func TestBitbucketCloudPRChecks(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	output := session.Run(t, "", "pr", "checks", prURL, "--json", "*")

	var payload struct {
		Workspace   string                   `json:"workspace"`
		Repo        string                   `json:"repo"`
		PullRequest int                      `json:"pull_request"`
		Statuses    []bitbucket.CommitStatus `json:"statuses"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pr checks JSON: %v\n%s", err, output)
	}
	if payload.Workspace != session.Workspace || payload.Repo != fixture.PrimaryRepo.Slug || payload.PullRequest != fixture.PrimaryPRID {
		t.Fatalf("unexpected pr checks payload %+v", payload)
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

	editBody := commentBody + " updated"
	editHuman := session.Run(t, "", "pr", "comment", "edit", strconv.Itoa(payload.ID), "--pr", prURL, "--body", editBody)
	assertContainsOrdered(t, string(editHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Comment:",
		"Action:",
		"edited",
		"State:",
		"Next: bb pr comment view "+strconv.Itoa(payload.ID)+" --pr "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)

	deleteHuman := session.Run(t, "", "pr", "comment", "delete", strconv.Itoa(payload.ID), "--pr", prURL, "--yes")
	assertContainsOrdered(t, string(deleteHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Comment:",
		"Action:",
		"deleted",
		"State:",
		"deleted",
		"Next: bb pr view "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)
}

func TestBitbucketCloudPRTaskFlow(t *testing.T) {
	session := newIntegrationSession(t)
	fixture := session.Fixture(t)

	commentBody := fmt.Sprintf("integration task comment %d", time.Now().UTC().UnixNano())
	comment, err := session.Client.CreatePullRequestComment(context.Background(), session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, commentBody)
	if err != nil {
		t.Fatalf("create fixture pull request comment: %v", err)
	}
	commentURL := comment.Links.HTML.Href
	if commentURL == "" {
		commentURL = fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d#comment-%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, comment.ID)
	}
	prURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID)
	taskBody := fmt.Sprintf("integration task %d", time.Now().UTC().UnixNano())

	createOutput := session.Run(t, "", "pr", "task", "create", prURL, "--comment", commentURL, "--body", taskBody, "--json", "*")

	var created struct {
		Workspace   string `json:"workspace"`
		Repo        string `json:"repo"`
		PullRequest int    `json:"pull_request"`
		Task        struct {
			ID      int    `json:"id"`
			State   string `json:"state"`
			Content struct {
				Raw string `json:"raw"`
			} `json:"content"`
			Comment *struct {
				ID int `json:"id"`
			} `json:"comment"`
		} `json:"task"`
	}
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse pr task create JSON: %v\n%s", err, createOutput)
	}
	if created.Task.ID == 0 || created.Task.Content.Raw != taskBody || created.Task.Comment == nil || created.Task.Comment.ID != comment.ID {
		t.Fatalf("unexpected pr task create payload %+v", created)
	}

	listOutput := session.Run(t, "", "pr", "task", "list", prURL, "--json", "*")
	var listed struct {
		Workspace   string `json:"workspace"`
		Repo        string `json:"repo"`
		PullRequest int    `json:"pull_request"`
		Tasks       []struct {
			ID int `json:"id"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(listOutput, &listed); err != nil {
		t.Fatalf("parse pr task list JSON: %v\n%s", err, listOutput)
	}
	var found bool
	for _, task := range listed.Tasks {
		if task.ID == created.Task.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected created task %d in task list payload %+v", created.Task.ID, listed)
	}

	viewOutput := session.Run(t, "", "pr", "task", "view", strconv.Itoa(created.Task.ID), "--pr", prURL, "--json", "*")
	var viewed struct {
		Task struct {
			ID int `json:"id"`
		} `json:"task"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse pr task view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Task.ID != created.Task.ID {
		t.Fatalf("unexpected pr task view payload %+v", viewed)
	}

	editedBody := taskBody + " updated"
	editOutput := session.Run(t, "", "pr", "task", "edit", strconv.Itoa(created.Task.ID), "--pr", prURL, "--body", editedBody, "--json", "*")
	var edited struct {
		Task struct {
			Content struct {
				Raw string `json:"raw"`
			} `json:"content"`
		} `json:"task"`
	}
	if err := json.Unmarshal(editOutput, &edited); err != nil {
		t.Fatalf("parse pr task edit JSON: %v\n%s", err, editOutput)
	}
	if edited.Task.Content.Raw != editedBody {
		t.Fatalf("unexpected pr task edit payload %+v", edited)
	}

	resolveOutput := session.Run(t, "", "pr", "task", "resolve", strconv.Itoa(created.Task.ID), "--pr", prURL, "--json", "*")
	var resolved struct {
		Task struct {
			State string `json:"state"`
		} `json:"task"`
	}
	if err := json.Unmarshal(resolveOutput, &resolved); err != nil {
		t.Fatalf("parse pr task resolve JSON: %v\n%s", err, resolveOutput)
	}
	if resolved.Task.State != "RESOLVED" {
		t.Fatalf("unexpected pr task resolve payload %+v", resolved)
	}

	resolveHuman := session.Run(t, "", "pr", "task", "resolve", strconv.Itoa(created.Task.ID), "--pr", prURL)
	assertContainsOrdered(t, string(resolveHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Task:",
		"Action:",
		"resolved",
		"State:",
		"resolved",
		"Next: bb pr task reopen "+strconv.Itoa(created.Task.ID)+" --pr "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)

	reopenOutput := session.Run(t, "", "pr", "task", "reopen", strconv.Itoa(created.Task.ID), "--pr", prURL, "--json", "*")
	var reopened struct {
		Task struct {
			State string `json:"state"`
		} `json:"task"`
	}
	if err := json.Unmarshal(reopenOutput, &reopened); err != nil {
		t.Fatalf("parse pr task reopen JSON: %v\n%s", err, reopenOutput)
	}
	if reopened.Task.State != "UNRESOLVED" {
		t.Fatalf("unexpected pr task reopen payload %+v", reopened)
	}

	reopenHuman := session.Run(t, "", "pr", "task", "reopen", strconv.Itoa(created.Task.ID), "--pr", prURL)
	assertContainsOrdered(t, string(reopenHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Task:",
		"Action:",
		"reopened",
		"State:",
		"open",
		"Next: bb pr task resolve "+strconv.Itoa(created.Task.ID)+" --pr "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)

	deleteOutput := session.Run(t, "", "pr", "task", "delete", strconv.Itoa(created.Task.ID), "--pr", prURL, "--yes", "--json", "*")
	var deleted struct {
		Deleted bool `json:"deleted"`
		Task    struct {
			ID int `json:"id"`
		} `json:"task"`
	}
	if err := json.Unmarshal(deleteOutput, &deleted); err != nil {
		t.Fatalf("parse pr task delete JSON: %v\n%s", err, deleteOutput)
	}
	if !deleted.Deleted || deleted.Task.ID != created.Task.ID {
		t.Fatalf("unexpected pr task delete payload %+v", deleted)
	}

	humanDeleteTask, err := session.Client.CreatePullRequestTask(context.Background(), session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, bitbucket.CreatePullRequestTaskOptions{
		Body: "human delete smoke task",
	})
	if err != nil {
		t.Fatalf("create human delete smoke task: %v", err)
	}
	deleteHuman := session.Run(t, "", "pr", "task", "delete", strconv.Itoa(humanDeleteTask.ID), "--pr", prURL, "--yes")
	assertContainsOrdered(t, string(deleteHuman),
		"Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
		"Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID),
		"Task:",
		"Action:",
		"deleted",
		"State:",
		"deleted",
		"Next: bb pr task list "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug,
	)

	if _, err := session.Client.GetPullRequestTask(context.Background(), session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, created.Task.ID); err == nil {
		t.Fatalf("expected deleted task %d to be unavailable", created.Task.ID)
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

func TestBitbucketCloudIssueCommentFlow(t *testing.T) {
	session := newIntegrationSession(t)
	issueRepo := ensureIssueRepository(t, session.Client, session.Workspace)
	issueID := ensureOpenIssue(t, session.Client, session.Workspace, issueRepo.Slug)
	repoTarget := session.Workspace + "/" + issueRepo.Slug

	createOutput := session.Run(t, "", "issue", "comment", "create", fmt.Sprintf("%d", issueID), "--repo", repoTarget, "--body", "bb cli issue comment integration flow", "--json", "*")

	var created struct {
		Comment bitbucket.IssueComment `json:"comment"`
	}
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse issue comment create JSON: %v\n%s", err, createOutput)
	}
	if created.Comment.ID == 0 || created.Comment.Content.Raw == "" {
		t.Fatalf("unexpected issue comment create payload %+v", created)
	}

	listOutput := session.Run(t, "", "issue", "comment", "list", fmt.Sprintf("%d", issueID), "--repo", repoTarget, "--json", "*")
	var listed struct {
		Comments []bitbucket.IssueComment `json:"comments"`
	}
	if err := json.Unmarshal(listOutput, &listed); err != nil {
		t.Fatalf("parse issue comment list JSON: %v\n%s", err, listOutput)
	}
	var found bool
	for _, comment := range listed.Comments {
		if comment.ID == created.Comment.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected created issue comment in list %+v", listed.Comments)
	}

	issueURL := fmt.Sprintf("https://bitbucket.org/%s/%s/issues/%d", session.Workspace, issueRepo.Slug, issueID)
	viewOutput := session.Run(t, "", "issue", "comment", "view", fmt.Sprintf("%d", created.Comment.ID), "--issue", issueURL, "--json", "*")
	var viewed struct {
		Comment bitbucket.IssueComment `json:"comment"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse issue comment view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Comment.ID != created.Comment.ID {
		t.Fatalf("unexpected issue comment view payload %+v", viewed)
	}

	editOutput := session.Run(t, "", "issue", "comment", "edit", fmt.Sprintf("%d", created.Comment.ID), "--issue", issueURL, "--body", "updated issue comment body", "--json", "*")
	var edited struct {
		Comment bitbucket.IssueComment `json:"comment"`
	}
	if err := json.Unmarshal(editOutput, &edited); err != nil {
		t.Fatalf("parse issue comment edit JSON: %v\n%s", err, editOutput)
	}
	if edited.Comment.Content.Raw != "updated issue comment body" {
		t.Fatalf("unexpected edited issue comment %+v", edited)
	}

	deleteOutput := session.Run(t, "", "issue", "comment", "delete", fmt.Sprintf("%d", created.Comment.ID), "--issue", issueURL, "--yes", "--json", "*")
	var deleted struct {
		Deleted bool                   `json:"deleted"`
		Comment bitbucket.IssueComment `json:"comment"`
	}
	if err := json.Unmarshal(deleteOutput, &deleted); err != nil {
		t.Fatalf("parse issue comment delete JSON: %v\n%s", err, deleteOutput)
	}
	if !deleted.Deleted || deleted.Comment.ID != created.Comment.ID {
		t.Fatalf("unexpected deleted issue comment %+v", deleted)
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

func TestBitbucketCloudPipelineRun(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)

	output := session.Run(t, "", "pipeline", "run", "--repo", session.Workspace+"/"+pipelines.Repo.Slug, "--ref", "main", "--json", "*")

	var payload struct {
		Workspace string             `json:"workspace"`
		Repo      string             `json:"repo"`
		Pipeline  bitbucket.Pipeline `json:"pipeline"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline run JSON: %v\n%s", err, output)
	}
	if payload.Workspace != session.Workspace || payload.Repo != pipelines.Repo.Slug {
		t.Fatalf("unexpected pipeline run identity %+v", payload)
	}
	if payload.Pipeline.BuildNumber <= 0 || payload.Pipeline.Target.RefName != "main" {
		t.Fatalf("unexpected triggered pipeline %+v", payload.Pipeline)
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

func TestBitbucketCloudPipelineTestReports(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)

	if len(pipelines.PipelineSteps) == 0 {
		t.Skip("pipeline fixture has no steps to inspect for test reports")
	}
	_, err := session.Client.GetPipelineTestReports(context.Background(), session.Workspace, pipelines.Repo.Slug, pipelines.Pipeline.UUID, pipelines.PipelineSteps[0].UUID)
	if err != nil {
		if apiErr, ok := bitbucket.AsAPIError(err); ok && apiErr.StatusCode == http.StatusNotFound {
			t.Skip("Bitbucket did not expose test reports for the pipeline fixture step")
		}
		t.Fatalf("probe pipeline test reports: %v", err)
	}

	output := session.Run(
		t,
		"",
		"pipeline", "test-reports", fmt.Sprintf("%d", pipelines.Pipeline.BuildNumber),
		"--repo", session.Workspace+"/"+pipelines.Repo.Slug,
		"--step", pipelines.PipelineSteps[0].UUID,
		"--json", "*",
	)

	var payload struct {
		Workspace string                              `json:"workspace"`
		Repo      string                              `json:"repo"`
		Pipeline  bitbucket.Pipeline                  `json:"pipeline"`
		Step      bitbucket.PipelineStep              `json:"step"`
		Summary   bitbucket.PipelineTestReportSummary `json:"summary"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline test reports JSON: %v\n%s", err, output)
	}
	if payload.Workspace != session.Workspace || payload.Repo != pipelines.Repo.Slug {
		t.Fatalf("unexpected pipeline test reports identity %+v", payload)
	}
	if payload.Pipeline.UUID != pipelines.Pipeline.UUID || payload.Step.UUID != pipelines.PipelineSteps[0].UUID {
		t.Fatalf("unexpected pipeline test reports payload %+v", payload)
	}
	if len(payload.Summary) == 0 {
		t.Fatalf("expected non-empty pipeline test report summary %+v", payload)
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

func TestBitbucketCloudPipelineVariableFlow(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	repoTarget := session.Workspace + "/" + pipelines.Repo.Slug
	variableKey := fmt.Sprintf("BB_CLI_IT_%d", time.Now().UTC().UnixNano())

	createOutput, err := session.RunAllowFailure(t, "", "pipeline", "variable", "create", "--repo", repoTarget, "--key", variableKey, "--value", "created-value", "--json", "*")
	if err != nil {
		if bytes.Contains(createOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("pipeline variable create requires broader repository administration scopes:\n%s", createOutput)
		}
		t.Fatalf("bb pipeline variable create failed: %v\n%s", err, createOutput)
	}

	var created struct {
		Workspace string                     `json:"workspace"`
		Repo      string                     `json:"repo"`
		Variable  bitbucket.PipelineVariable `json:"variable"`
	}
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse pipeline variable create JSON: %v\n%s", err, createOutput)
	}
	if created.Workspace != session.Workspace || created.Repo != pipelines.Repo.Slug || created.Variable.Key != variableKey {
		t.Fatalf("unexpected created pipeline variable %+v", created)
	}

	viewOutput := session.Run(t, "", "pipeline", "variable", "view", created.Variable.UUID, "--repo", repoTarget, "--json", "*")
	var viewed struct {
		Variable bitbucket.PipelineVariable `json:"variable"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse pipeline variable view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Variable.UUID != created.Variable.UUID {
		t.Fatalf("expected viewed variable UUID %s, got %+v", created.Variable.UUID, viewed)
	}

	editOutput := session.Run(t, "", "pipeline", "variable", "edit", created.Variable.UUID, "--repo", repoTarget, "--value", "updated-value", "--secured", "false", "--json", "*")
	var edited struct {
		Variable bitbucket.PipelineVariable `json:"variable"`
	}
	if err := json.Unmarshal(editOutput, &edited); err != nil {
		t.Fatalf("parse pipeline variable edit JSON: %v\n%s", err, editOutput)
	}
	if edited.Variable.UUID != created.Variable.UUID || edited.Variable.Value != "updated-value" {
		t.Fatalf("unexpected edited pipeline variable %+v", edited)
	}

	var listed struct {
		Variables []bitbucket.PipelineVariable `json:"variables"`
	}
	var found bool
	for attempt := 0; attempt < 12; attempt++ {
		listOutput := session.Run(t, "", "pipeline", "variable", "list", "--repo", repoTarget, "--json", "*")
		if err := json.Unmarshal(listOutput, &listed); err != nil {
			t.Fatalf("parse pipeline variable list JSON: %v\n%s", err, listOutput)
		}
		for _, variable := range listed.Variables {
			if variable.UUID == created.Variable.UUID {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if !found {
		t.Fatalf("expected created pipeline variable in list %+v", listed.Variables)
	}

	deleteOutput := session.Run(t, "", "pipeline", "variable", "delete", created.Variable.UUID, "--repo", repoTarget, "--yes", "--json", "*")
	var deleted struct {
		Deleted  bool                       `json:"deleted"`
		Variable bitbucket.PipelineVariable `json:"variable"`
	}
	if err := json.Unmarshal(deleteOutput, &deleted); err != nil {
		t.Fatalf("parse pipeline variable delete JSON: %v\n%s", err, deleteOutput)
	}
	if !deleted.Deleted || deleted.Variable.UUID != created.Variable.UUID {
		t.Fatalf("unexpected deleted pipeline variable payload %+v", deleted)
	}
}

func TestBitbucketCloudPipelineScheduleFlow(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	repoTarget := session.Workspace + "/" + pipelines.Repo.Slug
	cronPattern := "0 0 12 1 1 ? 2099"

	createOutput, err := session.RunAllowFailure(t, "", "pipeline", "schedule", "create", "--repo", repoTarget, "--ref", "main", "--cron", cronPattern, "--enabled=false", "--json", "*")
	if err != nil {
		if bytes.Contains(createOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("pipeline schedule create requires broader repository administration scopes:\n%s", createOutput)
		}
		t.Fatalf("bb pipeline schedule create failed: %v\n%s", err, createOutput)
	}

	var created struct {
		Schedule bitbucket.PipelineSchedule `json:"schedule"`
	}
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse pipeline schedule create JSON: %v\n%s", err, createOutput)
	}
	if created.Schedule.UUID == "" || created.Schedule.CronPattern != cronPattern || created.Schedule.Enabled {
		t.Fatalf("unexpected created pipeline schedule %+v", created)
	}

	viewOutput := session.Run(t, "", "pipeline", "schedule", "view", created.Schedule.UUID, "--repo", repoTarget, "--json", "*")
	var viewed struct {
		Schedule bitbucket.PipelineSchedule `json:"schedule"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse pipeline schedule view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Schedule.UUID != created.Schedule.UUID {
		t.Fatalf("unexpected viewed pipeline schedule %+v", viewed)
	}

	enableOutput := session.Run(t, "", "pipeline", "schedule", "enable", created.Schedule.UUID, "--repo", repoTarget, "--json", "*")
	var enabled struct {
		Schedule bitbucket.PipelineSchedule `json:"schedule"`
	}
	if err := json.Unmarshal(enableOutput, &enabled); err != nil {
		t.Fatalf("parse pipeline schedule enable JSON: %v\n%s", err, enableOutput)
	}
	if !enabled.Schedule.Enabled {
		t.Fatalf("expected enabled pipeline schedule, got %+v", enabled)
	}

	var listed struct {
		Schedules []bitbucket.PipelineSchedule `json:"schedules"`
	}
	var found bool
	for attempt := 0; attempt < 12; attempt++ {
		listOutput := session.Run(t, "", "pipeline", "schedule", "list", "--repo", repoTarget, "--json", "*")
		if err := json.Unmarshal(listOutput, &listed); err != nil {
			t.Fatalf("parse pipeline schedule list JSON: %v\n%s", err, listOutput)
		}
		for _, schedule := range listed.Schedules {
			if schedule.UUID == created.Schedule.UUID {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if !found {
		t.Fatalf("expected created pipeline schedule in list %+v", listed.Schedules)
	}

	deleteOutput := session.Run(t, "", "pipeline", "schedule", "delete", created.Schedule.UUID, "--repo", repoTarget, "--yes", "--json", "*")
	var deleted struct {
		Deleted  bool                       `json:"deleted"`
		Schedule bitbucket.PipelineSchedule `json:"schedule"`
	}
	if err := json.Unmarshal(deleteOutput, &deleted); err != nil {
		t.Fatalf("parse pipeline schedule delete JSON: %v\n%s", err, deleteOutput)
	}
	if !deleted.Deleted || deleted.Schedule.UUID != created.Schedule.UUID {
		t.Fatalf("unexpected deleted pipeline schedule %+v", deleted)
	}
}

func TestBitbucketCloudPipelineRunnerList(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)

	output := session.Run(t, "", "pipeline", "runner", "list", "--repo", session.Workspace+"/"+pipelines.Repo.Slug, "--json", "*")

	var payload struct {
		Runners []bitbucket.PipelineRunner `json:"runners"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline runner list JSON: %v\n%s", err, output)
	}
}

func TestBitbucketCloudPipelineCacheList(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)

	output := session.Run(t, "", "pipeline", "cache", "list", "--repo", session.Workspace+"/"+pipelines.Repo.Slug, "--json", "*")

	var payload struct {
		Caches []bitbucket.PipelineCache `json:"caches"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		t.Fatalf("parse pipeline cache list JSON: %v\n%s", err, output)
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
	if !strings.Contains(string(prListOutput), "tsk") || !strings.Contains(string(prListOutput), "cmt") {
		t.Fatalf("expected task/comment count columns in pr list output:\n%s", prListOutput)
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
	if !strings.Contains(string(searchPRsOutput), "tsk") || !strings.Contains(string(searchPRsOutput), "cmt") {
		t.Fatalf("expected task/comment count columns in search prs output:\n%s", searchPRsOutput)
	}

	commentBody := fmt.Sprintf("integration resolve comment %d", time.Now().UTC().UnixNano())
	comment, err := session.Client.CreatePullRequestComment(context.Background(), session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, commentBody)
	if err != nil {
		t.Fatalf("create smoke pull request comment: %v", err)
	}
	commentURL := comment.Links.HTML.Href
	if commentURL == "" {
		commentURL = fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d#comment-%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, comment.ID)
	}

	resolveOutput := session.Run(t, "", "resolve", commentURL)
	if !strings.Contains(string(resolveOutput), "Type: pull-request-comment") {
		t.Fatalf("expected pull-request-comment type in resolve output:\n%s", resolveOutput)
	}
	if !strings.Contains(string(resolveOutput), "Next: bb pr comment view "+strconv.Itoa(comment.ID)+" --pr "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected resolve next step for comment URL:\n%s", resolveOutput)
	}
	messyCommentURL := fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d/?smoke=1#comment-%d", session.Workspace, fixture.PrimaryRepo.Slug, fixture.PrimaryPRID, comment.ID)
	messyResolveOutput := session.Run(t, "", "resolve", messyCommentURL)
	if !strings.Contains(string(messyResolveOutput), "Canonical URL: https://bitbucket.org/"+session.Workspace+"/"+fixture.PrimaryRepo.Slug+"/pull-requests/"+strconv.Itoa(fixture.PrimaryPRID)+"#comment-"+strconv.Itoa(comment.ID)) {
		t.Fatalf("expected canonical URL for messy resolve input:\n%s", messyResolveOutput)
	}
	if !strings.Contains(string(messyResolveOutput), "Next: bb pr comment view "+strconv.Itoa(comment.ID)+" --pr "+strconv.Itoa(fixture.PrimaryPRID)+" --repo "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected resolve next step for messy comment URL:\n%s", messyResolveOutput)
	}

	prCommentViewOutput := session.Run(t, "", "pr", "comment", "view", commentURL)
	if !strings.Contains(string(prCommentViewOutput), "Repository: "+session.Workspace+"/"+fixture.PrimaryRepo.Slug) {
		t.Fatalf("expected repo header in pr comment view output:\n%s", prCommentViewOutput)
	}
	if !strings.Contains(string(prCommentViewOutput), "Pull Request: #"+strconv.Itoa(fixture.PrimaryPRID)) {
		t.Fatalf("expected pull request id in pr comment view output:\n%s", prCommentViewOutput)
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
	runExternal(t, repoRoot, false, "go", "build", "-o", binary, "./cmd/bb")
	return binary
}

func cloneRepository(t *testing.T, hostConfig config.HostConfig, workspace, repoSlug, dir string) {
	t.Helper()

	cloneURL := authenticatedCloneURL(hostConfig, workspace, repoSlug)
	runExternal(t, "", true, "git", "clone", cloneURL, dir)
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
	runGit(t, repoDir, "config", "user.email", "bb-cli-integration@example.invalid")
}

func hasCommit(t *testing.T, repoDir string) bool {
	t.Helper()

	if _, err := runExternalAllowFailure(t, repoDir, true, "git", "rev-parse", "--verify", "HEAD"); err != nil {
		return false
	}
	return true
}

func remoteBranchExists(t *testing.T, repoDir, branch string) bool {
	t.Helper()

	_, err := runExternalAllowFailure(t, repoDir, true, "git", "ls-remote", "--exit-code", "--heads", "origin", branch)
	return err == nil
}

func currentGitBranch(t *testing.T, repoDir string) string {
	t.Helper()

	output := runExternal(t, repoDir, true, "git", "branch", "--show-current")
	return strings.TrimSpace(string(output))
}

func gitOutput(t *testing.T, repoDir string, args ...string) string {
	t.Helper()

	output := runExternal(t, repoDir, true, "git", args...)
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

	output := runExternal(t, repoDir, true, "git", "status", "--porcelain")
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

	runExternal(t, repoDir, true, "git", args...)
}

func runGitAllowFailure(t *testing.T, repoDir string, args ...string) {
	t.Helper()

	_, _ = runExternalAllowFailure(t, repoDir, true, "git", args...)
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
