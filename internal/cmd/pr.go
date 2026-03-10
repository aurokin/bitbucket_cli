package cmd

import "github.com/spf13/cobra"

func newPRCmd() *cobra.Command {
	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Work with pull requests",
	}

	prCmd.AddCommand(
		newStubCommand("list", "List pull requests", "pr list"),
		newStubCommand("view", "View a pull request", "pr view"),
		newStubCommand("create", "Create a pull request", "pr create"),
		newStubCommand("checkout", "Check out a pull request locally", "pr checkout"),
	)

	return prCmd
}
