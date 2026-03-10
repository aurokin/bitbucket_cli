package cmd

import "testing"

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
