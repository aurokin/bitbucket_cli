package cmd

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GenerateCLIReference renders a markdown command reference directly from the
// Cobra command tree so the written docs can stay task-oriented.
func GenerateCLIReference() (string, error) {
	root := NewRootCmd()

	var buf bytes.Buffer
	buf.WriteString("# CLI Reference\n\n")
	buf.WriteString("Generated from the `bb` command tree.\n\n")
	buf.WriteString("Use this file for the full command surface. Keep [README.md](../README.md) focused on workflows.\n")

	commands := visibleSubcommands(root.Commands())
	if len(commands) > 0 {
		buf.WriteString("\n## Command Tree\n\n")
		for _, command := range commands {
			writeCommandTree(&buf, command, 0)
		}
	}

	for _, command := range commands {
		writeCommandReference(&buf, command)
	}

	return buf.String(), nil
}

func writeCommandTree(buf *bytes.Buffer, command *cobra.Command, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(buf, "%s- `%s`\n", indent, command.CommandPath())
	for _, child := range visibleSubcommands(command.Commands()) {
		writeCommandTree(buf, child, depth+1)
	}
}

func writeCommandReference(buf *bytes.Buffer, command *cobra.Command) {
	path := command.CommandPath()
	buf.WriteString("\n## `")
	buf.WriteString(path)
	buf.WriteString("`\n\n")

	if command.Short != "" {
		buf.WriteString(command.Short)
		buf.WriteString("\n\n")
	}

	if long := strings.TrimSpace(command.Long); long != "" && long != command.Short {
		buf.WriteString(long)
		buf.WriteString("\n\n")
	}

	if aliases := command.Aliases; len(aliases) > 0 {
		buf.WriteString("Aliases: `")
		buf.WriteString(strings.Join(aliases, "`, `"))
		buf.WriteString("`\n\n")
	}

	buf.WriteString("Usage:\n\n```text\n")
	buf.WriteString(command.UseLine())
	buf.WriteString("\n```\n")

	if examples := strings.TrimSpace(command.Example); examples != "" {
		buf.WriteString("\nExamples:\n\n```bash\n")
		buf.WriteString(formatCommandExamples(examples))
		buf.WriteString("\n```\n")
	}

	if flags := collectFlags(command); len(flags) > 0 {
		buf.WriteString("\nFlags:\n\n")
		for _, line := range flags {
			buf.WriteString("- ")
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	children := visibleSubcommands(command.Commands())
	if len(children) > 0 {
		buf.WriteString("\nSubcommands:\n\n")
		for _, child := range children {
			fmt.Fprintf(buf, "- `%s`: %s\n", child.CommandPath(), child.Short)
		}
		for _, child := range children {
			writeCommandReference(buf, child)
		}
	}
}

func collectFlags(command *cobra.Command) []string {
	lines := make([]string, 0)
	seen := map[string]struct{}{}

	appendFlags := func(flags *pflag.FlagSet) {
		if flags == nil {
			return
		}
		flags.VisitAll(func(flag *pflag.Flag) {
			if flag.Name == "help" {
				return
			}
			if _, ok := seen[flag.Name]; ok {
				return
			}
			seen[flag.Name] = struct{}{}

			var prefix string
			if flag.Shorthand != "" {
				prefix = fmt.Sprintf("`-%s`, `--%s`", flag.Shorthand, flag.Name)
			} else {
				prefix = fmt.Sprintf("`--%s`", flag.Name)
			}

			usage := strings.TrimSpace(flag.Usage)
			if usage == "" {
				lines = append(lines, prefix)
				return
			}
			lines = append(lines, fmt.Sprintf("%s: %s", prefix, usage))
		})
	}

	appendFlags(command.NonInheritedFlags())
	appendFlags(command.InheritedFlags())
	slices.Sort(lines)
	return lines
}

func visibleSubcommands(commands []*cobra.Command) []*cobra.Command {
	visible := make([]*cobra.Command, 0, len(commands))
	for _, command := range commands {
		if command.Hidden {
			continue
		}
		if command.Name() == "help" {
			continue
		}
		visible = append(visible, command)
	}
	return visible
}

func formatCommandExamples(examples string) string {
	lines := strings.Split(examples, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, "  ")
	}
	return strings.Join(lines, "\n")
}
