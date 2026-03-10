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
