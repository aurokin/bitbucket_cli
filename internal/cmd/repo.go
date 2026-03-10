package cmd

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Work with Bitbucket repositories",
	}

	repoCmd.AddCommand(
		newRepoViewCmd(),
		newRepoCreateCmd(),
	)

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

func newRepoCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var projectKey string
	var description string
	var private bool
	var reuseExisting bool
	var name string

	cmd := &cobra.Command{
		Use:   "create <slug>",
		Short: "Create a repository in Bitbucket Cloud",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			resolvedHost, err := cfg.ResolveHost(host)
			if err != nil {
				return err
			}
			hostConfig, ok := cfg.Hosts[resolvedHost]
			if !ok {
				return fmt.Errorf("no stored credentials found for %s", resolvedHost)
			}

			client, err := bitbucket.NewClient(resolvedHost, hostConfig)
			if err != nil {
				return err
			}

			resolvedWorkspace, err := resolveWorkspaceForCreate(context.Background(), client, workspace)
			if err != nil {
				return err
			}

			repo, err := client.CreateRepository(context.Background(), resolvedWorkspace, args[0], bitbucket.CreateRepositoryOptions{
				Name:          name,
				Description:   description,
				ProjectKey:    projectKey,
				IsPrivate:     private,
				ReuseExisting: reuseExisting,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, repo, func(w io.Writer) error {
				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintf(tw, "Workspace:\t%s\n", resolvedWorkspace); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Repository:\t%s\n", repo.Slug); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Name:\t%s\n", repo.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Private:\t%t\n", repo.IsPrivate); err != nil {
					return err
				}
				if repo.Project.Key != "" {
					if _, err := fmt.Fprintf(tw, "Project:\t%s\n", repo.Project.Key); err != nil {
						return err
					}
				}
				if repo.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", repo.Links.HTML.Href); err != nil {
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
	cmd.Flags().StringVar(&workspace, "workspace", "", "Bitbucket workspace slug; inferred when only one workspace is available")
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Bitbucket project key for the repository")
	cmd.Flags().StringVar(&description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&private, "private", true, "Create the repository as private")
	cmd.Flags().BoolVar(&reuseExisting, "reuse-existing", false, "Return the existing repository instead of failing when it already exists")
	cmd.Flags().StringVar(&name, "name", "", "Display name for the repository")

	return cmd
}

func resolveWorkspaceForCreate(ctx context.Context, client *bitbucket.Client, workspace string) (string, error) {
	if workspace != "" {
		return workspace, nil
	}

	workspaces, err := client.ListWorkspaces(ctx)
	if err != nil {
		return "", err
	}
	if len(workspaces) == 1 {
		return workspaces[0].Slug, nil
	}
	if len(workspaces) == 0 {
		return "", fmt.Errorf("no Bitbucket workspaces available")
	}

	return "", fmt.Errorf("multiple workspaces available; specify --workspace")
}
