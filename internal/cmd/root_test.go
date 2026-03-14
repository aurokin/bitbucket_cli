package cmd

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeCLIArgsLeavesRegularArgsUntouched(t *testing.T) {
	got := normalizeCLIArgs([]string{"pr", "list", "--limit", "5"})
	want := []string{"pr", "list", "--limit", "5"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestNormalizeCLIArgsExpandsBareJSONFlag(t *testing.T) {
	got := normalizeCLIArgs([]string{"pr", "list", "--json"})
	want := []string{"pr", "list", "--json=*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestNormalizeCLIArgsConsumesJSONValue(t *testing.T) {
	got := normalizeCLIArgs([]string{"pr", "view", "1", "--json", "id,title"})
	want := []string{"pr", "view", "1", "--json=id,title"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestNormalizeCLIArgsExpandsAliasBeforeJSONNormalization(t *testing.T) {
	got, err := normalizeCLIArgsWithAliases([]string{"pv", "7", "--json"}, map[string]string{
		"pv": "pr view",
	})
	if err != nil {
		t.Fatalf("normalizeCLIArgsWithAliases returned error: %v", err)
	}
	want := []string{"pr", "view", "7", "--json=*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestExpandAliasArgsStopsOnAliasLoop(t *testing.T) {
	got, err := expandAliasArgs([]string{"a"}, map[string]string{
		"a": "b",
		"b": "a",
	})
	if err != nil {
		t.Fatalf("expandAliasArgs returned error: %v", err)
	}
	want := []string{"a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestExpandAliasArgsPreservesQuotedArguments(t *testing.T) {
	got, err := expandAliasArgs([]string{"ship"}, map[string]string{
		"ship": `pr create --title "Add feature" --body 'Needs review'`,
	})
	if err != nil {
		t.Fatalf("expandAliasArgs returned error: %v", err)
	}

	want := []string{"pr", "create", "--title", "Add feature", "--body", "Needs review"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestExpandAliasArgsPreservesEscapedWhitespace(t *testing.T) {
	got, err := expandAliasArgs([]string{"open"}, map[string]string{
		"open": `repo view --repo acme/widgets\ beta`,
	})
	if err != nil {
		t.Fatalf("expandAliasArgs returned error: %v", err)
	}

	want := []string{"repo", "view", "--repo", "acme/widgets beta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestExpandAliasArgsReturnsErrorForInvalidQuoting(t *testing.T) {
	_, err := expandAliasArgs([]string{"broken"}, map[string]string{
		"broken": `pr view "unterminated`,
	})
	if err == nil {
		t.Fatal("expected alias quoting error")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, `invalid alias "broken"`) {
		t.Fatalf("expected invalid alias guidance, got %q", got)
	}
}

func TestShouldRunExtension(t *testing.T) {
	t.Parallel()

	if !shouldRunExtension(userFacingError(errUnknownCommand("hello")), []string{"hello"}) {
		t.Fatal("expected unknown command to fall through to extension lookup")
	}
	if shouldRunExtension(nil, []string{"hello"}) {
		t.Fatal("did not expect nil error to trigger extension lookup")
	}
	if shouldRunExtension(userFacingError(errUnknownCommand("hello")), nil) {
		t.Fatal("did not expect empty args to trigger extension lookup")
	}
	if shouldRunExtension(assertErrString("other error"), []string{"hello"}) {
		t.Fatal("did not expect unrelated error to trigger extension lookup")
	}
}

func errUnknownCommand(name string) error {
	return assertErrString(`unknown command "` + name + `" for "bb"`)
}

type assertErrString string

func (e assertErrString) Error() string { return string(e) }

func TestExecuteVersionCommand(t *testing.T) {
	lockCommandTestHooks(t)

	t.Setenv("BB_CONFIG_DIR", t.TempDir())

	previousArgs := os.Args
	t.Cleanup(func() {
		os.Args = previousArgs
	})
	os.Args = []string{"bb", "version"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}
