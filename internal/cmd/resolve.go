package cmd

import (
	"io"
	"strconv"

	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newResolveCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "resolve <url>",
		Short: "Resolve a Bitbucket URL into a structured entity",
		Long:  "Resolve Bitbucket repository, pull request, pull request comment, issue, commit, and source URLs into a structured entity payload without making an API request.",
		Example: "  bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7\n" +
			"  bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'\n" +
			"  bb resolve https://bitbucket.org/workspace-slug/repo-slug/src/main/README.md#lines-12 --json type,repo,path,line",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			entity, err := parseBitbucketEntityURL(args[0])
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, entity, func(w io.Writer) error {
				return writeResolveSummary(w, entity)
			})
		},
	}

	addFormatFlags(cmd, &flags)

	return cmd
}

func nextResolveCommand(entity resolvedEntity) string {
	repoTarget := entity.Workspace + "/" + entity.Repo

	switch entity.Type {
	case "repository":
		return "bb repo view --repo " + repoTarget
	case "pull-request":
		return "bb pr view " + strconv.Itoa(entity.PR) + " --repo " + repoTarget
	case "pull-request-comment":
		return "bb pr comment view " + strconv.Itoa(entity.Comment) + " --pr " + strconv.Itoa(entity.PR) + " --repo " + repoTarget
	case "issue":
		return "bb issue view " + strconv.Itoa(entity.Issue) + " --repo " + repoTarget
	case "commit":
		return "bb browse " + entity.Commit + " --repo " + repoTarget + " --no-browser"
	case "path":
		target := entity.Path
		if entity.Line > 0 {
			target += ":" + strconv.Itoa(entity.Line)
		}
		if target == "" {
			return "bb browse --repo " + repoTarget + " --no-browser"
		}
		return "bb browse " + target + " --repo " + repoTarget + " --no-browser"
	default:
		return ""
	}
}

func writeResolveSummary(w io.Writer, entity resolvedEntity) error {
	if err := writeTargetHeader(w, "Repository", entity.Workspace, entity.Repo); err != nil {
		return err
	}
	for _, row := range resolveSummaryRows(entity) {
		if err := writeLabelValue(w, row.Label, row.Value); err != nil {
			return err
		}
	}
	if err := writeLabelValue(w, "Canonical URL", entity.CanonicalURL); err != nil {
		return err
	}
	return writeNextStep(w, nextResolveCommand(entity))
}

func resolveSummaryRows(entity resolvedEntity) []summaryRow {
	rows := []summaryRow{
		{Label: "Type", Value: entity.Type},
	}
	if entity.PR > 0 {
		rows = append(rows, summaryRow{Label: "Pull Request", Value: strconv.Itoa(entity.PR)})
	}
	if entity.Comment > 0 {
		rows = append(rows, summaryRow{Label: "Comment", Value: strconv.Itoa(entity.Comment)})
	}
	if entity.Issue > 0 {
		rows = append(rows, summaryRow{Label: "Issue", Value: strconv.Itoa(entity.Issue)})
	}
	rows = appendSummaryRow(rows, "Commit", entity.Commit)
	rows = appendSummaryRow(rows, "Ref", entity.Ref)
	rows = appendSummaryRow(rows, "Path", entity.Path)
	if entity.Line > 0 {
		rows = append(rows, summaryRow{Label: "Line", Value: strconv.Itoa(entity.Line)})
	}
	rows = appendSummaryRow(rows, "URL", entity.URL)
	return rows
}
