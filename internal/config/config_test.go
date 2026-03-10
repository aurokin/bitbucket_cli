package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingConfig(t *testing.T) {
	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cfg.Hosts) != 0 {
		t.Fatalf("expected empty host set, got %+v", cfg.Hosts)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BB_CONFIG_DIR", dir)

	cfg := Config{}
	cfg.SetHost("bitbucket.org", HostConfig{
		Username:  "auro",
		Token:     "secret",
		TokenType: "bearer",
	}, true)

	if err := Save(cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.DefaultHost != "bitbucket.org" {
		t.Fatalf("unexpected default host %q", loaded.DefaultHost)
	}

	host, ok := loaded.Hosts["bitbucket.org"]
	if !ok {
		t.Fatalf("expected saved host in config")
	}
	if host.Username != "auro" || host.Token != "secret" || host.TokenType != "bearer" {
		t.Fatalf("unexpected host config %+v", host)
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected config mode %o", info.Mode().Perm())
	}
}

func TestResolveHost(t *testing.T) {
	cfg := Config{
		DefaultHost: "bitbucket.org",
		Hosts: map[string]HostConfig{
			"bitbucket.org": {},
			"example.com":   {},
		},
	}

	host, err := cfg.ResolveHost("")
	if err != nil {
		t.Fatalf("ResolveHost returned error: %v", err)
	}
	if host != "bitbucket.org" {
		t.Fatalf("unexpected resolved host %q", host)
	}
}

func TestPathUsesOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BB_CONFIG_DIR", dir)

	path, err := Path()
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}

	expected := filepath.Join(dir, defaultConfigFile)
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}
