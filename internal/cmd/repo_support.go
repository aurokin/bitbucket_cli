package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/spf13/cobra"
)

func writeRepoViewSummary(w io.Writer, payload repoViewPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.RepoSlug); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Host", payload.Host); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Private)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.ProjectKey); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project Name", payload.ProjectName); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Main Branch", payload.MainBranch); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.HTMLURL); err != nil {
		return err
	}
	if err := writeLabelValue(w, "HTTPS Clone", payload.HTTPSClone); err != nil {
		return err
	}
	if err := writeLabelValue(w, "SSH Clone", payload.SSHClone); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Remote", payload.RemoteName); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Local Clone URL", payload.LocalCloneURL); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Local Root", payload.RootDir); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Description", payload.Description); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Full Name", payload.FullName); err != nil {
		return err
	}
	return writeNextStep(w, repoViewNextStep(payload))
}

func cloneURLForName(targets []bitbucket.NamedCloneTarget, name string) string {
	for _, target := range targets {
		if target.Name == name {
			return target.Href
		}
	}
	return ""
}

func repoViewNextStep(payload repoViewPayload) string {
	if payload.RootDir != "" {
		return fmt.Sprintf("bb pr list --repo %s/%s", payload.Workspace, payload.RepoSlug)
	}
	return fmt.Sprintf("bb repo clone %s/%s", payload.Workspace, payload.RepoSlug)
}

func repoVisibilityLabel(private bool) string {
	if private {
		return "private"
	}
	return "public"
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

func buildRepoCreatePayload(ctx context.Context, host, workspace, repo, repoArg string, options bitbucket.CreateRepositoryOptions) (repoCreatePayload, error) {
	resolved, err := resolveRepoCommandTargetInput(ctx, host, workspace, repo, repoArg, false)
	if err != nil {
		return repoCreatePayload{}, err
	}

	createdRepo, err := resolved.Client.CreateRepository(ctx, resolved.Target.Workspace, resolved.Target.Repo, options)
	if err != nil {
		return repoCreatePayload{}, err
	}

	return repoCreatePayload{
		Host:       resolved.Target.Host,
		Workspace:  resolved.Target.Workspace,
		Repository: createdRepo,
	}, nil
}

func writeRepoCreateSummary(w io.Writer, payload repoCreatePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repository.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Repository.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Repository.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.Repository.Project.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Repository.Links.HTML.Href); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo clone %s/%s", payload.Workspace, payload.Repository.Slug))
}

func buildRepoClonePayload(ctx context.Context, host, workspace, repo string, args []string) (repoClonePayload, error) {
	repoArg, targetDir, err := resolveRepoCloneInput(args, repo)
	if err != nil {
		return repoClonePayload{}, err
	}

	resolved, err := resolveRepoCommandTargetInput(ctx, host, workspace, repo, repoArg, false)
	if err != nil {
		return repoClonePayload{}, err
	}

	resolvedHost, hostConfig, err := resolveAuthenticatedHostConfig(resolved.Target.Host)
	if err != nil {
		return repoClonePayload{}, err
	}

	client, err := bitbucket.NewClient(resolvedHost, hostConfig)
	if err != nil {
		return repoClonePayload{}, err
	}

	repository, err := client.GetRepository(ctx, resolved.Target.Workspace, resolved.Target.Repo)
	if err != nil {
		return repoClonePayload{}, err
	}

	httpsCloneURL := cloneURLForName(repository.Links.Clone, "https")
	if httpsCloneURL == "" {
		return repoClonePayload{}, fmt.Errorf("repository %s/%s does not expose an HTTPS clone URL", resolved.Target.Workspace, repository.Slug)
	}

	if targetDir == "" {
		targetDir = repository.Slug
	}

	if err := gitrepo.CloneRepository(ctx, httpsCloneURL, hostConfig.Token, targetDir); err != nil {
		return repoClonePayload{}, err
	}

	absoluteDir, err := filepath.Abs(targetDir)
	if err != nil {
		return repoClonePayload{}, fmt.Errorf("resolve clone directory: %w", err)
	}

	return repoClonePayload{
		Host:      resolvedHost,
		Workspace: resolved.Target.Workspace,
		RepoSlug:  repository.Slug,
		Name:      repository.Name,
		Directory: absoluteDir,
		CloneURL:  httpsCloneURL,
	}, nil
}

func writeRepoCloneSummary(w io.Writer, payload repoClonePayload) error {
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
}

func buildRepoDeletePayload(ctx context.Context, cmd *cobra.Command, host, workspace, repo string, args []string, yes bool) (repoDeletePayload, error) {
	resolved, err := resolveRepoCommandTargetInput(ctx, host, workspace, repo, firstArg(args), false)
	if err != nil {
		return repoDeletePayload{}, err
	}

	repository, err := resolved.Client.GetRepository(ctx, resolved.Target.Workspace, resolved.Target.Repo)
	if err != nil {
		return repoDeletePayload{}, err
	}

	if err := confirmRepoDeletion(cmd, resolved.Target.Workspace, repository.Slug, yes); err != nil {
		return repoDeletePayload{}, err
	}

	if err := resolved.Client.DeleteRepository(ctx, resolved.Target.Workspace, repository.Slug); err != nil {
		return repoDeletePayload{}, err
	}

	return repoDeletePayload{
		Host:      resolved.Target.Host,
		Workspace: resolved.Target.Workspace,
		RepoSlug:  repository.Slug,
		Name:      repository.Name,
		Deleted:   true,
	}, nil
}

func confirmRepoDeletion(cmd *cobra.Command, workspace, repo string, yes bool) error {
	if yes {
		return nil
	}
	if !promptsEnabled(cmd) {
		return fmt.Errorf("repository deletion requires confirmation; pass --yes or run in an interactive terminal")
	}
	return confirmExactMatch(cmd, workspace+"/"+repo)
}

func writeRepoDeleteSummary(w io.Writer, payload repoDeletePayload) error {
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
}

func repoDeletionStatus(deleted bool) string {
	if deleted {
		return "deleted"
	}
	return "present"
}
