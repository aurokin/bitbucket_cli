package cmd

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type crossRepoStatusPayload struct {
	User               string                 `json:"user,omitempty"`
	Workspaces         []string               `json:"workspaces"`
	Repositories       int                    `json:"repositories_scanned"`
	AuthoredPRs        []crossRepoPullRequest `json:"authored_prs"`
	ReviewRequestedPRs []crossRepoPullRequest `json:"review_requested_prs"`
	YourIssues         []crossRepoIssue       `json:"your_issues"`
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
			"  bb status --workspace OhBizzle\n" +
			"  bb status --json authored_prs,review_requested_prs,your_issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(host)
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

			payload, err := buildCrossRepoStatus(context.Background(), client, currentUser, resolvedHost, workspaces, repoLimit, itemLimit)
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
				return writeCrossRepoIssueTable(w, payload.YourIssues)
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
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

func buildCrossRepoStatus(ctx context.Context, client *bitbucket.Client, currentUser bitbucket.CurrentUser, host string, workspaces []string, repoLimit, itemLimit int) (crossRepoStatusPayload, error) {
	payload := crossRepoStatusPayload{
		User:       coalesce(currentUser.DisplayName, currentUser.Username, currentUser.AccountID),
		Workspaces: append([]string(nil), workspaces...),
	}

	for _, workspace := range workspaces {
		repos, err := client.ListRepositories(ctx, workspace, bitbucket.ListRepositoriesOptions{
			Sort:  "-updated_on",
			Limit: repoLimit,
		})
		if err != nil {
			return crossRepoStatusPayload{}, err
		}

		for _, repo := range repos {
			payload.Repositories++

			prs, err := client.ListPullRequests(ctx, workspace, repo.Slug, bitbucket.ListPullRequestsOptions{
				State: "OPEN",
				Limit: itemLimit,
				Sort:  "-updated_on",
			})
			if err != nil {
				return crossRepoStatusPayload{}, err
			}

			for _, pr := range prs {
				if sameActor(currentUser, pr.Author) {
					payload.AuthoredPRs = append(payload.AuthoredPRs, crossRepoPullRequest{
						Workspace:   workspace,
						Repo:        repo.Slug,
						PullRequest: pr,
					})
				}
				if reviewRequestedFromUser(currentUser, pr) {
					payload.ReviewRequestedPRs = append(payload.ReviewRequestedPRs, crossRepoPullRequest{
						Workspace:   workspace,
						Repo:        repo.Slug,
						PullRequest: pr,
					})
				}
			}

			issues, err := client.ListIssues(ctx, workspace, repo.Slug, bitbucket.ListIssuesOptions{
				Limit: itemLimit,
				Sort:  "-updated_on",
			})
			if err != nil {
				if isNoIssueTrackerError(err) {
					continue
				}
				return crossRepoStatusPayload{}, err
			}

			for _, issue := range issues {
				if !issueNeedsAttention(issue) {
					continue
				}
				if !issueInvolvesUser(currentUser, issue) {
					continue
				}
				payload.YourIssues = append(payload.YourIssues, crossRepoIssue{
					Workspace: workspace,
					Repo:      repo.Slug,
					Issue:     issue,
				})
			}
		}
	}

	sortCrossRepoStatus(&payload)
	payload.AuthoredPRs = limitCrossRepoPRs(payload.AuthoredPRs, itemLimit)
	payload.ReviewRequestedPRs = limitCrossRepoPRs(payload.ReviewRequestedPRs, itemLimit)
	payload.YourIssues = limitCrossRepoIssues(payload.YourIssues, itemLimit)

	return payload, nil
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

func writeCrossRepoPRTable(w io.Writer, prs []crossRepoPullRequest) error {
	if len(prs) == 0 {
		_, err := fmt.Fprintln(w, "none")
		return err
	}

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "repo\t#\ttitle\tstate\tsrc\tdst\tupdated"); err != nil {
		return err
	}
	for _, item := range prs {
		if _, err := fmt.Fprintf(
			tw,
			"%s/%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			item.Workspace,
			item.Repo,
			item.PullRequest.ID,
			output.Truncate(item.PullRequest.Title, 32),
			output.Truncate(item.PullRequest.State, 12),
			output.TruncateMiddle(item.PullRequest.Source.Branch.Name, 12),
			output.TruncateMiddle(item.PullRequest.Destination.Branch.Name, 12),
			formatPRUpdated(item.PullRequest.UpdatedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writeCrossRepoIssueTable(w io.Writer, issues []crossRepoIssue) error {
	if len(issues) == 0 {
		_, err := fmt.Fprintln(w, "none")
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
