package cmd

import "github.com/spf13/cobra"

func newAPICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Make an authenticated Bitbucket API request",
		RunE:  newStubCommand("api", "Make an authenticated Bitbucket API request", "api").RunE,
	}
}
