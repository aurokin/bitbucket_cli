package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type repoHookListPayload struct {
	Host      string                        `json:"host"`
	Workspace string                        `json:"workspace"`
	Repo      string                        `json:"repo"`
	Hooks     []bitbucket.RepositoryWebhook `json:"hooks"`
}

type repoHookPayload struct {
	Host      string                      `json:"host"`
	Workspace string                      `json:"workspace"`
	Repo      string                      `json:"repo"`
	Action    string                      `json:"action,omitempty"`
	Deleted   bool                        `json:"deleted,omitempty"`
	Hook      bitbucket.RepositoryWebhook `json:"hook"`
}

func newRepoHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Work with repository webhooks",
		Long:  "List, view, create, edit, and delete Bitbucket repository webhooks.",
		Example: "  bb repo hook list --repo workspace-slug/repo-slug\n" +
			"  bb repo hook create --repo workspace-slug/repo-slug --url https://example.com/hook --event repo:push\n" +
			"  bb repo hook delete {hook-uuid} --repo workspace-slug/repo-slug --yes",
	}
	cmd.AddCommand(
		newRepoHookListCmd(),
		newRepoHookViewCmd(),
		newRepoHookCreateCmd(),
		newRepoHookEditCmd(),
		newRepoHookDeleteCmd(),
	)
	return cmd
}

func newRepoHookListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repository webhooks",
		Example: "  bb repo hook list --repo workspace-slug/repo-slug\n" +
			"  bb repo hook list --repo workspace-slug/repo-slug --json hooks",
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
			hooks, err := resolved.Client.ListRepositoryWebhooks(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := repoHookListPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Hooks: hooks}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoHookListSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of repository webhooks to return")
	return cmd
}

func newRepoHookViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	cmd := &cobra.Command{
		Use:   "view <webhook-id>",
		Short: "View one repository webhook",
		Example: "  bb repo hook view {hook-uuid} --repo workspace-slug/repo-slug\n" +
			"  bb repo hook view {hook-uuid} --repo workspace-slug/repo-slug --json hook",
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
			hook, err := resolved.Client.GetRepositoryWebhook(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, strings.TrimSpace(args[0]))
			if err != nil {
				return err
			}
			payload := repoHookPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "viewed", Hook: hook}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoHookSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newRepoHookCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, hookURL, description, secret string
	var events []string
	var active bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a repository webhook",
		Example: "  bb repo hook create --repo workspace-slug/repo-slug --url https://example.com/hook --event repo:push\n" +
			"  bb repo hook create --repo workspace-slug/repo-slug --url https://example.com/hook --event pullrequest:created --event pullrequest:updated --json hook",
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
			activeValue := active
			options := bitbucket.RepositoryWebhookMutationOptions{
				URL:         hookURL,
				Description: description,
				Active:      &activeValue,
				Events:      append([]string(nil), events...),
			}
			if secret != "" {
				options.Secret = &secret
			}
			hook, err := resolved.Client.CreateRepositoryWebhook(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, options)
			if err != nil {
				return err
			}
			payload := repoHookPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "created", Hook: hook}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoHookSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&hookURL, "url", "", "Webhook delivery URL")
	cmd.Flags().StringVar(&description, "description", "", "Webhook description")
	cmd.Flags().StringSliceVar(&events, "event", nil, "Webhook event to subscribe to; repeat for multiple events")
	cmd.Flags().BoolVar(&active, "active", true, "Create the webhook as active")
	cmd.Flags().StringVar(&secret, "secret", "", "Optional webhook secret")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("event")
	return cmd
}

func newRepoHookEditCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, hookURL, description, secret string
	var events []string
	var active bool
	var clearSecret bool
	cmd := &cobra.Command{
		Use:   "edit <webhook-id>",
		Short: "Edit a repository webhook",
		Example: "  bb repo hook edit {hook-uuid} --repo workspace-slug/repo-slug --description 'Updated hook'\n" +
			"  bb repo hook edit {hook-uuid} --repo workspace-slug/repo-slug --event repo:push --event pullrequest:created --json hook",
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
			options := bitbucket.RepositoryWebhookMutationOptions{
				URL:         hookURL,
				Description: description,
				ClearSecret: clearSecret,
			}
			if cmd.Flags().Changed("active") {
				activeValue := active
				options.Active = &activeValue
			}
			if cmd.Flags().Changed("event") {
				options.Events = append([]string(nil), events...)
			}
			if cmd.Flags().Changed("secret") {
				options.Secret = &secret
			}
			if !cmd.Flags().Changed("url") && description == "" && options.Active == nil && options.Secret == nil && !clearSecret && !cmd.Flags().Changed("event") {
				return fmt.Errorf("at least one webhook field must be updated")
			}
			hook, err := resolved.Client.UpdateRepositoryWebhook(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, strings.TrimSpace(args[0]), options)
			if err != nil {
				return err
			}
			payload := repoHookPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "edited", Hook: hook}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoHookSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&hookURL, "url", "", "Updated webhook delivery URL")
	cmd.Flags().StringVar(&description, "description", "", "Updated webhook description")
	cmd.Flags().StringSliceVar(&events, "event", nil, "Replace the webhook events with the provided values")
	cmd.Flags().BoolVar(&active, "active", false, "Set whether the webhook is active")
	cmd.Flags().StringVar(&secret, "secret", "", "Set a new webhook secret")
	cmd.Flags().BoolVar(&clearSecret, "clear-secret", false, "Remove the webhook secret")
	cmd.MarkFlagsMutuallyExclusive("secret", "clear-secret")
	return cmd
}

func newRepoHookDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <webhook-id>",
		Short: "Delete a repository webhook",
		Long:  "Delete a Bitbucket repository webhook. Humans must confirm the exact repository and webhook unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb repo hook delete {hook-uuid} --repo workspace-slug/repo-slug --yes\n" +
			"  bb --no-prompt repo hook delete {hook-uuid} --repo workspace-slug/repo-slug --yes --json '*'",
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
			hookID := strings.TrimSpace(args[0])
			hook, err := resolved.Client.GetRepositoryWebhook(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, hookID)
			if err != nil {
				return err
			}
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("repository webhook deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", resolved.Target.Workspace, resolved.Target.Repo, hookID)); err != nil {
					return err
				}
			}
			if err := resolved.Client.DeleteRepositoryWebhook(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, hookID); err != nil {
				return err
			}
			payload := repoHookPayload{Host: resolved.Target.Host, Workspace: resolved.Target.Workspace, Repo: resolved.Target.Repo, Action: "deleted", Deleted: true, Hook: hook}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeRepoHookSummary(w, payload)
			})
		},
	}
	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm repository webhook deletion without prompting")
	return cmd
}

func writeRepoHookListSummary(w io.Writer, payload repoHookListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Hooks) == 0 {
		_, err := fmt.Fprintf(w, "No repository webhooks found for %s/%s.\n", payload.Workspace, payload.Repo)
		return err
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "id\tactive\tdescription\turl"); err != nil {
		return err
	}
	for _, hook := range payload.Hooks {
		if _, err := fmt.Fprintf(tw, "%s\t%t\t%s\t%s\n",
			output.Truncate(hook.UUID, 24),
			hook.Active,
			output.Truncate(hook.Description, 28),
			output.Truncate(hook.URL, 40),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb repo hook view %s --repo %s/%s", payload.Hooks[0].UUID, payload.Workspace, payload.Repo))
}

func writeRepoHookSummary(w io.Writer, payload repoHookPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Hook", payload.Hook.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if err := writeLabelValue(w, "State", repoHookState(payload)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Description", payload.Hook.Description); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Hook.URL); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Active", strconv.FormatBool(payload.Hook.Active)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Events", strings.Join(payload.Hook.Events, ", ")); err != nil {
		return err
	}
	if payload.Deleted {
		return writeNextStep(w, fmt.Sprintf("bb repo hook list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	return writeNextStep(w, fmt.Sprintf("bb repo hook view %s --repo %s/%s", payload.Hook.UUID, payload.Workspace, payload.Repo))
}

func repoHookState(payload repoHookPayload) string {
	if payload.Deleted {
		return "deleted"
	}
	if payload.Hook.Active {
		return "active"
	}
	return "inactive"
}
