package cmd

import "github.com/spf13/cobra"

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
	return NewRootCmd().Execute()
}
