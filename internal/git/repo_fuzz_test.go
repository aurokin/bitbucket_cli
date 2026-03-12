package git

import "testing"

func FuzzParseRemoteURL(f *testing.F) {
	seeds := []string{
		"https://bitbucket.org/acme/widgets.git",
		"git@bitbucket.org:acme/widgets.git",
		"ssh://git@bitbucket.org/acme/widgets.git",
		"https://bitbucket.org/acme/widgets",
		"",
		"not-a-url",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		parsed, err := ParseRemoteURL(raw)
		if err != nil {
			return
		}
		if parsed.Host == "" || parsed.Workspace == "" || parsed.RepoSlug == "" {
			t.Fatalf("expected full parsed remote, got %+v", parsed)
		}
	})
}
