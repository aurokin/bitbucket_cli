package cmd

import (
	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
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
