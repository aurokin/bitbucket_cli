package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
)

var (
	getWorkingDirectory = os.Getwd
	resolveRepoAtDir    = gitrepo.ResolveRepoContext
)

type workspaceResolver interface {
	ListWorkspaces(ctx context.Context) ([]bitbucket.Workspace, error)
}

type repoSelector struct {
	Host      string
	Workspace string
	Repo      string
	Explicit  bool
}

type resolvedRepoTarget struct {
	LocalRepo *gitrepo.RepoContext
	Warnings  []string
	Host      string
	Workspace string
	Repo      string
	Explicit  bool
}

type pullRequestSelector struct {
	Repo repoSelector
	ID   int
}

type resolvedPullRequestTarget struct {
	RepoTarget resolvedRepoTarget
	ID         int
}

func parseRepoSelector(hostFlag, workspaceFlag, repoFlag string) (repoSelector, error) {
	hostFlag = strings.TrimSpace(hostFlag)
	workspaceFlag = strings.TrimSpace(workspaceFlag)
	repoFlag = strings.TrimSpace(repoFlag)

	if repoFlag == "" {
		if workspaceFlag != "" {
			return repoSelector{}, fmt.Errorf("--workspace requires --repo")
		}
		return repoSelector{Host: hostFlag}, nil
	}

	repoTarget, err := parseRepositoryReference(repoFlag)
	if err != nil {
		return repoSelector{}, err
	}

	if workspaceFlag != "" {
		if repoTarget.Workspace != "" && repoTarget.Workspace != workspaceFlag {
			return repoSelector{}, fmt.Errorf("--workspace %q does not match repository target %q", workspaceFlag, repoFlag)
		}
		repoTarget.Workspace = workspaceFlag
	}

	if hostFlag != "" {
		if repoTarget.Host != "" && repoTarget.Host != hostFlag {
			return repoSelector{}, fmt.Errorf("--host %q does not match repository target %q", hostFlag, repoFlag)
		}
		repoTarget.Host = hostFlag
	}

	repoTarget.Explicit = true
	return repoTarget, nil
}

func parseRepoTargetInput(hostFlag, workspaceFlag, repoFlag, positional string) (repoSelector, error) {
	hostFlag = strings.TrimSpace(hostFlag)
	workspaceFlag = strings.TrimSpace(workspaceFlag)
	repoFlag = strings.TrimSpace(repoFlag)
	positional = strings.TrimSpace(positional)

	selector := repoSelector{Host: hostFlag}
	if workspaceFlag != "" {
		selector.Workspace = workspaceFlag
	}

	if repoFlag != "" {
		repoTarget, err := parseRepositoryReference(repoFlag)
		if err != nil {
			return repoSelector{}, err
		}

		selector, err = mergeRepoSelectors(selector, repoTarget)
		if err != nil {
			return repoSelector{}, err
		}
	}

	if positional != "" {
		positionalTarget, err := parseRepositoryReference(positional)
		if err != nil {
			return repoSelector{}, err
		}

		selector, err = mergeRepoSelectors(selector, positionalTarget)
		if err != nil {
			return repoSelector{}, err
		}
	}

	if selector.Repo != "" {
		selector.Explicit = true
	}

	return selector, nil
}

func requireExplicitRepoTarget(selector repoSelector) error {
	if strings.TrimSpace(selector.Repo) != "" {
		return nil
	}

	return fmt.Errorf("repository is required; pass <repo>, <workspace>/<repo>, or --repo")
}

func parseRepositoryReference(raw string) (repoSelector, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return repoSelector{}, fmt.Errorf("repository is required")
	}

	if strings.Contains(raw, "://") {
		return parseRepositoryURL(raw)
	}

	if strings.Contains(raw, "@") && strings.Contains(raw, ":") {
		parsed, err := gitrepo.ParseRemoteURL(raw)
		if err != nil {
			return repoSelector{}, fmt.Errorf("parse repository target %q: %w", raw, err)
		}
		return repoSelector{
			Host:      parsed.Host,
			Workspace: parsed.Workspace,
			Repo:      parsed.RepoSlug,
		}, nil
	}

	switch strings.Count(raw, "/") {
	case 0:
		return repoSelector{Repo: raw}, nil
	case 1:
		parts := strings.SplitN(raw, "/", 2)
		if parts[0] == "" || parts[1] == "" {
			return repoSelector{}, fmt.Errorf("repository must be provided as <repo>, <workspace>/<repo>, or a repository URL")
		}
		return repoSelector{
			Workspace: parts[0],
			Repo:      parts[1],
		}, nil
	default:
		return repoSelector{}, fmt.Errorf("repository must be provided as <repo>, <workspace>/<repo>, or a repository URL")
	}
}

func parsePullRequestSelector(raw string) (pullRequestSelector, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pullRequestSelector{}, fmt.Errorf("pull request reference is required")
	}

	if id, err := strconv.Atoi(raw); err == nil {
		if id <= 0 {
			return pullRequestSelector{}, fmt.Errorf("invalid pull request ID %q", raw)
		}
		return pullRequestSelector{ID: id}, nil
	}

	parsedURL, err := url.Parse(raw)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return pullRequestSelector{}, fmt.Errorf("pull request must be provided as an ID or Bitbucket pull request URL")
	}

	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] != "pull-requests" {
		return pullRequestSelector{}, fmt.Errorf("pull request URL %q must point to a Bitbucket pull request", raw)
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil || id <= 0 {
		return pullRequestSelector{}, fmt.Errorf("pull request URL %q does not contain a valid pull request ID", raw)
	}

	return pullRequestSelector{
		Repo: repoSelector{
			Host:      parsedURL.Hostname(),
			Workspace: parts[0],
			Repo:      strings.TrimSuffix(parts[1], ".git"),
			Explicit:  true,
		},
		ID: id,
	}, nil
}

func parseRepositoryURL(raw string) (repoSelector, error) {
	parsedURL, err := url.Parse(raw)
	if err != nil {
		return repoSelector{}, fmt.Errorf("parse repository URL %q: %w", raw, err)
	}

	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return repoSelector{}, fmt.Errorf("repository URL %q must point to a repository", raw)
	}

	return repoSelector{
		Host:      parsedURL.Hostname(),
		Workspace: parts[0],
		Repo:      strings.TrimSuffix(parts[1], ".git"),
	}, nil
}

func mergeRepoSelectors(base, extra repoSelector) (repoSelector, error) {
	merged := base

	if extra.Host != "" {
		if merged.Host != "" && merged.Host != extra.Host {
			return repoSelector{}, fmt.Errorf("repository host %q does not match %q", merged.Host, extra.Host)
		}
		merged.Host = extra.Host
	}
	if extra.Workspace != "" {
		if merged.Workspace != "" && merged.Workspace != extra.Workspace {
			return repoSelector{}, fmt.Errorf("repository workspace %q does not match %q", merged.Workspace, extra.Workspace)
		}
		merged.Workspace = extra.Workspace
	}
	if extra.Repo != "" {
		if merged.Repo != "" && merged.Repo != extra.Repo {
			return repoSelector{}, fmt.Errorf("repository %q does not match %q", merged.Repo, extra.Repo)
		}
		merged.Repo = extra.Repo
	}
	merged.Explicit = merged.Explicit || extra.Explicit

	return merged, nil
}

func resolveRepoTarget(ctx context.Context, selector repoSelector, client workspaceResolver, allowLocal bool) (resolvedRepoTarget, error) {
	selector.Host = strings.TrimSpace(selector.Host)
	selector.Workspace = strings.TrimSpace(selector.Workspace)
	selector.Repo = strings.TrimSpace(selector.Repo)

	var local *gitrepo.RepoContext
	warnings := make([]string, 0, 1)
	if allowLocal {
		if localRepo, err := resolveLocalRepoContext(ctx); err == nil {
			local = &localRepo
		} else if selector.Repo != "" {
			warnings = append(warnings, localRepoContextWarning(err))
		}
	}

	target := resolvedRepoTarget{
		Warnings:  warnings,
		Host:      selector.Host,
		Repo:      selector.Repo,
		Workspace: selector.Workspace,
		Explicit:  selector.Explicit,
	}

	if selector.Repo != "" && selector.Workspace != "" {
		if local != nil && repoSelectorMatchesLocal(selector, *local) {
			target.LocalRepo = local
			if target.Host == "" {
				target.Host = local.Host
			}
		}
		return target, nil
	}

	if selector.Repo != "" && selector.Workspace == "" {
		if local != nil && local.RepoSlug == selector.Repo {
			target.LocalRepo = local
			target.Workspace = local.Workspace
			if target.Host == "" {
				target.Host = local.Host
			}
			return target, nil
		}

		if client == nil {
			return resolvedRepoTarget{}, fmt.Errorf("repository target %q requires --workspace", selector.Repo)
		}

		workspace, err := resolveWorkspaceForCreate(ctx, client, "")
		if err != nil {
			return resolvedRepoTarget{}, err
		}

		target.Workspace = workspace
		return target, nil
	}

	if local != nil {
		return resolvedRepoTarget{
			LocalRepo: local,
			Warnings:  warnings,
			Host:      coalesce(selector.Host, local.Host),
			Workspace: local.Workspace,
			Repo:      local.RepoSlug,
			Explicit:  false,
		}, nil
	}

	return resolvedRepoTarget{}, fmt.Errorf("could not determine the repository from the current directory; run inside a Bitbucket git checkout or pass --repo")
}

func resolvePullRequestTarget(ctx context.Context, base repoSelector, client workspaceResolver, ref string, allowLocal bool) (resolvedPullRequestTarget, error) {
	prSelector, err := parsePullRequestSelector(ref)
	if err != nil {
		return resolvedPullRequestTarget{}, err
	}

	repoSelector, err := mergeRepoSelectors(base, prSelector.Repo)
	if err != nil {
		return resolvedPullRequestTarget{}, err
	}

	repoTarget, err := resolveRepoTarget(ctx, repoSelector, client, allowLocal)
	if err != nil {
		return resolvedPullRequestTarget{}, err
	}

	return resolvedPullRequestTarget{
		RepoTarget: repoTarget,
		ID:         prSelector.ID,
	}, nil
}

func resolveLocalRepoContext(ctx context.Context) (gitrepo.RepoContext, error) {
	currentDir, err := getWorkingDirectory()
	if err != nil {
		return gitrepo.RepoContext{}, fmt.Errorf("get working directory: %w", err)
	}

	repoContext, err := resolveRepoAtDir(ctx, currentDir)
	if err != nil {
		return gitrepo.RepoContext{}, err
	}

	return repoContext, nil
}

func repoSelectorMatchesLocal(selector repoSelector, local gitrepo.RepoContext) bool {
	if selector.Workspace != local.Workspace || selector.Repo != local.RepoSlug {
		return false
	}

	return selector.Host == "" || selector.Host == local.Host
}

func coalesce(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func localRepoContextWarning(err error) string {
	if err == nil {
		return ""
	}

	return fmt.Sprintf("local repository context unavailable; continuing without local checkout metadata (%v)", err)
}
