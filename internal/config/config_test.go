package config

import "testing"

func TestNormalizeSettings(t *testing.T) {
	t.Parallel()

	settings := NormalizeSettings(Settings{
		Browser:      "  firefox  ",
		Editor:       "  vim  ",
		Pager:        "  less -FRX  ",
		OutputFormat: "json",
	})

	if settings.Browser != "firefox" || settings.Editor != "vim" || settings.Pager != "less -FRX" {
		t.Fatalf("unexpected trimmed settings %+v", settings)
	}
	if settings.OutputFormat != OutputFormatJSON {
		t.Fatalf("expected json output format, got %q", settings.OutputFormat)
	}
}

func TestConfigPromptEnabledDefaultsTrue(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	if !cfg.PromptEnabled() {
		t.Fatal("expected prompts to be enabled by default")
	}

	disabled := false
	cfg.Settings.Prompt = &disabled
	if cfg.PromptEnabled() {
		t.Fatal("expected prompts to be disabled")
	}
}

func TestConfigEffectiveOutputFormatDefaultsTable(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	if got := cfg.EffectiveOutputFormat(); got != OutputFormatTable {
		t.Fatalf("expected table output, got %q", got)
	}

	cfg.Settings.OutputFormat = OutputFormatJSON
	if got := cfg.EffectiveOutputFormat(); got != OutputFormatJSON {
		t.Fatalf("expected json output, got %q", got)
	}
}
