package cmd

import "testing"

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
			args:     []string{"OhBizzle/widgets"},
			wantRepo: "OhBizzle/widgets",
		},
		{
			name:     "positional repo and directory",
			args:     []string{"OhBizzle/widgets", "./tmp/widgets"},
			wantRepo: "OhBizzle/widgets",
			wantDir:  "./tmp/widgets",
		},
		{
			name:     "repo flag only",
			repoFlag: "OhBizzle/widgets",
		},
		{
			name:     "repo flag with directory",
			args:     []string{"./tmp/widgets"},
			repoFlag: "OhBizzle/widgets",
			wantDir:  "./tmp/widgets",
		},
		{
			name:      "missing repository",
			wantError: "repository is required; pass <repo>, <workspace>/<repo>, or --repo",
		},
		{
			name:      "too many args with repo flag",
			args:      []string{"./tmp/widgets", "./extra"},
			repoFlag:  "OhBizzle/widgets",
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
