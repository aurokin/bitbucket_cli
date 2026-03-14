package cmd

import (
	"context"
	"fmt"
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

type pullRequestCommentSelector struct {
	PR        pullRequestSelector
	CommentID int
}

type resolvedPullRequestCommentTarget struct {
	PRTarget  resolvedPullRequestTarget
	CommentID int
}

func resolveRepoTarget(ctx context.Context, selector repoSelector, client workspaceResolver, allowLocal bool) (resolvedRepoTarget, error) {
	selector.Host = strings.TrimSpace(selector.Host)
	selector.Workspace = strings.TrimSpace(selector.Workspace)
	selector.Repo = strings.TrimSpace(selector.Repo)

	local, warnings := resolveRepoTargetLocalContext(ctx, selector, allowLocal)

	target := resolvedRepoTarget{
		Warnings:  warnings,
		Host:      selector.Host,
		Repo:      selector.Repo,
		Workspace: selector.Workspace,
		Explicit:  selector.Explicit,
	}

	if selector.Repo != "" && selector.Workspace != "" {
		return resolveExplicitRepoTarget(target, selector, local), nil
	}

	if selector.Repo != "" && selector.Workspace == "" {
		return resolveBareRepoTarget(ctx, target, selector, client, local)
	}

	if local != nil {
		return resolveLocalOnlyRepoTarget(selector, warnings, local), nil
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

func resolvePullRequestCommentTarget(ctx context.Context, base repoSelector, client workspaceResolver, prRef, commentRef string, allowLocal bool) (resolvedPullRequestCommentTarget, error) {
	commentSelector, err := parsePullRequestCommentSelector(commentRef)
	if err != nil {
		return resolvedPullRequestCommentTarget{}, err
	}

	if commentSelector.PR.ID > 0 {
		repoSelector, err := mergeRepoSelectors(base, commentSelector.PR.Repo)
		if err != nil {
			return resolvedPullRequestCommentTarget{}, err
		}
		prTarget, err := resolvePullRequestTarget(ctx, repoSelector, client, strconv.Itoa(commentSelector.PR.ID), allowLocal)
		if err != nil {
			return resolvedPullRequestCommentTarget{}, err
		}
		if strings.TrimSpace(prRef) != "" {
			explicitPRTarget, err := resolvePullRequestTarget(ctx, base, client, prRef, allowLocal)
			if err != nil {
				return resolvedPullRequestCommentTarget{}, err
			}
			if explicitPRTarget.ID != prTarget.ID ||
				explicitPRTarget.RepoTarget.Workspace != prTarget.RepoTarget.Workspace ||
				explicitPRTarget.RepoTarget.Repo != prTarget.RepoTarget.Repo {
				return resolvedPullRequestCommentTarget{}, fmt.Errorf("--pr %q does not match comment target %q", prRef, commentRef)
			}
		}
		return resolvedPullRequestCommentTarget{
			PRTarget:  prTarget,
			CommentID: commentSelector.CommentID,
		}, nil
	}

	if strings.TrimSpace(prRef) == "" {
		return resolvedPullRequestCommentTarget{}, fmt.Errorf("pull request comment ID %d requires --pr <id-or-url>", commentSelector.CommentID)
	}

	prTarget, err := resolvePullRequestTarget(ctx, base, client, prRef, allowLocal)
	if err != nil {
		return resolvedPullRequestCommentTarget{}, err
	}

	return resolvedPullRequestCommentTarget{
		PRTarget:  prTarget,
		CommentID: commentSelector.CommentID,
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

func resolveRepoTargetLocalContext(ctx context.Context, selector repoSelector, allowLocal bool) (*gitrepo.RepoContext, []string) {
	var local *gitrepo.RepoContext
	warnings := make([]string, 0, 1)
	if !allowLocal {
		return nil, warnings
	}
	if localRepo, err := resolveLocalRepoContext(ctx); err == nil {
		local = &localRepo
	} else if selector.Repo != "" {
		warnings = append(warnings, localRepoContextWarning(err))
	}
	return local, warnings
}

func resolveExplicitRepoTarget(target resolvedRepoTarget, selector repoSelector, local *gitrepo.RepoContext) resolvedRepoTarget {
	if local != nil && repoSelectorMatchesLocal(selector, *local) {
		target.LocalRepo = local
		if target.Host == "" {
			target.Host = local.Host
		}
	}
	return target
}

func resolveBareRepoTarget(ctx context.Context, target resolvedRepoTarget, selector repoSelector, client workspaceResolver, local *gitrepo.RepoContext) (resolvedRepoTarget, error) {
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

func resolveLocalOnlyRepoTarget(selector repoSelector, warnings []string, local *gitrepo.RepoContext) resolvedRepoTarget {
	return resolvedRepoTarget{
		LocalRepo: local,
		Warnings:  warnings,
		Host:      coalesce(selector.Host, local.Host),
		Workspace: local.Workspace,
		Repo:      local.RepoSlug,
		Explicit:  false,
	}
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
