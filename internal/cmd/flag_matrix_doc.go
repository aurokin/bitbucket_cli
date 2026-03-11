package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func GenerateFlagMatrixDoc() (string, error) {
	root := NewRootCmd()

	var b strings.Builder
	b.WriteString("# Flag Matrix\n\n")
	b.WriteString("Generated from the `bb` command tree.\n\n")
	b.WriteString("This is a compact view of common automation and targeting flags across executable commands.\n\n")
	b.WriteString("| Command | `--json` | `--jq` | `--no-prompt` | `--host` | `--repo` | `--workspace` | Other notable flags |\n")
	b.WriteString("|---|---|---|---|---|---|---|---|\n")

	for _, command := range executableCommands(root) {
		flags := commandFlagSet(command)
		others := otherFlagNames(flags, "help", "json", "jq", "no-prompt", "host", "repo", "workspace")
		fmt.Fprintf(
			&b,
			"| `%s` | %s | %s | %s | %s | %s | %s | %s |\n",
			command.CommandPath(),
			flagMarker(flags, "json"),
			flagMarker(flags, "jq"),
			flagMarker(flags, "no-prompt"),
			flagMarker(flags, "host"),
			flagMarker(flags, "repo"),
			flagMarker(flags, "workspace"),
			joinFlagList(others),
		)
	}

	return b.String(), nil
}

func executableCommands(root *cobra.Command) []*cobra.Command {
	commands := make([]*cobra.Command, 0)
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		children := visibleSubcommands(command.Commands())
		if len(children) == 0 && command != root {
			commands = append(commands, command)
			return
		}
		for _, child := range children {
			walk(child)
		}
	}
	walk(root)
	slices.SortFunc(commands, func(a, b *cobra.Command) int {
		return strings.Compare(a.CommandPath(), b.CommandPath())
	})
	return commands
}

func commandFlagSet(command *cobra.Command) map[string]struct{} {
	flags := map[string]struct{}{}
	appendFlags := func(flagSet *pflag.FlagSet) {
		if flagSet == nil {
			return
		}
		flagSet.VisitAll(func(flag *pflag.Flag) {
			flags[flag.Name] = struct{}{}
		})
	}
	appendFlags(command.NonInheritedFlags())
	appendFlags(command.InheritedFlags())
	return flags
}

func flagMarker(flags map[string]struct{}, name string) string {
	if _, ok := flags[name]; ok {
		return "yes"
	}
	return ""
}

func otherFlagNames(flags map[string]struct{}, excluded ...string) []string {
	exclude := map[string]struct{}{}
	for _, name := range excluded {
		exclude[name] = struct{}{}
	}

	names := make([]string, 0)
	for name := range flags {
		if _, ok := exclude[name]; ok {
			continue
		}
		names = append(names, fmt.Sprintf("`--%s`", name))
	}
	slices.Sort(names)
	return names
}

func joinFlagList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}
