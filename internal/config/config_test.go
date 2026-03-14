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

func TestConfigRemoveHostUpdatesDefault(t *testing.T) {
	t.Parallel()

	cfg := Config{
		DefaultHost: "bitbucket.org",
		Hosts: map[string]HostConfig{
			"bitbucket.org": {},
			"example.org":   {},
		},
	}

	cfg.RemoveHost("bitbucket.org")
	if _, ok := cfg.Hosts["bitbucket.org"]; ok {
		t.Fatal("expected host to be removed")
	}
	if cfg.DefaultHost != "example.org" {
		t.Fatalf("expected default host to move to example.org, got %q", cfg.DefaultHost)
	}

	cfg.RemoveHost("example.org")
	if cfg.DefaultHost != "" {
		t.Fatalf("expected default host to clear, got %q", cfg.DefaultHost)
	}
}

func TestConfigResolveHost(t *testing.T) {
	t.Parallel()

	cfg := Config{
		DefaultHost: "bitbucket.org",
		Hosts: map[string]HostConfig{
			"bitbucket.org": {},
			"example.org":   {},
		},
	}

	if got, err := cfg.ResolveHost("override.org"); err != nil || got != "override.org" {
		t.Fatalf("expected explicit host, got %q %v", got, err)
	}
	if got, err := cfg.ResolveHost(""); err != nil || got != "bitbucket.org" {
		t.Fatalf("expected default host, got %q %v", got, err)
	}

	cfg.DefaultHost = ""
	if _, err := cfg.ResolveHost(""); err == nil || err.Error() != "multiple authenticated hosts configured; specify --host" {
		t.Fatalf("expected multiple-host error, got %v", err)
	}

	cfg.Hosts = map[string]HostConfig{"only.org": {}}
	if got, err := cfg.ResolveHost(""); err != nil || got != "only.org" {
		t.Fatalf("expected single host fallback, got %q %v", got, err)
	}

	cfg.Hosts = map[string]HostConfig{}
	if _, err := cfg.ResolveHost(""); err == nil || err.Error() != "no authenticated hosts configured" {
		t.Fatalf("expected no-host error, got %v", err)
	}
}

func TestConfigHostNamesSorted(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Hosts: map[string]HostConfig{
			"z.example": {},
			"a.example": {},
		},
	}

	got := cfg.HostNames()
	want := []string{"a.example", "z.example"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
