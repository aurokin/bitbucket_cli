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
				if err := writeTargetHeader(w, "Repository", entity.Workspace, entity.Repo); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Type", entity.Type); err != nil {
					return err
				}
				if entity.PR > 0 {
					if err := writeLabelValue(w, "Pull Request", strconv.Itoa(entity.PR)); err != nil {
						return err
					}
				}
				if entity.Comment > 0 {
					if err := writeLabelValue(w, "Comment", strconv.Itoa(entity.Comment)); err != nil {
						return err
					}
				}
				if entity.Issue > 0 {
					if err := writeLabelValue(w, "Issue", strconv.Itoa(entity.Issue)); err != nil {
						return err
					}
				}
				if err := writeLabelValue(w, "Commit", entity.Commit); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Ref", entity.Ref); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Path", entity.Path); err != nil {
					return err
				}
				if entity.Line > 0 {
					if err := writeLabelValue(w, "Line", strconv.Itoa(entity.Line)); err != nil {
						return err
					}
				}
				if err := writeLabelValue(w, "URL", entity.URL); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Canonical URL", entity.CanonicalURL); err != nil {
					return err
				}
				return writeNextStep(w, nextResolveCommand(entity))
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
	case "pull-request", "pull-request-comment":
		return "bb pr view " + strconv.Itoa(entity.PR) + " --repo " + repoTarget
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
