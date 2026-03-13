package cmd

import "testing"

func TestParseBitbucketEntityURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		want    resolvedEntity
		wantErr string
	}{
		{
			name: "repository",
			raw:  "https://bitbucket.org/acme/widgets",
			want: resolvedEntity{
				Host:         "bitbucket.org",
				Workspace:    "acme",
				Repo:         "widgets",
				Type:         "repository",
				URL:          "https://bitbucket.org/acme/widgets",
				CanonicalURL: "https://bitbucket.org/acme/widgets",
			},
		},
		{
			name: "pull request",
			raw:  "https://bitbucket.org/acme/widgets/pull-requests/42",
			want: resolvedEntity{
				Host:         "bitbucket.org",
				Workspace:    "acme",
				Repo:         "widgets",
				Type:         "pull-request",
				URL:          "https://bitbucket.org/acme/widgets/pull-requests/42",
				CanonicalURL: "https://bitbucket.org/acme/widgets/pull-requests/42",
				PR:           42,
			},
		},
		{
			name: "pull request comment",
			raw:  "https://bitbucket.org/acme/widgets/pull-requests/42#comment-15",
			want: resolvedEntity{
				Host:         "bitbucket.org",
				Workspace:    "acme",
				Repo:         "widgets",
				Type:         "pull-request-comment",
				URL:          "https://bitbucket.org/acme/widgets/pull-requests/42#comment-15",
				CanonicalURL: "https://bitbucket.org/acme/widgets/pull-requests/42#comment-15",
				PR:           42,
				Comment:      15,
			},
		},
		{
			name: "issue",
			raw:  "https://bitbucket.org/acme/widgets/issues/5",
			want: resolvedEntity{
				Host:         "bitbucket.org",
				Workspace:    "acme",
				Repo:         "widgets",
				Type:         "issue",
				URL:          "https://bitbucket.org/acme/widgets/issues/5",
				CanonicalURL: "https://bitbucket.org/acme/widgets/issues/5",
				Issue:        5,
			},
		},
		{
			name: "commit",
			raw:  "https://bitbucket.org/acme/widgets/commits/abcdef1",
			want: resolvedEntity{
				Host:         "bitbucket.org",
				Workspace:    "acme",
				Repo:         "widgets",
				Type:         "commit",
				URL:          "https://bitbucket.org/acme/widgets/commits/abcdef1",
				CanonicalURL: "https://bitbucket.org/acme/widgets/commits/abcdef1",
				Commit:       "abcdef1",
			},
		},
		{
			name: "source path with line",
			raw:  "https://bitbucket.org/acme/widgets/src/main/README.md#lines-12:14",
			want: resolvedEntity{
				Host:         "bitbucket.org",
				Workspace:    "acme",
				Repo:         "widgets",
				Type:         "path",
				URL:          "https://bitbucket.org/acme/widgets/src/main/README.md#lines-12:14",
				CanonicalURL: "https://bitbucket.org/acme/widgets/src/main/README.md#lines-12",
				Ref:          "main",
				Path:         "README.md",
				Line:         12,
			},
		},
		{
			name:    "unsupported repo settings url",
			raw:     "https://bitbucket.org/acme/widgets/admin",
			wantErr: `Bitbucket URL "https://bitbucket.org/acme/widgets/admin" is not a supported repository, pull request, comment, issue, commit, or source URL`,
		},
		{
			name:    "invalid raw",
			raw:     "not-a-url",
			wantErr: `Bitbucket URL "not-a-url" is invalid`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseBitbucketEntityURL(tc.raw)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("expected error %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("did not expect error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestNextResolveCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		entity resolvedEntity
		want   string
	}{
		{
			name: "pull request comment",
			entity: resolvedEntity{
				Type:      "pull-request-comment",
				Workspace: "acme",
				Repo:      "widgets",
				PR:        7,
				Comment:   15,
			},
			want: "bb pr comment view 15 --pr 7 --repo acme/widgets",
		},
		{
			name: "source path with line",
			entity: resolvedEntity{
				Type:      "path",
				Workspace: "acme",
				Repo:      "widgets",
				Path:      "README.md",
				Line:      12,
			},
			want: "bb browse README.md:12 --repo acme/widgets --no-browser",
		},
		{
			name: "repository",
			entity: resolvedEntity{
				Type:      "repository",
				Workspace: "acme",
				Repo:      "widgets",
			},
			want: "bb repo view --repo acme/widgets",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := nextResolveCommand(tc.entity); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
