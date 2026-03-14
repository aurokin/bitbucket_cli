package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMarkdownExamplesReferenceValidCommands(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("..", "..", "README.md"),
		filepath.Join("..", "..", "docs", "examples.md"),
		filepath.Join("..", "..", "docs", "workflows.md"),
		filepath.Join("..", "..", "docs", "automation.md"),
		filepath.Join("..", "..", "docs", "json-shapes.md"),
		filepath.Join("..", "..", "docs", "recovery.md"),
	}

	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}

			examples := markdownExamples(string(data))
			if len(examples) == 0 {
				t.Fatalf("no bb examples found in %s", file)
			}

			for _, example := range examples {
				example := example
				t.Run(example.name, func(t *testing.T) {
					t.Parallel()

					root := newDocsValidationRootCmd()
					root.SetIn(strings.NewReader(""))
					root.SetOut(ioDiscard{})
					root.SetErr(ioDiscard{})
					root.SetArgs(normalizeCLIArgs(example.args))

					if err := root.Execute(); err != nil {
						t.Fatalf("example %q in %s is invalid: %v", example.command, file, err)
					}
				})
			}
		})
	}
}

type markdownExample struct {
	name    string
	command string
	args    []string
}

func markdownExamples(markdown string) []markdownExample {
	lines := strings.Split(markdown, "\n")
	insideBlock := false
	blockLines := make([]string, 0)
	examples := make([]markdownExample, 0)

	flushBlock := func() {
		examples = append(examples, commandExamplesFromBlock(blockLines)...)
		blockLines = blockLines[:0]
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if insideBlock {
				flushBlock()
				insideBlock = false
				continue
			}
			lang := strings.TrimPrefix(trimmed, "```")
			if lang == "bash" || lang == "sh" {
				insideBlock = true
				continue
			}
		}
		if insideBlock {
			blockLines = append(blockLines, line)
		}
	}

	return examples
}

func commandExamplesFromBlock(lines []string) []markdownExample {
	joined := make([]string, 0)
	current := ""

	flush := func() {
		if strings.TrimSpace(current) != "" {
			joined = append(joined, strings.TrimSpace(current))
		}
		current = ""
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			flush()
			continue
		}
		if strings.HasSuffix(line, "\\") {
			current += strings.TrimSpace(strings.TrimSuffix(line, "\\")) + " "
			continue
		}
		current += line
		flush()
	}
	flush()

	examples := make([]markdownExample, 0)
	for index, line := range joined {
		command := extractBBCommand(line)
		if command == "" {
			continue
		}
		args, err := splitCommandLine(strings.TrimSpace(strings.TrimPrefix(command, "bb ")))
		if err != nil {
			continue
		}
		examples = append(examples, markdownExample{
			name:    fmt.Sprintf("%02d", index+1),
			command: command,
			args:    args,
		})
	}
	return examples
}

func extractBBCommand(line string) string {
	if strings.HasPrefix(line, "bb ") {
		return line
	}
	if index := strings.Index(line, "| bb "); index >= 0 {
		return strings.TrimSpace(line[index+2:])
	}
	return ""
}

func newDocsValidationRootCmd() *cobra.Command {
	root := NewRootCmd()
	disableCommandExecution(root)
	return root
}

func disableCommandExecution(command *cobra.Command) {
	command.Run = func(*cobra.Command, []string) {}
	command.RunE = func(*cobra.Command, []string) error { return nil }
	command.PreRun = nil
	command.PreRunE = nil
	command.PostRun = nil
	command.PostRunE = nil
	command.PersistentPreRun = nil
	command.PersistentPreRunE = nil
	command.PersistentPostRun = nil
	command.PersistentPostRunE = nil

	for _, child := range command.Commands() {
		disableCommandExecution(child)
	}
}

func renderHelp(t *testing.T, args ...string) string {
	t.Helper()

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatalf("render help %v: %v", args, err)
	}

	return out.String()
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}
