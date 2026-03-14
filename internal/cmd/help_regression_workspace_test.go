package cmd

import (
	"strings"
	"testing"
)

func TestWorkspaceHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "workspace", "member", "list", "--help")
	for _, fragment := range []string{
		"bb workspace member list workspace-slug",
		"--workspace string",
		"--query string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("workspace help missing %q\n%s", fragment, output)
		}
	}
}

func TestProjectHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "project", "create", "--help")
	for _, fragment := range []string{
		"bb project create BBCLI --workspace workspace-slug --name 'bb cli integration'",
		"--name string",
		"--visibility string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("project help missing %q\n%s", fragment, output)
		}
	}
}

func TestDeploymentHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "deployment", "environment", "list", "--help")
	for _, fragment := range []string{
		"bb deployment environment list --repo workspace-slug/pipelines-repo-slug",
		"--repo string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("deployment help missing %q\n%s", fragment, output)
		}
	}

	output = renderHelp(t, "deployment", "environment", "view", "--help")
	for _, fragment := range []string{
		"bb deployment environment view '{environment-uuid}' --repo workspace-slug/pipelines-repo-slug --json environment",
		"Show deployment environment information",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("deployment help missing %q\n%s", fragment, output)
		}
	}

	output = renderHelp(t, "deployment", "environment", "variable", "list", "--help")
	for _, fragment := range []string{
		"bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test",
		"--environment string",
		"--limit int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("deployment help missing %q\n%s", fragment, output)
		}
	}

	output = renderHelp(t, "deployment", "environment", "variable", "create", "--help")
	for _, fragment := range []string{
		"bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key APP_ENV --value production",
		"--key string",
		"--value-file string",
		"--secured",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("deployment help missing %q\n%s", fragment, output)
		}
	}
}
