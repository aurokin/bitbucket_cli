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

type deploymentListPayload struct {
	Host        string                 `json:"host"`
	Workspace   string                 `json:"workspace"`
	Repo        string                 `json:"repo"`
	Warnings    []string               `json:"warnings,omitempty"`
	Deployments []bitbucket.Deployment `json:"deployments"`
}

type deploymentPayload struct {
	Host       string               `json:"host"`
	Workspace  string               `json:"workspace"`
	Repo       string               `json:"repo"`
	Warnings   []string             `json:"warnings,omitempty"`
	Deployment bitbucket.Deployment `json:"deployment"`
}

type deploymentEnvironmentListPayload struct {
	Host         string                            `json:"host"`
	Workspace    string                            `json:"workspace"`
	Repo         string                            `json:"repo"`
	Warnings     []string                          `json:"warnings,omitempty"`
	Environments []bitbucket.DeploymentEnvironment `json:"environments"`
}

type deploymentEnvironmentPayload struct {
	Host        string                          `json:"host"`
	Workspace   string                          `json:"workspace"`
	Repo        string                          `json:"repo"`
	Warnings    []string                        `json:"warnings,omitempty"`
	Environment bitbucket.DeploymentEnvironment `json:"environment"`
}

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

func newDeploymentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: "Inspect Bitbucket deployments and environments",
		Long:  "Inspect Bitbucket deployment history, deployment environments, and deployment environment variables backed by the official Bitbucket Cloud deployments APIs.",
	}
	cmd.AddCommand(
		newDeploymentListCmd(),
		newDeploymentViewCmd(),
		newDeploymentEnvironmentCmd(),
	)
	return cmd
}

func newDeploymentListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments in one repository",
		Example: "  bb deployment list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb deployment list --repo workspace-slug/pipelines-repo-slug --json deployments",
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
			deployments, err := resolved.Client.ListDeployments(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := deploymentListPayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Deployments: deployments,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of deployments to return")
	return cmd
}

func newDeploymentViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <deployment-uuid>",
		Short: "Show deployment information",
		Example: "  bb deployment view '{deployment-uuid}' --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb deployment view '{deployment-uuid}' --repo workspace-slug/pipelines-repo-slug --json deployment",
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
			item, err := resolved.Client.GetDeployment(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}
			payload := deploymentPayload{
				Host:       resolved.Target.Host,
				Workspace:  resolved.Target.Workspace,
				Repo:       resolved.Target.Repo,
				Warnings:   append([]string(nil), resolved.Target.Warnings...),
				Deployment: item,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
}

func newDeploymentEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"environments"},
		Short:   "Inspect deployment environments",
	}
	cmd.AddCommand(
		newDeploymentEnvironmentListCmd(),
		newDeploymentEnvironmentViewCmd(),
		newDeploymentEnvironmentVariableCmd(),
	)
	return cmd
}

func newDeploymentEnvironmentListCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployment environments in one repository",
		Example: "  bb deployment environment list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb deployment environment list --repo workspace-slug/pipelines-repo-slug --json environments",
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
			environments, err := resolved.Client.ListDeploymentEnvironments(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, limit)
			if err != nil {
				return err
			}
			payload := deploymentEnvironmentListPayload{
				Host:         resolved.Target.Host,
				Workspace:    resolved.Target.Workspace,
				Repo:         resolved.Target.Repo,
				Warnings:     append([]string(nil), resolved.Target.Warnings...),
				Environments: environments,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentEnvironmentListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of environments to return")
	return cmd
}

func newDeploymentEnvironmentViewCmd() *cobra.Command {
	var flags formatFlags
	var host, workspace, repo string

	cmd := &cobra.Command{
		Use:   "view <environment>",
		Short: "Show deployment environment information",
		Example: "  bb deployment environment view test --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb deployment environment view '{environment-uuid}' --repo workspace-slug/pipelines-repo-slug --json environment",
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
			environment, err := resolveDeploymentEnvironment(context.Background(), resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, args[0])
			if err != nil {
				return err
			}
			payload := deploymentEnvironmentPayload{
				Host:        resolved.Target.Host,
				Workspace:   resolved.Target.Workspace,
				Repo:        resolved.Target.Repo,
				Warnings:    append([]string(nil), resolved.Target.Warnings...),
				Environment: environment,
			}
			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writeDeploymentEnvironmentSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	return cmd
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
			if strings.TrimSpace(key) == "" {
				key = existing.Key
			}
			nextSecured := existing.Secured
			if strings.TrimSpace(secured) != "" {
				nextSecured, err = parseBoolString(secured)
				if err != nil {
					return fmt.Errorf("--secured must be true or false")
				}
			}
			variable, err := resolved.Client.UpdateDeploymentVariable(context.Background(), resolved.Target.Workspace, resolved.Target.Repo, env.UUID, existing.UUID, bitbucket.DeploymentVariable{
				Key:     strings.TrimSpace(key),
				Value:   resolvedValue,
				Secured: nextSecured,
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
			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("deployment environment variable deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s:%s", resolved.Target.Workspace, resolved.Target.Repo, env.Slug, variable.Key)); err != nil {
					return err
				}
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

func resolveDeploymentEnvironment(ctx context.Context, client *bitbucket.Client, workspace, repo, raw string) (bitbucket.DeploymentEnvironment, error) {
	reference := strings.TrimSpace(raw)
	if reference == "" {
		return bitbucket.DeploymentEnvironment{}, fmt.Errorf("deployment environment reference is required")
	}
	if strings.HasPrefix(reference, "{") {
		return client.GetDeploymentEnvironment(ctx, workspace, repo, reference)
	}
	environments, err := client.ListDeploymentEnvironments(ctx, workspace, repo, 200)
	if err != nil {
		return bitbucket.DeploymentEnvironment{}, err
	}
	for _, environment := range environments {
		if strings.EqualFold(environment.Name, reference) || strings.EqualFold(environment.Slug, reference) || strings.EqualFold(strings.Trim(environment.UUID, "{}"), strings.Trim(reference, "{}")) {
			return environment, nil
		}
	}
	return bitbucket.DeploymentEnvironment{}, fmt.Errorf("deployment environment %q not found in %s/%s", reference, workspace, repo)
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

func writeDeploymentListSummary(w io.Writer, payload deploymentListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if len(payload.Deployments) == 0 {
		if _, err := fmt.Fprintf(w, "No deployments found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb deployment environment list --repo %s/%s", payload.Workspace, payload.Repo))
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "uuid\tstate\tenvironment\trelease"); err != nil {
		return err
	}
	for _, deployment := range payload.Deployments {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			output.Truncate(deployment.UUID, 18),
			output.Truncate(deployment.State.Name, 16),
			output.Truncate(deployment.Environment.Name, 20),
			output.Truncate(deployment.Release.Name, 20),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb deployment view %s --repo %s/%s", payload.Deployments[0].UUID, payload.Workspace, payload.Repo))
}

func writeDeploymentSummary(w io.Writer, payload deploymentPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Deployment", payload.Deployment.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "State", payload.Deployment.State.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Environment", payload.Deployment.Environment.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Release", payload.Deployment.Release.Name); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb deployment environment view %s --repo %s/%s", payload.Deployment.Environment.UUID, payload.Workspace, payload.Repo))
}

func writeDeploymentEnvironmentListSummary(w io.Writer, payload deploymentEnvironmentListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if len(payload.Environments) == 0 {
		if _, err := fmt.Fprintf(w, "No deployment environments found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "name\tslug\ttype\tlock"); err != nil {
		return err
	}
	for _, environment := range payload.Environments {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			output.Truncate(environment.Name, 20),
			output.Truncate(environment.Slug, 20),
			output.Truncate(environment.Category.Name, 16),
			output.Truncate(environment.Lock.Name, 12),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb deployment environment view %s --repo %s/%s", payload.Environments[0].Slug, payload.Workspace, payload.Repo))
}

func writeDeploymentEnvironmentSummary(w io.Writer, payload deploymentEnvironmentPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeWarnings(w, payload.Warnings); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Environment", payload.Environment.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Slug", payload.Environment.Slug); err != nil {
		return err
	}
	if err := writeLabelValue(w, "UUID", payload.Environment.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Type", payload.Environment.Category.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Lock", payload.Environment.Lock.Name); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb deployment environment variable list --repo %s/%s --environment %s", payload.Workspace, payload.Repo, payload.Environment.Slug))
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
