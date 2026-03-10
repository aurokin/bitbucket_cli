package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newPRCmd() *cobra.Command {
	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Work with pull requests",
	}

	prCmd.AddCommand(
		newPRListCmd(),
		newStubCommand("view", "View a pull request", "pr view"),
		newStubCommand("create", "Create a pull request", "pr create"),
		newStubCommand("checkout", "Check out a pull request locally", "pr checkout"),
	)

	return prCmd
}

func newPRListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var state string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests for a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolvedHost, resolvedWorkspace, resolvedRepo, err := resolvePRRepository(context.Background(), host, workspace, repo)
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			resolvedConfigHost, err := cfg.ResolveHost(resolvedHost)
			if err != nil {
				return err
			}
			hostConfig, ok := cfg.Hosts[resolvedConfigHost]
			if !ok {
				return fmt.Errorf("no stored credentials found for %s", resolvedConfigHost)
			}

			client, err := bitbucket.NewClient(resolvedConfigHost, hostConfig)
			if err != nil {
				return err
			}

			prs, err := client.ListPullRequests(context.Background(), resolvedWorkspace, resolvedRepo, bitbucket.ListPullRequestsOptions{
				State: state,
				Limit: limit,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, prs, func(w io.Writer) error {
				if len(prs) == 0 {
					_, err := fmt.Fprintf(w, "No pull requests found for %s/%s.\n", resolvedWorkspace, resolvedRepo)
					return err
				}

				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintln(tw, "ID\tTITLE\tSTATE\tAUTHOR\tSOURCE\tDESTINATION\tUPDATED"); err != nil {
					return err
				}
				for _, pr := range prs {
					updated := pr.UpdatedOn
					if updated != "" {
						if parsed, parseErr := time.Parse(time.RFC3339, pr.UpdatedOn); parseErr == nil {
							updated = parsed.Local().Format("2006-01-02 15:04")
						}
					}
					if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n", pr.ID, pr.Title, pr.State, pr.Author.DisplayName, pr.Source.Branch.Name, pr.Destination.Branch.Name, updated); err != nil {
						return err
					}
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Bitbucket workspace slug")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository slug")
	cmd.Flags().StringVar(&state, "state", "OPEN", "Filter pull requests by state: OPEN, MERGED, DECLINED, SUPERSEDED, or ALL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pull requests to return")

	return cmd
}

func resolvePRRepository(ctx context.Context, host, workspace, repo string) (string, string, string, error) {
	if workspace != "" && repo != "" {
		return host, workspace, repo, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("get working directory: %w", err)
	}

	repoContext, err := gitrepo.ResolveRepoContext(ctx, currentDir)
	if err != nil {
		return "", "", "", fmt.Errorf("resolve repository from flags or git remote: %w", err)
	}

	resolvedHost := host
	if resolvedHost == "" {
		resolvedHost = repoContext.Host
	}

	resolvedWorkspace := workspace
	if resolvedWorkspace == "" {
		resolvedWorkspace = repoContext.Workspace
	}

	resolvedRepo := repo
	if resolvedRepo == "" {
		resolvedRepo = repoContext.RepoSlug
	}

	return resolvedHost, resolvedWorkspace, resolvedRepo, nil
}
