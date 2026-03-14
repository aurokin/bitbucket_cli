package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newIssueCmd() *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Work with repository issues",
		Long:  "List, view, create, edit, close, and reopen Bitbucket Cloud repository issues, and manage issue comments, attachments, milestones, and components.",
	}

	issueCmd.AddCommand(
		newIssueAttachmentCmd(),
		newIssueCommentCmd(),
		newIssueComponentCmd(),
		newIssueListCmd(),
		newIssueMilestoneCmd(),
		newIssueViewCmd(),
		newIssueCreateCmd(),
		newIssueEditCmd(),
		newIssueCloseCmd(),
		newIssueReopenCmd(),
	)

	return issueCmd
}

func newIssueListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var state string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues for a repository",
		Example: "  bb issue list --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue list --repo workspace-slug/repo-slug\n" +
			"  bb issue list --state open --json id,title,state",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			target, client, err := resolveIssueTarget(host, workspace, repo)
			if err != nil {
				return err
			}

			issues, err := client.ListIssues(context.Background(), target.Workspace, target.Repo, bitbucket.ListIssuesOptions{
				State: state,
				Sort:  "-updated_on",
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issues, func(w io.Writer) error {
				return writeIssueListSummary(w, target, issues)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&state, "state", "ALL", "Filter issues by state")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of issues to return")

	return cmd
}

func newIssueViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View one issue",
		Example: "  bb issue view 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue view 1 --repo workspace-slug/repo-slug --json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			target, client, issueID, err := resolveIssueTargetAndID(host, workspace, repo, args[0])
			if err != nil {
				return err
			}

			issue, err := client.GetIssue(context.Background(), target.Workspace, target.Repo, issueID)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issue, func(w io.Writer) error {
				return writeIssueViewSummary(w, target, issue)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newIssueCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var title string
	var body string
	var kind string
	var priority string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue",
		Example: "  bb issue create --repo workspace-slug/issues-repo-slug --title 'Broken flow'\n" +
			"  bb issue create --repo workspace-slug/repo-slug --title 'Broken flow' --body 'Needs investigation'\n" +
			"  bb issue create --title 'Request' --kind proposal --priority major --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			target, client, err := resolveIssueTarget(host, workspace, repo)
			if err != nil {
				return err
			}

			if title == "" && promptsEnabled(cmd) {
				title, err = promptRequiredString(cmd, "Title", "")
				if err != nil {
					return err
				}
			}
			if title == "" {
				return fmt.Errorf("issue title is required; pass --title or run in an interactive terminal")
			}

			issue, err := client.CreateIssue(context.Background(), target.Workspace, target.Repo, bitbucket.CreateIssueOptions{
				Title:    title,
				Body:     body,
				Kind:     kind,
				Priority: priority,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issue, func(w io.Writer) error {
				return writeIssueMutationSummary(w, "Created", target.Workspace, target.Repo, issue, true)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&title, "title", "", "Issue title")
	cmd.Flags().StringVar(&body, "body", "", "Issue body text")
	cmd.Flags().StringVar(&kind, "kind", "", "Issue kind")
	cmd.Flags().StringVar(&priority, "priority", "", "Issue priority")

	return cmd
}

func newIssueEditCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var title string
	var body string
	var state string
	var kind string
	var priority string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an issue",
		Example: "  bb issue edit 1 --repo workspace-slug/issues-repo-slug --title 'Updated title'\n" +
			"  bb issue edit 1 --repo workspace-slug/repo-slug --state open --priority major --json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			target, client, issueID, err := resolveIssueTargetAndID(host, workspace, repo, args[0])
			if err != nil {
				return err
			}

			issue, err := client.UpdateIssue(context.Background(), target.Workspace, target.Repo, issueID, bitbucket.UpdateIssueOptions{
				Title:    title,
				Body:     body,
				State:    state,
				Kind:     kind,
				Priority: priority,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issue, func(w io.Writer) error {
				return writeIssueMutationSummary(w, "Updated", target.Workspace, target.Repo, issue, false)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&title, "title", "", "Issue title")
	cmd.Flags().StringVar(&body, "body", "", "Issue body text")
	cmd.Flags().StringVar(&state, "state", "", "Issue state")
	cmd.Flags().StringVar(&kind, "kind", "", "Issue kind")
	cmd.Flags().StringVar(&priority, "priority", "", "Issue priority")

	return cmd
}

func newIssueCloseCmd() *cobra.Command {
	return newIssueStateTransitionCmd("close", "resolved", "Close an issue", "Resolve an issue by moving it to the resolved state.")
}

func newIssueReopenCmd() *cobra.Command {
	return newIssueStateTransitionCmd("reopen", "new", "Reopen an issue", "Reopen an issue by moving it back to the new state.")
}

func newIssueStateTransitionCmd(use, defaultState, short, long string) *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var state string
	var message string

	cmd := &cobra.Command{
		Use:   use + " <id>",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			target, client, issueID, err := resolveIssueTargetAndID(host, workspace, repo, args[0])
			if err != nil {
				return err
			}

			newState := state
			if newState == "" {
				newState = defaultState
			}

			if err := client.ChangeIssueState(context.Background(), target.Workspace, target.Repo, issueID, bitbucket.IssueChangeOptions{
				State:   newState,
				Message: message,
			}); err != nil {
				return err
			}

			issue, err := client.GetIssue(context.Background(), target.Workspace, target.Repo, issueID)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, issue, func(w io.Writer) error {
				return writeIssueMutationSummary(w, "Updated", target.Workspace, target.Repo, issue, false)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&state, "state", "", "Target issue state")
	cmd.Flags().StringVar(&message, "message", "", "Optional issue change message")

	return cmd
}
