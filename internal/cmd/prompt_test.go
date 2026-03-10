package cmd

import (
	"bytes"
	"testing"

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
