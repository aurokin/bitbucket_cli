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
	"github.com/spf13/cobra"
)

var crossRepoStatusMaxConcurrentRepoScans = 8

type crossRepoStatusPayload struct {
	User                            string                 `json:"user,omitempty"`
	Workspaces                      []string               `json:"workspaces"`
	Repositories                    int                    `json:"repositories_scanned"`
	RepoLimitPerWorkspace           int                    `json:"repo_limit_per_workspace,omitempty"`
	ItemLimitPerSection             int                    `json:"item_limit_per_section,omitempty"`
	AuthoredPRsTotal                int                    `json:"authored_prs_total"`
	ReviewRequestedPRsTotal         int                    `json:"review_requested_prs_total"`
	YourIssuesTotal                 int                    `json:"your_issues_total"`
	RepositoriesWithoutIssueTracker int                    `json:"repositories_without_issue_tracker"`
	WorkspacesAtRepoLimit           []string               `json:"workspaces_at_repo_limit,omitempty"`
	Warnings                        []string               `json:"warnings,omitempty"`
	AuthoredPRs                     []crossRepoPullRequest `json:"authored_prs"`
	ReviewRequestedPRs              []crossRepoPullRequest `json:"review_requested_prs"`
	YourIssues                      []crossRepoIssue       `json:"your_issues"`
}

type crossRepoPullRequest struct {
	Workspace   string                `json:"workspace"`
	Repo        string                `json:"repo"`
	PullRequest bitbucket.PullRequest `json:"pull_request"`
}

type crossRepoIssue struct {
	Workspace string          `json:"workspace"`
	Repo      string          `json:"repo"`
	Issue     bitbucket.Issue `json:"issue"`
}

type crossRepoStatusClient interface {
	ListRepositories(ctx context.Context, workspace string, options bitbucket.ListRepositoriesOptions) ([]bitbucket.Repository, error)
	ListPullRequests(ctx context.Context, workspace, repoSlug string, options bitbucket.ListPullRequestsOptions) ([]bitbucket.PullRequest, error)
	ListIssues(ctx context.Context, workspace, repoSlug string, options bitbucket.ListIssuesOptions) ([]bitbucket.Issue, error)
}

type crossRepoScanResult struct {
	repository           int
	issueTrackerDisabled bool
	authoredPRs          []crossRepoPullRequest
	reviewRequestedPRs   []crossRepoPullRequest
	yourIssues           []crossRepoIssue
	err                  error
}

func newStatusCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repoLimit int
	var itemLimit int

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cross-repository pull request and issue status",
		Long:  "Show authored pull requests, pull requests requesting your review, and open issues that involve you across accessible repositories.",
		Example: "  bb status\n" +
			"  bb status --workspace workspace-slug --limit 10\n" +
			"  bb status --json authored_prs,review_requested_prs,your_issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			_, client, err := resolveAuthenticatedClient(host)
			if err != nil {
				return err
			}

			currentUser, err := client.CurrentUser(context.Background())
			if err != nil {
				return err
			}

			workspaces, err := resolveStatusWorkspaces(context.Background(), client, workspace)
			if err != nil {
				return err
			}

			payload, err := buildCrossRepoStatus(context.Background(), client, currentUser, workspaces, repoLimit, itemLimit)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if _, err := fmt.Fprintf(w, "User: %s\n", payload.User); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Workspaces: %s\n", strings.Join(payload.Workspaces, ", ")); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Repositories Scanned: %d\n\n", payload.Repositories); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(w, "Authored Pull Requests"); err != nil {
					return err
				}
				if err := writeCrossRepoPRTable(w, payload.AuthoredPRs); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(w, "\nReview Requested"); err != nil {
					return err
				}
				if err := writeCrossRepoPRTable(w, payload.ReviewRequestedPRs); err != nil {
					return err
				}

				if _, err := fmt.Fprintln(w, "\nYour Issues"); err != nil {
					return err
				}
				if err := writeCrossRepoIssueTable(w, payload.YourIssues); err != nil {
					return err
				}

				if len(payload.Warnings) == 0 {
					return nil
				}

				if _, err := fmt.Fprintln(w, "\nNotes"); err != nil {
					return err
				}
				for _, warning := range payload.Warnings {
					if _, err := fmt.Fprintf(w, "- %s\n", warning); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Limit status aggregation to one workspace")
	cmd.Flags().IntVar(&repoLimit, "repo-limit", 100, "Maximum repositories to scan per workspace")
	cmd.Flags().IntVar(&itemLimit, "limit", 20, "Maximum items to return per status section")

	return cmd
}

func resolveStatusWorkspaces(ctx context.Context, client *bitbucket.Client, selected string) ([]string, error) {
	if selected != "" {
		return []string{selected}, nil
	}

	workspaces, err := client.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	slugs := make([]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		if workspace.Slug != "" {
			slugs = append(slugs, workspace.Slug)
		}
	}
	slices.Sort(slugs)
	return slugs, nil
}

func buildCrossRepoStatus(ctx context.Context, client crossRepoStatusClient, currentUser bitbucket.CurrentUser, workspaces []string, repoLimit, itemLimit int) (crossRepoStatusPayload, error) {
	payload := crossRepoStatusPayload{
		User:                  coalesce(currentUser.DisplayName, currentUser.Username, currentUser.AccountID),
		Workspaces:            append([]string(nil), workspaces...),
		RepoLimitPerWorkspace: repoLimit,
		ItemLimitPerSection:   itemLimit,
	}

	for _, workspace := range workspaces {
		repos, err := client.ListRepositories(ctx, workspace, bitbucket.ListRepositoriesOptions{
			Sort:  "-updated_on",
			Limit: repoLimit,
		})
		if err != nil {
			return crossRepoStatusPayload{}, err
		}
		if repoLimit > 0 && len(repos) == repoLimit {
			payload.WorkspacesAtRepoLimit = append(payload.WorkspacesAtRepoLimit, workspace)
		}

		results, err := scanWorkspaceRepositories(ctx, client, currentUser, workspace, repos, itemLimit)
		if err != nil {
			return crossRepoStatusPayload{}, err
		}

		for _, result := range results {
			payload.Repositories += result.repository
			payload.RepositoriesWithoutIssueTracker += boolToInt(result.issueTrackerDisabled)
			payload.AuthoredPRs = append(payload.AuthoredPRs, result.authoredPRs...)
			payload.ReviewRequestedPRs = append(payload.ReviewRequestedPRs, result.reviewRequestedPRs...)
			payload.YourIssues = append(payload.YourIssues, result.yourIssues...)
		}
	}

	sortCrossRepoStatus(&payload)
	payload.AuthoredPRsTotal = len(payload.AuthoredPRs)
	payload.ReviewRequestedPRsTotal = len(payload.ReviewRequestedPRs)
	payload.YourIssuesTotal = len(payload.YourIssues)
	payload.AuthoredPRs = limitCrossRepoPRs(payload.AuthoredPRs, itemLimit)
	payload.ReviewRequestedPRs = limitCrossRepoPRs(payload.ReviewRequestedPRs, itemLimit)
	payload.YourIssues = limitCrossRepoIssues(payload.YourIssues, itemLimit)
	payload.Warnings = buildCrossRepoStatusWarnings(payload)

	return payload, nil
}

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

func issueInvolvesUser(user bitbucket.CurrentUser, issue bitbucket.Issue) bool {
	return sameIssueActor(user, issue.Reporter) || sameIssueActor(user, issue.Assignee)
}

func sameIssueActor(user bitbucket.CurrentUser, actor bitbucket.IssueActor) bool {
	switch {
	case user.AccountID != "" && actor.AccountID != "":
		return user.AccountID == actor.AccountID
	case user.Username != "" && actor.Nickname != "":
		return user.Username == actor.Nickname
	case user.DisplayName != "" && actor.DisplayName != "":
		return user.DisplayName == actor.DisplayName
	default:
		return false
	}
}

func issueNeedsAttention(issue bitbucket.Issue) bool {
	switch strings.ToLower(strings.TrimSpace(issue.State)) {
	case "resolved", "closed", "invalid", "duplicate", "wontfix":
		return false
	default:
		return true
	}
}

func isNoIssueTrackerError(err error) bool {
	apiErr, ok := bitbucket.AsAPIError(err)
	if !ok {
		return false
	}
	return strings.Contains(strings.ToLower(apiErrorDetail(apiErr)), "no issue tracker")
}
