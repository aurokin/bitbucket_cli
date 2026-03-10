package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	repoCmd := &cobra.Command{
		Use:     "repo",
		Aliases: []string{"repos", "repository"},
		Short:   "Work with Bitbucket repositories",
		Long:    "Inspect and create Bitbucket repositories.",
	}

	repoCmd.AddCommand(
		newRepoViewCmd(),
		newRepoCreateCmd(),
		newRepoCloneCmd(),
	)

	return repoCmd
}

func newRepoViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view",
		Short: "Show repository information",
		Long:  "Show repository information from Bitbucket Cloud. When run inside a git checkout, local remote details are included in the output.",
		Example: "  bb repo view\n" +
			"  bb repo view --workspace OhBizzle --repo bb-cli-integration-primary\n" +
			"  bb repo view --json name,project_key,main_branch",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			localRepo, resolvedHost, resolvedWorkspace, resolvedRepo, err := resolveRepoViewTarget(context.Background(), host, workspace, repo)
			if err != nil {
				return err
			}

			_, client, err := resolveAuthenticatedClient(resolvedHost)
			if err != nil {
				return err
			}

			repository, err := client.GetRepository(context.Background(), resolvedWorkspace, resolvedRepo)
			if err != nil {
				return err
			}

			payload := repoViewPayload{
				Host:        resolvedHost,
				Workspace:   resolvedWorkspace,
				RepoSlug:    repository.Slug,
				Name:        repository.Name,
				FullName:    repository.FullName,
				Description: repository.Description,
				Private:     repository.IsPrivate,
				ProjectKey:  repository.Project.Key,
				ProjectName: repository.Project.Name,
				MainBranch:  repository.MainBranch.Name,
				HTMLURL:     repository.Links.HTML.Href,
				HTTPSClone:  cloneURLForName(repository.Links.Clone, "https"),
				SSHClone:    cloneURLForName(repository.Links.Clone, "ssh"),
			}
			if localRepo != nil {
				payload.RemoteName = localRepo.RemoteName
				payload.RootDir = localRepo.RootDir
				payload.LocalCloneURL = localRepo.CloneURL
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintf(tw, "Workspace:\t%s\n", payload.Workspace); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Repository:\t%s\n", payload.RepoSlug); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Name:\t%s\n", payload.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Host:\t%s\n", payload.Host); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Private:\t%t\n", payload.Private); err != nil {
					return err
				}
				if payload.ProjectKey != "" {
					if _, err := fmt.Fprintf(tw, "Project:\t%s\n", payload.ProjectKey); err != nil {
						return err
					}
				}
				if payload.MainBranch != "" {
					if _, err := fmt.Fprintf(tw, "Main Branch:\t%s\n", payload.MainBranch); err != nil {
						return err
					}
				}
				if payload.HTMLURL != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", payload.HTMLURL); err != nil {
						return err
					}
				}
				if payload.HTTPSClone != "" {
					if _, err := fmt.Fprintf(tw, "HTTPS Clone:\t%s\n", payload.HTTPSClone); err != nil {
						return err
					}
				}
				if payload.SSHClone != "" {
					if _, err := fmt.Fprintf(tw, "SSH Clone:\t%s\n", payload.SSHClone); err != nil {
						return err
					}
				}
				if payload.RemoteName != "" {
					if _, err := fmt.Fprintf(tw, "Remote:\t%s\n", payload.RemoteName); err != nil {
						return err
					}
				}
				if payload.LocalCloneURL != "" {
					if _, err := fmt.Fprintf(tw, "Local Clone URL:\t%s\n", payload.LocalCloneURL); err != nil {
						return err
					}
				}
				if payload.RootDir != "" {
					if _, err := fmt.Fprintf(tw, "Root:\t%s\n", payload.RootDir); err != nil {
						return err
					}
				}
				if payload.Description != "" {
					if _, err := fmt.Fprintf(tw, "Description:\t%s\n", payload.Description); err != nil {
						return err
					}
				}
				if payload.FullName != "" {
					if _, err := fmt.Fprintf(tw, "Full Name:\t%s\n", payload.FullName); err != nil {
						return err
					}
				}
				if payload.ProjectName != "" {
					if _, err := fmt.Fprintf(tw, "Project Name:\t%s\n", payload.ProjectName); err != nil {
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

type repoViewPayload struct {
	Host          string `json:"host"`
	Workspace     string `json:"workspace"`
	RepoSlug      string `json:"repo"`
	Name          string `json:"name,omitempty"`
	FullName      string `json:"full_name,omitempty"`
	Description   string `json:"description,omitempty"`
	Private       bool   `json:"private"`
	ProjectKey    string `json:"project_key,omitempty"`
	ProjectName   string `json:"project_name,omitempty"`
	MainBranch    string `json:"main_branch,omitempty"`
	HTMLURL       string `json:"html_url,omitempty"`
	HTTPSClone    string `json:"https_clone,omitempty"`
	SSHClone      string `json:"ssh_clone,omitempty"`
	RemoteName    string `json:"remote,omitempty"`
	LocalCloneURL string `json:"local_clone_url,omitempty"`
	RootDir       string `json:"root,omitempty"`
}

func resolveRepoViewTarget(ctx context.Context, host, workspace, repo string) (*gitrepo.RepoContext, string, string, string, error) {
	if err := validateRepoSelector(workspace, repo); err != nil {
		return nil, "", "", "", err
	}

	if workspace != "" && repo != "" {
		return nil, host, workspace, repo, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", "", "", fmt.Errorf("get working directory: %w", err)
	}

	localRepo, err := gitrepo.ResolveRepoContext(ctx, currentDir)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("could not determine the repository from the current directory; run inside a Bitbucket git checkout or pass --workspace and --repo")
	}

	resolvedHost := host
	if resolvedHost == "" {
		resolvedHost = localRepo.Host
	}

	resolvedWorkspace := workspace
	if resolvedWorkspace == "" {
		resolvedWorkspace = localRepo.Workspace
	}

	resolvedRepo := repo
	if resolvedRepo == "" {
		resolvedRepo = localRepo.RepoSlug
	}

	return &localRepo, resolvedHost, resolvedWorkspace, resolvedRepo, nil
}

func cloneURLForName(targets []bitbucket.NamedCloneTarget, name string) string {
	for _, target := range targets {
		if target.Name == name {
			return target.Href
		}
	}
	return ""
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
		Long:  "Create a repository in Bitbucket Cloud. Use --reuse-existing when the command may be run repeatedly.",
		Example: "  bb repo create my-repo --workspace OhBizzle --project-key BBCLI\n" +
			"  bb repo create my-repo --workspace OhBizzle --reuse-existing --json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			_, client, err := resolveAuthenticatedClient(host)
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

func newRepoCloneCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string

	cmd := &cobra.Command{
		Use:   "clone <repository> [directory]",
		Short: "Clone a Bitbucket repository locally",
		Long:  "Clone a Bitbucket repository over HTTPS using the configured API token. The origin remote is rewritten after cloning so the token is not stored in git config.",
		Example: "  bb repo clone OhBizzle/bb-cli-integration-primary\n" +
			"  bb repo clone bb-cli-integration-primary --workspace OhBizzle\n" +
			"  bb repo clone OhBizzle/bb-cli-integration-primary ./tmp/repo --json",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolvedHost, hostConfig, err := resolveAuthenticatedHostConfig(host)
			if err != nil {
				return err
			}

			client, err := bitbucket.NewClient(resolvedHost, hostConfig)
			if err != nil {
				return err
			}

			resolvedWorkspace, repoSlug, err := resolveRepoCloneTarget(context.Background(), client, workspace, args[0])
			if err != nil {
				return err
			}

			repository, err := client.GetRepository(context.Background(), resolvedWorkspace, repoSlug)
			if err != nil {
				return err
			}

			httpsCloneURL := cloneURLForName(repository.Links.Clone, "https")
			if httpsCloneURL == "" {
				return fmt.Errorf("repository %s/%s does not expose an HTTPS clone URL", resolvedWorkspace, repository.Slug)
			}

			targetDir := repository.Slug
			if len(args) == 2 {
				targetDir = args[1]
			}

			if err := gitrepo.CloneRepository(context.Background(), httpsCloneURL, hostConfig.Token, targetDir); err != nil {
				return err
			}

			absoluteDir, err := filepath.Abs(targetDir)
			if err != nil {
				return fmt.Errorf("resolve clone directory: %w", err)
			}

			payload := repoClonePayload{
				Host:      resolvedHost,
				Workspace: resolvedWorkspace,
				RepoSlug:  repository.Slug,
				Name:      repository.Name,
				Directory: absoluteDir,
				CloneURL:  httpsCloneURL,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintf(tw, "Workspace:\t%s\n", payload.Workspace); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Repository:\t%s\n", payload.RepoSlug); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Directory:\t%s\n", payload.Directory); err != nil {
					return err
				}
				if payload.CloneURL != "" {
					if _, err := fmt.Fprintf(tw, "Clone URL:\t%s\n", payload.CloneURL); err != nil {
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

	return cmd
}

type repoClonePayload struct {
	Host      string `json:"host"`
	Workspace string `json:"workspace"`
	RepoSlug  string `json:"repo"`
	Name      string `json:"name,omitempty"`
	Directory string `json:"directory"`
	CloneURL  string `json:"clone_url,omitempty"`
}

func resolveRepoCloneTarget(ctx context.Context, client *bitbucket.Client, workspaceFlag, target string) (string, string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", "", fmt.Errorf("repository is required")
	}

	if strings.Count(target, "/") > 1 {
		return "", "", fmt.Errorf("repository must be provided as <repo> or <workspace>/<repo>")
	}

	if strings.Contains(target, "/") {
		parts := strings.SplitN(target, "/", 2)
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("repository must be provided as <repo> or <workspace>/<repo>")
		}
		if workspaceFlag != "" && workspaceFlag != parts[0] {
			return "", "", fmt.Errorf("--workspace %q does not match repository target %q", workspaceFlag, target)
		}
		return parts[0], parts[1], nil
	}

	resolvedWorkspace, err := resolveWorkspaceForCreate(ctx, client, workspaceFlag)
	if err != nil {
		return "", "", err
	}

	return resolvedWorkspace, target, nil
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
