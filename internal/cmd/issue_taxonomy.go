package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type issueMilestoneListPayload struct {
	Host       string                     `json:"host"`
	Workspace  string                     `json:"workspace"`
	Repo       string                     `json:"repo"`
	Milestones []bitbucket.IssueMilestone `json:"milestones"`
}

type issueMilestonePayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Milestone bitbucket.IssueMilestone `json:"milestone"`
}

type issueComponentListPayload struct {
	Host       string                     `json:"host"`
	Workspace  string                     `json:"workspace"`
	Repo       string                     `json:"repo"`
	Components []bitbucket.IssueComponent `json:"components"`
}

type issueComponentPayload struct {
	Host      string                   `json:"host"`
	Workspace string                   `json:"workspace"`
	Repo      string                   `json:"repo"`
	Component bitbucket.IssueComponent `json:"component"`
}

func newIssueMilestoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone",
		Short: "List and view issue milestones",
		Long:  "List and view Bitbucket issue tracker milestones. Bitbucket Cloud only exposes milestone read APIs in the official REST surface.",
	}
	cmd.AddCommand(newIssueMilestoneListCmd(), newIssueMilestoneViewCmd())
	return cmd
}

func newIssueComponentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "component",
		Short: "List and view issue components",
		Long:  "List and view Bitbucket issue tracker components. Bitbucket Cloud only exposes component read APIs in the official REST surface.",
	}
	cmd.AddCommand(newIssueComponentListCmd(), newIssueComponentViewCmd())
	return cmd
}

func newIssueMilestoneListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue milestones for a repository",
		Example: "  bb issue milestone list --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue milestone list --repo workspace-slug/issues-repo-slug --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, err := resolveIssueTarget(host, workspace, repo)
			if err != nil {
				return err
			}
			items, err := client.ListIssueMilestones(context.Background(), target.Workspace, target.Repo, limit)
			if err != nil {
				return err
			}
			payload := issueMilestoneListPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Milestones: items}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueMilestoneListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of milestones to return")
	return cmd
}

func newIssueMilestoneViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View one issue milestone",
		Example: "  bb issue milestone view 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue milestone view 1 --repo workspace-slug/issues-repo-slug --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, err := resolveIssueTarget(host, workspace, repo)
			if err != nil {
				return err
			}
			id, err := parsePositiveInt("issue milestone", args[0])
			if err != nil {
				return err
			}
			item, err := client.GetIssueMilestone(context.Background(), target.Workspace, target.Repo, id)
			if err != nil {
				return err
			}
			payload := issueMilestonePayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Milestone: item}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueMilestoneSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newIssueComponentListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue components for a repository",
		Example: "  bb issue component list --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue component list --repo workspace-slug/issues-repo-slug --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, err := resolveIssueTarget(host, workspace, repo)
			if err != nil {
				return err
			}
			items, err := client.ListIssueComponents(context.Background(), target.Workspace, target.Repo, limit)
			if err != nil {
				return err
			}
			payload := issueComponentListPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Components: items}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueComponentListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of components to return")
	return cmd
}

func newIssueComponentViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View one issue component",
		Example: "  bb issue component view 1 --repo workspace-slug/issues-repo-slug\n" +
			"  bb issue component view 1 --repo workspace-slug/issues-repo-slug --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			target, client, err := resolveIssueTarget(host, workspace, repo)
			if err != nil {
				return err
			}
			id, err := parsePositiveInt("issue component", args[0])
			if err != nil {
				return err
			}
			item, err := client.GetIssueComponent(context.Background(), target.Workspace, target.Repo, id)
			if err != nil {
				return err
			}
			payload := issueComponentPayload{Host: target.Host, Workspace: target.Workspace, Repo: target.Repo, Component: item}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeIssueComponentSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func writeIssueMilestoneListSummary(w io.Writer, payload issueMilestoneListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Milestones) == 0 {
		if _, err := fmt.Fprintf(w, "No issue milestones found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tname"); err != nil {
		return err
	}
	for _, item := range payload.Milestones {
		if _, err := fmt.Fprintf(tw, "%d\t%s\n", item.ID, output.Truncate(item.Name, 40)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue milestone view %d --repo %s/%s", payload.Milestones[0].ID, payload.Workspace, payload.Repo))
}

func writeIssueMilestoneSummary(w io.Writer, payload issueMilestonePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Milestone", fmt.Sprintf("%d", payload.Milestone.ID)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Milestone.Name); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue milestone list --repo %s/%s", payload.Workspace, payload.Repo))
}

func writeIssueComponentListSummary(w io.Writer, payload issueComponentListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Components) == 0 {
		if _, err := fmt.Fprintf(w, "No issue components found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tname"); err != nil {
		return err
	}
	for _, item := range payload.Components {
		if _, err := fmt.Fprintf(tw, "%d\t%s\n", item.ID, output.Truncate(item.Name, 40)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue component view %d --repo %s/%s", payload.Components[0].ID, payload.Workspace, payload.Repo))
}

func writeIssueComponentSummary(w io.Writer, payload issueComponentPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Component", fmt.Sprintf("%d", payload.Component.ID)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Component.Name); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb issue component list --repo %s/%s", payload.Workspace, payload.Repo))
}
