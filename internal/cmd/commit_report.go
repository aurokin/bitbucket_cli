package cmd

import (
	"context"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type commitReportListPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Warnings  []string                 `json:"warnings,omitempty"`
	Commit    string                   `json:"commit"`
	Reports   []bitbucket.CommitReport `json:"reports"`
}

type commitReportPayload struct {
	Host      string                 `json:"host"`
	Workspace string                 `json:"workspace"`
	Repo      string                 `json:"repo"`
	Warnings  []string               `json:"warnings,omitempty"`
	Commit    string                 `json:"commit"`
	Report    bitbucket.CommitReport `json:"report"`
}

func newCommitReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Work with commit code-insight reports",
		Long:  "List and inspect Bitbucket code-insight reports attached to one commit.",
	}
	cmd.AddCommand(newCommitReportListCmd(), newCommitReportViewCmd())
	return cmd
}

func newCommitReportListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list <hash-or-url>",
		Short: "List code-insight reports for a commit",
		Long:  "List Bitbucket code-insight reports for one commit. Accepts a commit SHA or a commit URL.",
		Example: "  bb commit report list abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit report list https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json reports\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommandTarget(context.Background(), host, workspace, repo, args[0], true)
			if err != nil {
				return err
			}

			reports, err := resolved.Client.ListCommitReports(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit, bitbucket.ListCommitReportsOptions{Limit: limit})
			if err != nil {
				return err
			}

			payload := commitReportListPayload{
				Host:      resolved.Target.RepoTarget.Host,
				Workspace: resolved.Target.RepoTarget.Workspace,
				Repo:      resolved.Target.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				Commit:    resolved.Target.Commit,
				Reports:   reports,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitReportListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of commit reports to return")
	return cmd
}

func newCommitReportViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, commitRef string

	cmd := &cobra.Command{
		Use:   "view <report-id>",
		Short: "View one code-insight report",
		Long:  "View one Bitbucket code-insight report. Pass the commit with --commit as a SHA or commit URL.",
		Example: "  bb commit report view my-report --commit abc1234 --repo workspace-slug/repo-slug\n" +
			"  bb commit report view my-report --commit https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveCommitCommandTarget(context.Background(), host, workspace, repo, commitRef, true)
			if err != nil {
				return err
			}

			report, err := resolved.Client.GetCommitReport(context.Background(), resolved.Target.RepoTarget.Workspace, resolved.Target.RepoTarget.Repo, resolved.Target.Commit, args[0])
			if err != nil {
				return err
			}

			payload := commitReportPayload{
				Host:      resolved.Target.RepoTarget.Host,
				Workspace: resolved.Target.RepoTarget.Workspace,
				Repo:      resolved.Target.RepoTarget.Repo,
				Warnings:  append([]string(nil), resolved.Target.RepoTarget.Warnings...),
				Commit:    resolved.Target.Commit,
				Report:    report,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeCommitReportSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&commitRef, "commit", "", "Commit SHA or commit URL")
	_ = cmd.MarkFlagRequired("commit")
	return cmd
}
