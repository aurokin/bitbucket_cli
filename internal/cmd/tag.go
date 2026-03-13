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

type tagListPayload struct {
	Host      string                    `json:"host"`
	Workspace string                    `json:"workspace"`
	Repo      string                    `json:"repo"`
	Warnings  []string                  `json:"warnings,omitempty"`
	Query     string                    `json:"query,omitempty"`
	Tags      []bitbucket.RepositoryTag `json:"tags"`
}

type tagPayload struct {
	Host      string                  `json:"host"`
	Workspace string                  `json:"workspace"`
	Repo      string                  `json:"repo"`
	Warnings  []string                `json:"warnings,omitempty"`
	Action    string                  `json:"action,omitempty"`
	Tag       bitbucket.RepositoryTag `json:"tag"`
}

type tagDeletePayload struct {
	Host      string   `json:"host"`
	Workspace string   `json:"workspace"`
	Repo      string   `json:"repo"`
	Warnings  []string `json:"warnings,omitempty"`
	Tag       string   `json:"tag"`
	Deleted   bool     `json:"deleted"`
}

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Work with repository tags",
		Long:  "List, inspect, create, and delete Bitbucket repository tags.",
	}
	cmd.AddCommand(newTagListCmd(), newTagViewCmd(), newTagCreateCmd(), newTagDeleteCmd())
	return cmd
}

func newTagListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, query, sort string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [repository]",
		Short: "List repository tags",
		Long:  "List tags in a Bitbucket repository.",
		Example: "  bb tag list workspace-slug/repo-slug\n" +
			"  bb tag list --repo workspace-slug/repo-slug --limit 50\n" +
			"  bb tag list --repo workspace-slug/repo-slug --query 'name ~ \"v1\"' --json tags",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, formatErr := flags.options()
			if formatErr != nil {
				return formatErr
			}
			var resolved resolvedRepoCommandTarget
			var err error
			if firstArg(args) != "" || strings.TrimSpace(repo) != "" {
				resolved, err = resolveRepoCommandTargetInput(context.Background(), host, workspace, repo, firstArg(args), true)
			} else {
				resolved, err = resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			}
			if err != nil {
				return err
			}

			tags, err := resolved.Client.ListTags(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.ListTagsOptions{
				Limit: limit,
				Query: query,
				Sort:  sort,
			})
			if err != nil {
				return err
			}

			payload := tagListPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Query:     query,
				Tags:      tags,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeTagListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket tag query filter")
	cmd.Flags().StringVar(&sort, "sort", "name", "Bitbucket tag sort expression")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of tags to return")
	return cmd
}

func newTagViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <name>",
		Short: "View one tag",
		Long:  "View one Bitbucket repository tag.",
		Example: "  bb tag view v1.0.0 --repo workspace-slug/repo-slug\n" +
			"  bb tag view release-2026 --repo workspace-slug/repo-slug --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			tag, err := resolved.Client.GetTag(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}

			payload := tagPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Tag:       tag,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeTagSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newTagCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, target, message string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a tag",
		Long:  "Create a Bitbucket tag from one commit hash or branch name. When run inside a matching checkout, bb defaults the target to the current branch if --target is omitted.",
		Example: "  bb tag create v1.0.0 --repo workspace-slug/repo-slug --target main --message 'release'\n" +
			"  bb tag create v1.0.0 --repo workspace-slug/repo-slug --target abc1234 --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			targetRef, err := resolveRefCreateTarget(resolved.Target, target)
			if err != nil {
				return err
			}

			tag, err := resolved.Client.CreateTag(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.CreateTagOptions{
				Name:    args[0],
				Target:  targetRef,
				Message: message,
			})
			if err != nil {
				return err
			}

			payload := tagPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Action:    "created",
				Tag:       tag,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeTagCreateSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&target, "target", "", "Source commit hash or branch name to tag")
	cmd.Flags().StringVar(&message, "message", "", "Optional tag message for annotated tags")
	return cmd
}

func newTagDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a tag",
		Long:  "Delete a Bitbucket tag. Humans must confirm the exact repository and tag unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb tag delete v1.0.0 --repo workspace-slug/repo-slug --yes\n" +
			"  bb --no-prompt tag delete v1.0.0 --repo workspace-slug/repo-slug --yes --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			name := strings.TrimSpace(args[0])
			confirmationTarget := fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, name)
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("tag deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, confirmationTarget); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeleteTag(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, name); err != nil {
				return err
			}

			payload := tagDeletePayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Tag:       name,
				Deleted:   true,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeTagDeleteSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func writeTagListSummary(w io.Writer, payload tagListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", payload.Query); err != nil {
		return err
	}
	if len(payload.Tags) == 0 {
		if _, err := fmt.Fprintln(w, "No tags found."); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\ttarget\tdate"); err != nil {
		return err
	}
	for _, tag := range payload.Tags {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(tag.Name, 28),
			output.Truncate(tag.Target.Hash, 12),
			formatPRUpdated(tag.Date),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb tag view %s --repo %s/%s", payload.Tags[0].Name, payload.Workspace, payload.Repo))
}

func writeTagSummary(w io.Writer, payload tagPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Tag", payload.Tag.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Target", payload.Tag.Target.Hash); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Date", formatPRUpdated(payload.Tag.Date)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Message", firstLine(strings.TrimSpace(payload.Tag.Message))); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb browse --repo %s/%s --commit %s --no-browser", payload.Workspace, payload.Repo, payload.Tag.Target.Hash))
}

func writeTagCreateSummary(w io.Writer, payload tagPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Tag", payload.Tag.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Target", payload.Tag.Target.Hash); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb tag view %s --repo %s/%s", payload.Tag.Name, payload.Workspace, payload.Repo))
}

func writeTagDeleteSummary(w io.Writer, payload tagDeletePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Tag", payload.Tag); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", "deleted"); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Status", "deleted"); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb tag list --repo %s/%s", payload.Workspace, payload.Repo))
}
