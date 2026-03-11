package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bb",
		Short:         "Bitbucket CLI",
		Long:          "bb is a Bitbucket Cloud CLI for humans and automation. Use --json and --jq for structured output, and --no-prompt for deterministic non-interactive runs.",
		Example:       "  bb auth login --username you@example.com --with-token\n  bb repo view\n  bb pr list --json id,title,state\n  bb --no-prompt pr create --source feature --destination main --title 'Add feature'\n  bb api /user",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.PersistentFlags().Bool("no-prompt", false, "Do not prompt for missing input, even in an interactive terminal")

	rootCmd.AddCommand(
		newVersionCmd(),
		newConfigCmd(),
		newAliasCmd(),
		newExtensionCmd(),
		newSearchCmd(),
		newIssueCmd(),
		newRepoCmd(),
		newAuthCmd(),
		newPRCmd(),
		newAPICmd(),
	)

	return rootCmd
}

func Execute() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	rootCmd := NewRootCmd()
	args := normalizeCLIArgsWithAliases(os.Args[1:], cfg.Aliases)
	rootCmd.SetArgs(args)

	err = rootCmd.Execute()
	if err != nil && shouldRunExtension(err, args) {
		return runExtensionCommand(args[0], args[1:])
	}

	return userFacingError(err)
}

func normalizeCLIArgs(args []string) []string {
	return normalizeCLIArgsWithAliases(args, nil)
}

func normalizeCLIArgsWithAliases(args []string, aliases map[string]string) []string {
	expanded := expandAliasArgs(args, aliases)
	normalized := make([]string, 0, len(args))

	for i := 0; i < len(expanded); i++ {
		arg := expanded[i]
		if arg != "--json" {
			normalized = append(normalized, arg)
			continue
		}

		if i+1 < len(expanded) && !strings.HasPrefix(expanded[i+1], "-") {
			normalized = append(normalized, "--json="+expanded[i+1])
			i++
			continue
		}

		normalized = append(normalized, "--json=*")
	}

	return normalized
}

func expandAliasArgs(args []string, aliases map[string]string) []string {
	if len(args) == 0 || len(aliases) == 0 {
		return args
	}

	expanded := append([]string(nil), args...)
	seen := map[string]struct{}{}

	for depth := 0; depth < 8 && len(expanded) > 0; depth++ {
		replacement, ok := aliases[expanded[0]]
		if !ok || strings.TrimSpace(replacement) == "" {
			return expanded
		}
		if _, ok := seen[expanded[0]]; ok {
			return expanded
		}
		seen[expanded[0]] = struct{}{}

		fields := strings.Fields(replacement)
		if len(fields) == 0 {
			return expanded
		}
		expanded = append(fields, expanded[1:]...)
	}

	return expanded
}

func shouldRunExtension(err error, args []string) bool {
	return err != nil && len(args) > 0 && strings.HasPrefix(err.Error(), fmt.Sprintf("unknown command %q for \"bb\"", args[0]))
}

var (
	execLookPath        = exec.LookPath
	executeExternalFunc = executeExternalCommand
)

func runExtensionCommand(name string, args []string) error {
	executable, err := execLookPath("bb-" + name)
	if err != nil {
		return userFacingError(fmt.Errorf("unknown command %q for \"bb\"", name))
	}

	return executeExternalFunc(executable, args)
}
