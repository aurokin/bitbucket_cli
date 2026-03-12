package cmd

import "testing"

func FuzzParseRepoSelector(f *testing.F) {
	seeds := [][3]string{
		{"", "", ""},
		{"", "", "widgets"},
		{"", "OhBizzle", "widgets"},
		{"bitbucket.org", "", "OhBizzle/widgets"},
		{"", "", "https://bitbucket.org/OhBizzle/widgets"},
		{"", "", "ssh://git@bitbucket.org/OhBizzle/widgets.git"},
		{"bitbucket.org", "OhBizzle", "https://bitbucket.org/OhBizzle/widgets"},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2])
	}

	f.Fuzz(func(t *testing.T, hostFlag, workspaceFlag, repoFlag string) {
		selector, err := parseRepoSelector(hostFlag, workspaceFlag, repoFlag)
		if err != nil {
			return
		}
		if repoFlag == "" && selector.Explicit {
			t.Fatalf("selector should not be explicit without a repo flag: %+v", selector)
		}
		if selector.Repo == "" && selector.Workspace != "" && repoFlag == "" {
			t.Fatalf("selector should not resolve a workspace without a repo flag: %+v", selector)
		}
	})
}

func FuzzParseRepoTargetInput(f *testing.F) {
	seeds := [][4]string{
		{"", "", "", ""},
		{"", "OhBizzle", "", "widgets"},
		{"bitbucket.org", "", "OhBizzle/widgets", ""},
		{"", "", "", "https://bitbucket.org/OhBizzle/widgets"},
		{"bitbucket.org", "OhBizzle", "widgets", "OhBizzle/widgets"},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}

	f.Fuzz(func(t *testing.T, hostFlag, workspaceFlag, repoFlag, positional string) {
		selector, err := parseRepoTargetInput(hostFlag, workspaceFlag, repoFlag, positional)
		if err != nil {
			return
		}
		if selector.Repo == "" && selector.Explicit {
			t.Fatalf("selector cannot be explicit without a repo: %+v", selector)
		}
	})
}

func FuzzParsePullRequestSelector(f *testing.F) {
	seeds := []string{
		"1",
		"42",
		"https://bitbucket.org/OhBizzle/widgets/pull-requests/7",
		"https://bitbucket.org/OhBizzle/widgets/src/main/README.md",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		selector, err := parsePullRequestSelector(raw)
		if err != nil {
			return
		}
		if selector.ID <= 0 {
			t.Fatalf("expected positive pull request ID, got %+v", selector)
		}
	})
}
