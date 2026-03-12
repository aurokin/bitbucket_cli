package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type configValue struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage bb behavior defaults",
		Long:  "Manage persistent bb settings that affect runtime behavior today, such as prompt behavior, browser selection for `bb browse`, and the default output format.",
	}

	configCmd.AddCommand(
		newConfigListCmd(),
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigUnsetCmd(),
		newConfigPathCmd(),
	)

	return configCmd
}

func newConfigListCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List effective bb settings",
		Example: "  bb config list\n" +
			"  bb config list --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			values := effectiveConfigValues(cfg)
			return output.Render(cmd.OutOrStdout(), opts, values, func(w io.Writer) error {
				tw := output.NewTableWriter(w)
				if _, err := fmt.Fprintln(tw, "key\tvalue\tsource"); err != nil {
					return err
				}
				for _, value := range values {
					if _, err := fmt.Fprintf(tw, "%s\t%v\t%s\n", value.Key, value.Value, value.Source); err != nil {
						return err
					}
				}
				return tw.Flush()
			})
		},
	}

	addFormatFlags(cmd, &flags)

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get one effective bb setting",
		Example: "  bb config get prompt\n" +
			"  bb config get browser\n" +
			"  bb config get output.format --json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			value, err := configValueForKey(cfg, args[0])
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, value, func(w io.Writer) error {
				_, err := fmt.Fprintf(w, "%v\n", value.Value)
				return err
			})
		},
	}

	addFormatFlags(cmd, &flags)

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a persistent bb setting",
		Example: "  bb config set prompt false\n" +
			"  bb config set browser 'firefox --new-window'\n" +
			"  bb config set output.format json\n" +
			"  bb config get output.format",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := setConfigValue(&cfg, args[0], args[1]); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}

			value, err := configValueForKey(cfg, args[0])
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Set %s=%v\n", value.Key, value.Value); err != nil {
				return err
			}
			return writeNextStep(cmd.OutOrStdout(), fmt.Sprintf("bb config get %s", value.Key))
		},
	}

	return cmd
}

func newConfigUnsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a persistent bb setting",
		Example: "  bb config unset prompt\n" +
			"  bb config unset browser\n" +
			"  bb config unset output.format",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := unsetConfigValue(&cfg, args[0]); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}

			key := normalizeConfigKey(args[0])
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Unset %s\n", key); err != nil {
				return err
			}
			return writeNextStep(cmd.OutOrStdout(), "bb config list")
		},
	}

	return cmd
}

func newConfigPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show the bb config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}

	return cmd
}

func effectiveConfigValues(cfg config.Config) []configValue {
	values := make([]configValue, 0, 3)
	for _, key := range []string{"prompt", "browser", "output.format"} {
		value, err := configValueForKey(cfg, key)
		if err == nil {
			values = append(values, value)
		}
	}
	return values
}

func configValueForKey(cfg config.Config, key string) (configValue, error) {
	switch normalizeConfigKey(key) {
	case "prompt":
		source := "default"
		if cfg.Settings.Prompt != nil {
			source = "config"
		}
		return configValue{Key: "prompt", Value: cfg.PromptEnabled(), Source: source}, nil
	case "output.format":
		source := "default"
		if cfg.Settings.OutputFormat != "" {
			source = "config"
		}
		return configValue{Key: "output.format", Value: cfg.EffectiveOutputFormat(), Source: source}, nil
	case "browser":
		source := "default"
		value := "system"
		if cfg.Settings.Browser != "" {
			source = "config"
			value = cfg.Settings.Browser
		}
		return configValue{Key: "browser", Value: value, Source: source}, nil
	default:
		return configValue{}, unsupportedConfigKeyError(key)
	}
}

func setConfigValue(cfg *config.Config, key, rawValue string) error {
	switch normalizeConfigKey(key) {
	case "prompt":
		value, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("prompt must be true or false")
		}
		cfg.Settings.Prompt = &value
	case "output.format":
		switch rawValue {
		case config.OutputFormatTable:
			cfg.Settings.OutputFormat = ""
		case config.OutputFormatJSON:
			cfg.Settings.OutputFormat = config.OutputFormatJSON
		default:
			return fmt.Errorf("output.format must be %q or %q", config.OutputFormatTable, config.OutputFormatJSON)
		}
	case "browser":
		value := strings.TrimSpace(rawValue)
		if value == "" {
			return fmt.Errorf("browser cannot be empty")
		}
		parts, err := splitCommandLine(value)
		if err != nil {
			return fmt.Errorf("browser command is invalid: %w", err)
		}
		if len(parts) == 0 {
			return fmt.Errorf("browser command cannot be empty")
		}
		cfg.Settings.Browser = value
	default:
		return unsupportedConfigKeyError(key)
	}

	cfg.Settings = config.NormalizeSettings(cfg.Settings)
	return nil
}

func unsetConfigValue(cfg *config.Config, key string) error {
	switch normalizeConfigKey(key) {
	case "prompt":
		cfg.Settings.Prompt = nil
	case "output.format":
		cfg.Settings.OutputFormat = ""
	case "browser":
		cfg.Settings.Browser = ""
	default:
		return unsupportedConfigKeyError(key)
	}

	cfg.Settings = config.NormalizeSettings(cfg.Settings)
	return nil
}

func unsupportedConfigKeyError(key string) error {
	switch normalizeConfigKey(key) {
	case "editor", "pager":
		return fmt.Errorf("config key %q is planned but not supported yet. Supported keys today: prompt, browser, output.format", key)
	default:
		return fmt.Errorf("unknown config key %q; supported keys: prompt, browser, output.format", key)
	}
}

func normalizeConfigKey(key string) string {
	switch key {
	case "output", "format":
		return "output.format"
	default:
		return key
	}
}
