package cmd

import (
	"bytes"
	"strings"
	"testing"

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
				DisplayName:   "Hunter Sadler",
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
		"Hunter Sadler",
		"example.com auth error: 401 Unauthorized",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}
