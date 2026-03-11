package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type aliasEntry struct {
	Name      string `json:"name"`
	Expansion string `json:"expansion"`
}

type extensionEntry struct {
	Name       string `json:"name"`
	Executable string `json:"executable"`
}

func newAliasCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage command aliases",
		Long:  "Manage persistent command aliases stored in the bb config file.",
	}

	aliasCmd.AddCommand(
		newAliasListCmd(),
		newAliasGetCmd(),
		newAliasSetCmd(),
		newAliasDeleteCmd(),
	)

	return aliasCmd
}

func newAliasListCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			aliases := listAliasEntries(cfg.Aliases)
			return output.Render(cmd.OutOrStdout(), opts, aliases, func(w io.Writer) error {
				if len(aliases) == 0 {
					_, err := fmt.Fprintln(w, "No aliases configured.")
					return err
				}

				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintln(tw, "name\texpansion"); err != nil {
					return err
				}
				for _, alias := range aliases {
					if _, err := fmt.Fprintf(tw, "%s\t%s\n", alias.Name, alias.Expansion); err != nil {
						return err
					}
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")

	return cmd
}

func newAliasGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Show one configured alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			expansion, ok := cfg.Aliases[args[0]]
			if !ok {
				return fmt.Errorf("no alias named %q", args[0])
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), expansion)
			return err
		},
	}

	return cmd
}

func newAliasSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name> <expansion...>",
		Short: "Create or replace an alias",
		Example: "  bb alias set pv 'pr view'\n" +
			"  bb alias set rls 'pr list --state OPEN'",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("alias name is required")
			}
			if cfg.Aliases == nil {
				cfg.Aliases = map[string]string{}
			}

			expansion := strings.Join(args[1:], " ")
			cfg.Aliases[name] = strings.TrimSpace(expansion)
			if err := config.Save(cfg); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Set alias %s=%s\n", name, cfg.Aliases[name])
			return err
		},
	}

	return cmd
}

func newAliasDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"remove", "rm"},
		Short:   "Delete an alias",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if _, ok := cfg.Aliases[args[0]]; !ok {
				return fmt.Errorf("no alias named %q", args[0])
			}

			delete(cfg.Aliases, args[0])
			if err := config.Save(cfg); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted alias %s\n", args[0])
			return err
		},
	}

	return cmd
}

func newExtensionCmd() *cobra.Command {
	extCmd := &cobra.Command{
		Use:   "extension",
		Short: "Discover and run external bb commands",
		Long:  "Discover and run external commands named bb-<name> from PATH.",
	}

	extCmd.AddCommand(
		newExtensionListCmd(),
		newExtensionExecCmd(),
	)

	return extCmd
}

func newExtensionListCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered external bb commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			entries, err := discoverExtensions()
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, entries, func(w io.Writer) error {
				if len(entries) == 0 {
					_, err := fmt.Fprintln(w, "No extensions found in PATH.")
					return err
				}

				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintln(tw, "name\texecutable"); err != nil {
					return err
				}
				for _, entry := range entries {
					if _, err := fmt.Fprintf(tw, "%s\t%s\n", entry.Name, entry.Executable); err != nil {
						return err
					}
				}
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")

	return cmd
}

func newExtensionExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <name> [args...]",
		Short: "Run an external bb command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtensionCommand(args[0], args[1:])
		},
	}

	return cmd
}

func listAliasEntries(aliases map[string]string) []aliasEntry {
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]aliasEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, aliasEntry{Name: name, Expansion: aliases[name]})
	}
	return entries
}

func discoverExtensions() ([]extensionEntry, error) {
	seen := map[string]string{}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if strings.TrimSpace(dir) == "" {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, "bb-") || name == "bb" {
				continue
			}

			fullPath := filepath.Join(dir, name)
			info, err := entry.Info()
			if err != nil || info.Mode()&0o111 == 0 {
				continue
			}

			commandName := strings.TrimPrefix(name, "bb-")
			if _, ok := seen[commandName]; !ok {
				seen[commandName] = fullPath
			}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)

	extensions := make([]extensionEntry, 0, len(names))
	for _, name := range names {
		extensions = append(extensions, extensionEntry{Name: name, Executable: seen[name]})
	}
	return extensions, nil
}

func executeExternalCommand(executable string, args []string) error {
	command := exec.Command(executable, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
