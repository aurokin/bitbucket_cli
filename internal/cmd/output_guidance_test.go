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

func TestWriteLabelValue(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeLabelValue(&buf, "Query", "fix auth"); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if got := buf.String(); got != "Query: fix auth\n" {
		t.Fatalf("unexpected label/value output %q", got)
	}
}

func TestWriteWarnings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeWarnings(&buf, []string{"first warning", "", "second warning"}); err != nil {
		t.Fatalf("writeWarnings returned error: %v", err)
	}
	if got := buf.String(); got != "Warning: first warning\nWarning: second warning\n" {
		t.Fatalf("unexpected warning output %q", got)
	}
}
