package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
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
		Long:  "List, view, create, and check out Bitbucket pull requests.",
	}

	prCmd.AddCommand(
		newPRListCmd(),
		newPRViewCmd(),
		newPRCreateCmd(),
		newPRCheckoutCmd(),
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
		Example: "  bb pr list\n" +
			"  bb pr list --workspace OhBizzle --repo bb-cli-integration-primary\n" +
			"  bb pr list --state ALL --json id,title,state",
		Args: cobra.NoArgs,
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

func newPRCheckoutCmd() *cobra.Command {
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "checkout <id>",
		Short: "Check out a pull request locally",
		Long:  "Fetch the pull request source branch from the current repository's remote and switch to it locally.",
		Example: "  bb pr checkout 1\n" +
			"  bb pr checkout 1 --workspace OhBizzle --repo bb-cli-integration-primary",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prID, err := strconv.Atoi(args[0])
			if err != nil || prID <= 0 {
				return fmt.Errorf("invalid pull request ID %q", args[0])
			}

			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			repoContext, err := gitrepo.ResolveRepoContext(context.Background(), currentDir)
			if err != nil {
				return fmt.Errorf("pr checkout must be run inside a local git checkout of the target repository")
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

			pr, err := client.GetPullRequest(context.Background(), resolvedWorkspace, resolvedRepo, prID)
			if err != nil {
				return err
			}

			if err := gitrepo.CheckoutRemoteBranch(context.Background(), repoContext.RootDir, repoContext.RemoteName, pr.Source.Branch.Name); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Checked out %s for PR #%d\n", pr.Source.Branch.Name, pr.ID)
			return err
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Bitbucket workspace slug")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository slug")

	return cmd
}

func newPRCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var title string
	var description string
	var source string
	var destination string
	var closeSourceBranch bool
	var draft bool
	var reuseExisting bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		Long:  "Create a pull request in Bitbucket Cloud. The source branch defaults to the current branch and the destination defaults to the repository main branch.",
		Example: "  bb pr create --title 'Add feature'\n" +
			"  bb pr create --source feature --destination main --description 'Ready for review'\n" +
			"  bb pr create --reuse-existing --json",
		Args: cobra.NoArgs,
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

			sourceBranch, err := resolveSourceBranch(source)
			if err != nil {
				return err
			}

			destinationBranch, err := resolveDestinationBranch(context.Background(), client, resolvedWorkspace, resolvedRepo, destination)
			if err != nil {
				return err
			}

			if title == "" {
				title = defaultPRTitle(sourceBranch)
			}

			pr, err := client.CreatePullRequest(context.Background(), resolvedWorkspace, resolvedRepo, bitbucket.CreatePullRequestOptions{
				Title:             title,
				Description:       description,
				SourceBranch:      sourceBranch,
				DestinationBranch: destinationBranch,
				CloseSourceBranch: closeSourceBranch,
				Draft:             draft,
				ReuseExisting:     reuseExisting,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, pr, func(w io.Writer) error {
				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintf(tw, "ID:\t%d\n", pr.ID); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Title:\t%s\n", pr.Title); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "State:\t%s\n", pr.State); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Source:\t%s\n", pr.Source.Branch.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", pr.Destination.Branch.Name); err != nil {
					return err
				}
				if pr.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", pr.Links.HTML.Href); err != nil {
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
	cmd.Flags().StringVar(&title, "title", "", "Pull request title; defaults to the source branch name")
	cmd.Flags().StringVar(&description, "description", "", "Pull request description")
	cmd.Flags().StringVar(&source, "source", "", "Source branch; defaults to the current git branch")
	cmd.Flags().StringVar(&destination, "destination", "", "Destination branch; defaults to the repository main branch")
	cmd.Flags().BoolVar(&closeSourceBranch, "close-source-branch", false, "Close the source branch when the pull request is merged")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create the pull request as a draft")
	cmd.Flags().BoolVar(&reuseExisting, "reuse-existing", false, "Return an existing matching open pull request instead of creating a new one")

	return cmd
}

func newPRViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a pull request",
		Example: "  bb pr view 1\n" +
			"  bb pr view 1 --json title,state,source,destination\n" +
			"  bb pr view 1 --workspace OhBizzle --repo bb-cli-integration-primary",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			prID, err := strconv.Atoi(args[0])
			if err != nil || prID <= 0 {
				return fmt.Errorf("invalid pull request ID %q", args[0])
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

			pr, err := client.GetPullRequest(context.Background(), resolvedWorkspace, resolvedRepo, prID)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, pr, func(w io.Writer) error {
				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
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

	return cmd
}

func resolvePRRepository(ctx context.Context, host, workspace, repo string) (string, string, string, error) {
	if err := validateRepoSelector(workspace, repo); err != nil {
		return "", "", "", err
	}

	if workspace != "" && repo != "" {
		return host, workspace, repo, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("get working directory: %w", err)
	}

	repoContext, err := gitrepo.ResolveRepoContext(ctx, currentDir)
	if err != nil {
		return "", "", "", fmt.Errorf("could not determine the repository from the current directory; run inside a Bitbucket git checkout or pass --workspace and --repo")
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

func resolveSourceBranch(source string) (string, error) {
	if source != "" {
		return source, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	branch, err := gitrepo.CurrentBranch(context.Background(), currentDir)
	if err != nil {
		return "", fmt.Errorf("resolve source branch: %w", err)
	}
	if branch == "" {
		return "", fmt.Errorf("could not determine current branch; pass --source")
	}

	return branch, nil
}

func resolveDestinationBranch(ctx context.Context, client *bitbucket.Client, workspace, repo, destination string) (string, error) {
	if destination != "" {
		return destination, nil
	}

	repository, err := client.GetRepository(ctx, workspace, repo)
	if err != nil {
		return "", err
	}
	if repository.MainBranch.Name == "" {
		return "", fmt.Errorf("repository main branch is unknown; pass --destination")
	}

	return repository.MainBranch.Name, nil
}

func defaultPRTitle(sourceBranch string) string {
	return sourceBranch
}
