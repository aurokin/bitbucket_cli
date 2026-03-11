package cmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
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
		Long:  "Manage persistent bb settings such as prompt behavior, browser, editor, pager, and default output format.",
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

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get one effective bb setting",
		Example: "  bb config get prompt\n" +
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

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a persistent bb setting",
		Example: "  bb config set prompt false\n" +
			"  bb config set output.format json\n" +
			"  bb config set editor 'code --wait'",
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

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Set %s=%v\n", value.Key, value.Value)
			return err
		},
	}

	return cmd
}

func newConfigUnsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a persistent bb setting",
		Example: "  bb config unset prompt\n" +
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

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Unset %s\n", normalizeConfigKey(args[0]))
			return err
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
	values := make([]configValue, 0, 5)
	for _, key := range []string{"prompt", "browser", "editor", "pager", "output.format"} {
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
	case "browser":
		return configStringValue("browser", cfg.Settings.Browser), nil
	case "editor":
		return configStringValue("editor", cfg.Settings.Editor), nil
	case "pager":
		return configStringValue("pager", cfg.Settings.Pager), nil
	case "output.format":
		source := "default"
		if cfg.Settings.OutputFormat != "" {
			source = "config"
		}
		return configValue{Key: "output.format", Value: cfg.EffectiveOutputFormat(), Source: source}, nil
	default:
		return configValue{}, fmt.Errorf("unknown config key %q; supported keys: prompt, browser, editor, pager, output.format", key)
	}
}

func configStringValue(key, value string) configValue {
	source := "default"
	if value != "" {
		source = "config"
	}
	return configValue{Key: key, Value: value, Source: source}
}

func setConfigValue(cfg *config.Config, key, rawValue string) error {
	switch normalizeConfigKey(key) {
	case "prompt":
		value, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("prompt must be true or false")
		}
		cfg.Settings.Prompt = &value
	case "browser":
		cfg.Settings.Browser = rawValue
	case "editor":
		cfg.Settings.Editor = rawValue
	case "pager":
		cfg.Settings.Pager = rawValue
	case "output.format":
		switch rawValue {
		case config.OutputFormatTable:
			cfg.Settings.OutputFormat = ""
		case config.OutputFormatJSON:
			cfg.Settings.OutputFormat = config.OutputFormatJSON
		default:
			return fmt.Errorf("output.format must be %q or %q", config.OutputFormatTable, config.OutputFormatJSON)
		}
	default:
		return fmt.Errorf("unknown config key %q; supported keys: prompt, browser, editor, pager, output.format", key)
	}

	cfg.Settings = config.NormalizeSettings(cfg.Settings)
	return nil
}

func unsetConfigValue(cfg *config.Config, key string) error {
	switch normalizeConfigKey(key) {
	case "prompt":
		cfg.Settings.Prompt = nil
	case "browser":
		cfg.Settings.Browser = ""
	case "editor":
		cfg.Settings.Editor = ""
	case "pager":
		cfg.Settings.Pager = ""
	case "output.format":
		cfg.Settings.OutputFormat = ""
	default:
		return fmt.Errorf("unknown config key %q; supported keys: prompt, browser, editor, pager, output.format", key)
	}

	cfg.Settings = config.NormalizeSettings(cfg.Settings)
	return nil
}

func normalizeConfigKey(key string) string {
	switch key {
	case "output", "format":
		return "output.format"
	default:
		return key
	}
}
