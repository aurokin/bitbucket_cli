package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/spf13/cobra"
)

func resolveSourceBranch(source string) (string, error) {
	if source != "" {
		return source, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	branch, err := gitrepo.CurrentBranch(context.Background(), currentDir)
	if err != nil {
		return "", fmt.Errorf("resolve source branch: %w", err)
	}
	if branch == "" {
		return "", fmt.Errorf("could not determine current branch; pass --source")
	}

	return branch, nil
}

func resolveSourceBranchInput(cmd *cobra.Command, source string, interactive bool, explicitRepoSelector bool, workspace, repo string) (string, error) {
	if source != "" {
		return source, nil
	}

	if explicitRepoSelector {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}

		localRepo, err := gitrepo.ResolveRepoContext(context.Background(), currentDir)
		if err != nil || localRepo.Workspace != workspace || localRepo.RepoSlug != repo {
			if interactive {
				return promptRequiredString(cmd, "Source branch", "")
			}
			return "", fmt.Errorf("could not determine the source branch for %s/%s from the current directory; pass --source or run in an interactive terminal", workspace, repo)
		}
	}

	defaultSource, err := resolveSourceBranch(source)
	if err == nil {
		if interactive {
			return promptRequiredString(cmd, "Source branch", defaultSource)
		}
		return defaultSource, nil
	}

	if interactive {
		return promptRequiredString(cmd, "Source branch", "")
	}

	return "", fmt.Errorf("could not determine the source branch; pass --source or run in an interactive terminal")
}

func resolveDestinationBranch(ctx context.Context, client *bitbucket.Client, workspace, repo, destination string) (string, error) {
	if destination != "" {
		return destination, nil
	}

	repository, err := client.GetRepository(ctx, workspace, repo)
	if err != nil {
		return "", err
	}
	if repository.MainBranch.Name == "" {
		return "", fmt.Errorf("repository main branch is unknown; pass --destination")
	}

	return repository.MainBranch.Name, nil
}

func resolveDestinationBranchInput(cmd *cobra.Command, client *bitbucket.Client, workspace, repo, destination string, interactive bool) (string, error) {
	if destination != "" {
		return destination, nil
	}

	defaultDestination, err := resolveDestinationBranch(context.Background(), client, workspace, repo, "")
	if err == nil {
		if interactive {
			return promptRequiredString(cmd, "Destination branch", defaultDestination)
		}
		return defaultDestination, nil
	}

	if interactive {
		return promptRequiredString(cmd, "Destination branch", "")
	}

	return "", fmt.Errorf("could not determine the destination branch; pass --destination or run in an interactive terminal")
}

func defaultPRTitle(sourceBranch string) string {
	return sourceBranch
}
