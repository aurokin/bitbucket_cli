package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
)

type fakeStatusClient struct {
	repositories map[string][]bitbucket.Repository
	pullRequests map[string][]bitbucket.PullRequest
	issues       map[string]statusIssueResult

	repoGate       chan struct{}
	repoRelease    chan struct{}
	activeScans    atomic.Int32
	maxActiveScans atomic.Int32
	prCalls        atomic.Int32
	issueCalls     atomic.Int32
}

type statusIssueResult struct {
	items []bitbucket.Issue
	err   error
}

func (f *fakeStatusClient) ListRepositories(_ context.Context, workspace string, _ bitbucket.ListRepositoriesOptions) ([]bitbucket.Repository, error) {
	return append([]bitbucket.Repository(nil), f.repositories[workspace]...), nil
}

func (f *fakeStatusClient) ListPullRequests(_ context.Context, workspace, repoSlug string, _ bitbucket.ListPullRequestsOptions) ([]bitbucket.PullRequest, error) {
	f.prCalls.Add(1)
	if f.repoGate != nil {
		current := f.activeScans.Add(1)
		updateMaxAtomic(&f.maxActiveScans, current)
		f.repoGate <- struct{}{}
		<-f.repoRelease
		f.activeScans.Add(-1)
	}
	return append([]bitbucket.PullRequest(nil), f.pullRequests[workspace+"/"+repoSlug]...), nil
}

func (f *fakeStatusClient) ListIssues(_ context.Context, workspace, repoSlug string, _ bitbucket.ListIssuesOptions) ([]bitbucket.Issue, error) {
	f.issueCalls.Add(1)
	result, ok := f.issues[workspace+"/"+repoSlug]
	if !ok {
		return nil, nil
	}
	return append([]bitbucket.Issue(nil), result.items...), result.err
}

func TestIssueNeedsAttention(t *testing.T) {
	t.Parallel()

	if issueNeedsAttention(bitbucket.Issue{State: "resolved"}) {
		t.Fatal("expected resolved issue to be filtered out")
	}
	if !issueNeedsAttention(bitbucket.Issue{State: "new"}) {
		t.Fatal("expected new issue to need attention")
	}
}

func TestIssueInvolvesUser(t *testing.T) {
	t.Parallel()

	user := bitbucket.CurrentUser{AccountID: "user-1"}
	issue := bitbucket.Issue{
		Reporter: bitbucket.IssueActor{AccountID: "user-1"},
	}
	if !issueInvolvesUser(user, issue) {
		t.Fatal("expected reporter match")
	}
}

func TestBuildCrossRepoStatusAddsWarningsAndTotals(t *testing.T) {
	t.Parallel()

	client := &fakeStatusClient{
		repositories: map[string][]bitbucket.Repository{
			"acme": {
				{Slug: "repo-one"},
				{Slug: "repo-two"},
			},
		},
		pullRequests: map[string][]bitbucket.PullRequest{
			"acme/repo-one": {
				openPR(7, "Owned PR", "user-1", "", "2026-03-11T00:00:00Z"),
				openReviewPR(9, "Needs review", "user-2", "user-1", "2026-03-11T01:00:00Z"),
			},
			"acme/repo-two": {
				openPR(10, "Second owned PR", "user-1", "", "2026-03-11T02:00:00Z"),
			},
		},
		issues: map[string]statusIssueResult{
			"acme/repo-one": {
				items: []bitbucket.Issue{
					{
						ID:        5,
						Title:     "Assigned issue",
						State:     "new",
						UpdatedOn: "2026-03-11T03:00:00Z",
						Assignee:  bitbucket.IssueActor{AccountID: "user-1"},
					},
				},
			},
			"acme/repo-two": {
				err: bitbucket.NewAPIError(400, "Bad Request", []byte(`{"error":{"message":"This repository has no issue tracker"}}`)),
			},
		},
	}

	payload, err := buildCrossRepoStatus(context.Background(), client, bitbucket.CurrentUser{
		AccountID:   "user-1",
		DisplayName: "Test User",
	}, []string{"acme"}, 2, 1)
	if err != nil {
		t.Fatalf("buildCrossRepoStatus returned error: %v", err)
	}

	if payload.Repositories != 2 {
		t.Fatalf("expected 2 repositories scanned, got %d", payload.Repositories)
	}
	if payload.AuthoredPRsTotal != 2 || len(payload.AuthoredPRs) != 1 {
		t.Fatalf("expected authored PR totals to be tracked, got total=%d shown=%d", payload.AuthoredPRsTotal, len(payload.AuthoredPRs))
	}
	if payload.ReviewRequestedPRsTotal != 1 || len(payload.ReviewRequestedPRs) != 1 {
		t.Fatalf("expected review PR totals to be tracked, got total=%d shown=%d", payload.ReviewRequestedPRsTotal, len(payload.ReviewRequestedPRs))
	}
	if payload.YourIssuesTotal != 1 || len(payload.YourIssues) != 1 {
		t.Fatalf("expected issue totals to be tracked, got total=%d shown=%d", payload.YourIssuesTotal, len(payload.YourIssues))
	}
	if payload.RepositoriesWithoutIssueTracker != 1 {
		t.Fatalf("expected one repo without issue tracker, got %d", payload.RepositoriesWithoutIssueTracker)
	}
	if len(payload.WorkspacesAtRepoLimit) != 1 || payload.WorkspacesAtRepoLimit[0] != "acme" {
		t.Fatalf("expected repo limit warning metadata, got %+v", payload.WorkspacesAtRepoLimit)
	}
	if len(payload.Warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %+v", payload.Warnings)
	}
	if !containsSubstring(payload.Warnings, "bb pr list --repo <workspace>/<repo>") {
		t.Fatalf("expected warning to direct users to repo-specific detail, got %+v", payload.Warnings)
	}
}

func TestBuildCrossRepoStatusUsesBoundedConcurrency(t *testing.T) {
	lockCommandTestHooks(t)

	client := &fakeStatusClient{
		repositories: map[string][]bitbucket.Repository{
			"acme": {
				{Slug: "repo-one"},
				{Slug: "repo-two"},
				{Slug: "repo-three"},
				{Slug: "repo-four"},
			},
		},
		pullRequests: map[string][]bitbucket.PullRequest{
			"acme/repo-one":   nil,
			"acme/repo-two":   nil,
			"acme/repo-three": nil,
			"acme/repo-four":  nil,
		},
		issues:      map[string]statusIssueResult{},
		repoGate:    make(chan struct{}, 16),
		repoRelease: make(chan struct{}, 16),
	}

	previousLimit := crossRepoStatusMaxConcurrentRepoScans
	crossRepoStatusMaxConcurrentRepoScans = 2
	defer func() {
		crossRepoStatusMaxConcurrentRepoScans = previousLimit
	}()

	done := make(chan error, 1)
	go func() {
		_, err := buildCrossRepoStatus(context.Background(), client, bitbucket.CurrentUser{AccountID: "user-1"}, []string{"acme"}, 10, 10)
		done <- err
	}()

	for i := 0; i < 2; i++ {
		<-client.repoGate
	}
	if got := client.maxActiveScans.Load(); got != 2 {
		t.Fatalf("expected 2 concurrent repo scans, got %d", got)
	}

	for i := 0; i < 4; i++ {
		client.repoRelease <- struct{}{}
	}
	if err := <-done; err != nil {
		t.Fatalf("buildCrossRepoStatus returned error: %v", err)
	}
	if got := client.maxActiveScans.Load(); got > 2 {
		t.Fatalf("expected bounded concurrency, got %d", got)
	}
	if client.prCalls.Load() != 4 {
		t.Fatalf("expected 4 PR calls, got %d", client.prCalls.Load())
	}
	if client.issueCalls.Load() != 4 {
		t.Fatalf("expected 4 issue calls, got %d", client.issueCalls.Load())
	}
}

func TestResolveStatusWorkspaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/workspaces" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"slug":"zeta"},{"slug":"acme"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client, err := bitbucket.NewClient("bitbucket.org", config.HostConfig{
		AuthType: config.AuthTypeAPIToken,
		Username: "agent@example.com",
		Token:    "secret",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	workspaces, err := resolveStatusWorkspaces(context.Background(), client, "")
	if err != nil {
		t.Fatalf("resolveStatusWorkspaces returned error: %v", err)
	}
	if len(workspaces) != 2 || workspaces[0] != "acme" || workspaces[1] != "zeta" {
		t.Fatalf("unexpected workspaces %+v", workspaces)
	}

	selected, err := resolveStatusWorkspaces(context.Background(), client, "only-this")
	if err != nil {
		t.Fatalf("resolveStatusWorkspaces with selection returned error: %v", err)
	}
	if len(selected) != 1 || selected[0] != "only-this" {
		t.Fatalf("unexpected selected workspaces %+v", selected)
	}
}

func TestStatusOutputIncludesNotes(t *testing.T) {
	t.Parallel()

	payload := crossRepoStatusPayload{
		User:                "Test User",
		Workspaces:          []string{"acme"},
		Repositories:        1,
		ItemLimitPerSection: 20,
		Warnings: []string{
			"Showing up to 20 items per section. Use `bb pr list --repo <workspace>/<repo>` or `bb issue list --repo <workspace>/<repo>` for complete repository-level detail.",
		},
	}

	var buf bytes.Buffer
	if err := writeCrossRepoStatusSummary(&buf, payload); err != nil {
		t.Fatalf("writeCrossRepoStatusSummary returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Notes") || !strings.Contains(got, "bb pr list --repo <workspace>/<repo>") {
		t.Fatalf("expected notes to direct next steps, got %q", got)
	}
}

func TestStatusOutputSectionOrder(t *testing.T) {
	t.Parallel()

	payload := crossRepoStatusPayload{
		User:         "Test User",
		Workspaces:   []string{"acme"},
		Repositories: 1,
		AuthoredPRs: []crossRepoPullRequest{
			{
				Workspace: "acme",
				Repo:      "widgets",
				PullRequest: bitbucket.PullRequest{
					ID:          7,
					Title:       "Owned PR",
					State:       "OPEN",
					Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/owned"}},
					Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
				},
			},
		},
		ReviewRequestedPRs: []crossRepoPullRequest{
			{
				Workspace: "acme",
				Repo:      "widgets",
				PullRequest: bitbucket.PullRequest{
					ID:          9,
					Title:       "Needs Review",
					State:       "OPEN",
					Source:      bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/review"}},
					Destination: bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
				},
			},
		},
		YourIssues: []crossRepoIssue{
			{
				Workspace: "acme",
				Repo:      "widgets",
				Issue: bitbucket.Issue{
					ID:    5,
					Title: "Assigned issue",
					State: "new",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := writeCrossRepoStatusSummary(&buf, payload); err != nil {
		t.Fatalf("writeCrossRepoStatusSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"User: Test User",
		"Workspaces: acme",
		"Authored Pull Requests",
		"Review Requested",
		"Your Issues",
	)
}

func TestWriteCrossRepoPRTableIncludesTaskAndCommentCounts(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	prs := []crossRepoPullRequest{
		{
			Workspace: "acme",
			Repo:      "widgets",
			PullRequest: bitbucket.PullRequest{
				ID:           7,
				Title:        "Needs follow-up",
				State:        "OPEN",
				TaskCount:    3,
				CommentCount: 5,
				Source:       bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "feature/tasks"}},
				Destination:  bitbucket.PullRequestRef{Branch: bitbucket.PullRequestBranch{Name: "main"}},
				UpdatedOn:    "2026-03-11T00:00:00Z",
			},
		},
	}

	if err := writeCrossRepoPRTable(&buf, prs); err != nil {
		t.Fatalf("writeCrossRepoPRTable returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"repo",
		"tsk",
		"cmt",
		"Needs follow-up",
		"3",
		"5",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestBuildCrossRepoStatusWarnings(t *testing.T) {
	t.Parallel()

	warnings := buildCrossRepoStatusWarnings(crossRepoStatusPayload{
		RepoLimitPerWorkspace:           25,
		ItemLimitPerSection:             10,
		WorkspacesAtRepoLimit:           []string{"acme", "other"},
		RepositoriesWithoutIssueTracker: 3,
		AuthoredPRsTotal:                11,
		AuthoredPRs:                     make([]crossRepoPullRequest, 10),
	})
	if len(warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %+v", warnings)
	}
	assertOrderedSubstrings(t, strings.Join(warnings, "\n"),
		"Reached --repo-limit=25 in acme, other.",
		"Skipped issue status for 3 repositories",
		"Showing up to 10 items per section.",
	)
}

func TestBuildCrossRepoStatusWarningsEmpty(t *testing.T) {
	t.Parallel()

	warnings := buildCrossRepoStatusWarnings(crossRepoStatusPayload{
		RepoLimitPerWorkspace: 25,
		ItemLimitPerSection:   10,
		AuthoredPRsTotal:      1,
		AuthoredPRs:           make([]crossRepoPullRequest, 1),
	})
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %+v", warnings)
	}
}

func BenchmarkBuildCrossRepoStatus(b *testing.B) {
	client := &fakeStatusClient{
		repositories: map[string][]bitbucket.Repository{
			"acme": makeBenchmarkRepositories(200),
		},
		pullRequests: makeBenchmarkPRs(200),
		issues:       makeBenchmarkIssues(200),
	}
	user := bitbucket.CurrentUser{AccountID: "user-1", DisplayName: "Bench User"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := buildCrossRepoStatus(context.Background(), client, user, []string{"acme"}, 200, 50); err != nil {
			b.Fatalf("buildCrossRepoStatus returned error: %v", err)
		}
	}
}

func openPR(id int, title, authorID, reviewerID, updated string) bitbucket.PullRequest {
	pr := bitbucket.PullRequest{
		ID:        id,
		Title:     title,
		State:     "OPEN",
		UpdatedOn: updated,
		Author: bitbucket.PullRequestActor{
			AccountID: authorID,
		},
	}
	if reviewerID != "" {
		pr.Reviewers = []bitbucket.PullRequestActor{{AccountID: reviewerID}}
	}
	return pr
}

func openReviewPR(id int, title, authorID, reviewerID, updated string) bitbucket.PullRequest {
	return openPR(id, title, authorID, reviewerID, updated)
}

func containsSubstring(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func updateMaxAtomic(target *atomic.Int32, candidate int32) {
	for {
		current := target.Load()
		if candidate <= current {
			return
		}
		if target.CompareAndSwap(current, candidate) {
			return
		}
	}
}

func makeBenchmarkRepositories(count int) []bitbucket.Repository {
	repos := make([]bitbucket.Repository, 0, count)
	for i := 0; i < count; i++ {
		repos = append(repos, bitbucket.Repository{Slug: fmt.Sprintf("repo-%d", i)})
	}
	return repos
}

func makeBenchmarkPRs(count int) map[string][]bitbucket.PullRequest {
	prs := make(map[string][]bitbucket.PullRequest, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("acme/repo-%d", i)
		prs[key] = []bitbucket.PullRequest{
			openPR(i+1, fmt.Sprintf("Owned %d", i), "user-1", "", "2026-03-11T00:00:00Z"),
			openReviewPR(i+1000, fmt.Sprintf("Review %d", i), "user-2", "user-1", "2026-03-11T00:00:00Z"),
		}
	}
	return prs
}

func makeBenchmarkIssues(count int) map[string]statusIssueResult {
	issues := make(map[string]statusIssueResult, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("acme/repo-%d", i)
		issues[key] = statusIssueResult{
			items: []bitbucket.Issue{
				{
					ID:        i + 1,
					Title:     fmt.Sprintf("Issue %d", i),
					State:     "new",
					UpdatedOn: "2026-03-11T00:00:00Z",
					Assignee:  bitbucket.IssueActor{AccountID: "user-1"},
				},
			},
		}
	}
	return issues
}
