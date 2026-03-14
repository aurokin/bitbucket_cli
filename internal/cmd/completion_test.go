package cmd

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

func TestWriteCompletionIncludesShellMarkersAndCommands(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	cases := []struct {
		shell    string
		markers  []string
		unwanted []string
	}{
		{
			shell:   "bash",
			markers: []string{"# bash completion V2 for bb", "__start_bb()", "__complete"},
		},
		{
			shell:   "zsh",
			markers: []string{"#compdef bb", "compdef _bb bb", "__complete"},
		},
		{
			shell:   "fish",
			markers: []string{"# fish completion for bb", "complete -c bb", "__complete"},
		},
		{
			shell:    "powershell",
			markers:  []string{"# powershell completion for bb", "Register-ArgumentCompleter", "__complete"},
			unwanted: []string{"# bash completion V2 for bb"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.shell, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			if err := writeCompletion(root, tc.shell, &out); err != nil {
				t.Fatalf("writeCompletion(%q) returned error: %v", tc.shell, err)
			}

			rendered := out.String()
			for _, marker := range tc.markers {
				if !strings.Contains(rendered, marker) {
					t.Fatalf("%s completion missing %q", tc.shell, marker)
				}
			}
			for _, marker := range tc.unwanted {
				if strings.Contains(rendered, marker) {
					t.Fatalf("%s completion unexpectedly included %q", tc.shell, marker)
				}
			}
		})
	}
}

func TestWriteCompletionRejectsUnsupportedShell(t *testing.T) {
	t.Parallel()

	err := writeCompletion(NewRootCmd(), "tcsh", &bytes.Buffer{})
	if err == nil || err.Error() != `unsupported shell "tcsh"` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateCompletionFilesReturnsExpectedPaths(t *testing.T) {
	t.Parallel()

	files, err := GenerateCompletionFiles()
	if err != nil {
		t.Fatalf("GenerateCompletionFiles returned error: %v", err)
	}

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
		if len(file.Content) == 0 {
			t.Fatalf("generated completion %s was empty", file.Path)
		}
	}
	slices.Sort(paths)

	expected := []string{
		"docs/completions/_bb",
		"docs/completions/bb.bash",
		"docs/completions/bb.fish",
		"docs/completions/bb.ps1",
	}
	if !slices.Equal(paths, expected) {
		t.Fatalf("unexpected completion paths: got %v want %v", paths, expected)
	}
}

func TestCompletionCommandExecution(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args   []string
		marker string
	}{
		{args: []string{"completion", "bash"}, marker: "# bash completion V2 for bb"},
		{args: []string{"completion", "zsh"}, marker: "#compdef bb"},
		{args: []string{"completion", "fish"}, marker: "# fish completion for bb"},
		{args: []string{"completion", "powershell"}, marker: "# powershell completion for bb"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(strings.Join(tc.args, "-"), func(t *testing.T) {
			t.Parallel()

			root := NewRootCmd()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(tc.args)

			if err := root.Execute(); err != nil {
				t.Fatalf("root.Execute(%v) returned error: %v", tc.args, err)
			}
			if !strings.Contains(out.String(), tc.marker) {
				t.Fatalf("output for %v missing %q", tc.args, tc.marker)
			}
		})
	}
}
