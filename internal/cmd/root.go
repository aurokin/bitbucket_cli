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
