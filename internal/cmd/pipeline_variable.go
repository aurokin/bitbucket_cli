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

type pipelineVariablePayload struct {
	Host      string                     `json:"host"`
	Workspace string                     `json:"workspace"`
	Repo      string                     `json:"repo"`
	Action    string                     `json:"action,omitempty"`
	Deleted   bool                       `json:"deleted,omitempty"`
	Variable  bitbucket.PipelineVariable `json:"variable"`
}

type pipelineVariableListPayload struct {
	Host      string                       `json:"host"`
	Workspace string                       `json:"workspace"`
	Repo      string                       `json:"repo"`
	Variables []bitbucket.PipelineVariable `json:"variables"`
}

func newPipelineVariableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage repository pipeline variables",
		Long:  "List, inspect, create, edit, and delete Bitbucket repository pipeline variables.",
		Example: "  bb pipeline variable list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured\n" +
			"  bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes",
	}

	cmd.AddCommand(
		newPipelineVariableListCmd(),
		newPipelineVariableViewCmd(),
		newPipelineVariableCreateCmd(),
		newPipelineVariableEditCmd(),
		newPipelineVariableDeleteCmd(),
	)

	return cmd
}

func newPipelineVariableListCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repository pipeline variables",
		Example: "  bb pipeline variable list --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline variable list --repo workspace-slug/pipelines-repo-slug --json '*'",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			target := resolved.Target

			variables, err := resolved.Client.ListPipelineVariables(context.Background(), target.Workspace, target.Repo, bitbucket.ListPipelineVariablesOptions{Limit: limit})
			if err != nil {
				return err
			}

			payload := pipelineVariableListPayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Variables: variables,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableListSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of pipeline variables to return")

	return cmd
}

func newPipelineVariableViewCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string

	cmd := &cobra.Command{
		Use:   "view <key-or-uuid>",
		Short: "View one repository pipeline variable",
		Example: "  bb pipeline variable view CI_TOKEN --repo workspace-slug/pipelines-repo-slug\n" +
			"  bb pipeline variable view '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'",
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
			target := resolved.Target

			variable, err := resolvePipelineVariableReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "viewed",
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")

	return cmd
}

func newPipelineVariableCreateCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var key string
	var value string
	var valueFile string
	var secured bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a repository pipeline variable",
		Example: "  bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured\n" +
			"  printf 'token-value\\n' | bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file - --json '*'\n" +
			"  bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value production",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			target := resolved.Target

			resolvedValue, err := resolvePipelineVariableValue(cmd.InOrStdin(), value, valueFile)
			if err != nil {
				return err
			}

			variable, err := resolved.Client.CreatePipelineVariable(context.Background(), target.Workspace, target.Repo, bitbucket.PipelineVariable{
				Key:     strings.TrimSpace(key),
				Value:   resolvedValue,
				Secured: secured,
			})
			if err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "created",
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&key, "key", "", "Pipeline variable key")
	cmd.Flags().StringVar(&value, "value", "", "Pipeline variable value")
	cmd.Flags().StringVar(&valueFile, "value-file", "", "Read the pipeline variable value from a file, or '-' for stdin")
	cmd.Flags().BoolVar(&secured, "secured", false, "Mark the variable as secured")
	cmd.MarkFlagsMutuallyExclusive("value", "value-file")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func newPipelineVariableEditCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var key string
	var value string
	var valueFile string
	var secured string

	cmd := &cobra.Command{
		Use:   "edit <key-or-uuid>",
		Short: "Edit a repository pipeline variable",
		Long:  "Edit a Bitbucket repository pipeline variable by key or UUID. By default the existing secured flag is preserved unless --secured true or --secured false is provided.",
		Example: "  bb pipeline variable edit CI_TOKEN --repo workspace-slug/pipelines-repo-slug --value-file secret.txt --secured true\n" +
			"  bb pipeline variable edit '{uuid}' --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value staging --json '*'\n" +
			"  bb pipeline variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --value production",
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
			target := resolved.Target

			existing, err := resolvePipelineVariableReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
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

			variable, err := resolved.Client.UpdatePipelineVariable(context.Background(), target.Workspace, target.Repo, existing.UUID, bitbucket.PipelineVariable{
				Key:     strings.TrimSpace(key),
				Value:   resolvedValue,
				Secured: nextSecured,
			})
			if err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "edited",
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().StringVar(&key, "key", "", "Override the pipeline variable key")
	cmd.Flags().StringVar(&value, "value", "", "Pipeline variable value")
	cmd.Flags().StringVar(&valueFile, "value-file", "", "Read the pipeline variable value from a file, or '-' for stdin")
	cmd.Flags().StringVar(&secured, "secured", "", "Set secured to true or false; defaults to the existing value")
	cmd.MarkFlagsMutuallyExclusive("value", "value-file")

	return cmd
}

func newPipelineVariableDeleteCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <key-or-uuid>",
		Short: "Delete a repository pipeline variable",
		Long:  "Delete a Bitbucket repository pipeline variable by key or UUID. Humans must confirm the exact repository and variable unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.",
		Example: "  bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes\n" +
			"  bb --no-prompt pipeline variable delete '{uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'\n" +
			"  bb pipeline variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --yes",
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
			target := resolved.Target

			variable, err := resolvePipelineVariableReference(context.Background(), resolved.Client, target.Workspace, target.Repo, args[0])
			if err != nil {
				return err
			}

			if !yes {
				if !promptsEnabled(cmd) {
					return fmt.Errorf("pipeline variable deletion requires confirmation; pass --yes or run in an interactive terminal")
				}
				if err := confirmExactMatch(cmd, fmt.Sprintf("%s/%s:%s", target.Workspace, target.Repo, variable.Key)); err != nil {
					return err
				}
			}

			if err := resolved.Client.DeletePipelineVariable(context.Background(), target.Workspace, target.Repo, variable.UUID); err != nil {
				return err
			}

			payload := pipelineVariablePayload{
				Host:      target.Host,
				Workspace: target.Workspace,
				Repo:      target.Repo,
				Action:    "deleted",
				Deleted:   true,
				Variable:  variable,
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				return writePipelineVariableSummary(w, payload)
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

func resolvePipelineVariableReference(ctx context.Context, client *bitbucket.Client, workspace, repo, raw string) (bitbucket.PipelineVariable, error) {
	return resolveVariableReference(
		raw,
		"pipeline variable",
		func(reference string) (bitbucket.PipelineVariable, error) {
			return client.GetPipelineVariable(ctx, workspace, repo, reference)
		},
		func() ([]bitbucket.PipelineVariable, error) {
			return client.ListPipelineVariables(ctx, workspace, repo, bitbucket.ListPipelineVariablesOptions{Limit: 200})
		},
		func(reference string, variable bitbucket.PipelineVariable) bool {
			return variable.Key == reference || strings.Trim(variable.UUID, "{}") == strings.Trim(reference, "{}") || variable.UUID == reference
		},
	)
}

func writePipelineVariableListSummary(w io.Writer, payload pipelineVariableListPayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if len(payload.Variables) == 0 {
		if _, err := fmt.Fprintf(w, "No pipeline variables found for %s/%s.\n", payload.Workspace, payload.Repo); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb pipeline variable create --repo %s/%s --key EXAMPLE --value value", payload.Workspace, payload.Repo))
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "key\tsecured\tuuid"); err != nil {
		return err
	}
	for _, variable := range payload.Variables {
		if _, err := fmt.Fprintf(tw, "%s\t%t\t%s\n", output.Truncate(variable.Key, 32), variable.Secured, output.Truncate(variable.UUID, 40)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline variable view %s --repo %s/%s", payload.Variables[0].Key, payload.Workspace, payload.Repo))
}

func writePipelineVariableSummary(w io.Writer, payload pipelineVariablePayload) error {
	if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Variable", payload.Variable.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
	} else {
		if err := writeLabelValue(w, "Secured", fmt.Sprintf("%t", payload.Variable.Secured)); err != nil {
			return err
		}
		if err := writeLabelValue(w, "UUID", payload.Variable.UUID); err != nil {
			return err
		}
		if !payload.Variable.Secured {
			if err := writeLabelValue(w, "Value", payload.Variable.Value); err != nil {
				return err
			}
		}
	}
	return writeNextStep(w, fmt.Sprintf("bb pipeline variable list --repo %s/%s", payload.Workspace, payload.Repo))
}
