package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bb",
		Short:         "Bitbucket CLI",
		Long:          "bb is a Bitbucket Cloud CLI for humans and automation. Use --json and --jq for structured output.",
		Example:       "  bb auth login --username you@example.com --with-token\n  bb repo view\n  bb pr list --json id,title,state\n  bb api /user",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.AddCommand(
		newVersionCmd(),
		newRepoCmd(),
		newAuthCmd(),
		newPRCmd(),
		newAPICmd(),
	)

	return rootCmd
}

func Execute() error {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs(normalizeCLIArgs(os.Args[1:]))
	return rootCmd.Execute()
}

func normalizeCLIArgs(args []string) []string {
	normalized := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg != "--json" {
			normalized = append(normalized, arg)
			continue
		}

		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			normalized = append(normalized, "--json="+args[i+1])
			i++
			continue
		}

		normalized = append(normalized, "--json=*")
	}

	return normalized
}
