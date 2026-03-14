package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteAPIResponse(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeAPIResponse(&buf, []byte(`{"name":"repo"}`), ""); err != nil {
		t.Fatalf("writeAPIResponse returned error: %v", err)
	}
	if got := buf.String(); got != "{\"name\":\"repo\"}\n" {
		t.Fatalf("unexpected raw API response %q", got)
	}
}

func TestWriteAPIResponseWithJQ(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeAPIResponse(&buf, []byte(`{"name":"repo","state":"OPEN"}`), ".name"); err != nil {
		t.Fatalf("writeAPIResponse returned error: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != `"repo"` {
		t.Fatalf("unexpected jq API response %q", got)
	}
}

func TestWriteAPIResponseRejectsNonJSONForJQ(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := writeAPIResponse(&buf, []byte(`not-json`), ".name"); err == nil || !strings.Contains(err.Error(), "cannot apply --jq to non-JSON response") {
		t.Fatalf("expected non-JSON jq error, got %v", err)
	}
}
