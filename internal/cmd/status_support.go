package cmd

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func scanWorkspaceRepositories(ctx context.Context, client crossRepoStatusClient, currentUser bitbucket.CurrentUser, workspace string, repos []bitbucket.Repository, itemLimit int) ([]crossRepoScanResult, error) {
	if len(repos) == 0 {
		return nil, nil
	}

	workerCount := crossRepoStatusMaxConcurrentRepoScans
	if workerCount <= 0 {
		workerCount = 1
	}
	if workerCount > len(repos) {
		workerCount = len(repos)
	}

	jobs := make(chan bitbucket.Repository)
	results := make(chan crossRepoScanResult, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				results <- scanRepositoryStatus(ctx, client, currentUser, workspace, repo.Slug, itemLimit)
			}
		}()
	}

	for _, repo := range repos {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		case jobs <- repo:
		}
	}
	close(jobs)
	wg.Wait()
	close(results)

	collected := make([]crossRepoScanResult, 0, len(repos))
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		collected = append(collected, result)
	}
	return collected, nil
}

func scanRepositoryStatus(ctx context.Context, client crossRepoStatusClient, currentUser bitbucket.CurrentUser, workspace, repoSlug string, itemLimit int) crossRepoScanResult {
	result := crossRepoScanResult{repository: 1}

	prs, err := client.ListPullRequests(ctx, workspace, repoSlug, bitbucket.ListPullRequestsOptions{
		State: "OPEN",
		Limit: itemLimit,
		Sort:  "-updated_on",
	})
	if err != nil {
		result.err = err
		return result
	}

	for _, pr := range prs {
		if sameActor(currentUser, pr.Author) {
			result.authoredPRs = append(result.authoredPRs, crossRepoPullRequest{
				Workspace:   workspace,
				Repo:        repoSlug,
				PullRequest: pr,
			})
		}
		if reviewRequestedFromUser(currentUser, pr) {
			result.reviewRequestedPRs = append(result.reviewRequestedPRs, crossRepoPullRequest{
				Workspace:   workspace,
				Repo:        repoSlug,
				PullRequest: pr,
			})
		}
	}

	issues, err := client.ListIssues(ctx, workspace, repoSlug, bitbucket.ListIssuesOptions{
		Limit: itemLimit,
		Sort:  "-updated_on",
	})
	if err != nil {
		if isNoIssueTrackerError(err) {
			result.issueTrackerDisabled = true
			return result
		}
		result.err = err
		return result
	}

	for _, issue := range issues {
		if !issueNeedsAttention(issue) {
			continue
		}
		if !issueInvolvesUser(currentUser, issue) {
			continue
		}
		result.yourIssues = append(result.yourIssues, crossRepoIssue{
			Workspace: workspace,
			Repo:      repoSlug,
			Issue:     issue,
		})
	}

	return result
}

func sortCrossRepoStatus(payload *crossRepoStatusPayload) {
	slices.SortFunc(payload.AuthoredPRs, func(a, b crossRepoPullRequest) int {
		return strings.Compare(coalesce(b.PullRequest.UpdatedOn, b.PullRequest.CreatedOn), coalesce(a.PullRequest.UpdatedOn, a.PullRequest.CreatedOn))
	})
	slices.SortFunc(payload.ReviewRequestedPRs, func(a, b crossRepoPullRequest) int {
		return strings.Compare(coalesce(b.PullRequest.UpdatedOn, b.PullRequest.CreatedOn), coalesce(a.PullRequest.UpdatedOn, a.PullRequest.CreatedOn))
	})
	slices.SortFunc(payload.YourIssues, func(a, b crossRepoIssue) int {
		return strings.Compare(coalesce(b.Issue.UpdatedOn, b.Issue.CreatedOn), coalesce(a.Issue.UpdatedOn, a.Issue.CreatedOn))
	})
}

func limitCrossRepoPRs(items []crossRepoPullRequest, limit int) []crossRepoPullRequest {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitCrossRepoIssues(items []crossRepoIssue, limit int) []crossRepoIssue {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func buildCrossRepoStatusWarnings(payload crossRepoStatusPayload) []string {
	warnings := make([]string, 0, 4)
	if len(payload.WorkspacesAtRepoLimit) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"Reached --repo-limit=%d in %s. Only the most recently updated repositories were scanned there; use --repo-limit or narrow with --workspace for a more complete view.",
			payload.RepoLimitPerWorkspace,
			strings.Join(payload.WorkspacesAtRepoLimit, ", "),
		))
	}
	if payload.RepositoriesWithoutIssueTracker > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"Skipped issue status for %d repositories with issue tracking disabled. Use `bb repo view --repo <workspace>/<repo>` to inspect repository settings.",
			payload.RepositoriesWithoutIssueTracker,
		))
	}
	if payload.AuthoredPRsTotal > len(payload.AuthoredPRs) || payload.ReviewRequestedPRsTotal > len(payload.ReviewRequestedPRs) || payload.YourIssuesTotal > len(payload.YourIssues) {
		warnings = append(warnings, fmt.Sprintf(
			"Showing up to %d items per section. Use `bb pr list --repo <workspace>/<repo>` or `bb issue list --repo <workspace>/<repo>` for complete repository-level detail.",
			payload.ItemLimitPerSection,
		))
	}
	return warnings
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func writeCrossRepoPRTable(w io.Writer, prs []crossRepoPullRequest) error {
	if len(prs) == 0 {
		_, err := fmt.Fprintln(w, "None.")
		return err
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "repo\t#\ttitle\tstate\tsrc\tdst\ttsk\tcmt\tupdated"); err != nil {
		return err
	}
	for _, item := range prs {
		if _, err := fmt.Fprintf(
			tw,
			"%s/%s\t%d\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
			item.Workspace,
			item.Repo,
			item.PullRequest.ID,
			output.Truncate(item.PullRequest.Title, 32),
			output.Truncate(item.PullRequest.State, 12),
			output.TruncateMiddle(item.PullRequest.Source.Branch.Name, 12),
			output.TruncateMiddle(item.PullRequest.Destination.Branch.Name, 12),
			item.PullRequest.TaskCount,
			item.PullRequest.CommentCount,
			formatPRUpdated(item.PullRequest.UpdatedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writeCrossRepoIssueTable(w io.Writer, issues []crossRepoIssue) error {
	if len(issues) == 0 {
		_, err := fmt.Fprintln(w, "None.")
		return err
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "repo\t#\ttitle\tstate\treporter\tupdated"); err != nil {
		return err
	}
	for _, item := range issues {
		if _, err := fmt.Fprintf(
			tw,
			"%s/%s\t%d\t%s\t%s\t%s\t%s\n",
			item.Workspace,
			item.Repo,
			item.Issue.ID,
			output.Truncate(item.Issue.Title, 32),
			output.Truncate(item.Issue.State, 12),
			output.Truncate(item.Issue.Reporter.DisplayName, 16),
			formatPRUpdated(item.Issue.UpdatedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
