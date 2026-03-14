package output

import (
	"bytes"
	"testing"
)

func TestNewTableWriterAndTerminalWidthFallback(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tw := NewTableWriter(&buf)
	if _, err := tw.Write([]byte("col1\tcol2\n")); err != nil {
		t.Fatalf("table writer write failed: %v", err)
	}
	if err := tw.Flush(); err != nil {
		t.Fatalf("table writer flush failed: %v", err)
	}
	if got := TerminalWidth(&buf); got != 0 {
		t.Fatalf("expected non-terminal width 0, got %d", got)
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	if got := Truncate("feature/add-a-very-long-branch-name", 12); got != "feature/add…" {
		t.Fatalf("unexpected truncated value %q", got)
	}

	if got := Truncate("short", 12); got != "short" {
		t.Fatalf("unexpected short value %q", got)
	}
}

func TestTruncateMiddle(t *testing.T) {
	t.Parallel()

	if got := TruncateMiddle("src/components/very/long/path/file.go", 16); got != "src/com…/file.go" {
		t.Fatalf("unexpected middle truncation %q", got)
	}

	if got := TruncateMiddle("file.go", 16); got != "file.go" {
		t.Fatalf("unexpected short value %q", got)
	}
}
