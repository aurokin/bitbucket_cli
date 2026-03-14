package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long:  "Generate shell completion scripts for bash, zsh, fish, and PowerShell.",
	}

	cmd.AddCommand(
		newCompletionShellCmd("bash", "Generate a bash completion script"),
		newCompletionShellCmd("zsh", "Generate a zsh completion script"),
		newCompletionShellCmd("fish", "Generate a fish completion script"),
		newCompletionShellCmd("powershell", "Generate a PowerShell completion script"),
	)

	return cmd
}

func newCompletionShellCmd(shell, short string) *cobra.Command {
	return &cobra.Command{
		Use:     shell,
		Short:   short,
		Args:    cobra.NoArgs,
		Example: completionExample(shell),
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCompletion(cmd.Root(), shell, cmd.OutOrStdout())
		},
	}
}

func completionExample(shell string) string {
	switch shell {
	case "bash":
		return "  bb completion bash"
	case "zsh":
		return "  bb completion zsh"
	case "fish":
		return "  bb completion fish"
	case "powershell":
		return "  bb completion powershell"
	default:
		return ""
	}
}

func writeCompletion(root *cobra.Command, shell string, out interface{ Write([]byte) (int, error) }) error {
	switch shell {
	case "bash":
		return root.GenBashCompletionV2(out, true)
	case "zsh":
		return root.GenZshCompletion(out)
	case "fish":
		return root.GenFishCompletion(out, true)
	case "powershell":
		return root.GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}
