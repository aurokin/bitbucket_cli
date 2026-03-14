package cmd

import (
	"strings"
	"testing"
)

func FuzzParseBitbucketEntityURL(f *testing.F) {
	seeds := []string{
		"https://bitbucket.org/acme/widgets",
		"https://bitbucket.org/acme/widgets/pull-requests/42#comment-15",
		"https://bitbucket.org/acme/widgets/issues/5",
		"https://bitbucket.org/acme/widgets/commits/abcdef1",
		"https://bitbucket.org/acme/widgets/src/main/README.md#lines-12:14",
		"https://bitbucket.org/acme/widgets/src/release%2F1.0/docs/guide.md?at=ignored#lines-8",
		"not-a-url",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		entity, err := parseBitbucketEntityURL(raw)
		if err != nil {
			return
		}

		if entity.Host == "" || entity.Workspace == "" || entity.Repo == "" || entity.Type == "" {
			t.Fatalf("parsed entity missing core fields: %+v", entity)
		}
		if entity.URL != strings.TrimSpace(raw) {
			t.Fatalf("expected normalized URL %q, got %q", strings.TrimSpace(raw), entity.URL)
		}
	})
}
