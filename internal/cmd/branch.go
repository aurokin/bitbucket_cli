package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type branchListPayload struct {
	Host      string                       `json:"host"`
	Workspace string                       `json:"workspace"`
	Repo      string                       `json:"repo"`
	Warnings  []string                     `json:"warnings,omitempty"`
	Query     string                       `json:"query,omitempty"`
	Branches  []bitbucket.RepositoryBranch `json:"branches"`
}

type branchPayload struct {
	Host      string                     `json:"host"`
	Workspace string                     `json:"workspace"`
	Repo      string                     `json:"repo"`
	Warnings  []string                   `json:"warnings,omitempty"`
	Action    string                     `json:"action,omitempty"`
	Branch    bitbucket.RepositoryBranch `json:"branch"`
}

type branchDeletePayload struct {
	Host      string   `json:"host"`
	Workspace string   `json:"workspace"`
	Repo      string   `json:"repo"`
	Warnings  []string `json:"warnings,omitempty"`
	Branch    string   `json:"branch"`
	Deleted   bool     `json:"deleted"`
}

func newBranchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Work with repository branches",
		Long:  "List, inspect, create, and delete Bitbucket repository branches.",
	}
	cmd.AddCommand(newBranchListCmd(), newBranchViewCmd(), newBranchCreateCmd(), newBranchDeleteCmd())
	return cmd
}

func newBranchListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, query, sort string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [repository]",
		Short: "List repository branches",
		Long:  "List open branches in a Bitbucket repository.",
		Example: "  bb branch list workspace-slug/repo-slug\n" +
			"  bb branch list --repo workspace-slug/repo-slug --limit 50\n" +
			"  bb branch list --repo workspace-slug/repo-slug --query 'name ~ \"release\"' --json branches",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, formatErr := flags.options()
			if formatErr != nil {
				return formatErr
			}
			var resolved resolvedRepoCommandTarget
			var err error
			if firstArg(args) != "" || strings.TrimSpace(repo) != "" {
				resolved, err = resolveRepoCommandTargetInput(context.Background(), host, workspace, repo, firstArg(args), true)
			} else {
				resolved, err = resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			}
			if err != nil {
				return err
			}

			branches, err := resolved.Client.ListBranches(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.ListBranchesOptions{
				Limit: limit,
				Query: query,
				Sort:  sort,
			})
			if err != nil {
				return err
			}

			payload := branchListPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Query:     query,
				Branches:  branches,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeBranchListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&query, "query", "", "Bitbucket branch query filter")
	cmd.Flags().StringVar(&sort, "sort", "name", "Bitbucket branch sort expression")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of branches to return")
	return cmd
}

func newBranchViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <name>",
		Short: "View one branch",
		Long:  "View one Bitbucket repository branch.",
		Example: "  bb branch view main --repo workspace-slug/repo-slug\n" +
			"  bb branch view feature/demo --repo workspace-slug/repo-slug --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			branch, err := resolved.Client.GetBranch(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}

			payload := branchPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Branch:    branch,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeBranchSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newBranchCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, target string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a branch",
		Long:  "Create a Bitbucket branch from one commit hash or branch name. When run inside a matching checkout, bb defaults the target to the current branch if --target is omitted.",
		Example: "  bb branch create feature/demo --repo workspace-slug/repo-slug --target main\n" +
			"  bb branch create feature/demo --repo workspace-slug/repo-slug --target abc1234 --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			targetRef, err := resolveRefCreateTarget(resolved.Target, target)
			if err != nil {
				return err
			}

			branch, err := resolved.Client.CreateBranch(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.CreateBranchOptions{
				Name:   args[0],
				Target: targetRef,
			})
			if err != nil {
				return err
			}

			payload := branchPayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Action:    "created",
				Branch:    branch,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeBranchCreateSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&target, "target", "", "Source commit hash or branch name to branch from")
	return cmd
}

func newBranchDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a branch",
		Long:  "Delete a Bitbucket branch. Humans must confirm the exact repository and branch unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb branch delete feature/demo --repo workspace-slug/repo-slug --yes\n" +
			"  bb --no-prompt branch delete feature/demo --repo workspace-slug/repo-slug --yes --json '*'\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}

			name := strings.TrimSpace(args[0])
			confirmationTarget := fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, name)
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("branch deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, confirmationTarget); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeleteBranch(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, name); err != nil {
				return err
			}

			payload := branchDeletePayload{
				Host:      resolved.Target.Host,
				Workspace: resolved.Target.Workspace,
				Repo:      resolved.Target.Repo,
				Warnings:  append([]string(nil), resolved.Target.Warnings...),
				Branch:    name,
				Deleted:   true,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeBranchDeleteSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func writeBranchListSummary(w io.Writer, payload branchListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Query", payload.Query); err != nil {
		return err
	}
	if len(payload.Branches) == 0 {
		if _, err := fmt.Fprintln(w, "No branches found."); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\ttarget\tdefault"); err != nil {
		return err
	}
	for _, branch := range payload.Branches {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(branch.Name, 36),
			output.Truncate(branch.Target.Hash, 12),
			output.Truncate(branch.DefaultMergeStrategy, 18),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb branch view %s --repo %s/%s", payload.Branches[0].Name, payload.Workspace, payload.Repo))
}

func writeBranchSummary(w io.Writer, payload branchPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Branch", payload.Branch.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Target", payload.Branch.Target.Hash); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Default Merge", payload.Branch.DefaultMergeStrategy); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb browse --repo %s/%s --branch %s --no-browser", payload.Workspace, payload.Repo, payload.Branch.Name))
}

func writeBranchCreateSummary(w io.Writer, payload branchPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Branch", payload.Branch.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Target", payload.Branch.Target.Hash); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb branch view %s --repo %s/%s", payload.Branch.Name, payload.Workspace, payload.Repo))
}

func writeBranchDeleteSummary(w io.Writer, payload branchDeletePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Branch", payload.Branch); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", "deleted"); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Status", "deleted"); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb branch list --repo %s/%s", payload.Workspace, payload.Repo))
}

func resolveRefCreateTarget(target resolvedRepoTarget, provided string) (string, error) {
	if value := strings.TrimSpace(provided); value != "" {
		return value, nil
	}
	if target.LocalRepo != nil {
		branch, err := gitrepo.CurrentBranch(context.Background(), target.LocalRepo.RootDir)
		if err == nil && strings.TrimSpace(branch) != "" {
			return strings.TrimSpace(branch), nil
		}
	}
	return "", fmt.Errorf("could not determine the source ref for %s/%s from the current directory; pass --target or run inside a matching checkout", target.Workspace, target.Repo)
}
