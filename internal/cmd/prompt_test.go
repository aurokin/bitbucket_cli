package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/spf13/cobra"
)

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
	cmd.SetIn(bytes.NewBufferString("OhBizzle/widgets\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := confirmExactMatch(cmd, "OhBizzle/widgets"); err != nil {
		t.Fatalf("confirmExactMatch returned error: %v", err)
	}
}

func TestConfirmExactMatchRejectsMismatch(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("wrong/value\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := confirmExactMatch(cmd, "OhBizzle/widgets")
	if err == nil || err.Error() != "confirmation did not match OhBizzle/widgets" {
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
