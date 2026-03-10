package cmd

import "fmt"

func validateRepoSelector(workspace, repo string) error {
	if (workspace == "" && repo == "") || (workspace != "" && repo != "") {
		return nil
	}

	if workspace == "" {
		return fmt.Errorf("--repo requires --workspace")
	}

	return fmt.Errorf("--workspace requires --repo")
}
