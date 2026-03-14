package cmd

import (
	"context"
	"io"
	"slices"
	"strings"

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
				return writeCrossRepoStatusSummary(w, payload)
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
