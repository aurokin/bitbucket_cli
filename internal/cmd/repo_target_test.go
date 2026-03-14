package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestResolveRepoCloneInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		args      []string
		repoFlag  string
		wantRepo  string
		wantDir   string
		wantError string
	}{
		{
			name:     "positional repo only",
			args:     []string{"acme/widgets"},
			wantRepo: "acme/widgets",
		},
		{
			name:     "positional repo and directory",
			args:     []string{"acme/widgets", "./tmp/widgets"},
			wantRepo: "acme/widgets",
			wantDir:  "./tmp/widgets",
		},
		{
			name:     "repo flag only",
			repoFlag: "acme/widgets",
		},
		{
			name:     "repo flag with directory",
			args:     []string{"./tmp/widgets"},
			repoFlag: "acme/widgets",
			wantDir:  "./tmp/widgets",
		},
		{
			name:      "missing repository",
			wantError: "repository is required; pass <repo>, <workspace>/<repo>, or --repo",
		},
		{
			name:      "too many args with repo flag",
			args:      []string{"./tmp/widgets", "./extra"},
			repoFlag:  "acme/widgets",
			wantError: "when --repo is provided, pass at most one clone directory argument",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotRepo, gotDir, err := resolveRepoCloneInput(tc.args, tc.repoFlag)
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError {
					t.Fatalf("expected error %q, got %v", tc.wantError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("did not expect error: %v", err)
			}
			if gotRepo != tc.wantRepo || gotDir != tc.wantDir {
				t.Fatalf("expected repo=%q dir=%q, got repo=%q dir=%q", tc.wantRepo, tc.wantDir, gotRepo, gotDir)
			}
		})
	}
}

func TestRepoViewNextStep(t *testing.T) {
	t.Parallel()

	if got := repoViewNextStep(repoViewPayload{Workspace: "acme", RepoSlug: "widgets"}); got != "bb repo clone acme/widgets" {
		t.Fatalf("unexpected non-local repo next step %q", got)
	}
	if got := repoViewNextStep(repoViewPayload{Workspace: "acme", RepoSlug: "widgets", RootDir: "/tmp/widgets"}); got != "bb pr list --repo acme/widgets" {
		t.Fatalf("unexpected local repo next step %q", got)
	}
}

func TestRepoDeletionStatus(t *testing.T) {
	t.Parallel()

	if got := repoDeletionStatus(true); got != "deleted" {
		t.Fatalf("unexpected deleted status %q", got)
	}
	if got := repoDeletionStatus(false); got != "present" {
		t.Fatalf("unexpected present status %q", got)
	}
}

func TestWriteRepoCloneSummaryIncludesNextStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoClonePayload{
		Workspace: "acme",
		RepoSlug:  "widgets",
		Directory: "/tmp/widgets",
		CloneURL:  "https://bitbucket.org/acme/widgets.git",
	}

	if err := writeRepoCloneSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoCloneSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Directory: /tmp/widgets",
		"Clone URL: https://bitbucket.org/acme/widgets.git",
		"Next: bb repo view --repo acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestWriteRepoDeleteSummaryIncludesNextStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoDeletePayload{
		Workspace: "acme",
		RepoSlug:  "widgets",
		Name:      "Widget Service",
		Deleted:   true,
	}

	if err := writeRepoDeleteSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoDeleteSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Name: Widget Service",
		"Status: deleted",
		"Next: bb repo create acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
}

func TestRepoUtilityHelpers(t *testing.T) {
	t.Parallel()

	targets := []bitbucket.NamedCloneTarget{
		{Name: "ssh", Href: "git@bitbucket.org:acme/widgets.git"},
		{Name: "https", Href: "https://bitbucket.org/acme/widgets.git"},
	}
	if got := cloneURLForName(targets, "https"); got != "https://bitbucket.org/acme/widgets.git" {
		t.Fatalf("unexpected https clone URL %q", got)
	}
	if got := cloneURLForName(targets, "missing"); got != "" {
		t.Fatalf("expected missing clone URL to be empty, got %q", got)
	}
	if got := firstArg([]string{"one", "two"}); got != "one" {
		t.Fatalf("unexpected first arg %q", got)
	}
	if got := firstArg(nil); got != "" {
		t.Fatalf("expected empty first arg, got %q", got)
	}
	if got := previousRepoSlug("widgets", "widgets"); got != "" {
		t.Fatalf("expected empty previous repo slug, got %q", got)
	}
	if got := previousRepoSlug("widgets-old", "widgets"); got != "widgets-old" {
		t.Fatalf("unexpected previous repo slug %q", got)
	}
}

func TestRepoVisibilityLabel(t *testing.T) {
	t.Parallel()

	if got := repoVisibilityLabel(true); got != "private" {
		t.Fatalf("unexpected private visibility label %q", got)
	}
	if got := repoVisibilityLabel(false); got != "public" {
		t.Fatalf("unexpected public visibility label %q", got)
	}
}

func TestWriteRepoViewSummaryIncludesWarningsAndNextStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := repoViewPayload{
		Host:        "bitbucket.org",
		Workspace:   "acme",
		RepoSlug:    "widgets",
		Warnings:    []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
		Name:        "Widgets",
		Private:     true,
		ProjectKey:  "BBCLI",
		MainBranch:  "main",
		HTMLURL:     "https://bitbucket.org/acme/widgets",
		HTTPSClone:  "https://bitbucket.org/acme/widgets.git",
		Description: "Fixture repository",
	}

	if err := writeRepoViewSummary(&buf, payload); err != nil {
		t.Fatalf("writeRepoViewSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Visibility: private",
		"URL: https://bitbucket.org/acme/widgets",
		"Next: bb repo clone acme/widgets",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Visibility: private",
		"URL: https://bitbucket.org/acme/widgets",
		"Next: bb repo clone acme/widgets",
	)
}
