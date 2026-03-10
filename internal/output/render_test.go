package output

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type samplePayload struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

func TestParseFormatOptionsRequiresJSONForJQ(t *testing.T) {
	t.Parallel()

	_, err := ParseFormatOptions("", ".name")
	if err == nil || !strings.Contains(err.Error(), "--jq requires --json") {
		t.Fatalf("expected jq validation error, got %v", err)
	}
}

func TestRenderHumanOutput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Render(&buf, FormatOptions{}, samplePayload{Name: "repo"}, func(w io.Writer) error {
		_, err := io.WriteString(w, "human output\n")
		return err
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got := buf.String(); got != "human output\n" {
		t.Fatalf("unexpected human output %q", got)
	}
}

func TestRenderSelectedJSONFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Render(&buf, FormatOptions{JSONFields: []string{"name"}}, samplePayload{Name: "repo", State: "OPEN"}, func(w io.Writer) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `"name": "repo"`) {
		t.Fatalf("expected selected field in output: %s", got)
	}
	if strings.Contains(got, `"state"`) {
		t.Fatalf("did not expect omitted field in output: %s", got)
	}
}

func TestRenderJQProjection(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Render(&buf, FormatOptions{AllFields: true, JQ: ".name"}, samplePayload{Name: "repo", State: "OPEN"}, func(w io.Writer) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != `"repo"` {
		t.Fatalf("unexpected jq result %q", got)
	}
}

func TestRenderSelectedFieldsForArrays(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := []samplePayload{
		{Name: "one", State: "OPEN"},
		{Name: "two", State: "MERGED"},
	}

	err := Render(&buf, FormatOptions{JSONFields: []string{"state"}}, payload, func(w io.Writer) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `"state": "OPEN"`) || !strings.Contains(got, `"state": "MERGED"`) {
		t.Fatalf("expected states in output: %s", got)
	}
	if strings.Contains(got, `"name"`) {
		t.Fatalf("did not expect names in output: %s", got)
	}
}
