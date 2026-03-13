package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

type resolvedRepoCommandTarget struct {
	Client *bitbucket.Client
	Target resolvedRepoTarget
}

type resolvedPullRequestCommandTarget struct {
	Client *bitbucket.Client
	Target resolvedPullRequestTarget
}

type resolvedPullRequestCommentCommandTarget struct {
	Client *bitbucket.Client
	Target resolvedPullRequestCommentTarget
}

type resolvedCommitTarget struct {
	RepoTarget resolvedRepoTarget
	Commit     string
}

type resolvedCommitCommandTarget struct {
	Client *bitbucket.Client
	Target resolvedCommitTarget
}

type resolvedCommitCommentTarget struct {
	CommitTarget resolvedCommitTarget
	CommentID    int
}

type resolvedCommitCommentCommandTarget struct {
	Client *bitbucket.Client
	Target resolvedCommitCommentTarget
}

type resolvedPullRequestTaskTarget struct {
	PRTarget resolvedPullRequestTarget
	TaskID   int
}

type resolvedPullRequestTaskCommandTarget struct {
	Client *bitbucket.Client
	Target resolvedPullRequestTaskTarget
}

func resolveRepoCommandTarget(ctx context.Context, host, workspace, repo string, allowLocal bool) (resolvedRepoCommandTarget, error) {
	selector, err := parseRepoSelector(host, workspace, repo)
	if err != nil {
		return resolvedRepoCommandTarget{}, err
	}

	resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
	if err != nil {
		return resolvedRepoCommandTarget{}, err
	}

	selector.Host = resolvedHost
	target, err := resolveRepoTarget(ctx, selector, client, allowLocal)
	if err != nil {
		return resolvedRepoCommandTarget{}, err
	}

	return resolvedRepoCommandTarget{
		Client: client,
		Target: target,
	}, nil
}

func resolveRepoCommandTargetInput(ctx context.Context, host, workspace, repo, positional string, allowLocal bool) (resolvedRepoCommandTarget, error) {
	selector, err := parseRepoTargetInput(host, workspace, repo, positional)
	if err != nil {
		return resolvedRepoCommandTarget{}, err
	}
	if err := requireExplicitRepoTarget(selector); err != nil {
		return resolvedRepoCommandTarget{}, err
	}

	resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
	if err != nil {
		return resolvedRepoCommandTarget{}, err
	}

	selector.Host = resolvedHost
	target, err := resolveRepoTarget(ctx, selector, client, allowLocal)
	if err != nil {
		return resolvedRepoCommandTarget{}, err
	}

	return resolvedRepoCommandTarget{
		Client: client,
		Target: target,
	}, nil
}

func resolvePullRequestCommandTarget(ctx context.Context, host, workspace, repo, ref string, allowLocal bool) (resolvedPullRequestCommandTarget, error) {
	selector, err := parseRepoSelector(host, workspace, repo)
	if err != nil {
		return resolvedPullRequestCommandTarget{}, err
	}

	resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
	if err != nil {
		return resolvedPullRequestCommandTarget{}, err
	}

	selector.Host = resolvedHost
	target, err := resolvePullRequestTarget(ctx, selector, client, ref, allowLocal)
	if err != nil {
		return resolvedPullRequestCommandTarget{}, err
	}

	return resolvedPullRequestCommandTarget{
		Client: client,
		Target: target,
	}, nil
}

func resolvePullRequestCommentCommandTarget(ctx context.Context, host, workspace, repo, prRef, commentRef string, allowLocal bool) (resolvedPullRequestCommentCommandTarget, error) {
	selector, err := parseRepoSelector(host, workspace, repo)
	if err != nil {
		return resolvedPullRequestCommentCommandTarget{}, err
	}

	resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
	if err != nil {
		return resolvedPullRequestCommentCommandTarget{}, err
	}

	selector.Host = resolvedHost
	target, err := resolvePullRequestCommentTarget(ctx, selector, client, prRef, commentRef, allowLocal)
	if err != nil {
		return resolvedPullRequestCommentCommandTarget{}, err
	}

	return resolvedPullRequestCommentCommandTarget{
		Client: client,
		Target: target,
	}, nil
}

func resolveCommitCommandTarget(ctx context.Context, host, workspace, repo, ref string, allowLocal bool) (resolvedCommitCommandTarget, error) {
	selector, err := parseRepoSelector(host, workspace, repo)
	if err != nil {
		return resolvedCommitCommandTarget{}, err
	}

	commitRef := strings.TrimSpace(ref)
	if commitRef == "" {
		return resolvedCommitCommandTarget{}, fmt.Errorf("commit reference is required")
	}

	if entity, err := parseBitbucketEntityURL(commitRef); err == nil {
		if entity.Type != "commit" {
			return resolvedCommitCommandTarget{}, fmt.Errorf("commit must be provided as a commit SHA or commit URL")
		}
		selector, err = mergeRepoSelectors(selector, repoSelector{
			Host:      entity.Host,
			Workspace: entity.Workspace,
			Repo:      entity.Repo,
			Explicit:  true,
		})
		if err != nil {
			return resolvedCommitCommandTarget{}, err
		}
		commitRef = entity.Commit
	}

	resolvedHost, client, err := resolveAuthenticatedClient(selector.Host)
	if err != nil {
		return resolvedCommitCommandTarget{}, err
	}

	selector.Host = resolvedHost
	repoTarget, err := resolveRepoTarget(ctx, selector, client, allowLocal)
	if err != nil {
		return resolvedCommitCommandTarget{}, err
	}

	return resolvedCommitCommandTarget{
		Client: client,
		Target: resolvedCommitTarget{
			RepoTarget: repoTarget,
			Commit:     commitRef,
		},
	}, nil
}

func resolveCommitCommentCommandTarget(ctx context.Context, host, workspace, repo, commitRef, commentRef string, allowLocal bool) (resolvedCommitCommentCommandTarget, error) {
	commentID, err := parsePositiveID(commentRef, "commit comment")
	if err != nil {
		return resolvedCommitCommentCommandTarget{}, fmt.Errorf("commit comment must be provided as a numeric comment ID")
	}

	resolved, err := resolveCommitCommandTarget(ctx, host, workspace, repo, commitRef, allowLocal)
	if err != nil {
		return resolvedCommitCommentCommandTarget{}, err
	}

	return resolvedCommitCommentCommandTarget{
		Client: resolved.Client,
		Target: resolvedCommitCommentTarget{
			CommitTarget: resolved.Target,
			CommentID:    commentID,
		},
	}, nil
}

func resolvePullRequestTaskCommandTarget(ctx context.Context, host, workspace, repo, prRef, taskRef string, allowLocal bool) (resolvedPullRequestTaskCommandTarget, error) {
	taskID, err := parsePullRequestTaskID(taskRef)
	if err != nil {
		return resolvedPullRequestTaskCommandTarget{}, err
	}

	resolved, err := resolvePullRequestCommandTarget(ctx, host, workspace, repo, prRef, allowLocal)
	if err != nil {
		return resolvedPullRequestTaskCommandTarget{}, err
	}

	return resolvedPullRequestTaskCommandTarget{
		Client: resolved.Client,
		Target: resolvedPullRequestTaskTarget{
			PRTarget: resolved.Target,
			TaskID:   taskID,
		},
	}, nil
}

func parsePullRequestTaskID(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("pull request task reference is required")
	}
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("pull request task must be provided as a numeric task ID")
	}
	return id, nil
}
