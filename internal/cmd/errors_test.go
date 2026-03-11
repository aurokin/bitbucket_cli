package cmd

import (
	"strings"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
)

func TestUserFacingErrorUnauthorized(t *testing.T) {
	t.Parallel()

	err := userFacingError(bitbucket.NewAPIError(401, "401 Unauthorized", []byte(`{"type":"error","error":{"message":"Token is invalid or expired"}}`)))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid, expired, or revoked") {
		t.Fatalf("expected auth remediation hint, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "bb auth status --check") {
		t.Fatalf("expected auth verification hint, got %q", err.Error())
	}
}

func TestUserFacingErrorForbidden(t *testing.T) {
	t.Parallel()

	err := userFacingError(bitbucket.NewAPIError(403, "403 Forbidden", []byte(`{"type":"error","error":{"message":"Access denied. You may not have the required scopes."}}`)))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing required Bitbucket scopes") {
		t.Fatalf("expected scope remediation hint, got %q", err.Error())
	}
}

func TestUserFacingErrorNotFound(t *testing.T) {
	t.Parallel()

	err := userFacingError(bitbucket.NewAPIError(404, "404 Not Found", []byte(`{"type":"error","error":{"message":"Repository not found"}}`)))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "repository target") && !strings.Contains(err.Error(), "workspace slug") {
		t.Fatalf("expected target remediation hint, got %q", err.Error())
	}
}

func TestUserFacingErrorRepoResolutionHints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{
			input: "multiple workspaces available; specify --workspace",
			want:  "pass --repo <workspace>/<repo> or add --workspace",
		},
		{
			input: `repository target "widgets" requires --workspace`,
			want:  "pass --repo <workspace>/<repo> or add --workspace",
		},
		{
			input: "could not determine the repository from the current directory; run inside a Bitbucket git checkout or pass --repo",
			want:  "bb repo view --repo <workspace>/<repo>",
		},
		{
			input: "--workspace requires --repo",
			want:  "--workspace only disambiguates a repository target",
		},
	}

	for _, tc := range cases {
		err := userFacingError(assertErr(tc.input))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("expected %q to contain %q, got %v", tc.input, tc.want, err)
		}
	}
}

type staticError string

func (e staticError) Error() string { return string(e) }

func assertErr(message string) error {
	return staticError(message)
}
