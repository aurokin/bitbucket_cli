package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIReferenceMatchesGenerator(t *testing.T) {
	t.Parallel()

	generated, err := GenerateCLIReference()
	if err != nil {
		t.Fatalf("GenerateCLIReference returned error: %v", err)
	}

	path := filepath.Join("..", "..", "docs", "cli-reference.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cli reference: %v", err)
	}

	if string(existing) != generated {
		t.Fatalf("docs/cli-reference.md is out of date; run `go run ./cmd/gen-docs`")
	}
}

func TestRootHelpHighlightsHumanAndAgentPaths(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "--help")
	for _, fragment := range []string{
		"Prefer --repo <workspace>/<repo> for explicit targeting.",
		"bb issue create --repo OhBizzle/bb-cli-integration-issues --title 'Broken flow'",
		"bb status --json authored_prs,review_requested_prs,your_issues",
		"--no-prompt",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("root help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueCreateHelpShowsExplicitRepoExamples(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "create", "--help")
	for _, fragment := range []string{
		"bb issue create --repo OhBizzle/bb-cli-integration-issues --title 'Broken flow'",
		"--repo string",
		"--no-prompt",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("issue create help missing %q\n%s", fragment, output)
		}
	}
}

func TestStatusHelpShowsBoundedExamples(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "status", "--help")
	for _, fragment := range []string{
		"bb status --workspace OhBizzle --limit 10",
		"--repo-limit int",
		"Maximum items to return per status section",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("status help missing %q\n%s", fragment, output)
		}
	}
}
