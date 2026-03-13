package cmd

import (
	"context"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type prReviewPayload struct {
	Host        string                           `json:"host"`
	Workspace   string                           `json:"workspace"`
	Repo        string                           `json:"repo"`
	PullRequest int                              `json:"pull_request"`
	Action      string                           `json:"action"`
	Reviewer    bitbucket.PullRequestActor       `json:"reviewer"`
	ReviewState string                           `json:"review_state,omitempty"`
	Participant *bitbucket.PullRequestParticipant `json:"participant,omitempty"`
}

type prActivityPayload struct {
	Host        string                      `json:"host"`
	Workspace   string                      `json:"workspace"`
	Repo        string                      `json:"repo"`
	Warnings    []string                    `json:"warnings,omitempty"`
	PullRequest int                         `json:"pull_request"`
	Activity    []bitbucket.PullRequestActivity `json:"activity"`
}

type prCommitsPayload struct {
	Host        string                       `json:"host"`
	Workspace   string                       `json:"workspace"`
	Repo        string                       `json:"repo"`
	Warnings    []string                     `json:"warnings,omitempty"`
	PullRequest int                          `json:"pull_request"`
	Commits     []bitbucket.RepositoryCommit `json:"commits"`
}

type prChecksPayload struct {
	Host        string                    `json:"host"`
	Workspace   string                    `json:"workspace"`
	Repo        string                    `json:"repo"`
	Warnings    []string                  `json:"warnings,omitempty"`
	PullRequest int                       `json:"pull_request"`
	Statuses    []bitbucket.CommitStatus  `json:"statuses"`
}

func newPRReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review a pull request",
		Long:  "Review a pull request using Bitbucket Cloud review actions such as approve, request-changes, and clearing your own prior review state.",
		Example: "  bb pr review approve 7 --repo workspace-slug/repo-slug\n" +
			"  bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'\n" +
			"  bb pr review clear-request-changes 7 --repo workspace-slug/repo-slug",
	}

	cmd.AddCommand(
		newPRReviewActionCmd("approve", "Approve a pull request", bitbucket.PullRequestReviewApprove),
		newPRReviewActionCmd("unapprove", "Withdraw your approval of a pull request", bitbucket.PullRequestReviewUnapprove),
		newPRReviewActionCmd("request-changes", "Request changes on a pull request", bitbucket.PullRequestReviewRequestChanges),
		newPRReviewActionCmd("clear-request-changes", "Clear your prior request for changes", bitbucket.PullRequestReviewClearRequestChanges),
	)

	return cmd
}

func newPRReviewActionCmd(use, short string, action bitbucket.PullRequestReviewAction) *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   use + " <id-or-url>",
		Short: short,
		Long:  reviewActionLongDescription(action),
		Example: reviewActionExample(action),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			participant, err := resolved.Client.ReviewPullRequest(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.ID, action)
			if err != nil {
				return err
			}

			reviewer := participant.User
			if reviewer == (bitbucket.PullRequestActor{}) {
				currentUser, err := resolved.Client.CurrentUser(context.Background())
				if err != nil {
					return err
				}
				reviewer = bitbucket.PullRequestActor{
					DisplayName: currentUser.DisplayName,
					Nickname:    currentUser.Username,
					AccountID:   currentUser.AccountID,
				}
			}

			payload := prReviewPayload{
				Host:        resolved.Target.RepoTarget.Host,
				Workspace:   resolved.Target.RepoTarget.Workspace,
				Repo:        resolved.Target.RepoTarget.Repo,
				PullRequest: resolved.Target.ID,
				Action:      string(action),
				Reviewer:    reviewer,
				ReviewState: reviewStateForAction(action, participant),
			}
			if participant != (bitbucket.PullRequestParticipant{}) {
				participantCopy := participant
				payload.Participant = &participantCopy
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePRReviewSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newPRActivityCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "activity <id-or-url>",
		Short: "Show recent pull request activity",
		Long:  "Show recent pull request activity including comments, updates, approvals, and change requests. Accepts a numeric ID, pull request URL, or pull request comment URL.",
		Example: "  bb pr activity 7 --repo workspace-slug/repo-slug\n" +
			"  bb pr activity https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --limit 50 --json '*'\n" +
			"  bb pr activity https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			activity, err := resolved.Client.ListPullRequestActivity(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.ID, bitbucket.ListPullRequestActivityOptions{
				Limit: limit,
			})
			if err != nil {
				return err
			}

			payload := prActivityPayload{
				Host:        resolved.Target.RepoTarget.Host,
				Workspace:   resolved.Target.RepoTarget.Workspace,
				Repo:        resolved.Target.RepoTarget.Repo,
				Warnings:    append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				PullRequest: resolved.Target.ID,
				Activity:    activity,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePRActivitySummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of activity entries to return")

	return cmd
}

func newPRCommitsCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "commits <id-or-url>",
		Short: "List commits on a pull request",
		Long:  "List the commits that would be merged by the pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.",
		Example: "  bb pr commits 7 --repo workspace-slug/repo-slug\n" +
			"  bb pr commits https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'\n" +
			"  bb pr commits https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --limit 50",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			commits, err := resolved.Client.ListPullRequestCommits(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.ID, bitbucket.ListPullRequestCommitsOptions{
				Limit: limit,
			})
			if err != nil {
				return err
			}

			payload := prCommitsPayload{
				Host:        resolved.Target.RepoTarget.Host,
				Workspace:   resolved.Target.RepoTarget.Workspace,
				Repo:        resolved.Target.RepoTarget.Repo,
				Warnings:    append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				PullRequest: resolved.Target.ID,
				Commits:     commits,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePRCommitsSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of commits to return")

	return cmd
}

func newPRChecksCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:     "checks <id-or-url>",
		Aliases: []string{"statuses"},
		Short:   "Show commit statuses for a pull request",
		Long:    "Show commit statuses for a pull request. This is the Bitbucket Cloud equivalent of PR checks backed by commit statuses. Accepts a numeric ID, pull request URL, or pull request comment URL.",
		Example: "  bb pr checks 7 --repo workspace-slug/repo-slug\n" +
			"  bb pr checks https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'\n" +
			"  bb pr statuses https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolvePullRequestCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			statuses, err := resolved.Client.ListPullRequestStatuses(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.ID, bitbucket.ListPullRequestStatusesOptions{
				Limit: limit,
				Sort:  "-updated_on",
			})
			if err != nil {
				return err
			}

			payload := prChecksPayload{
				Host:        resolved.Target.RepoTarget.Host,
				Workspace:   resolved.Target.RepoTarget.Workspace,
				Repo:        resolved.Target.RepoTarget.Repo,
				Warnings:    append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				PullRequest: resolved.Target.ID,
				Statuses:    statuses,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePRChecksSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of commit statuses to return")

	return cmd
}

func reviewActionLongDescription(action bitbucket.PullRequestReviewAction) string {
	switch action {
	case bitbucket.PullRequestReviewApprove:
		return "Approve a pull request as the authenticated user. Accepts a numeric ID, pull request URL, or pull request comment URL."
	case bitbucket.PullRequestReviewUnapprove:
		return "Withdraw your own approval of a pull request. Accepts a numeric ID, pull request URL, or pull request comment URL."
	case bitbucket.PullRequestReviewRequestChanges:
		return "Request changes on a pull request as the authenticated user. Accepts a numeric ID, pull request URL, or pull request comment URL."
	case bitbucket.PullRequestReviewClearRequestChanges:
		return "Clear your own prior request for changes on a pull request. Accepts a numeric ID, pull request URL, or pull request comment URL."
	default:
		return "Review a pull request."
	}
}

func reviewActionExample(action bitbucket.PullRequestReviewAction) string {
	switch action {
	case bitbucket.PullRequestReviewApprove:
		return "  bb pr review approve 7 --repo workspace-slug/repo-slug\n  bb pr review approve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'\n  bb pr review approve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7"
	case bitbucket.PullRequestReviewUnapprove:
		return "  bb pr review unapprove 7 --repo workspace-slug/repo-slug\n  bb pr review unapprove https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'\n  bb pr review unapprove https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7"
	case bitbucket.PullRequestReviewRequestChanges:
		return "  bb pr review request-changes 7 --repo workspace-slug/repo-slug\n  bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'\n  bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7"
	case bitbucket.PullRequestReviewClearRequestChanges:
		return "  bb pr review clear-request-changes 7 --repo workspace-slug/repo-slug\n  bb pr review clear-request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'\n  bb pr review clear-request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7"
	default:
		return ""
	}
}

func reviewStateForAction(action bitbucket.PullRequestReviewAction, participant bitbucket.PullRequestParticipant) string {
	if participant.State != "" {
		return participant.State
	}
	if participant.Approved {
		return "approved"
	}

	switch action {
	case bitbucket.PullRequestReviewApprove:
		return "approved"
	case bitbucket.PullRequestReviewUnapprove:
		return "unapproved"
	case bitbucket.PullRequestReviewRequestChanges:
		return "changes_requested"
	case bitbucket.PullRequestReviewClearRequestChanges:
		return "changes_request_cleared"
	default:
		return ""
	}
}
