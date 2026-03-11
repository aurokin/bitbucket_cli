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

func TestJSONShapesMatchesGenerator(t *testing.T) {
	t.Parallel()

	generated, err := GenerateJSONShapesDoc()
	if err != nil {
		t.Fatalf("GenerateJSONShapesDoc returned error: %v", err)
	}

	path := filepath.Join("..", "..", "docs", "json-shapes.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json shapes: %v", err)
	}

	if string(existing) != generated {
		t.Fatalf("docs/json-shapes.md is out of date; run `go run ./cmd/gen-docs`")
	}
}

func TestRecoveryDocMatchesGenerator(t *testing.T) {
	t.Parallel()

	generated, err := GenerateRecoveryDoc()
	if err != nil {
		t.Fatalf("GenerateRecoveryDoc returned error: %v", err)
	}

	path := filepath.Join("..", "..", "docs", "recovery.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read recovery doc: %v", err)
	}

	if string(existing) != generated {
		t.Fatalf("docs/recovery.md is out of date; run `go run ./cmd/gen-docs`")
	}
}

func TestFlagMatrixMatchesGenerator(t *testing.T) {
	t.Parallel()

	generated, err := GenerateFlagMatrixDoc()
	if err != nil {
		t.Fatalf("GenerateFlagMatrixDoc returned error: %v", err)
	}

	path := filepath.Join("..", "..", "docs", "flag-matrix.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read flag matrix: %v", err)
	}

	if string(existing) != generated {
		t.Fatalf("docs/flag-matrix.md is out of date; run `go run ./cmd/gen-docs`")
	}
}

func TestJSONFieldsMatchesGenerator(t *testing.T) {
	t.Parallel()

	generated, err := GenerateJSONFieldsDoc()
	if err != nil {
		t.Fatalf("GenerateJSONFieldsDoc returned error: %v", err)
	}

	path := filepath.Join("..", "..", "docs", "json-fields.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json fields: %v", err)
	}

	if string(existing) != generated {
		t.Fatalf("docs/json-fields.md is out of date; run `go run ./cmd/gen-docs`")
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
