package cmd

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Work with Bitbucket repositories",
	}

	repoCmd.AddCommand(newRepoViewCmd())

	return repoCmd
}

func newRepoViewCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "view",
		Short: "Show repository information for the current git checkout",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			repo, err := gitrepo.ResolveRepoContext(context.Background(), ".")
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, repo, func(w io.Writer) error {
				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintf(tw, "Workspace:\t%s\n", repo.Workspace); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Repository:\t%s\n", repo.RepoSlug); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Host:\t%s\n", repo.Host); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Remote:\t%s\n", repo.RemoteName); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Clone URL:\t%s\n", repo.CloneURL); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Root:\t%s\n", repo.RootDir); err != nil {
					return err
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")

	return cmd
}
