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

type repoListPayload struct {
	Host      string                 `json:"host"`
	Workspace string                 `json:"workspace"`
	Query     string                 `json:"query,omitempty"`
	Repos     []bitbucket.Repository `json:"repos"`
}

type repoEditPayload struct {
	Host         string               `json:"host"`
	Workspace    string               `json:"workspace"`
	RepoSlug     string               `json:"repo"`
	PreviousRepo string               `json:"previous_repo,omitempty"`
	Action       string               `json:"action"`
	Repository   bitbucket.Repository `json:"repository"`
}

type repoForkPayload struct {
	Host            string               `json:"host"`
	SourceWorkspace string               `json:"source_workspace"`
	SourceRepo      string               `json:"source_repo"`
	Action          string               `json:"action"`
	Repository      bitbucket.Repository `json:"repository"`
}

type repoForkInput struct {
	DestinationWorkspace string
	Name                 string
	Description          string
	Visibility           string
	ReuseExisting        bool
}

func newRepoListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, query, sort string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [workspace]",
		Short: "List repositories in a workspace",
		Long:  "List Bitbucket repositories in one workspace. If you have access to exactly one workspace, the workspace can be omitted.",
		Example: "  bb repo list workspace-slug\n" +
			"  bb repo list --workspace workspace-slug --limit 50\n" +
			"  bb repo list workspace-slug --query 'name ~ \"bb\"' --json repos",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			selectedWorkspace, err := resolveWorkspaceInput(workspace, firstArg(args))
			if err != nil {
				return err
			}
			resolvedHost, client, err := resolveAuthenticatedClient(host)
			if err != nil {
				return err
			}
			selectedWorkspace, err = resolveWorkspaceForCreate(context.Background(), client, selectedWorkspace)
			if err != nil {
				return err
			}

			repos, err := client.ListRepositories(context.Background(), selectedWorkspace, bitbucket.ListRepositoriesOptions{
				Query: query,
				Sort:  sort,
				Limit: limit,
			})
			if err != nil {
				return err
			}

			payload := repoListPayload{
				Host:      resolvedHost,
				Workspace: selectedWorkspace,
				Query:     query,
				Repos:     repos,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Workspace slug to list repositories from")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket repository query filter")
	cmd.Flags().StringVar(&sort, "sort", "-updated_on", "Bitbucket repository sort expression")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of repositories to return")
	return cmd
}

func newRepoEditCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, visibility, name, description string

	cmd := &cobra.Command{
		Use:   "edit [repository]",
		Short: "Edit repository metadata",
		Long:  "Edit a Bitbucket repository's name, description, or visibility. Use --repo <workspace>/<repo> for deterministic targeting.",
		Example: "  bb repo edit workspace-slug/repo-slug --description 'Updated description'\n" +
			"  bb repo edit --repo workspace-slug/repo-slug --visibility public --json '*'\n" +
			"  bb repo edit repo-slug --workspace workspace-slug --name 'Renamed repo'",
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

			private, err := parseRepoVisibility(visibility)
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" && strings.TrimSpace(description) == "" && private == nil {
				return fmt.Errorf("at least one of --name, --description, or --visibility must be provided")
			}

			updated, err := resolved.Client.UpdateRepository(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.UpdateRepositoryOptions{
				Name:        name,
				Description: description,
				IsPrivate:   private,
			})
			if err != nil {
				return err
			}

			payload := repoEditPayload{
				Host:         resolved.Target.Host,
				Workspace:    resolved.Target.Workspace,
				RepoSlug:     updated.Slug,
				PreviousRepo: previousRepoSlug(resolved.Target.Repo, updated.Slug),
				Action:       "updated",
				Repository:   updated,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoEditSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&name, "name", "", "New repository display name")
	cmd.Flags().StringVar(&description, "description", "", "New repository description")
	cmd.Flags().StringVar(&visibility, "visibility", "", "New repository visibility: private or public")
	return cmd
}

func newRepoForkCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, destinationWorkspace, name, description, visibility string
	var reuseExisting bool

	cmd := &cobra.Command{
		Use:   "fork [repository]",
		Short: "Fork a repository",
		Long:  "Fork a Bitbucket repository. When forking into the same workspace, Bitbucket requires a new fork name because the slug is derived from the name.",
		Example: "  bb repo fork workspace-slug/repo-slug --to-workspace workspace-slug --name repo-slug-fork\n" +
			"  bb repo fork --repo workspace-slug/repo-slug --to-workspace other-workspace --reuse-existing --json '*'\n" +
			"  bb repo fork repo-slug --workspace workspace-slug --to-workspace workspace-slug --name repo-slug-fork",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			payload, err := buildRepoForkPayload(context.Background(), host, workspace, repo, firstArg(args), repoForkInput{
				DestinationWorkspace: destinationWorkspace,
				Name:                 name,
				Description:          description,
				Visibility:           visibility,
				ReuseExisting:        reuseExisting,
			})
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoForkSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&destinationWorkspace, "to-workspace", "", "Workspace to create the fork in; defaults to the source workspace")
	cmd.Flags().StringVar(&name, "name", "", "Fork display name; required when forking into the same workspace")
	cmd.Flags().StringVar(&description, "description", "", "Fork description override")
	cmd.Flags().StringVar(&visibility, "visibility", "", "Fork visibility override: private or public")
	cmd.Flags().BoolVar(&reuseExisting, "reuse-existing", false, "Return an existing matching fork instead of failing")
	return cmd
}

func writeRepoListSummary(w io.Writer, payload repoListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", payload.Query); err != nil {
		return err
	}
	if len(payload.Repos) == 0 {
		if _, err := fmt.Fprintf(w, "No repositories found in %s.\n", payload.Workspace); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "slug\tname\tvis\tproject\tupdated"); err != nil {
		return err
	}
	for _, repo := range payload.Repos {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			output.Truncate(repo.Slug, 24),
			output.Truncate(repo.Name, 28),
			repoVisibilityLabel(repo.IsPrivate),
			output.Truncate(repo.Project.Key, 10),
			output.Truncate(repo.UpdatedOn, 20),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo view --repo %s/%s", payload.Workspace, payload.Repos[0].Slug))
}

func writeRepoEditSummary(w io.Writer, payload repoEditPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repository.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Previous Repository", payload.PreviousRepo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Repository.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Repository.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Repository.Links.HTML.Href); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Description", payload.Repository.Description); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo view --repo %s/%s", payload.Workspace, payload.Repository.Slug))
}

func writeRepoForkSummary(w io.Writer, payload repoForkPayload) error {
	workspace := repoWorkspace(payload.Repository)
	if workspace == "" {
		workspace = payload.SourceWorkspace
	}
	if err := writeTargetHeader(w, "Repository", workspace, payload.Repository.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Source", payload.SourceWorkspace+"/"+payload.SourceRepo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Repository.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Repository.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Repository.Links.HTML.Href); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo clone %s/%s", workspace, payload.Repository.Slug))
}

func buildRepoForkPayload(ctx context.Context, host, workspace, repo, repoArg string, input repoForkInput) (repoForkPayload, error) {
	resolved, err := resolveRepoCommandTargetInput(ctx, host, workspace, repo, repoArg, false)
	if err != nil {
		return repoForkPayload{}, err
	}

	private, err := parseRepoVisibility(input.Visibility)
	if err != nil {
		return repoForkPayload{}, err
	}
	forked, err := resolved.Client.ForkRepository(ctx, resolved.Target.Workspace, resolved.Target.Repo, bitbucket.ForkRepositoryOptions{
		Workspace:     input.DestinationWorkspace,
		Name:          input.Name,
		Description:   input.Description,
		IsPrivate:     private,
		ReuseExisting: input.ReuseExisting,
	})
	if err != nil {
		return repoForkPayload{}, err
	}

	return repoForkPayload{
		Host:            resolved.Target.Host,
		SourceWorkspace: resolved.Target.Workspace,
		SourceRepo:      resolved.Target.Repo,
		Action:          repoForkAction(forked, resolved.Target.Workspace, input.DestinationWorkspace, input.Name, input.ReuseExisting),
		Repository:      forked,
	}, nil
}

func repoForkAction(forked bitbucket.Repository, sourceWorkspace, destinationWorkspace, name string, reuseExisting bool) string {
	action := "forked"
	if forked.Parent == nil || forked.Parent.FullName == "" || !reuseExisting {
		return action
	}
	expectedWorkspace := destinationWorkspace
	if expectedWorkspace == "" {
		expectedWorkspace = sourceWorkspace
	}
	expectedName := strings.TrimSpace(name)
	if expectedName == "" {
		expectedName = forked.Name
	}
	if strings.EqualFold(forked.Name, expectedName) && strings.HasPrefix(forked.FullName, expectedWorkspace+"/") {
		return "forked"
	}
	return action
}

func parseRepoVisibility(raw string) (*bool, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "":
		return nil, nil
	case "private":
		v := true
		return &v, nil
	case "public":
		v := false
		return &v, nil
	default:
		return nil, fmt.Errorf("repository visibility must be either private or public")
	}
}

func resolveWorkspaceInput(flagValue, positional string) (string, error) {
	flagValue = strings.TrimSpace(flagValue)
	positional = strings.TrimSpace(positional)
	if flagValue != "" && positional != "" && flagValue != positional {
		return "", fmt.Errorf("workspace %q does not match %q", flagValue, positional)
	}
	if flagValue != "" {
		return flagValue, nil
	}
	return positional, nil
}

func previousRepoSlug(previous, current string) string {
	if previous == current {
		return ""
	}
	return previous
}

func repoWorkspace(repo bitbucket.Repository) string {
	if repo.FullName == "" {
		return ""
	}
	parts := strings.SplitN(repo.FullName, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
