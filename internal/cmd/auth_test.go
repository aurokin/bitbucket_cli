package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/spf13/cobra"
)

func TestResolveUsernameValueUsesEnv(t *testing.T) {
	t.Setenv("BB_EMAIL", "agent@example.com")

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	value, err := resolveUsernameValue(cmd, "")
	if err != nil {
		t.Fatalf("resolveUsernameValue returned error: %v", err)
	}
	if value != "agent@example.com" {
		t.Fatalf("unexpected username %q", value)
	}
}

func TestResolveTokenValueUsesEnv(t *testing.T) {
	t.Setenv("BB_TOKEN", "secret-token")

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	value, err := resolveTokenValue(cmd, "", false)
	if err != nil {
		t.Fatalf("resolveTokenValue returned error: %v", err)
	}
	if value != "secret-token" {
		t.Fatalf("unexpected token %q", value)
	}
}

func TestResolveTokenValueUsesStdin(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("secret-token\n"))
	cmd.SetOut(bytes.NewBuffer(nil))

	value, err := resolveTokenValue(cmd, "", true)
	if err != nil {
		t.Fatalf("resolveTokenValue returned error: %v", err)
	}
	if value != "secret-token" {
		t.Fatalf("unexpected token %q", value)
	}
}

func TestResolveUsernameValueGuidesNonInteractiveUser(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	_, err := resolveUsernameValue(cmd, "")
	if err == nil || !strings.Contains(err.Error(), atlassianAPITokenManageURL) {
		t.Fatalf("expected guidance error, got %v", err)
	}
}

func TestResolveTokenValueGuidesNonInteractiveUser(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	_, err := resolveTokenValue(cmd, "", false)
	if err == nil || !strings.Contains(err.Error(), atlassianAPITokenManageURL) {
		t.Fatalf("expected guidance error, got %v", err)
	}
}

func TestWriteAuthStatusSummaryWithNoHosts(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeAuthStatusSummary(&buf, authStatusPayload{}); err != nil {
		t.Fatalf("writeAuthStatusSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"No authenticated hosts.",
		"Create Token: " + atlassianAPITokenManageURL,
		"Next: bb auth login",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"No authenticated hosts.",
		"Create Token: "+atlassianAPITokenManageURL,
		"Next: bb auth login",
	)
}

func TestWriteAuthStatusSummaryIncludesAuthErrors(t *testing.T) {
	t.Parallel()

	authenticated := true
	rejected := false

	var buf bytes.Buffer
	payload := authStatusPayload{
		DefaultHost: "bitbucket.org",
		Hosts: []authStatusHostRow{
			{
				Host:          "bitbucket.org",
				Default:       true,
				Username:      "user@example.com",
				AuthType:      "api-token",
				Authenticated: &authenticated,
				DisplayName:   "Example User",
				UpdatedAt:     "2026-03-11T00:00:00Z",
			},
			{
				Host:                "example.com",
				Username:            "user@example.com",
				AuthType:            "api-token",
				Authenticated:       &rejected,
				AuthenticationError: "401 Unauthorized",
				UpdatedAt:           "2026-03-11T00:00:00Z",
			},
		},
	}

	if err := writeAuthStatusSummary(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"bitbucket.org",
		"example.com",
		"Example User",
		"example.com auth error: 401 Unauthorized",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"host",
		"bitbucket.org",
		"example.com",
		"example.com auth error: 401 Unauthorized",
	)
}

func TestStatusHostNames(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Hosts: map[string]config.HostConfig{
			"bitbucket.org": {},
			"example.org":   {},
		},
	}

	names, err := statusHostNames(cfg, "")
	if err != nil {
		t.Fatalf("statusHostNames returned error: %v", err)
	}
	if strings.Join(names, ",") != "bitbucket.org,example.org" {
		t.Fatalf("unexpected host names %v", names)
	}

	names, err = statusHostNames(cfg, "example.org")
	if err != nil || len(names) != 1 || names[0] != "example.org" {
		t.Fatalf("unexpected selected host result %v %v", names, err)
	}

	if _, err := statusHostNames(cfg, "missing.org"); err == nil || !strings.Contains(err.Error(), "no stored credentials found") {
		t.Fatalf("expected missing host error, got %v", err)
	}
}

func TestBuildAuthStatusPayloadWithoutLiveCheck(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		DefaultHost: "bitbucket.org",
		Hosts: map[string]config.HostConfig{
			"bitbucket.org": {
				Username: "user@example.com",
				Token:    "token",
				AuthType: config.AuthTypeAPIToken,
			},
		},
	}

	payload, err := buildAuthStatusPayload(context.Background(), cfg, "", false)
	if err != nil {
		t.Fatalf("buildAuthStatusPayload returned error: %v", err)
	}
	if payload.DefaultHost != "bitbucket.org" || len(payload.Hosts) != 1 {
		t.Fatalf("unexpected payload %+v", payload)
	}
	row := payload.Hosts[0]
	if row.Host != "bitbucket.org" || !row.TokenConfigured || row.Username != "user@example.com" || row.AuthType != config.AuthTypeAPIToken || !row.Default {
		t.Fatalf("unexpected auth row %+v", row)
	}
}

func TestAuthLogoutCommandOutput(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("BB_CONFIG_DIR", configDir)

	cfg := config.Config{}
	cfg.SetHost("bitbucket.org", config.HostConfig{
		Username: "user@example.com",
		Token:    "token",
		AuthType: config.AuthTypeAPIToken,
	}, true)
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	output := renderCommand(t, "auth", "logout", "--host", "bitbucket.org")
	assertOrderedSubstrings(t, output,
		"Removed credentials for bitbucket.org",
		"Create Token: "+atlassianAPITokenManageURL,
		"Next: bb auth login --host bitbucket.org",
	)
}
