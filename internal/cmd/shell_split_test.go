package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitCommandLine(t *testing.T) {
	t.Parallel()

	got, err := splitCommandLine(`pr create --title "Add feature" --body 'Needs review' --repo acme/widgets\ beta`)
	if err != nil {
		t.Fatalf("splitCommandLine returned error: %v", err)
	}

	want := []string{"pr", "create", "--title", "Add feature", "--body", "Needs review", "--repo", "acme/widgets beta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestSplitCommandLineRejectsUnterminatedQuote(t *testing.T) {
	t.Parallel()

	_, err := splitCommandLine(`pr view "unterminated`)
	if err == nil {
		t.Fatal("expected splitCommandLine to fail")
	}
	if !strings.Contains(err.Error(), "unterminated double quote") {
		t.Fatalf("expected unterminated quote error, got %v", err)
	}
}

func TestSplitCommandLineRejectsDanglingEscape(t *testing.T) {
	t.Parallel()

	_, err := splitCommandLine(`pr view widgets\`)
	if err == nil || !strings.Contains(err.Error(), "dangling escape") {
		t.Fatalf("expected dangling escape error, got %v", err)
	}
}

func TestSplitCommandLinePreservesEscapedCharactersInDoubleQuotes(t *testing.T) {
	t.Parallel()

	got, err := splitCommandLine(`pr create --title "Quote: \"x\" and slash \\ path"`)
	if err != nil {
		t.Fatalf("splitCommandLine returned error: %v", err)
	}

	want := []string{"pr", "create", "--title", `Quote: "x" and slash \ path`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
