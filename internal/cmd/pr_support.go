package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func prViewNextStep(workspace, repo string, id int) string {
	return fmt.Sprintf("bb pr diff %d --repo %s/%s", id, workspace, repo)
}

func resolveMergeStrategy(pr bitbucket.PullRequest, requested string) (string, error) {
	available := uniqueNonEmptyStrings(pr.Destination.Branch.MergeStrategies)

	requested = strings.TrimSpace(requested)
	if requested != "" {
		if len(available) > 0 && !stringSliceContains(available, requested) {
			return "", fmt.Errorf("merge strategy %q is not allowed for destination branch %s; available: %s", requested, pr.Destination.Branch.Name, strings.Join(available, ", "))
		}
		return requested, nil
	}

	if defaultStrategy := strings.TrimSpace(pr.Destination.Branch.DefaultMergeStrategy); defaultStrategy != "" {
		return defaultStrategy, nil
	}

	if len(available) == 1 {
		return available[0], nil
	}
	if len(available) > 1 {
		return "", fmt.Errorf("multiple merge strategies are available for destination branch %s; pass --strategy (%s)", pr.Destination.Branch.Name, strings.Join(available, ", "))
	}

	return "", nil
}

type prStatusPayload struct {
	Host               string                  `json:"host"`
	Workspace          string                  `json:"workspace"`
	Repo               string                  `json:"repo"`
	Warnings           []string                `json:"warnings,omitempty"`
	CurrentUser        bitbucket.CurrentUser   `json:"current_user"`
	CurrentBranchName  string                  `json:"current_branch_name,omitempty"`
	CurrentBranchError string                  `json:"current_branch_error,omitempty"`
	CurrentBranch      *bitbucket.PullRequest  `json:"current_branch,omitempty"`
	Created            []bitbucket.PullRequest `json:"created"`
	ReviewRequested    []bitbucket.PullRequest `json:"review_requested"`
}

type prDiffPayload struct {
	Host      string                          `json:"host"`
	Workspace string                          `json:"workspace"`
	Repo      string                          `json:"repo"`
	Warnings  []string                        `json:"warnings,omitempty"`
	ID        int                             `json:"id"`
	Title     string                          `json:"title"`
	Patch     string                          `json:"patch"`
	Stats     []bitbucket.PullRequestDiffStat `json:"stats"`
}

func resolveCommentBody(stdin io.Reader, body, bodyFile string) (string, error) {
	if trimmed := strings.TrimSpace(body); trimmed != "" {
		return trimmed, nil
	}

	if strings.TrimSpace(bodyFile) != "" {
		data, err := readRequestBody(stdin, bodyFile)
		if err != nil {
			return "", err
		}
		trimmed := strings.TrimSpace(string(data))
		if trimmed == "" {
			return "", fmt.Errorf("comment body is empty")
		}
		return trimmed, nil
	}

	return "", fmt.Errorf("provide a comment body with --body or --body-file")
}

func sameActor(user bitbucket.CurrentUser, actor bitbucket.PullRequestActor) bool {
	switch {
	case user.AccountID != "" && actor.AccountID != "":
		return user.AccountID == actor.AccountID
	case user.Username != "" && actor.Nickname != "":
		return user.Username == actor.Nickname
	case user.DisplayName != "" && actor.DisplayName != "":
		return user.DisplayName == actor.DisplayName
	default:
		return false
	}
}

func reviewRequestedFromUser(user bitbucket.CurrentUser, pr bitbucket.PullRequest) bool {
	for _, reviewer := range pr.Reviewers {
		if sameActor(user, reviewer) {
			return true
		}
	}
	return false
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
