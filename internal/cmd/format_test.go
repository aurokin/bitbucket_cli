package cmd

import (
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestFormatFlagsUseConfigDefaultJSON(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())
	if err := config.Save(config.Config{
		Settings: config.Settings{
			OutputFormat: config.OutputFormatJSON,
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	opts, err := (&formatFlags{}).options()
	if err != nil {
		t.Fatalf("options returned error: %v", err)
	}
	if !opts.AllFields {
		t.Fatal("expected config default json output to enable all fields")
	}
}
