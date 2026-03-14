package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type projectDefaultReviewerListPayload struct {
	Host             string                      `json:"host"`
	Workspace        string                      `json:"workspace"`
	ProjectKey       string                      `json:"project_key"`
	DefaultReviewers []bitbucket.DefaultReviewer `json:"default_reviewers"`
}

func newProjectDefaultReviewerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "default-reviewer",
		Aliases: []string{"default-reviewers"},
		Short:   "Inspect project default reviewers",
	}
	cmd.AddCommand(newProjectDefaultReviewerListCmd())
	return cmd
}

func newProjectDefaultReviewerListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace string
	var limit int

	cmd := &cobra.Command{
		Use:   "list <project-key>",
		Short: "List default reviewers in a project",
		Example: "  bb project default-reviewer list BBCLI --workspace workspace-slug\n" +
			"  bb project default-reviewer list BBCLI --workspace workspace-slug --json default_reviewers",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			selectedWorkspace, resolvedHost, client, err := resolveWorkspaceCommandTarget(host, workspace, "")
			if err != nil {
				return err
			}
			projectKey := strings.TrimSpace(args[0])
			reviewers, err := client.ListProjectDefaultReviewers(context.Background(), selectedWorkspace, projectKey, limit)
			if err != nil {
				return err
			}
			payload := projectDefaultReviewerListPayload{
				Host:             resolvedHost,
				Workspace:        selectedWorkspace,
				ProjectKey:       projectKey,
				DefaultReviewers: reviewers,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeProjectDefaultReviewerListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to inspect")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of default reviewers to return")
	return cmd
}

func writeProjectDefaultReviewerListSummary(w io.Writer, payload projectDefaultReviewerListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.ProjectKey); err != nil {
		return err
	}
	if len(payload.DefaultReviewers) == 0 {
		if _, err := fmt.Fprintf(w, "No default reviewers found for %s/%s.\n", payload.Workspace, payload.ProjectKey); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "account-id\tuser\treviewer-type"); err != nil {
		return err
	}
	for _, reviewer := range payload.DefaultReviewers {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(reviewer.User.AccountID, 28),
			output.Truncate(reviewer.User.DisplayName, 24),
			reviewer.ReviewerType,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project permissions user list %s --workspace %s", payload.ProjectKey, payload.Workspace))
}
