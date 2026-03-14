package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestGenerateManPagesIncludesExpectedCommands(t *testing.T) {
	t.Parallel()

	files, err := GenerateManPages()
	if err != nil {
		t.Fatalf("GenerateManPages returned error: %v", err)
	}

	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	slices.Sort(paths)

	for _, expected := range []string{
		"docs/man/bb.1",
		"docs/man/bb-completion.1",
		"docs/man/bb-completion-bash.1",
		"docs/man/bb-pr-create.1",
		"docs/man/bb-pipeline-stop.1",
		"docs/man/bb-resolve.1",
	} {
		if !slices.Contains(paths, expected) {
			t.Fatalf("generated man pages missing %s", expected)
		}
	}
}

func TestGeneratedRootManPageContainsStableSections(t *testing.T) {
	t.Parallel()

	content := generatedManPageContent(t, "docs/man/bb.1")
	for _, fragment := range []string{
		`.SH NAME`,
		`bb - Bitbucket CLI`,
		`.SH SYNOPSIS`,
		`.SH DESCRIPTION`,
		`.SH OPTIONS`,
		`.SH EXAMPLE`,
		`bb auth login --username you@example.com --with-token`,
		`\fBbb-completion(1)\fP`,
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("root man page missing %q", fragment)
		}
	}
}

func TestGeneratedSubcommandManPagesContainStableDetails(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path      string
		fragments []string
	}{
		{
			path: "docs/man/bb-pr-create.1",
			fragments: []string{
				`.SH NAME`,
				`bb-pr-create - Create a pull request`,
				`.SH OPTIONS`,
				`--title`,
				`bb pr create --reuse-existing --json`,
			},
		},
		{
			path: "docs/man/bb-pipeline-stop.1",
			fragments: []string{
				`.SH NAME`,
				`bb-pipeline-stop - Stop a running pipeline`,
				`.SH OPTIONS INHERITED FROM PARENT COMMANDS`,
				`--yes`,
				`bb pipeline stop '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'`,
			},
		},
		{
			path: "docs/man/bb-completion-bash.1",
			fragments: []string{
				`.SH NAME`,
				`bb-completion-bash - Generate a bash completion script`,
				`.SH SYNOPSIS`,
				`bb completion bash`,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			content := generatedManPageContent(t, tc.path)
			for _, fragment := range tc.fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", tc.path, fragment)
				}
			}
		})
	}
}

func generatedManPageContent(t *testing.T, path string) string {
	t.Helper()

	files, err := GenerateManPages()
	if err != nil {
		t.Fatalf("GenerateManPages returned error: %v", err)
	}

	for _, file := range files {
		if file.Path == path {
			return string(file.Content)
		}
	}

	t.Fatalf("generated man page %s not found", path)
	return ""
}
