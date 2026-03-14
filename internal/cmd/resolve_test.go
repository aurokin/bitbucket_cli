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
	if err := writeResolveSummary(&buf, entity); err != nil {
		t.Fatalf("writeResolveSummary returned error: %v", err)
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
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Type: pull-request-comment",
		"Pull Request: 7",
		"Comment: 15",
		"URL: https://bitbucket.org/acme/widgets/pull-requests/7#comment-15",
		"Canonical URL: https://bitbucket.org/acme/widgets/pull-requests/7#comment-15",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	)
}

func TestResolveHumanOutputCanonicalizesMessyCommentURL(t *testing.T) {
	t.Parallel()

	entity, err := parseBitbucketEntityURL("https://bitbucket.org/acme/widgets/pull-requests/7/?foo=bar#comment-15")
	if err != nil {
		t.Fatalf("parseBitbucketEntityURL returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeResolveSummary(&buf, entity); err != nil {
		t.Fatalf("writeResolveSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Canonical URL: https://bitbucket.org/acme/widgets/pull-requests/7#comment-15",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Type: pull-request-comment",
		"Pull Request: 7",
		"Comment: 15",
		"URL: https://bitbucket.org/acme/widgets/pull-requests/7/?foo=bar#comment-15",
		"Canonical URL: https://bitbucket.org/acme/widgets/pull-requests/7#comment-15",
		"Next: bb pr comment view 15 --pr 7 --repo acme/widgets",
	)
}

func TestResolveHumanOutputForSourceURL(t *testing.T) {
	t.Parallel()

	entity, err := parseBitbucketEntityURL("https://bitbucket.org/acme/widgets/src/main/README.md#lines-12")
	if err != nil {
		t.Fatalf("parseBitbucketEntityURL returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeResolveSummary(&buf, entity); err != nil {
		t.Fatalf("writeResolveSummary returned error: %v", err)
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
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Type: path",
		"Ref: main",
		"Path: README.md",
		"Line: 12",
		"Canonical URL: https://bitbucket.org/acme/widgets/src/main/README.md#lines-12",
		"Next: bb browse README.md:12 --repo acme/widgets --no-browser",
	)
}

func TestResolveHumanOutputForIssueURL(t *testing.T) {
	t.Parallel()

	entity, err := parseBitbucketEntityURL("https://bitbucket.org/acme/widgets/issues/5/?tab=comments")
	if err != nil {
		t.Fatalf("parseBitbucketEntityURL returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeResolveSummary(&buf, entity); err != nil {
		t.Fatalf("writeResolveSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Type: issue",
		"Issue: 5",
		"URL: https://bitbucket.org/acme/widgets/issues/5/?tab=comments",
		"Canonical URL: https://bitbucket.org/acme/widgets/issues/5",
		"Next: bb issue view 5 --repo acme/widgets",
	)
}

func TestResolveHumanOutputForCommitURL(t *testing.T) {
	t.Parallel()

	entity, err := parseBitbucketEntityURL("https://bitbucket.org/acme/widgets/commits/deadbeef?foo=bar")
	if err != nil {
		t.Fatalf("parseBitbucketEntityURL returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeResolveSummary(&buf, entity); err != nil {
		t.Fatalf("writeResolveSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Type: commit",
		"Commit: deadbeef",
		"URL: https://bitbucket.org/acme/widgets/commits/deadbeef?foo=bar",
		"Canonical URL: https://bitbucket.org/acme/widgets/commits/deadbeef",
		"Next: bb browse deadbeef --repo acme/widgets --no-browser",
	)
}
