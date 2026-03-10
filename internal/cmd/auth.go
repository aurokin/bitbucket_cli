package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	authCmd.AddCommand(
		newStubCommand("login", "Authenticate with Bitbucket", "auth login"),
		newStubCommand("status", "Show authentication status", "auth status"),
		newStubCommand("logout", "Clear stored authentication", "auth logout"),
	)

	return authCmd
}

func newStubCommand(use, short, feature string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s is not implemented yet", feature)
		},
	}
}
