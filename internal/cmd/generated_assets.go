package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

type GeneratedDocFile struct {
	Path    string
	Content []byte
}

type commandMetadata struct {
	Path        string            `json:"path"`
	Use         string            `json:"use"`
	UseLine     string            `json:"use_line"`
	Short       string            `json:"short,omitempty"`
	Long        string            `json:"long,omitempty"`
	Aliases     []string          `json:"aliases,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
	Flags       []flagMetadata    `json:"flags,omitempty"`
	Subcommands []commandMetadata `json:"subcommands,omitempty"`
}

type flagMetadata struct {
	Name       string `json:"name"`
	Shorthand  string `json:"shorthand,omitempty"`
	Usage      string `json:"usage,omitempty"`
	Default    string `json:"default,omitempty"`
	Required   bool   `json:"required,omitempty"`
	Inherited  bool   `json:"inherited,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
	Deprecated string `json:"deprecated,omitempty"`
}

func GenerateExamplesDoc() (string, error) {
	root := NewRootCmd()

	var buf bytes.Buffer
	buf.WriteString("# Command Examples\n\n")
	buf.WriteString("Generated from Cobra `Example` fields and validated by the docs example tests.\n")

	for _, command := range executableCommands(root) {
		examples := exampleLines(command)
		if len(examples) == 0 {
			continue
		}
		fmt.Fprintf(&buf, "\n## `%s`\n\n", command.CommandPath())
		buf.WriteString("```bash\n")
		for _, example := range examples {
			buf.WriteString(example)
			buf.WriteByte('\n')
		}
		buf.WriteString("```\n")
	}

	return buf.String(), nil
}

func GenerateCommandMetadataJSON() (string, error) {
	root := NewRootCmd()

	payload := struct {
		Command commandMetadata `json:"command"`
	}{
		Command: buildCommandMetadata(root),
	}

	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw) + "\n", nil
}

func GenerateCompletionFiles() ([]GeneratedDocFile, error) {
	root := NewRootCmd()
	shells := []struct {
		Name string
		File string
	}{
		{Name: "bash", File: filepath.Join("docs", "completions", "bb.bash")},
		{Name: "zsh", File: filepath.Join("docs", "completions", "_bb")},
		{Name: "fish", File: filepath.Join("docs", "completions", "bb.fish")},
		{Name: "powershell", File: filepath.Join("docs", "completions", "bb.ps1")},
	}

	files := make([]GeneratedDocFile, 0, len(shells))
	for _, shell := range shells {
		var buf bytes.Buffer
		if err := writeCompletion(root, shell.Name, &buf); err != nil {
			return nil, err
		}
		files = append(files, GeneratedDocFile{
			Path:    shell.File,
			Content: buf.Bytes(),
		})
	}

	return files, nil
}

func GenerateManPages() ([]GeneratedDocFile, error) {
	root := NewRootCmd()

	tmpDir, err := os.MkdirTemp("", "bb-manpages-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	header := &doc.GenManHeader{
		Title:   "BB",
		Section: "1",
		Source:  "bb",
		Manual:  "Bitbucket CLI",
	}
	if err := doc.GenManTree(root, header, tmpDir); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, err
	}

	files := make([]GeneratedDocFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedDocFile{
			Path:    filepath.Join("docs", "man", entry.Name()),
			Content: data,
		})
	}

	slices.SortFunc(files, func(a, b GeneratedDocFile) int {
		return strings.Compare(a.Path, b.Path)
	})
	return files, nil
}

func buildCommandMetadata(command *cobra.Command) commandMetadata {
	children := visibleSubcommands(command.Commands())
	subcommands := make([]commandMetadata, 0, len(children))
	for _, child := range children {
		subcommands = append(subcommands, buildCommandMetadata(child))
	}

	return commandMetadata{
		Path:        command.CommandPath(),
		Use:         command.Use,
		UseLine:     command.UseLine(),
		Short:       command.Short,
		Long:        strings.TrimSpace(command.Long),
		Aliases:     slices.Clone(command.Aliases),
		Examples:    exampleLines(command),
		Flags:       collectFlagMetadata(command),
		Subcommands: subcommands,
	}
}

func collectFlagMetadata(command *cobra.Command) []flagMetadata {
	entries := make([]flagMetadata, 0)
	appendFlags := func(flagSet *pflag.FlagSet, inherited bool) {
		if flagSet == nil {
			return
		}
		flagSet.VisitAll(func(flag *pflag.Flag) {
			if flag.Name == "help" {
				return
			}
			entries = append(entries, flagMetadata{
				Name:       flag.Name,
				Shorthand:  flag.Shorthand,
				Usage:      strings.TrimSpace(flag.Usage),
				Default:    flag.DefValue,
				Required:   flagSet.Lookup(flag.Name).Annotations != nil && len(flagSet.Lookup(flag.Name).Annotations[cobra.BashCompOneRequiredFlag]) > 0,
				Inherited:  inherited,
				Hidden:     flag.Hidden,
				Deprecated: flag.Deprecated,
			})
		})
	}

	appendFlags(command.NonInheritedFlags(), false)
	appendFlags(command.InheritedFlags(), true)
	slices.SortFunc(entries, func(a, b flagMetadata) int {
		if a.Inherited != b.Inherited {
			if a.Inherited {
				return 1
			}
			return -1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return entries
}

func exampleLines(command *cobra.Command) []string {
	examples := strings.TrimSpace(command.Example)
	if examples == "" {
		return nil
	}
	lines := strings.Split(formatCommandExamples(examples), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}
