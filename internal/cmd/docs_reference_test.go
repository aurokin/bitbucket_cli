package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIReferenceMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "cli-reference.md"), GenerateCLIReference)
}

func TestExamplesDocMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "examples.md"), GenerateExamplesDoc)
}

func TestJSONShapesMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "json-shapes.md"), GenerateJSONShapesDoc)
}

func TestRecoveryDocMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "recovery.md"), GenerateRecoveryDoc)
}

func TestFlagMatrixMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "flag-matrix.md"), GenerateFlagMatrixDoc)
}

func TestErrorIndexMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "error-index.md"), GenerateErrorIndexDoc)
}

func TestJSONFieldsMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "json-fields.md"), GenerateJSONFieldsDoc)
}

func TestCommandMetadataMatchesGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedDocMatches(t, filepath.Join("..", "..", "docs", "command-metadata.json"), GenerateCommandMetadataJSON)
}

func TestCompletionFilesMatchGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedFileSetMatches(t, GenerateCompletionFiles)
}

func TestManPagesMatchGenerator(t *testing.T) {
	t.Parallel()
	assertGeneratedFileSetMatches(t, GenerateManPages)
}

func TestRootHelpHighlightsHumanAndAgentPaths(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "--help")
	for _, fragment := range []string{
		"Prefer --repo <workspace>/<repo> for explicit targeting.",
		"bb browse --repo workspace-slug/repo-slug --no-browser",
		"bb pipeline list --repo workspace-slug/pipelines-repo-slug",
		"bb issue create --repo workspace-slug/issues-repo-slug --title 'Broken flow'",
		"bb status --json authored_prs,review_requested_prs,your_issues",
		"--no-prompt",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("root help missing %q\n%s", fragment, output)
		}
	}
}

func TestPipelineHelpShowsExplicitRepoExamples(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "pipeline", "list", "--help")
	for _, fragment := range []string{
		"bb pipeline list --repo workspace-slug/repo-slug",
		"--repo string",
		"--json string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("pipeline list help missing %q\n%s", fragment, output)
		}
	}
}

func TestBrowseHelpShowsExplicitRepoExamples(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "browse", "--help")
	for _, fragment := range []string{
		"bb browse --repo workspace-slug/repo-slug",
		"--no-browser",
		"--pr int",
		"--issue int",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("browse help missing %q\n%s", fragment, output)
		}
	}
}

func TestResolveHelpShowsPRCommentExample(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "resolve", "--help")
	for _, fragment := range []string{
		"bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7",
		"#comment-15",
		"--json string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("resolve help missing %q\n%s", fragment, output)
		}
	}
}

func TestIssueCreateHelpShowsExplicitRepoExamples(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "issue", "create", "--help")
	for _, fragment := range []string{
		"bb issue create --repo workspace-slug/issues-repo-slug --title 'Broken flow'",
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
		"bb status --workspace workspace-slug --limit 10",
		"--repo-limit int",
		"Maximum items to return per status section",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("status help missing %q\n%s", fragment, output)
		}
	}
}

func assertGeneratedDocMatches(t *testing.T, path string, generate func() (string, error)) {
	t.Helper()

	generated, err := generate()
	if err != nil {
		t.Fatalf("generator returned error: %v", err)
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated doc %s: %v", path, err)
	}

	if string(existing) != generated {
		t.Fatalf("%s is out of date; run `go run ./cmd/gen-docs`", pathForError(path))
	}
}

func assertGeneratedFileSetMatches(t *testing.T, generate func() ([]GeneratedDocFile, error)) {
	t.Helper()

	generated, err := generate()
	if err != nil {
		t.Fatalf("file generator returned error: %v", err)
	}

	for _, file := range generated {
		existing, err := os.ReadFile(filepath.Join("..", "..", file.Path))
		if err != nil {
			t.Fatalf("read generated file %s: %v", file.Path, err)
		}
		if !bytes.Equal(existing, file.Content) {
			t.Fatalf("%s is out of date; run `go run ./cmd/gen-docs`", file.Path)
		}
	}
}

func pathForError(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(path), "../../")
}
