package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/aurokin/bitbucket_cli/internal/output"
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
		newRepoListCmd(),
		newRepoHookCmd(),
		newRepoDeployKeyCmd(),
		newRepoPermissionsCmd(),
		newRepoCreateCmd(),
		newRepoEditCmd(),
		newRepoForkCmd(),
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
			"  bb repo view --repo workspace-slug/repo-slug\n" +
			"  bb repo view --repo https://bitbucket.org/workspace-slug/repo-slug\n" +
			"  bb repo view --json name,project_key,main_branch",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			repository, err := client.GetRepository(context.Background(), target.Workspace, target.Repo)
			if err != nil {
				return err
			}

			payload := repoViewPayload{
				Host:        target.Host,
				Workspace:   target.Workspace,
				RepoSlug:    repository.Slug,
				Warnings:    append([]string(nil), target.Warnings...),
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
				return writeRepoViewSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

type repoViewPayload struct {
	Host          string   `json:"host"`
	Workspace     string   `json:"workspace"`
	RepoSlug      string   `json:"repo"`
	Warnings      []string `json:"warnings,omitempty"`
	Name          string   `json:"name,omitempty"`
	FullName      string   `json:"full_name,omitempty"`
	Description   string   `json:"description,omitempty"`
	Private       bool     `json:"private"`
	ProjectKey    string   `json:"project_key,omitempty"`
	ProjectName   string   `json:"project_name,omitempty"`
	MainBranch    string   `json:"main_branch,omitempty"`
	HTMLURL       string   `json:"html_url,omitempty"`
	HTTPSClone    string   `json:"https_clone,omitempty"`
	SSHClone      string   `json:"ssh_clone,omitempty"`
	RemoteName    string   `json:"remote,omitempty"`
	LocalCloneURL string   `json:"local_clone_url,omitempty"`
	RootDir       string   `json:"root,omitempty"`
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
		Example: "  bb repo create workspace-slug/my-repo --project-key BBCLI\n" +
			"  bb repo create --repo workspace-slug/my-repo --reuse-existing --json\n" +
			"  bb repo create my-repo --workspace workspace-slug",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTargetInput(context.Background(), host, workspace, repo, firstArg(args), false)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

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
				if err := writeTargetHeader(w, "Repository", target.Workspace, createdRepo.Slug); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Name", createdRepo.Name); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(createdRepo.IsPrivate)); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Project", createdRepo.Project.Key); err != nil {
					return err
				}
				if err := writeLabelValue(w, "URL", createdRepo.Links.HTML.Href); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb repo clone %s/%s", target.Workspace, createdRepo.Slug))
			})
		},
	}

	addFormatFlags(cmd, &flags)
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
		Example: "  bb repo clone workspace-slug/repo-slug\n" +
			"  bb repo clone --repo workspace-slug/repo-slug ./tmp/repo\n" +
			"  bb repo clone repo-slug --workspace workspace-slug\n" +
			"  bb repo clone https://bitbucket.org/workspace-slug/repo-slug\n" +
			"  bb repo clone workspace-slug/repo-slug ./tmp/repo --json",
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

			resolved, err := resolveRepoCommandTargetInput(context.Background(), host, workspace, repo, repoArg, false)
			if err != nil {
				return err
			}
			target := resolved.Target

			resolvedHost, hostConfig, err := resolveAuthenticatedHostConfig(target.Host)
			if err != nil {
				return err
			}

			client, err := bitbucket.NewClient(resolvedHost, hostConfig)
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
				if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.RepoSlug); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Directory", payload.Directory); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Clone URL", payload.CloneURL); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb repo view --repo %s/%s", payload.Workspace, payload.RepoSlug))
			})
		},
	}

	addFormatFlags(cmd, &flags)
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
		Example: "  bb repo delete workspace-slug/delete-repo-slug --yes\n" +
			"  bb repo delete --repo workspace-slug/delete-repo-slug --yes\n" +
			"  bb repo delete delete-repo-slug --workspace workspace-slug --yes\n" +
			"  bb repo delete https://bitbucket.org/workspace-slug/delete-repo-slug --json",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTargetInput(context.Background(), host, workspace, repo, firstArg(args), false)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

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
				if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.RepoSlug); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Name", payload.Name); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Status", repoDeletionStatus(payload.Deleted)); err != nil {
					return err
				}
				return writeNextStep(w, fmt.Sprintf("bb repo create %s/%s", payload.Workspace, payload.RepoSlug))
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
