package cmd

import (
	"context"

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
