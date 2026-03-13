package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveHumanOutputForCommentURL(t *testing.T) {
	t.Parallel()

	entity, err := parseBitbucketEntityURL("https://bitbucket.org/acme/widgets/pull-requests/7#comment-15")
	if err != nil {
		t.Fatalf("parseBitbucketEntityURL returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeTargetHeader(&buf, "Repository", entity.Workspace, entity.Repo); err != nil {
		t.Fatalf("writeTargetHeader returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Type", entity.Type); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Pull Request", "7"); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Comment", "15"); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "URL", entity.URL); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Canonical URL", entity.CanonicalURL); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeNextStep(&buf, nextResolveCommand(entity)); err != nil {
		t.Fatalf("writeNextStep returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Type: pull-request-comment",
		"Pull Request: 7",
		"Comment: 15",
		"Canonical URL: https://bitbucket.org/acme/widgets/pull-requests/7#comment-15",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestResolveHumanOutputForSourceURL(t *testing.T) {
	t.Parallel()

	entity, err := parseBitbucketEntityURL("https://bitbucket.org/acme/widgets/src/main/README.md#lines-12")
	if err != nil {
		t.Fatalf("parseBitbucketEntityURL returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeTargetHeader(&buf, "Repository", entity.Workspace, entity.Repo); err != nil {
		t.Fatalf("writeTargetHeader returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Type", entity.Type); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Ref", entity.Ref); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Path", entity.Path); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeLabelValue(&buf, "Line", "12"); err != nil {
		t.Fatalf("writeLabelValue returned error: %v", err)
	}
	if err := writeNextStep(&buf, nextResolveCommand(entity)); err != nil {
		t.Fatalf("writeNextStep returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Type: path",
		"Ref: main",
		"Path: README.md",
		"Line: 12",
		"Next: bb browse README.md:12 --repo acme/widgets --no-browser",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}
