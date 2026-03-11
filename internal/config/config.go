package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultConfigDirName = "bb"
	defaultConfigFile    = "config.json"
)

type Config struct {
	DefaultHost string                `json:"default_host,omitempty"`
	Hosts       map[string]HostConfig `json:"hosts,omitempty"`
	Settings    Settings              `json:"settings,omitempty"`
	Aliases     map[string]string     `json:"aliases,omitempty"`
}

type Settings struct {
	Prompt       *bool  `json:"prompt,omitempty"`
	Browser      string `json:"browser,omitempty"`
	Editor       string `json:"editor,omitempty"`
	Pager        string `json:"pager,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`
}

const (
	OutputFormatTable = "table"
	OutputFormatJSON  = "json"
)

type HostConfig struct {
	Username  string    `json:"username,omitempty"`
	Token     string    `json:"token,omitempty"`
	AuthType  string    `json:"auth_type,omitempty"`
	TokenType string    `json:"token_type,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

const (
	AuthTypeAPIToken = "api-token"
)

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{Hosts: map[string]HostConfig{}}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	if cfg.Hosts == nil {
		cfg.Hosts = map[string]HostConfig{}
	}
	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}

	for host, hostConfig := range cfg.Hosts {
		cfg.Hosts[host] = NormalizeHostConfig(hostConfig)
	}
	cfg.Settings = NormalizeSettings(cfg.Settings)

	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	if cfg.Hosts == nil {
		cfg.Hosts = map[string]HostConfig{}
	}
	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}

	for host, hostConfig := range cfg.Hosts {
		cfg.Hosts[host] = NormalizeHostConfig(hostConfig)
	}
	cfg.Settings = NormalizeSettings(cfg.Settings)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config file: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func Path() (string, error) {
	if override := strings.TrimSpace(os.Getenv("BB_CONFIG_DIR")); override != "" {
		return filepath.Join(override, defaultConfigFile), nil
	}

	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(baseDir, defaultConfigDirName, defaultConfigFile), nil
}

func (c *Config) SetHost(host string, hostConfig HostConfig, setDefault bool) {
	if c.Hosts == nil {
		c.Hosts = map[string]HostConfig{}
	}

	c.Hosts[host] = NormalizeHostConfig(hostConfig)
	if setDefault || c.DefaultHost == "" {
		c.DefaultHost = host
	}
}

func NormalizeHostConfig(hostConfig HostConfig) HostConfig {
	authType := strings.TrimSpace(hostConfig.AuthType)
	if authType == "" {
		switch strings.TrimSpace(hostConfig.TokenType) {
		case "", "app-password", "basic", "bearer", "token", "oauth", "api-token":
			authType = AuthTypeAPIToken
		default:
			authType = AuthTypeAPIToken
		}
	}

	hostConfig.Username = strings.TrimSpace(hostConfig.Username)
	hostConfig.Token = strings.TrimSpace(hostConfig.Token)
	hostConfig.AuthType = authType
	hostConfig.TokenType = ""

	return hostConfig
}

func NormalizeSettings(settings Settings) Settings {
	settings.Browser = strings.TrimSpace(settings.Browser)
	settings.Editor = strings.TrimSpace(settings.Editor)
	settings.Pager = strings.TrimSpace(settings.Pager)

	switch strings.TrimSpace(settings.OutputFormat) {
	case "", OutputFormatTable:
		settings.OutputFormat = ""
	case OutputFormatJSON:
		settings.OutputFormat = OutputFormatJSON
	default:
		settings.OutputFormat = ""
	}

	return settings
}

func (c Config) PromptEnabled() bool {
	if c.Settings.Prompt == nil {
		return true
	}
	return *c.Settings.Prompt
}

func (c Config) EffectiveOutputFormat() string {
	if c.Settings.OutputFormat == OutputFormatJSON {
		return OutputFormatJSON
	}
	return OutputFormatTable
}

func (c *Config) RemoveHost(host string) {
	delete(c.Hosts, host)

	if c.DefaultHost != host {
		return
	}

	names := c.HostNames()
	if len(names) == 0 {
		c.DefaultHost = ""
		return
	}

	c.DefaultHost = names[0]
}

func (c Config) ResolveHost(explicitHost string) (string, error) {
	if explicitHost != "" {
		return explicitHost, nil
	}

	if c.DefaultHost != "" {
		return c.DefaultHost, nil
	}

	names := c.HostNames()
	if len(names) == 1 {
		return names[0], nil
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no authenticated hosts configured")
	}

	return "", fmt.Errorf("multiple authenticated hosts configured; specify --host")
}

func (c Config) HostNames() []string {
	names := make([]string, 0, len(c.Hosts))
	for host := range c.Hosts {
		names = append(names, host)
	}
	sort.Strings(names)
	return names
}
