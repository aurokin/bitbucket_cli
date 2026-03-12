package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddFormatFlagsConfiguresBareJSONSupport(t *testing.T) {
	t.Parallel()

	var flags formatFlags
	cmd := &cobra.Command{Use: "test"}

	addFormatFlags(cmd, &flags)

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("expected json flag")
	}
	if jsonFlag.NoOptDefVal != "*" {
		t.Fatalf("expected bare json flag default '*', got %q", jsonFlag.NoOptDefVal)
	}

	jqFlag := cmd.Flags().Lookup("jq")
	if jqFlag == nil {
		t.Fatal("expected jq flag")
	}
}
