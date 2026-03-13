package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type issueAttachmentListPayload struct {
	Host        string                      `json:"host"`
	Workspace   string                      `json:"workspace"`
	Repo        string                      `json:"repo"`
	Issue       int                         `json:"issue"`
	Attachments []bitbucket.IssueAttachment `json:"attachments"`
}

type issueAttachmentUploadPayload struct {
	Host        string                      `json:"host"`
	Workspace   string                      `json:"workspace"`
	Repo        string                      `json:"repo"`
	Issue       int                         `json:"issue"`
	Action      string                      `json:"action,omitempty"`
	Files       []string                    `json:"files"`
	Attachments []bitbucket.IssueAttachment `json:"attachments"`
}

func newIssueAttachmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Work with issue attachments",
		Long:  "List and upload Bitbucket issue attachments. Attachment import and export jobs remain separate platform workflows.",
		Example: "  bb issue attachment list 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug",
	}
	cmd.AddCommand(newIssueAttachmentListCmd(), newIssueAttachmentUploadCmd())
	return cmd
}

func newIssueAttachmentListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list <issue-id-or-url>",
		Short: "List attachments on an issue",
		Example: "  bb issue attachment list 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue attachment list https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, issueID, err := resolveIssueReference(host, workspace, repo, args[0])
			if err != nil {
				return err
			}
			items, err := client.ListIssueAttachments(context.Background(), target.Workspace, target.Repo, issueID, limit)
			if err != nil {
				return err
			}
			payload := issueAttachmentListPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Issue: issueID, Attachments: items}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueAttachmentListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of issue attachments to return")
	return cmd
}

func newIssueAttachmentUploadCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "upload <issue-id-or-url> <file>...",
		Short: "Upload attachments to an issue",
		Long:  "Upload one or more files to a Bitbucket issue. Existing attachments with the same name are replaced by Bitbucket Cloud.",
		Example: "  bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue attachment upload https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 ./trace.txt ./screenshot.png --json '*'",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, issueID, err := resolveIssueReference(host, workspace, repo, args[0])
			if err != nil {
				return err
			}
			files := append([]string(nil), args[1:]...)
			if err := client.UploadIssueAttachments(context.Background(), target.Workspace, target.Repo, issueID, files); err != nil {
				return err
			}
			attachments, err := client.ListIssueAttachments(context.Background(), target.Workspace, target.Repo, issueID, 200)
			if err != nil {
				return err
			}
			filtered := filterIssueAttachmentsByName(attachments, files)
			payload := issueAttachmentUploadPayload{
				Host:        target.Host,
				Workspace:   target.Workspace,
				Repo:        target.Repo,
				Issue:       issueID,
				Action:      "uploaded",
				Files:       baseNames(files),
				Attachments: filtered,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueAttachmentUploadSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func filterIssueAttachmentsByName(attachments []bitbucket.IssueAttachment, paths []string) []bitbucket.IssueAttachment {
	names := map[string]struct{}{}
	for _, path := range paths {
		names[filepath.Base(path)] = struct{}{}
	}
	filtered := make([]bitbucket.IssueAttachment, 0, len(paths))
	for _, attachment := range attachments {
		if _, ok := names[attachment.Name]; ok {
			filtered = append(filtered, attachment)
		}
	}
	return filtered
}

func baseNames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, filepath.Base(path))
	}
	return names
}

func writeIssueAttachmentListSummary(w io.Writer, payload issueAttachmentListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Issue", fmt.Sprintf("%d", payload.Issue)); err != nil {
		return err
	}
	if len(payload.Attachments) == 0 {
		if _, err := fmt.Fprintf(w, "No attachments found on issue %s/%s#%d.\n", payload.Workspace, payload.Repo, payload.Issue); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\turl"); err != nil {
		return err
	}
	for _, attachment := range payload.Attachments {
		link := attachment.Links.HTML.Href
		if attachment.Links.Self.Href != "" {
			link = attachment.Links.Self.Href
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", output.Truncate(attachment.Name, 32), output.Truncate(link, 60)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue view %d --repo %s/%s", payload.Issue, payload.Workspace, payload.Repo))
}

func writeIssueAttachmentUploadSummary(w io.Writer, payload issueAttachmentUploadPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Issue", fmt.Sprintf("%d", payload.Issue)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Files", strings.Join(payload.Files, ", ")); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue attachment list %d --repo %s/%s", payload.Issue, payload.Workspace, payload.Repo))
}
