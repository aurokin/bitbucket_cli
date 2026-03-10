package cmd

import "github.com/auro/bitbucket_cli/internal/output"

type formatFlags struct {
	json string
	jq   string
}

func (f *formatFlags) options() (output.FormatOptions, error) {
	return output.ParseFormatOptions(f.json, f.jq)
}
