package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/aurokin/bitbucket_cli/internal/output"
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

func prViewNextStep(workspace, repo string, id int) string {
	return fmt.Sprintf("bb pr diff %d --repo %s/%s", id, workspace, repo)
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
	ID        int                             `json:"id"`
	Title     string                          `json:"title"`
	Patch     string                          `json:"patch"`
	Stats     []bitbucket.PullRequestDiffStat `json:"stats"`
}

func buildPRStatusPayload(target resolvedRepoTarget, currentUser bitbucket.CurrentUser, currentBranch, currentBranchError string, prs []bitbucket.PullRequest) prStatusPayload {
	payload := prStatusPayload{
		Host:               target.Host,
		Workspace:          target.Workspace,
		Repo:               target.Repo,
		Warnings:           append([]string(nil), target.Warnings...),
		CurrentUser:        currentUser,
		CurrentBranchName:  currentBranch,
		CurrentBranchError: currentBranchError,
		Created:            make([]bitbucket.PullRequest, 0),
		ReviewRequested:    make([]bitbucket.PullRequest, 0),
	}

	currentBranchID := 0
	for i := range prs {
		pr := prs[i]
		if payload.CurrentBranch == nil && currentBranch != "" && pr.Source.Branch.Name == currentBranch {
			prCopy := pr
			payload.CurrentBranch = &prCopy
			currentBranchID = pr.ID
			continue
		}
	}

	for _, pr := range prs {
		if currentBranchID != 0 && pr.ID == currentBranchID {
			continue
		}
		if sameActor(currentUser, pr.Author) {
			payload.Created = append(payload.Created, pr)
			continue
		}
		if reviewRequestedFromUser(currentUser, pr) {
			payload.ReviewRequested = append(payload.ReviewRequested, pr)
		}
	}

	return payload
}

func writePRStatusSummary(w io.Writer, payload prStatusPayload) error {
	if _, err := fmt.Fprintf(w, "Repository: %s/%s\n", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}

	if payload.CurrentBranchName != "" {
		if _, err := fmt.Fprintf(w, "Current Branch: %s\n", payload.CurrentBranchName); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(w, "Current Branch: unavailable"); err != nil {
		return err
	}
	if payload.CurrentBranchError != "" {
		if err := writeLabelValue(w, "Current Branch Error", payload.CurrentBranchError); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Current Branch Pull Request"); err != nil {
		return err
	}
	currentBranchPRs := make([]bitbucket.PullRequest, 0, 1)
	if payload.CurrentBranch != nil {
		currentBranchPRs = append(currentBranchPRs, *payload.CurrentBranch)
	}
	if err := writePRStatusSection(w, currentBranchPRs...); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Created By You"); err != nil {
		return err
	}
	if err := writePRStatusSection(w, payload.Created...); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Review Requested"); err != nil {
		return err
	}
	if err := writePRStatusSection(w, payload.ReviewRequested...); err != nil {
		return err
	}
	if len(payload.Created) == 0 && len(payload.ReviewRequested) == 0 && payload.CurrentBranch == nil {
		return writeNextStep(w, fmt.Sprintf("bb pr list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	return nil
}

func writePRStatusSection(w io.Writer, prs ...bitbucket.PullRequest) error {
	if len(prs) == 0 {
		_, err := fmt.Fprintln(w, "  None.")
		return err
	}

	for _, pr := range prs {
		line := fmt.Sprintf("  #%d  %s [%s] %s -> %s", pr.ID, pr.Title, pr.State, pr.Source.Branch.Name, pr.Destination.Branch.Name)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

func writePRDiffStatTable(w io.Writer, stats []bitbucket.PullRequestDiffStat) error {
	if len(stats) == 0 {
		_, err := fmt.Fprintln(w, "No changed files.")
		return err
	}

	pathWidth := diffStatPathWidth(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "status\tfile\t+add\t-rem"); err != nil {
		return err
	}

	totalAdded := 0
	totalRemoved := 0
	for _, stat := range stats {
		totalAdded += stat.LinesAdded
		totalRemoved += stat.LinesRemoved
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%d\n", output.Truncate(diffStatus(stat), 10), output.TruncateMiddle(diffPath(stat), pathWidth), stat.LinesAdded, stat.LinesRemoved); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "total\t%d files\t%d\t%d\n", len(stats), totalAdded, totalRemoved); err != nil {
		return err
	}

	return tw.Flush()
}

func writePRListTable(w io.Writer, prs []bitbucket.PullRequest) error {
	titleWidth, authorWidth, branchWidth := prListColumnWidths(output.TerminalWidth(w))

	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "#\ttitle\tstate\tauthor\tsrc\tdst\tupdated"); err != nil {
		return err
	}

	for _, pr := range prs {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			pr.ID,
			output.Truncate(pr.Title, titleWidth),
			output.Truncate(pr.State, 10),
			output.Truncate(pr.Author.DisplayName, authorWidth),
			output.TruncateMiddle(pr.Source.Branch.Name, branchWidth),
			output.TruncateMiddle(pr.Destination.Branch.Name, branchWidth),
			formatPRUpdated(pr.UpdatedOn),
		); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func formatPRUpdated(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}

	return parsed.Local().Format("2006-01-02 15:04")
}

func prListColumnWidths(termWidth int) (title, author, branch int) {
	switch {
	case termWidth >= 160:
		return 52, 18, 24
	case termWidth >= 132:
		return 40, 16, 18
	case termWidth >= 110:
		return 32, 14, 14
	default:
		return 24, 12, 12
	}
}

func diffStatPathWidth(termWidth int) int {
	switch {
	case termWidth >= 160:
		return 72
	case termWidth >= 132:
		return 56
	case termWidth >= 110:
		return 44
	default:
		return 32
	}
}

func diffPath(stat bitbucket.PullRequestDiffStat) string {
	switch {
	case stat.New != nil && stat.Old != nil && stat.New.Path != "" && stat.Old.Path != "" && stat.New.Path != stat.Old.Path:
		return stat.Old.Path + " -> " + stat.New.Path
	case stat.New != nil && stat.New.Path != "":
		return stat.New.Path
	case stat.Old != nil && stat.Old.Path != "":
		return stat.Old.Path
	default:
		return "(unknown)"
	}
}

func diffStatus(stat bitbucket.PullRequestDiffStat) string {
	status := strings.TrimSpace(stat.Status)
	if status == "" {
		return "changed"
	}
	return status
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

func defaultPRTitle(sourceBranch string) string {
	return sourceBranch
}

type pullRequestSummaryOptions struct {
	IncludeAuthor      bool
	IncludeUpdated     bool
	IncludeDescription bool
	Strategy           string
	MergeCommit        string
}

func writePullRequestSummaryTable(w io.Writer, pr bitbucket.PullRequest, options pullRequestSummaryOptions) error {
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintf(tw, "ID:\t%d\n", pr.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tw, "Title:\t%s\n", pr.Title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tw, "State:\t%s\n", pr.State); err != nil {
		return err
	}
	if options.IncludeAuthor && pr.Author.DisplayName != "" {
		if _, err := fmt.Fprintf(tw, "Author:\t%s\n", pr.Author.DisplayName); err != nil {
			return err
		}
	}
	if options.Strategy != "" {
		if _, err := fmt.Fprintf(tw, "Strategy:\t%s\n", options.Strategy); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tw, "Source:\t%s\n", pr.Source.Branch.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tw, "Destination:\t%s\n", pr.Destination.Branch.Name); err != nil {
		return err
	}
	if options.IncludeUpdated && pr.UpdatedOn != "" {
		if _, err := fmt.Fprintf(tw, "Updated:\t%s\n", pr.UpdatedOn); err != nil {
			return err
		}
	}
	if options.MergeCommit != "" {
		if _, err := fmt.Fprintf(tw, "Merge Commit:\t%s\n", options.MergeCommit); err != nil {
			return err
		}
	}
	if pr.Links.HTML.Href != "" {
		if _, err := fmt.Fprintf(tw, "URL:\t%s\n", pr.Links.HTML.Href); err != nil {
			return err
		}
	}
	if options.IncludeDescription && pr.Description != "" {
		if _, err := fmt.Fprintf(tw, "Description:\t%s\n", pr.Description); err != nil {
			return err
		}
	}
	return tw.Flush()
}
