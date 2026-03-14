package cmd

import (
	"fmt"
	"strconv"
	"strings"

	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
)

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

	entity, err := parseBitbucketEntityURL(raw)
	if err != nil {
		return pullRequestSelector{}, fmt.Errorf("pull request must be provided as an ID or Bitbucket pull request URL")
	}
	if entity.Type != "pull-request" && entity.Type != "pull-request-comment" {
		return pullRequestSelector{}, fmt.Errorf("pull request URL %q must point to a Bitbucket pull request", raw)
	}

	return pullRequestSelector{
		Repo: repoSelector{
			Host:      entity.Host,
			Workspace: entity.Workspace,
			Repo:      entity.Repo,
			Explicit:  true,
		},
		ID: entity.PR,
	}, nil
}

func parsePullRequestCommentSelector(raw string) (pullRequestCommentSelector, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pullRequestCommentSelector{}, fmt.Errorf("pull request comment reference is required")
	}

	if id, err := strconv.Atoi(raw); err == nil {
		if id <= 0 {
			return pullRequestCommentSelector{}, fmt.Errorf("invalid pull request comment ID %q", raw)
		}
		return pullRequestCommentSelector{CommentID: id}, nil
	}

	entity, err := parseBitbucketEntityURL(raw)
	if err != nil {
		return pullRequestCommentSelector{}, fmt.Errorf("pull request comment must be provided as an ID or Bitbucket pull request comment URL")
	}
	if entity.Type != "pull-request-comment" {
		return pullRequestCommentSelector{}, fmt.Errorf("pull request comment URL %q must point to a Bitbucket pull request comment", raw)
	}

	return pullRequestCommentSelector{
		PR: pullRequestSelector{
			Repo: repoSelector{
				Host:      entity.Host,
				Workspace: entity.Workspace,
				Repo:      entity.Repo,
				Explicit:  true,
			},
			ID: entity.PR,
		},
		CommentID: entity.Comment,
	}, nil
}

func parseRepositoryURL(raw string) (repoSelector, error) {
	entity, err := parseBitbucketEntityURL(raw)
	if err != nil {
		return repoSelector{}, fmt.Errorf("parse repository URL %q: %w", raw, err)
	}
	if entity.Type != "repository" {
		return repoSelector{}, fmt.Errorf("repository URL %q must point to a repository", raw)
	}

	return repoSelector{
		Host:      entity.Host,
		Workspace: entity.Workspace,
		Repo:      entity.Repo,
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
