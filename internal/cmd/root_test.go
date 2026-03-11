package cmd

import (
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
