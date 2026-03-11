package cmd

import (
	"bytes"
	"testing"
)

func TestWriteTargetHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeTargetHeader(&buf, "Repository", "acme", "widgets"); err != nil {
		t.Fatalf("writeTargetHeader returned error: %v", err)
	}
	if got := buf.String(); got != "Repository: acme/widgets\n" {
		t.Fatalf("unexpected header %q", got)
	}
}

func TestWriteNextStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeNextStep(&buf, "bb pr view 7 --repo acme/widgets"); err != nil {
		t.Fatalf("writeNextStep returned error: %v", err)
	}
	if got := buf.String(); got != "Next: bb pr view 7 --repo acme/widgets\n" {
		t.Fatalf("unexpected next step %q", got)
	}
}
