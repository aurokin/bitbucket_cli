package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowDocsCoverMajorCommandFamilies(t *testing.T) {
	t.Parallel()

	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "workflows.md"), []string{
		"bb auth login",
		"bb repo view",
		"bb pr view",
		"bb issue create",
		"bb search prs",
		"bb status",
	})
}

func TestAutomationDocsCoverMajorCommandFamilies(t *testing.T) {
	t.Parallel()

	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "automation.md"), []string{
		"bb auth login",
		"bb repo view",
		"bb pr list",
		"bb issue list",
		"bb search repos",
		"bb status",
		"bb config set",
		"bb alias set",
		"bb extension list",
	})
}

func TestSupportingDocsExistAndLinkKeyTopics(t *testing.T) {
	t.Parallel()

	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "json-shapes.md"), []string{
		"bb repo view",
		"bb pr status",
		"bb issue view",
		"bb status",
	})
	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "json-fields.md"), []string{
		"bb repo view",
		"bb pr status",
		"bb issue view",
		"bb auth status",
	})
	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "flag-matrix.md"), []string{
		"bb pr create",
		"--no-prompt",
		"--repo",
		"--json",
	})
	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "error-index.md"), []string{
		"authentication failed",
		"Missing Token Scopes Or Insufficient Access",
		"bb auth login",
	})
	assertFileContainsAll(t, filepath.Join("..", "..", "docs", "recovery.md"), []string{
		"bb auth login",
		"bb auth status --check",
		"--repo <workspace>/<repo>",
		"bb alias set",
	})
}

func assertFileContainsAll(t *testing.T, path string, fragments []string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	content := string(data)
	for _, fragment := range fragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("%s does not contain %q", path, fragment)
		}
	}
}
