package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newPRViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <id-or-url>",
		Short: "View a pull request",
		Example: "  bb pr view 1\n" +
			"  bb pr view 1 --json title,state,source,destination\n" +
			"  bb pr view https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1",
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
			prTarget := resolved.Target
			client := resolved.Client

			pr, err := client.GetPullRequest(context.Background(), prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, prTarget.ID)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, pr, func(w io.Writer) error {
				if err := writeTargetHeader(w, "Repository", prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo); err != nil {
					return err
				}
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", pr.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", pr.Title); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", pr.State); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Author:\t%s\n", pr.Author.DisplayName); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Source:\t%s\n", pr.Source.Branch.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", pr.Destination.Branch.Name); err != nil {
					return err
				}
				if pr.UpdatedOn != "" {
					if _, err := fmt.Fprintf(tw, "Updated:\t%s\n", pr.UpdatedOn); err != nil {
						return err
					}
				}
				if pr.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", pr.Links.HTML.Href); err != nil {
						return err
					}
				}
				if pr.Description != "" {
					if _, err := fmt.Fprintf(tw, "Description:\t%s\n", pr.Description); err != nil {
						return err
					}
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				return writeNextStep(w, prViewNextStep(prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, pr.ID))
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}
