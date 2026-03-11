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
