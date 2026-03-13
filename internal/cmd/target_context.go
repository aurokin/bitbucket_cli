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
