package cmd

import (
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestConfigValueForKeyUsesDefaults(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}

	value, err := configValueForKey(cfg, "prompt")
	if err != nil {
		t.Fatalf("configValueForKey returned error: %v", err)
	}
	if value.Value != true || value.Source != "default" {
		t.Fatalf("unexpected prompt value %+v", value)
	}

	value, err = configValueForKey(cfg, "output.format")
	if err != nil {
		t.Fatalf("configValueForKey returned error: %v", err)
	}
	if value.Value != config.OutputFormatTable || value.Source != "default" {
		t.Fatalf("unexpected output format value %+v", value)
	}

	value, err = configValueForKey(cfg, "browser")
	if err != nil {
		t.Fatalf("configValueForKey returned error: %v", err)
	}
	if value.Value != "system" || value.Source != "default" {
		t.Fatalf("unexpected browser value %+v", value)
	}
}

func TestSetAndUnsetConfigValue(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	if err := setConfigValue(&cfg, "prompt", "false"); err != nil {
		t.Fatalf("setConfigValue returned error: %v", err)
	}
	if cfg.PromptEnabled() {
		t.Fatal("expected prompts to be disabled")
	}

	if err := setConfigValue(&cfg, "output.format", "json"); err != nil {
		t.Fatalf("setConfigValue returned error: %v", err)
	}
	if cfg.EffectiveOutputFormat() != config.OutputFormatJSON {
		t.Fatalf("expected json output format, got %q", cfg.EffectiveOutputFormat())
	}

	if err := setConfigValue(&cfg, "browser", "firefox --new-window"); err != nil {
		t.Fatalf("setConfigValue returned error: %v", err)
	}
	if cfg.Settings.Browser != "firefox --new-window" {
		t.Fatalf("expected configured browser, got %q", cfg.Settings.Browser)
	}

	if err := unsetConfigValue(&cfg, "prompt"); err != nil {
		t.Fatalf("unsetConfigValue returned error: %v", err)
	}
	if !cfg.PromptEnabled() {
		t.Fatal("expected prompt default after unset")
	}

	if err := unsetConfigValue(&cfg, "browser"); err != nil {
		t.Fatalf("unsetConfigValue returned error: %v", err)
	}
	if cfg.Settings.Browser != "" {
		t.Fatalf("expected browser to be unset, got %q", cfg.Settings.Browser)
	}
}

func TestUnsupportedConfigKeyError(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	if err := setConfigValue(&cfg, "pager", "less"); err == nil || !strings.Contains(err.Error(), "planned but not supported yet") {
		t.Fatalf("expected planned-but-unsupported error, got %v", err)
	}
	if err := unsetConfigValue(&cfg, "editor"); err == nil || !strings.Contains(err.Error(), "planned but not supported yet") {
		t.Fatalf("expected planned-but-unsupported error, got %v", err)
	}
}

func TestSetConfigValueRejectsInvalidBrowserCommand(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	err := setConfigValue(&cfg, "browser", `"unterminated`)
	if err == nil || !strings.Contains(err.Error(), "browser command is invalid") {
		t.Fatalf("expected invalid browser command error, got %v", err)
	}
}
