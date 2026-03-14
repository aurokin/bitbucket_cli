package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func resolveIssueTarget(host, workspace, repo string) (resolvedRepoTarget, *bitbucket.Client, error) {
	resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
	if err != nil {
		return resolvedRepoTarget{}, nil, err
	}

	return resolved.Target, resolved.Client, nil
}

func resolveIssueTargetAndID(host, workspace, repo, rawID string) (resolvedRepoTarget, *bitbucket.Client, int, error) {
	target, client, issueID, err := resolveIssueReference(host, workspace, repo, rawID)
	if err != nil {
		return resolvedRepoTarget{}, nil, 0, err
	}
	return target, client, issueID, nil
}

func resolveIssueReference(host, workspace, repo, raw string) (resolvedRepoTarget, *bitbucket.Client, int, error) {
	raw = strings.TrimSpace(raw)
	if issueID, err := parsePositiveInt("issue", raw); err == nil {
		target, client, err := resolveIssueTarget(host, workspace, repo)
		if err != nil {
			return resolvedRepoTarget{}, nil, 0, err
		}
		return target, client, issueID, nil
	}

	entity, err := parseBitbucketEntityURL(raw)
	if err != nil || entity.Type != "issue" {
		return resolvedRepoTarget{}, nil, 0, fmt.Errorf("issue must be provided as an ID or Bitbucket issue URL")
	}

	resolved, err := resolveRepoCommandTarget(context.Background(), entity.Host, entity.Workspace, entity.Workspace+"/"+entity.Repo, false)
	if err != nil {
		return resolvedRepoTarget{}, nil, 0, err
	}

	if strings.TrimSpace(repo) != "" || strings.TrimSpace(workspace) != "" || strings.TrimSpace(host) != "" {
		explicit, _, err := resolveIssueTarget(host, workspace, repo)
		if err != nil {
			return resolvedRepoTarget{}, nil, 0, err
		}
		if explicit.Host != resolved.Target.Host || explicit.Workspace != resolved.Target.Workspace || explicit.Repo != resolved.Target.Repo {
			return resolvedRepoTarget{}, nil, 0, fmt.Errorf("issue URL %q does not match the explicit repository target", raw)
		}
	}

	return resolved.Target, resolved.Client, entity.Issue, nil
}

func writeIssueMutationSummary(w io.Writer, action, workspace, repo string, issue bitbucket.Issue, includeNext bool) error {
	if _, err := fmt.Fprintf(w, "%s issue %s/%s#%d: %s\n", action, workspace, repo, issue.ID, issue.Title); err != nil {
		return err
	}
	if issue.State != "" {
		if _, err := fmt.Fprintf(w, "State: %s\n", issue.State); err != nil {
			return err
		}
	}
	if issue.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(w, "URL: %s\n", issue.Links.HTML.Href); err != nil {
			return err
		}
	}
	if includeNext {
		if _, err := fmt.Fprintf(w, "Next: bb issue view %d --repo %s/%s\n", issue.ID, workspace, repo); err != nil {
			return err
		}
	}
	return nil
}

func writeIssueListSummary(w io.Writer, target resolvedRepoTarget, issues []bitbucket.Issue) error {
	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	if len(issues) == 0 {
		if _, err := fmt.Fprintf(w, "No issues found for %s/%s.\n", target.Workspace, target.Repo); err != nil {
			return err
		}
		return writeNextStep(w, issueListEmptyNextStep(target.Workspace, target.Repo))
	}
	return writeIssueTable(w, issues)
}

func writeIssueViewSummary(w io.Writer, target resolvedRepoTarget, issue bitbucket.Issue) error {
	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintf(tw, "ID:\t%d\n", issue.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tw, "Title:\t%s\n", issue.Title); err != nil {
		return err
	}
	if issue.State != "" {
		if _, err := fmt.Fprintf(tw, "State:\t%s\n", issue.State); err != nil {
			return err
		}
	}
	if issue.Kind != "" {
		if _, err := fmt.Fprintf(tw, "Kind:\t%s\n", issue.Kind); err != nil {
			return err
		}
	}
	if issue.Priority != "" {
		if _, err := fmt.Fprintf(tw, "Priority:\t%s\n", issue.Priority); err != nil {
			return err
		}
	}
	if issue.Reporter.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Reporter:\t%s\n", issue.Reporter.DisplayName); err != nil {
			return err
		}
	}
	if issue.Assignee.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Assignee:\t%s\n", issue.Assignee.DisplayName); err != nil {
			return err
		}
	}
	if issue.UpdatedOn != "" {
		if _, err := fmt.Fprintf(tw, "Updated:\t%s\n", issue.UpdatedOn); err != nil {
			return err
		}
	}
	if issue.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(tw, "URL:\t%s\n", issue.Links.HTML.Href); err != nil {
			return err
		}
	}
	if issue.Content.Raw != "" {
		if _, err := fmt.Fprintf(tw, "Body:\t%s\n", issue.Content.Raw); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, issueViewNextStep(target.Workspace, target.Repo, issue.ID))
}

func issueListEmptyNextStep(workspace, repo string) string {
	return fmt.Sprintf("bb issue create --repo %s/%s --title '<title>'", workspace, repo)
}

func issueViewNextStep(workspace, repo string, id int) string {
	return fmt.Sprintf("bb issue edit %d --repo %s/%s", id, workspace, repo)
}

func writeIssueTable(w io.Writer, issues []bitbucket.Issue) error {
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "#\ttitle\tstate\treporter\tupdated"); err != nil {
		return err
	}
	for _, issue := range issues {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\n",
			issue.ID,
			output.Truncate(issue.Title, 40),
			output.Truncate(issue.State, 12),
			output.Truncate(issue.Reporter.DisplayName, 16),
			formatPRUpdated(issue.UpdatedOn),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
