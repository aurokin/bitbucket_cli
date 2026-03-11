package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

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
		Long:    "Inspect, create, clone, and delete Bitbucket repositories.",
	}

	repoCmd.AddCommand(
		newRepoViewCmd(),
		newRepoCreateCmd(),
		newRepoCloneCmd(),
		newRepoDeleteCmd(),
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
			"  bb repo view --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb repo view --repo https://bitbucket.org/OhBizzle/bb-cli-integration-primary\n" +
			"  bb repo view --json name,project_key,main_branch",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoSelector(host, workspace, repo)
			if err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, true)
			if err != nil {
				return err
			}

			repository, err := client.GetRepository(context.Background(), target.Workspace, target.Repo)
			if err != nil {
				return err
			}

			payload := repoViewPayload{
				Host:        target.Host,
				Workspace:   target.Workspace,
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
			if target.LocalRepo != nil {
				payload.RemoteName = target.LocalRepo.RemoteName
				payload.RootDir = target.LocalRepo.RootDir
				payload.LocalCloneURL = target.LocalRepo.CloneURL
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				tw := output.NewTableWriter(w)
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
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

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
	var repo string
	var projectKey string
	var description string
	var private bool
	var reuseExisting bool
	var name string

	cmd := &cobra.Command{
		Use:   "create [repository]",
		Short: "Create a repository in Bitbucket Cloud",
		Long:  "Create a repository in Bitbucket Cloud. Prefer --repo <workspace>/<repo> for explicit targeting; use --workspace only to disambiguate a bare repository name. Use --reuse-existing when the command may be run repeatedly.",
		Example: "  bb repo create OhBizzle/my-repo --project-key BBCLI\n" +
			"  bb repo create --repo OhBizzle/my-repo --reuse-existing --json\n" +
			"  bb repo create my-repo --workspace OhBizzle",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoTargetInput(host, workspace, repo, firstArg(args))
			if err != nil {
				return err
			}
			if err := requireExplicitRepoTarget(selector); err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, false)
			if err != nil {
				return err
			}

			createdRepo, err := client.CreateRepository(context.Background(), target.Workspace, target.Repo, bitbucket.CreateRepositoryOptions{
				Name:          name,
				Description:   description,
				ProjectKey:    projectKey,
				IsPrivate:     private,
				ReuseExisting: reuseExisting,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, createdRepo, func(w io.Writer) error {
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "Workspace:\t%s\n", target.Workspace); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Repository:\t%s\n", createdRepo.Slug); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Name:\t%s\n", createdRepo.Name); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Private:\t%t\n", createdRepo.IsPrivate); err != nil {
					return err
				}
				if createdRepo.Project.Key != "" {
					if _, err := fmt.Fprintf(tw, "Project:\t%s\n", createdRepo.Project.Key); err != nil {
						return err
					}
				}
				if createdRepo.Links.HTML.Href != "" {
					if _, err := fmt.Fprintf(tw, "URL:\t%s\n", createdRepo.Links.HTML.Href); err != nil {
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
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
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
	var repo string

	cmd := &cobra.Command{
		Use:   "clone [repository] [directory]",
		Short: "Clone a Bitbucket repository locally",
		Long:  "Clone a Bitbucket repository over HTTPS using the configured API token. Prefer --repo <workspace>/<repo> for explicit targeting; use --workspace only to disambiguate a bare repository name. The origin remote is rewritten after cloning so the token is not stored in git config.",
		Example: "  bb repo clone OhBizzle/bb-cli-integration-primary\n" +
			"  bb repo clone --repo OhBizzle/bb-cli-integration-primary ./tmp/repo\n" +
			"  bb repo clone bb-cli-integration-primary --workspace OhBizzle\n" +
			"  bb repo clone https://bitbucket.org/OhBizzle/bb-cli-integration-primary\n" +
			"  bb repo clone OhBizzle/bb-cli-integration-primary ./tmp/repo --json",
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			repoArg, targetDir, err := resolveRepoCloneInput(args, repo)
			if err != nil {
				return err
			}

			selector, err := parseRepoTargetInput(host, workspace, repo, repoArg)
			if err != nil {
				return err
			}
			if err := requireExplicitRepoTarget(selector); err != nil {
				return err
			}

			resolvedHost, hostConfig, err := resolveAuthenticatedHostConfig(selector.Host)
			if err != nil {
				return err
			}

			client, err := bitbucket.NewClient(resolvedHost, hostConfig)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, false)
			if err != nil {
				return err
			}

			repository, err := client.GetRepository(context.Background(), target.Workspace, target.Repo)
			if err != nil {
				return err
			}

			httpsCloneURL := cloneURLForName(repository.Links.Clone, "https")
			if httpsCloneURL == "" {
				return fmt.Errorf("repository %s/%s does not expose an HTTPS clone URL", target.Workspace, repository.Slug)
			}

			if targetDir == "" {
				targetDir = repository.Slug
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
				Workspace: target.Workspace,
				RepoSlug:  repository.Slug,
				Name:      repository.Name,
				Directory: absoluteDir,
				CloneURL:  httpsCloneURL,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				tw := output.NewTableWriter(w)
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
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newRepoDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete [repository]",
		Short: "Delete a Bitbucket repository",
		Long:  "Delete a Bitbucket repository in Bitbucket Cloud. Prefer --repo <workspace>/<repo> for explicit targeting; use --workspace only to disambiguate a bare repository name. Humans must confirm the exact workspace/repository unless --yes is provided. Scripts and agents should use --yes together with --no-prompt when they need deterministic behavior.",
		Example: "  bb repo delete OhBizzle/bb-cli-delete-command-target --yes\n" +
			"  bb repo delete --repo OhBizzle/bb-cli-delete-command-target --yes\n" +
			"  bb repo delete bb-cli-delete-command-target --workspace OhBizzle --yes\n" +
			"  bb repo delete https://bitbucket.org/OhBizzle/bb-cli-delete-command-target --json",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selector, err := parseRepoTargetInput(host, workspace, repo, firstArg(args))
			if err != nil {
				return err
			}
			if err := requireExplicitRepoTarget(selector); err != nil {
				return err
			}

			resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
			if err != nil {
				return err
			}

			selector.Host = resolvedHost
			target, err := resolveRepoTarget(context.Background(), selector, client, false)
			if err != nil {
				return err
			}

			repository, err := client.GetRepository(context.Background(), target.Workspace, target.Repo)
			if err != nil {
				return err
			}

			confirmationTarget := target.Workspace + "/" + repository.Slug
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("repository deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, confirmationTarget); err != nil {
					return err
				}
			}

			if err := client.DeleteRepository(context.Background(), target.Workspace, repository.Slug); err != nil {
				return err
			}

			payload := repoDeletePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				RepoSlug:  repository.Slug,
				Name:      repository.Name,
				Deleted:   true,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintf(tw, "Workspace:\t%s\n", payload.Workspace); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(tw, "Repository:\t%s\n", payload.RepoSlug); err != nil {
					return err
				}
				if payload.Name != "" {
					if _, err := fmt.Fprintf(tw, "Name:\t%s\n", payload.Name); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintf(tw, "Deleted:\t%t\n", payload.Deleted); err != nil {
					return err
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")

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

type repoDeletePayload struct {
	Host      string `json:"host"`
	Workspace string `json:"workspace"`
	RepoSlug  string `json:"repo"`
	Name      string `json:"name,omitempty"`
	Deleted   bool   `json:"deleted"`
}

func resolveWorkspaceForCreate(ctx context.Context, client workspaceResolver, workspace string) (string, error) {
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

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func resolveRepoCloneInput(args []string, repoFlag string) (string, string, error) {
	if strings.TrimSpace(repoFlag) != "" {
		switch len(args) {
		case 0:
			return "", "", nil
		case 1:
			return "", args[0], nil
		default:
			return "", "", fmt.Errorf("when --repo is provided, pass at most one clone directory argument")
		}
	}

	switch len(args) {
	case 0:
		return "", "", fmt.Errorf("repository is required; pass <repo>, <workspace>/<repo>, or --repo")
	case 1:
		return args[0], "", nil
	default:
		return args[0], args[1], nil
	}
}
