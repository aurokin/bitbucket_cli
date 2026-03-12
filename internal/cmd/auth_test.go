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
