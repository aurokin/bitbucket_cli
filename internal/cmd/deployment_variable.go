package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type deploymentVariableListPayload struct {
	Host        string                          `json:"host"`
	Workspace   string                          `json:"workspace"`
	Repo        string                          `json:"repo"`
	Warnings    []string                        `json:"warnings,omitempty"`
	Environment bitbucket.DeploymentEnvironment `json:"environment"`
	Variables   []bitbucket.DeploymentVariable  `json:"variables"`
}

type deploymentVariablePayload struct {
	Host        string                          `json:"host"`
	Workspace   string                          `json:"workspace"`
	Repo        string                          `json:"repo"`
	Warnings    []string                        `json:"warnings,omitempty"`
	Action      string                          `json:"action,omitempty"`
	Deleted     bool                            `json:"deleted,omitempty"`
	Environment bitbucket.DeploymentEnvironment `json:"environment"`
	Variable    bitbucket.DeploymentVariable    `json:"variable"`
}

func newDeploymentEnvironmentVariableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variable",
		Aliases: []string{"variables"},
		Short:   "Manage deployment environment variables",
	}
	cmd.AddCommand(
		newDeploymentEnvironmentVariableListCmd(),
		newDeploymentEnvironmentVariableViewCmd(),
		newDeploymentEnvironmentVariableCreateCmd(),
		newDeploymentEnvironmentVariableEditCmd(),
		newDeploymentEnvironmentVariableDeleteCmd(),
	)
	return cmd
}

func newDeploymentEnvironmentVariableListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, environment string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List variables for one deployment environment",
		Example: "  bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test\n" +
			"  bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment '{environment-uuid}' --json variables",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			if strings.TrimSpace(environment) == "" {
				return fmt.Errorf("deployment environment variable listing requires --environment")
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			env, err := resolveDeploymentEnvironment(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, environment)
			if err != nil {
				return err
			}
			variables, err := resolved.Client.ListDeploymentVariables(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, env.UUID, bitbucket.ListDeploymentVariablesOptions{Limit: limit})
			if err != nil {
				return err
			}
			payload := deploymentVariableListPayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Environment: env,
				Variables:   variables,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentVariableListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&environment, "environment", "", "Deployment environment reference as a name, slug, or UUID")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of deployment variables to return")
	return cmd
}

func newDeploymentEnvironmentVariableViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, environment string

	cmd := &cobra.Command{
		Use:   "view <variable-uuid>",
		Short: "Show one deployment environment variable",
		Example: "  bb deployment environment variable view '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test\n" +
			"  bb deployment environment variable view '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test --json variable",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			if strings.TrimSpace(environment) == "" {
				return fmt.Errorf("deployment environment variable view requires --environment")
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			env, err := resolveDeploymentEnvironment(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, environment)
			if err != nil {
				return err
			}
			variable, err := resolveDeploymentVariableReference(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, env.UUID, args[0])
			if err != nil {
				return err
			}
			payload := deploymentVariablePayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Action:      "viewed",
				Environment: env,
				Variable:    variable,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&environment, "environment", "", "Deployment environment reference as a name, slug, or UUID")
	return cmd
}

func newDeploymentEnvironmentVariableCreateCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, environment, key, value, valueFile string
	var secured bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a deployment environment variable",
		Example: "  bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key APP_ENV --value production\n" +
			"  bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key SECRET_TOKEN --value-file secret.txt --secured\n" +
			"  printf 'secret\\n' | bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key SECRET_TOKEN --value-file - --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			if strings.TrimSpace(environment) == "" {
				return fmt.Errorf("deployment environment variable creation requires --environment")
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			env, err := resolveDeploymentEnvironment(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, environment)
			if err != nil {
				return err
			}
			resolvedValue, err := resolvePipelineVariableValue(cmd.InOrStdin(), value, valueFile)
			if err != nil {
				return err
			}
			variable, err := resolved.Client.CreateDeploymentVariable(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, env.UUID, bitbucket.DeploymentVariable{
				Key:     strings.TrimSpace(key),
				Value:   resolvedValue,
				Secured: secured,
			})
			if err != nil {
				return err
			}
			payload := deploymentVariablePayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Action:      "created",
				Environment: env,
				Variable:    variable,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&environment, "environment", "", "Deployment environment reference as a name, slug, or UUID")
	cmd.Flags().StringVar(&key, "key", "", "Deployment variable key")
	cmd.Flags().StringVar(&value, "value", "", "Deployment variable value")
	cmd.Flags().StringVar(&valueFile, "value-file", "", "Read the deployment variable value from a file, or '-' for stdin")
	cmd.Flags().BoolVar(&secured, "secured", false, "Mark the deployment variable as secured")
	cmd.MarkFlagsMutuallyExclusive("value", "value-file")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func newDeploymentEnvironmentVariableEditCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, environment, key, value, valueFile, secured string

	cmd := &cobra.Command{
		Use:   "edit <key-or-uuid>",
		Short: "Edit a deployment environment variable",
		Long:  "Edit a Bitbucket deployment environment variable by key or UUID. By default the existing secured flag is preserved unless --secured true or --secured false is provided.",
		Example: "  bb deployment environment variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --environment test --value staging\n" +
			"  bb deployment environment variable edit '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test --secured true --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			if strings.TrimSpace(environment) == "" {
				return fmt.Errorf("deployment environment variable editing requires --environment")
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			env, err := resolveDeploymentEnvironment(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, environment)
			if err != nil {
				return err
			}
			existing, err := resolveDeploymentVariableReference(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, env.UUID, args[0])
			if err != nil {
				return err
			}
			resolvedValue, err := resolvePipelineVariableValue(cmd.InOrStdin(), value, valueFile)
			if err != nil {
				return err
			}
			nextVariable, err := buildDeploymentVariableUpdate(existing, key, resolvedValue, secured)
			if err != nil {
				return err
			}
			variable, err := resolved.Client.UpdateDeploymentVariable(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, env.UUID, existing.UUID, bitbucket.DeploymentVariable{
				Key:     nextVariable.Key,
				Value:   nextVariable.Value,
				Secured: nextVariable.Secured,
			})
			if err != nil {
				return err
			}
			payload := deploymentVariablePayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Action:      "edited",
				Environment: env,
				Variable:    variable,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&environment, "environment", "", "Deployment environment reference as a name, slug, or UUID")
	cmd.Flags().StringVar(&key, "key", "", "Override the deployment variable key")
	cmd.Flags().StringVar(&value, "value", "", "Deployment variable value")
	cmd.Flags().StringVar(&valueFile, "value-file", "", "Read the deployment variable value from a file, or '-' for stdin")
	cmd.Flags().StringVar(&secured, "secured", "", "Set secured to true or false; defaults to the existing value")
	cmd.MarkFlagsMutuallyExclusive("value", "value-file")
	return cmd
}

func buildDeploymentVariableUpdate(existing bitbucket.DeploymentVariable, key, value, secured string) (bitbucket.DeploymentVariable, error) {
	nextKey := strings.TrimSpace(key)
	if nextKey == "" {
		nextKey = existing.Key
	}
	nextSecured := existing.Secured
	if strings.TrimSpace(secured) != "" {
		parsed, err := parseBoolString(secured)
		if err != nil {
			return bitbucket.DeploymentVariable{}, fmt.Errorf("--secured must be true or false")
		}
		nextSecured = parsed
	}
	return bitbucket.DeploymentVariable{
		Key:     nextKey,
		Value:   value,
		Secured: nextSecured,
	}, nil
}

func newDeploymentEnvironmentVariableDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo, environment string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <key-or-uuid>",
		Short: "Delete a deployment environment variable",
		Long:  "Delete a Bitbucket deployment environment variable by key or UUID. Humans must confirm the exact repository, environment, and variable unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb deployment environment variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --environment test --yes\n" +
			"  bb --no-prompt deployment environment variable delete '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test --yes --json '*'",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}
			if strings.TrimSpace(environment) == "" {
				return fmt.Errorf("deployment environment variable deletion requires --environment")
			}
			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			env, err := resolveDeploymentEnvironment(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, environment)
			if err != nil {
				return err
			}
			variable, err := resolveDeploymentVariableReference(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, env.UUID, args[0])
			if err != nil {
				return err
			}
			if err := confirmDeploymentVariableDeletion(cmd, resolved.Target.Workspace, resolved.Target.Repo, env.Slug, variable.Key, yes); err != nil {
				return err
			}
			if err := resolved.Client.DeleteDeploymentVariable(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, env.UUID, variable.UUID); err != nil {
				return err
			}
			payload := deploymentVariablePayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Action:      "deleted",
				Deleted:     true,
				Environment: env,
				Variable:    variable,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&environment, "environment", "", "Deployment environment reference as a name, slug, or UUID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func confirmDeploymentVariableDeletion(cmd *cobra.Command, workspace, repo, environmentSlug, key string, yes bool) error {
	if yes {
		return nil
	}
	if !promptsEnabled(cmd) {
		return fmt.Errorf("deployment environment variable deletion requires confirmation; pass --yes or run in an interactive terminal")
	}
	return confirmExactMatch(cmd, deploymentVariableDeletionConfirmationTarget(workspace, repo, environmentSlug, key))
}

func deploymentVariableDeletionConfirmationTarget(workspace, repo, environmentSlug, key string) string {
	return fmt.Sprintf("%s/%s:%s:%s", workspace, repo, environmentSlug, key)
}

func resolveDeploymentVariableReference(ctx context.Context, client *bitbucket.Client, workspace, repo, environmentUUID, raw string) (bitbucket.DeploymentVariable, error) {
	return resolveVariableReference(
		raw,
		"deployment variable",
		nil,
		func() ([]bitbucket.DeploymentVariable, error) {
			return client.ListDeploymentVariables(ctx, workspace, repo, environmentUUID, bitbucket.ListDeploymentVariablesOptions{Limit: 200})
		},
		func(reference string, variable bitbucket.DeploymentVariable) bool {
			return strings.EqualFold(variable.Key, reference) || strings.EqualFold(strings.Trim(variable.UUID, "{}"), strings.Trim(reference, "{}"))
		},
	)
}

func writeDeploymentVariableListSummary(w io.Writer, payload deploymentVariableListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Environment", payload.Environment.Name); err != nil {
		return err
	}
	if len(payload.Variables) == 0 {
		if _, err := fmt.Fprintf(w, "No deployment variables found for %s in %s/%s.\n", payload.Environment.Name, payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "uuid\tkey\tsecured"); err != nil {
		return err
	}
	for _, variable := range payload.Variables {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%t\n",
			output.Truncate(variable.UUID, 18),
			output.Truncate(variable.Key, 28),
			variable.Secured,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb deployment environment variable view %s --repo %s/%s --environment %s", payload.Variables[0].UUID, payload.Workspace, payload.Repo, payload.Environment.Slug))
}

func writeDeploymentVariableSummary(w io.Writer, payload deploymentVariablePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Environment", payload.Environment.Name); err != nil {
		return err
	}
	if payload.Action != "" {
		if err := writeLabelValue(w, "Action", payload.Action); err != nil {
			return err
		}
	}
	if err := writeLabelValue(w, "Variable", payload.Variable.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "UUID", payload.Variable.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Secured", fmt.Sprintf("%t", payload.Variable.Secured)); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb deployment environment variable list --repo %s/%s --environment %s", payload.Workspace, payload.Repo, payload.Environment.Slug))
}
