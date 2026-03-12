package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/spf13/cobra"
)

type secretPromptReader struct {
	*bytes.Buffer
}

func (secretPromptReader) Fd() uintptr {
	return 0
}

func TestPromptRequiredStringUsesDefaultOnEmptyInput(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	value, err := promptRequiredString(cmd, "Title", "feature-branch")
	if err != nil {
		t.Fatalf("promptRequiredString returned error: %v", err)
	}
	if value != "feature-branch" {
		t.Fatalf("expected default value, got %q", value)
	}
}

func TestPromptRequiredStringAcceptsTypedValue(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("Add feature\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	value, err := promptRequiredString(cmd, "Title", "feature-branch")
	if err != nil {
		t.Fatalf("promptRequiredString returned error: %v", err)
	}
	if value != "Add feature" {
		t.Fatalf("expected typed value, got %q", value)
	}
}

func TestPromptsDisabledWithInheritedFlag(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "bb"}
	root.PersistentFlags().Bool("no-prompt", false, "")

	child := &cobra.Command{Use: "pr"}
	root.AddCommand(child)

	flag := child.Flag("no-prompt")
	if flag == nil {
		t.Fatal("expected inherited no-prompt flag")
	}
	if err := flag.Value.Set("true"); err != nil {
		t.Fatalf("set no-prompt flag: %v", err)
	}

	if !promptsDisabled(child) {
		t.Fatal("expected promptsDisabled to respect inherited no-prompt flag")
	}
}

func TestConfirmExactMatch(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("acme/widgets\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := confirmExactMatch(cmd, "acme/widgets"); err != nil {
		t.Fatalf("confirmExactMatch returned error: %v", err)
	}
}

func TestConfirmExactMatchRejectsMismatch(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("wrong/value\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := confirmExactMatch(cmd, "acme/widgets")
	if err == nil || err.Error() != "confirmation did not match acme/widgets" {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestPromptsDisabledFromConfig(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("BB_CONFIG_DIR", configDir)

	disabled := false
	if err := config.Save(config.Config{
		Settings: config.Settings{
			Prompt: &disabled,
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("no-prompt", false, "")
	cmd.SetIn(bytes.NewBufferString(""))
	cmd.SetOut(bytes.NewBuffer(nil))

	if !promptsDisabled(cmd) {
		t.Fatal("expected promptsDisabled to honor config default")
	}

	if _, err := os.Stat(filepath.Join(configDir, "config.json")); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestPromptSecretStringTrimsInput(t *testing.T) {
	lockCommandTestHooks(t)

	previousReadSecretInput := readSecretInput
	t.Cleanup(func() { readSecretInput = previousReadSecretInput })

	readSecretInput = func(fd int) ([]byte, error) {
		return []byte("  secret-token  "), nil
	}

	cmd := &cobra.Command{}
	cmd.SetIn(secretPromptReader{Buffer: bytes.NewBuffer(nil)})
	var out bytes.Buffer
	cmd.SetOut(&out)

	value, err := promptSecretString(cmd, "Atlassian API token")
	if err != nil {
		t.Fatalf("promptSecretString returned error: %v", err)
	}
	if value != "secret-token" {
		t.Fatalf("expected trimmed secret value, got %q", value)
	}
	if got := out.String(); got != "Atlassian API token: \n" {
		t.Fatalf("unexpected prompt output %q", got)
	}
}

func TestPromptSecretStringRetriesOnEmptyInput(t *testing.T) {
	lockCommandTestHooks(t)

	previousReadSecretInput := readSecretInput
	t.Cleanup(func() { readSecretInput = previousReadSecretInput })

	attempts := 0
	readSecretInput = func(fd int) ([]byte, error) {
		attempts++
		if attempts == 1 {
			return []byte("   "), nil
		}
		return []byte("retry-token"), nil
	}

	cmd := &cobra.Command{}
	cmd.SetIn(secretPromptReader{Buffer: bytes.NewBuffer(nil)})
	var out bytes.Buffer
	cmd.SetOut(&out)

	value, err := promptSecretString(cmd, "Atlassian API token")
	if err != nil {
		t.Fatalf("promptSecretString returned error: %v", err)
	}
	if value != "retry-token" {
		t.Fatalf("expected retry token, got %q", value)
	}
	if attempts != 2 {
		t.Fatalf("expected two prompt attempts, got %d", attempts)
	}
	if got := out.String(); got != "Atlassian API token: \nA value is required.\nAtlassian API token: \n" {
		t.Fatalf("unexpected prompt output %q", got)
	}
}
