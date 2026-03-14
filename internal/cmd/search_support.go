package cmd

import (
	"fmt"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
)

func searchReposNextStep(workspace string, repos []bitbucket.Repository) string {
	if len(repos) == 1 {
		return fmt.Sprintf("bb repo view --repo %s/%s", workspace, repos[0].Slug)
	}
	if len(repos) > 1 {
		return fmt.Sprintf("bb repo view --repo %s/<repo>", workspace)
	}
	return fmt.Sprintf("bb repo create %s/<repo>", workspace)
}

func writeSearchRepoSummary(w io.Writer, workspace, query string, repos []bitbucket.Repository) error {
	if len(repos) == 0 {
		if _, err := fmt.Fprintf(w, "No repositories found in %s for %q.\n", workspace, query); err != nil {
			return err
		}
		return writeNextStep(w, searchReposNextStep(workspace, repos))
	}

	if err := writeLabelValue(w, "Workspace", workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", query); err != nil {
		return err
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\tslug\tprivate\tproject\tupdated"); err != nil {
		return err
	}
	for _, repo := range repos {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%t\t%s\t%s\n",
			output.Truncate(repo.Name, 32),
			output.Truncate(repo.Slug, 24),
			repo.IsPrivate,
			output.Truncate(repo.Project.Key, 12),
			formatPRUpdated(repo.UpdatedOn),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, searchReposNextStep(workspace, repos))
}

func searchPRsNextStep(workspace, repo string, prs []bitbucket.PullRequest) string {
	if len(prs) == 1 {
		return fmt.Sprintf("bb pr view %d --repo %s/%s", prs[0].ID, workspace, repo)
	}
	if len(prs) > 1 {
		return fmt.Sprintf("bb pr view <id> --repo %s/%s", workspace, repo)
	}
	return fmt.Sprintf("bb pr list --repo %s/%s", workspace, repo)
}

func writeSearchPRSummary(w io.Writer, target resolvedRepoTarget, query string, prs []bitbucket.PullRequest) error {
	if len(prs) == 0 {
		if err := writeWarnings(w, target.Warnings); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "No pull requests found for %s/%s matching %q.\n", target.Workspace, target.Repo, query); err != nil {
			return err
		}
		return writeNextStep(w, searchPRsNextStep(target.Workspace, target.Repo, prs))
	}
	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", query); err != nil {
		return err
	}
	if err := writePRListTable(w, prs); err != nil {
		return err
	}
	return writeNextStep(w, searchPRsNextStep(target.Workspace, target.Repo, prs))
}

func searchIssuesNextStep(workspace, repo string, issues []bitbucket.Issue) string {
	if len(issues) == 1 {
		return fmt.Sprintf("bb issue view %d --repo %s/%s", issues[0].ID, workspace, repo)
	}
	if len(issues) > 1 {
		return fmt.Sprintf("bb issue view <id> --repo %s/%s", workspace, repo)
	}
	return fmt.Sprintf("bb issue list --repo %s/%s", workspace, repo)
}

func writeSearchIssueSummary(w io.Writer, target resolvedRepoTarget, query string, issues []bitbucket.Issue) error {
	if len(issues) == 0 {
		if err := writeWarnings(w, target.Warnings); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "No issues found for %s/%s matching %q.\n", target.Workspace, target.Repo, query); err != nil {
			return err
		}
		return writeNextStep(w, searchIssuesNextStep(target.Workspace, target.Repo, issues))
	}

	if err := writeTargetHeader(w, "Repository", target.Workspace, target.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, target.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", query); err != nil {
		return err
	}
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
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, searchIssuesNextStep(target.Workspace, target.Repo, issues))
}
