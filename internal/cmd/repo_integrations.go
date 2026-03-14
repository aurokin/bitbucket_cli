package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type repoDeployKeyListPayload struct {
	Host      string                          `json:"host"`
	Workspace string                          `json:"workspace"`
	Repo      string                          `json:"repo"`
	Keys      []bitbucket.RepositoryDeployKey `json:"keys"`
}

type repoDeployKeyPayload struct {
	Host      string                        `json:"host"`
	Workspace string                        `json:"workspace"`
	Repo      string                        `json:"repo"`
	Action    string                        `json:"action,omitempty"`
	Deleted   bool                          `json:"deleted,omitempty"`
	Key       bitbucket.RepositoryDeployKey `json:"key"`
}

func newRepoDeployKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-key",
		Short: "Work with repository deploy keys",
		Long:  "List, view, create, and delete Bitbucket repository deploy keys. Bitbucket currently rejects deploy-key updates in the live API behavior we verified, so rotation should use delete plus create.",
		Example: "  bb repo deploy-key list --repo workspace-slug/repo-slug\n" +
			"  bb repo deploy-key create --repo workspace-slug/repo-slug --label ci --key-file ./id_ed25519.pub\n" +
			"  bb repo deploy-key delete 7 --repo workspace-slug/repo-slug --yes",
	}
	cmd.AddCommand(
		newRepoDeployKeyListCmd(),
		newRepoDeployKeyViewCmd(),
		newRepoDeployKeyCreateCmd(),
		newRepoDeployKeyDeleteCmd(),
	)
	return cmd
}

func newRepoDeployKeyListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repository deploy keys",
		Example: "  bb repo deploy-key list --repo workspace-slug/repo-slug\n" +
			"  bb repo deploy-key list --repo workspace-slug/repo-slug --json keys",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			keys, err := resolved.Client.ListRepositoryDeployKeys(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := repoDeployKeyListPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Keys: keys}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoDeployKeyListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of repository deploy keys to return")
	return cmd
}

func newRepoDeployKeyViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "view <key-id>",
		Short: "View one repository deploy key",
		Example: "  bb repo deploy-key view 7 --repo workspace-slug/repo-slug\n" +
			"  bb repo deploy-key view 7 --repo workspace-slug/repo-slug --json key",
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
			keyID, err := parsePositiveInt("repository deploy key", args[0])
			if err != nil {
				return err
			}
			key, err := resolved.Client.GetRepositoryDeployKey(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, keyID)
			if err != nil {
				return err
			}
			payload := repoDeployKeyPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "viewed", Key: key}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoDeployKeySummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newRepoDeployKeyCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, label, key, keyFile, comment string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a repository deploy key",
		Example: "  bb repo deploy-key create --repo workspace-slug/repo-slug --label ci --key-file ./id_ed25519.pub\n" +
			"  bb repo deploy-key create --repo workspace-slug/repo-slug --label ci --key 'ssh-ed25519 AAAA...' --json key",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			resolvedKey, err := resolveDeployKeyMaterial(key, keyFile)
			if err != nil {
				return err
			}
			deployKey, err := resolved.Client.CreateRepositoryDeployKey(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, bitbucket.CreateRepositoryDeployKeyOptions{
				Label:   label,
				Key:     resolvedKey,
				Comment: comment,
			})
			if err != nil {
				return err
			}
			payload := repoDeployKeyPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "created", Key: deployKey}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoDeployKeySummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&label, "label", "", "Deploy key label")
	cmd.Flags().StringVar(&key, "key", "", "Deploy key public key material")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "Read deploy key public key material from a file")
	cmd.Flags().StringVar(&comment, "comment", "", "Deploy key comment override")
	cmd.MarkFlagsMutuallyExclusive("key", "key-file")
	_ = cmd.MarkFlagRequired("label")
	return cmd
}

func newRepoDeployKeyDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <key-id>",
		Short: "Delete a repository deploy key",
		Long:  "Delete a Bitbucket repository deploy key. Humans must confirm the exact repository and deploy key unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb repo deploy-key delete 7 --repo workspace-slug/repo-slug --yes\n" +
			"  bb --no-prompt repo deploy-key delete 7 --repo workspace-slug/repo-slug --yes --json '*'",
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
			keyID, err := parsePositiveInt("repository deploy key", args[0])
			if err != nil {
				return err
			}
			deployKey, err := resolved.Client.GetRepositoryDeployKey(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, keyID)
			if err != nil {
				return err
			}
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("repository deploy key deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%d", resolved.Target.Workspace, resolved.Target.Repo, keyID)); err != nil {
					return err
				}
			}
			if err := resolved.Client.DeleteRepositoryDeployKey(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, keyID); err != nil {
				return err
			}
			payload := repoDeployKeyPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "deleted", Deleted: true, Key: deployKey}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoDeployKeySummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm repository deploy key deletion without prompting")
	return cmd
}

func writeRepoDeployKeyListSummary(w io.Writer, payload repoDeployKeyListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Keys) == 0 {
		_, err := fmt.Fprintf(w, "No repository deploy keys found for %s/%s.\n", payload.Workspace, payload.Repo)
		return err
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tlabel\tlast-used"); err != nil {
		return err
	}
	for _, key := range payload.Keys {
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\n", key.ID, output.Truncate(key.Label, 32), output.Truncate(key.LastUsed, 24)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo deploy-key view %d --repo %s/%s", payload.Keys[0].ID, payload.Workspace, payload.Repo))
}

func writeRepoDeployKeySummary(w io.Writer, payload repoDeployKeyPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Deploy Key", strconv.Itoa(payload.Key.ID)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "State", repoDeployKeyState(payload)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Label", payload.Key.Label); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Comment", payload.Key.Comment); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Last Used", payload.Key.LastUsed); err != nil {
		return err
	}
	if payload.Deleted {
		return writeNextStep(w, fmt.Sprintf("bb repo deploy-key list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	return writeNextStep(w, fmt.Sprintf("bb repo deploy-key view %d --repo %s/%s", payload.Key.ID, payload.Workspace, payload.Repo))
}

func repoDeployKeyState(payload repoDeployKeyPayload) string {
	if payload.Deleted {
		return "deleted"
	}
	return "present"
}

func resolveDeployKeyMaterial(raw, filePath string) (string, error) {
	raw = strings.TrimSpace(raw)
	filePath = strings.TrimSpace(filePath)
	if raw == "" && filePath == "" {
		return "", fmt.Errorf("repository deploy key is required; pass --key or --key-file")
	}
	if raw != "" {
		return raw, nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read deploy key file %q: %w", filePath, err)
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("deploy key file %q was empty", filePath)
	}
	return key, nil
}
