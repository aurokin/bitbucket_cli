package cmd

import (
	"context"
	"testing"
)

func TestValidateRepoSelector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		workspace string
		repo      string
		wantErr   bool
	}{
		{name: "none", wantErr: false},
		{name: "both", workspace: "OhBizzle", repo: "repo", wantErr: false},
		{name: "workspace only", workspace: "OhBizzle", wantErr: true},
		{name: "repo only", repo: "repo", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateRepoSelector(tc.workspace, tc.repo)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("did not expect error: %v", err)
			}
		})
	}
}

func TestResolveRepoCloneTarget(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		workspaceFlag string
		target        string
		wantWorkspace string
		wantRepo      string
		wantErr       bool
	}{
		{
			name:          "workspace slash repo",
			target:        "OhBizzle/widgets",
			wantWorkspace: "OhBizzle",
			wantRepo:      "widgets",
		},
		{
			name:          "repo with workspace flag",
			workspaceFlag: "OhBizzle",
			target:        "widgets",
			wantWorkspace: "OhBizzle",
			wantRepo:      "widgets",
		},
		{
			name:          "mismatched workspace flag",
			workspaceFlag: "Other",
			target:        "OhBizzle/widgets",
			wantErr:       true,
		},
		{
			name:    "too many slashes",
			target:  "one/two/three",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			workspace, repo, err := resolveRepoCloneTarget(context.Background(), nil, tc.workspaceFlag, tc.target)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("did not expect error: %v", err)
			}
			if workspace != tc.wantWorkspace || repo != tc.wantRepo {
				t.Fatalf("unexpected clone target: workspace=%q repo=%q", workspace, repo)
			}
		})
	}
}
