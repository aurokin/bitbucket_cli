package cmd

import (
	"context"
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
				if err := writePullRequestSummaryTable(w, pr, pullRequestSummaryOptions{
					IncludeAuthor:      true,
					IncludeUpdated:     true,
					IncludeDescription: true,
				}); err != nil {
					return err
				}
				return writeNextStep(w, prViewNextStep(prTarget.RepoTarget.Workspace, prTarget.RepoTarget.Repo, pr.ID))
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}
