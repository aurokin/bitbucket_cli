package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newIssueCmd() *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Work with repository issues",
		Long:  "List, view, create, edit, close, and reopen Bitbucket Cloud repository issues.",
	}

	issueCmd.AddCommand(
		newIssueListCmd(),
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
				if len(issues) == 0 {
					if _, err := fmt.Fprintf(w, "No issues found for %s/%s.\n", target.Workspace, target.Repo); err != nil {
						return err
					}
					return writeNextStep(w, issueListEmptyNextStep(target.Workspace, target.Repo))
				}
				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				return writeIssueTable(w, issues)
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
				if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", issue.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", issue.Title); err != nil {
					return err
				}
				if issue.State != "" {
					if _, err := fmt.Fprintf(tw, "State:\t%s\n", issue.State); err != nil {
						return err
					}
				}
				if issue.Kind != "" {
					if _, err := fmt.Fprintf(tw, "Kind:\t%s\n", issue.Kind); err != nil {
						return err
					}
				}
				if issue.Priority != "" {
					if _, err := fmt.Fprintf(tw, "Priority:\t%s\n", issue.Priority); err != nil {
						return err
					}
				}
				if issue.Reporter.DisplayName != "" {
					if _, err := fmt.Fprintf(tw, "Reporter:\t%s\n", issue.Reporter.DisplayName); err != nil {
						return err
					}
				}
				if issue.Assignee.DisplayName != "" {
					if _, err := fmt.Fprintf(tw, "Assignee:\t%s\n", issue.Assignee.DisplayName); err != nil {
						return err
					}
				}
				if issue.UpdatedOn != "" {
					if _, err := fmt.Fprintf(tw, "Updated:\t%s\n", issue.UpdatedOn); err != nil {
						return err
					}
				}
				if issue.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", issue.Links.HTML.Href); err != nil {
						return err
					}
				}
				if issue.Content.Raw != "" {
					if _, err := fmt.Fprintf(tw, "Body:\t%s\n", issue.Content.Raw); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, issueViewNextStep(target.Workspace, target.Repo, issue.ID))
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

func resolveIssueTarget(host, workspace, repo string) (resolvedRepoTarget, *bitbucket.Client, error) {
	resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
	if err != nil {
		return resolvedRepoTarget{}, nil, err
	}

	return resolved.Target, resolved.Client, nil
}

func resolveIssueTargetAndID(host, workspace, repo, rawID string) (resolvedRepoTarget, *bitbucket.Client, int, error) {
	target, client, err := resolveIssueTarget(host, workspace, repo)
	if err != nil {
		return resolvedRepoTarget{}, nil, 0, err
	}

	issueID, err := parsePositiveInt("issue", rawID)
	if err != nil {
		return resolvedRepoTarget{}, nil, 0, err
	}

	return target, client, issueID, nil
}

func writeIssueMutationSummary(w io.Writer, action, workspace, repo string, issue bitbucket.Issue, includeNext bool) error {
	if _, err := fmt.Fprintf(w, "%s issue %s/%s#%d: %s\n", action, workspace, repo, issue.ID, issue.Title); err != nil {
		return err
	}
	if issue.State != "" {
		if _, err := fmt.Fprintf(w, "State: %s\n", issue.State); err != nil {
			return err
		}
	}
	if issue.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(w, "URL: %s\n", issue.Links.HTML.Href); err != nil {
			return err
		}
	}
	if includeNext {
		if _, err := fmt.Fprintf(w, "Next: bb issue view %d --repo %s/%s\n", issue.ID, workspace, repo); err != nil {
			return err
		}
	}
	return nil
}

func issueListEmptyNextStep(workspace, repo string) string {
	return fmt.Sprintf("bb issue create --repo %s/%s --title '<title>'", workspace, repo)
}

func issueViewNextStep(workspace, repo string, id int) string {
	return fmt.Sprintf("bb issue edit %d --repo %s/%s", id, workspace, repo)
}

func writeIssueTable(w io.Writer, issues []bitbucket.Issue) error {
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "#\ttitle\tstate\treporter\tupdated"); err != nil {
		return err
	}
	for _, issue := range issues {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\n",
			issue.ID,
			output.Truncate(issue.Title, 40),
			output.Truncate(issue.State, 12),
			output.Truncate(issue.Reporter.DisplayName, 16),
			formatPRUpdated(issue.UpdatedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func parsePositiveInt(label, raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s ID must be a positive integer", label)
	}
	return value, nil
}
