package cmd

import (
	"strings"
	"testing"
)

func assertOrderedSubstrings(t *testing.T, got string, expected ...string) {
	t.Helper()

	searchFrom := 0
	for _, item := range expected {
		idx := strings.Index(got[searchFrom:], item)
		if idx < 0 {
			t.Fatalf("expected %q in output, got %q", item, got)
		}
		searchFrom += idx + len(item)
	}
}
