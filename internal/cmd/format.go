package cmd

import (
	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type formatFlags struct {
	json string
	jq   string
}

func (f *formatFlags) options() (output.FormatOptions, error) {
	if f.json == "" && f.jq == "" {
		cfg, err := config.Load()
		if err == nil && cfg.EffectiveOutputFormat() == config.OutputFormatJSON {
			return output.FormatOptions{AllFields: true}, nil
		}
	}

	return output.ParseFormatOptions(f.json, f.jq)
}

func addFormatFlags(cmd *cobra.Command, flags *formatFlags) {
	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
}
