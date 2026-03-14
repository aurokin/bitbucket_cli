package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	gitrepo "github.com/aurokin/bitbucket_cli/internal/git"
)

type stubWorkspaceResolver struct {
	workspaces []bitbucket.Workspace
	err        error
}

func (s stubWorkspaceResolver) ListWorkspaces(context.Context) ([]bitbucket.Workspace, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.workspaces, nil
}

func TestParseRepoSelector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		hostFlag      string
		workspaceFlag string
		repoFlag      string
		want          repoSelector
		wantErr       string
	}{
		{
			name: "none",
			want: repoSelector{},
		},
		{
			name:          "workspace without repo",
			repoFlag:      "",
			wantErr:       "--workspace requires --repo",
			workspaceFlag: "acme",
		},
		{
			name:     "bare repo",
			repoFlag: "widgets",
			want: repoSelector{
				Repo:     "widgets",
				Explicit: true,
			},
		},
		{
			name:          "bare repo with workspace flag",
			workspaceFlag: "acme",
			repoFlag:      "widgets",
			want: repoSelector{
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:     "workspace repo",
			repoFlag: "acme/widgets",
			want: repoSelector{
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:          "workspace mismatch",
			workspaceFlag: "Other",
			repoFlag:      "acme/widgets",
			wantErr:       `--workspace "Other" does not match repository target "acme/widgets"`,
		},
		{
			name:     "https repository url",
			repoFlag: "https://bitbucket.org/acme/widgets",
			want: repoSelector{
				Host:      "bitbucket.org",
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:     "ssh clone url",
			repoFlag: "ssh://git@bitbucket.org/acme/widgets.git",
			want: repoSelector{
				Host:      "bitbucket.org",
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:     "host mismatch",
			hostFlag: "example.com",
			repoFlag: "https://bitbucket.org/acme/widgets",
			wantErr:  `--host "example.com" does not match repository target "https://bitbucket.org/acme/widgets"`,
		},
		{
			name:     "invalid extra path",
			repoFlag: "https://bitbucket.org/acme/widgets/pull-requests/1",
			wantErr:  `repository URL "https://bitbucket.org/acme/widgets/pull-requests/1" must point to a repository`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRepoSelector(tc.hostFlag, tc.workspaceFlag, tc.repoFlag)
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

func TestParseRepoTargetInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		hostFlag      string
		workspaceFlag string
		repoFlag      string
		positional    string
		want          repoSelector
		wantErr       string
	}{
		{
			name:          "workspace flag applies to positional bare repo",
			workspaceFlag: "acme",
			positional:    "widgets",
			want: repoSelector{
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:       "repo flag only",
			repoFlag:   "acme/widgets",
			positional: "",
			want: repoSelector{
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:          "workspace flag merges with repo flag bare repo",
			workspaceFlag: "acme",
			repoFlag:      "widgets",
			want: repoSelector{
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:          "matching repo flag and positional",
			workspaceFlag: "acme",
			repoFlag:      "widgets",
			positional:    "acme/widgets",
			want: repoSelector{
				Workspace: "acme",
				Repo:      "widgets",
				Explicit:  true,
			},
		},
		{
			name:       "host mismatch with repo url",
			hostFlag:   "example.com",
			positional: "https://bitbucket.org/acme/widgets",
			wantErr:    `repository host "example.com" does not match "bitbucket.org"`,
		},
		{
			name:          "workspace mismatch",
			workspaceFlag: "Other",
			positional:    "acme/widgets",
			wantErr:       `repository workspace "Other" does not match "acme"`,
		},
		{
			name:          "workspace only is allowed until required later",
			workspaceFlag: "acme",
			want: repoSelector{
				Workspace: "acme",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRepoTargetInput(tc.hostFlag, tc.workspaceFlag, tc.repoFlag, tc.positional)
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

func TestParsePullRequestSelector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		want    pullRequestSelector
		wantErr string
	}{
		{
			name: "numeric id",
			raw:  "42",
			want: pullRequestSelector{ID: 42},
		},
		{
			name: "url",
			raw:  "https://bitbucket.org/acme/widgets/pull-requests/42",
			want: pullRequestSelector{
				Repo: repoSelector{
					Host:      "bitbucket.org",
					Workspace: "acme",
					Repo:      "widgets",
					Explicit:  true,
				},
				ID: 42,
			},
		},
		{
			name: "comment url resolves to parent pull request",
			raw:  "https://bitbucket.org/acme/widgets/pull-requests/42#comment-15",
			want: pullRequestSelector{
				Repo: repoSelector{
					Host:      "bitbucket.org",
					Workspace: "acme",
					Repo:      "widgets",
					Explicit:  true,
				},
				ID: 42,
			},
		},
		{
			name:    "invalid path",
			raw:     "https://bitbucket.org/acme/widgets/src/main.go",
			wantErr: `pull request URL "https://bitbucket.org/acme/widgets/src/main.go" must point to a Bitbucket pull request`,
		},
		{
			name:    "invalid raw",
			raw:     "feature-branch",
			wantErr: "pull request must be provided as an ID or Bitbucket pull request URL",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parsePullRequestSelector(tc.raw)
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

func TestParsePullRequestCommentSelector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		want    pullRequestCommentSelector
		wantErr string
	}{
		{
			name: "numeric id",
			raw:  "15",
			want: pullRequestCommentSelector{CommentID: 15},
		},
		{
			name: "comment url",
			raw:  "https://bitbucket.org/acme/widgets/pull-requests/42#comment-15",
			want: pullRequestCommentSelector{
				PR: pullRequestSelector{
					Repo: repoSelector{
						Host:      "bitbucket.org",
						Workspace: "acme",
						Repo:      "widgets",
						Explicit:  true,
					},
					ID: 42,
				},
				CommentID: 15,
			},
		},
		{
			name:    "pr url is not comment url",
			raw:     "https://bitbucket.org/acme/widgets/pull-requests/42",
			wantErr: `pull request comment URL "https://bitbucket.org/acme/widgets/pull-requests/42" must point to a Bitbucket pull request comment`,
		},
		{
			name:    "invalid raw",
			raw:     "branch-name",
			wantErr: "pull request comment must be provided as an ID or Bitbucket pull request comment URL",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parsePullRequestCommentSelector(tc.raw)
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

func TestMergeRepoSelectors(t *testing.T) {
	t.Parallel()

	base := repoSelector{Host: "bitbucket.org", Workspace: "acme", Repo: "widgets", Explicit: true}
	extra := repoSelector{Host: "bitbucket.org", Workspace: "acme", Repo: "widgets", Explicit: true}

	merged, err := mergeRepoSelectors(base, extra)
	if err != nil {
		t.Fatalf("mergeRepoSelectors returned error: %v", err)
	}
	if merged != base {
		t.Fatalf("expected %+v, got %+v", base, merged)
	}

	_, err = mergeRepoSelectors(base, repoSelector{Workspace: "Other"})
	if err == nil || err.Error() != `repository workspace "acme" does not match "Other"` {
		t.Fatalf("expected workspace mismatch error, got %v", err)
	}
}

func TestResolveRepoTargetFallsBackToLocalRepository(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{
		Host:      "bitbucket.org",
		Workspace: "acme",
		RepoSlug:  "widgets",
		RootDir:   "/tmp/widgets",
	}, nil)

	target, err := resolveRepoTarget(context.Background(), repoSelector{}, nil, true)
	if err != nil {
		t.Fatalf("resolveRepoTarget returned error: %v", err)
	}
	assertResolvedRepoTarget(t, target, "bitbucket.org", "acme", "widgets")
	if target.LocalRepo == nil || target.LocalRepo.RootDir != "/tmp/widgets" {
		t.Fatalf("expected local repo context, got %+v", target)
	}
}

func TestResolveRepoTargetInfersWorkspaceFromSingleAvailableWorkspace(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	target, err := resolveRepoTarget(context.Background(), repoSelector{
		Host:     "bitbucket.org",
		Repo:     "widgets",
		Explicit: true,
	}, stubWorkspaceResolver{
		workspaces: []bitbucket.Workspace{{Slug: "acme"}},
	}, false)
	if err != nil {
		t.Fatalf("resolveRepoTarget returned error: %v", err)
	}
	assertResolvedRepoTarget(t, target, "bitbucket.org", "acme", "widgets")
	if len(target.Warnings) != 0 {
		t.Fatalf("did not expect warnings when local inference is disabled, got %+v", target.Warnings)
	}
}

func TestResolveRepoTargetUsesMatchingLocalRepoForBareRepo(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{
		Host:      "bitbucket.org",
		Workspace: "acme",
		RepoSlug:  "widgets",
	}, nil)

	target, err := resolveRepoTarget(context.Background(), repoSelector{
		Repo:     "widgets",
		Explicit: true,
	}, stubWorkspaceResolver{
		workspaces: []bitbucket.Workspace{{Slug: "Other"}},
	}, true)
	if err != nil {
		t.Fatalf("resolveRepoTarget returned error: %v", err)
	}
	assertResolvedRepoTarget(t, target, "bitbucket.org", "acme", "widgets")
}

func TestResolveRepoTargetFailsWhenWorkspaceIsAmbiguous(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolveRepoTarget(context.Background(), repoSelector{
		Repo:     "widgets",
		Explicit: true,
	}, stubWorkspaceResolver{
		workspaces: []bitbucket.Workspace{{Slug: "One"}, {Slug: "Two"}},
	}, false)
	if err == nil || err.Error() != "multiple workspaces available; specify --workspace" {
		t.Fatalf("expected ambiguous workspace error, got %v", err)
	}
}

func TestResolveRepoTargetWarnsWhenExplicitRepoFallsBackWithoutLocalContext(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	target, err := resolveRepoTarget(context.Background(), repoSelector{
		Workspace: "acme",
		Repo:      "widgets",
		Explicit:  true,
	}, stubWorkspaceResolver{}, true)
	if err != nil {
		t.Fatalf("resolveRepoTarget returned error: %v", err)
	}
	if len(target.Warnings) != 1 || target.Warnings[0] != "local repository context unavailable; continuing without local checkout metadata (not a repo)" {
		t.Fatalf("expected local repo warning, got %+v", target.Warnings)
	}
}

func TestResolvePullRequestTargetUsesURLRepositoryContext(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	target, err := resolvePullRequestTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "https://bitbucket.org/acme/widgets/pull-requests/7", false)
	if err != nil {
		t.Fatalf("resolvePullRequestTarget returned error: %v", err)
	}
	assertResolvedPullRequestTarget(t, target, 7, "acme", "widgets")
}

func TestResolvePullRequestTargetUsesCommentURLRepositoryContext(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	target, err := resolvePullRequestTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15", false)
	if err != nil {
		t.Fatalf("resolvePullRequestTarget returned error: %v", err)
	}
	assertResolvedPullRequestTarget(t, target, 7, "acme", "widgets")
}

func TestResolvePullRequestTargetUsesLocalRepoForNumericID(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{
		Host:      "bitbucket.org",
		Workspace: "acme",
		RepoSlug:  "widgets",
	}, nil)

	target, err := resolvePullRequestTarget(context.Background(), repoSelector{}, nil, "7", true)
	if err != nil {
		t.Fatalf("resolvePullRequestTarget returned error: %v", err)
	}
	assertResolvedPullRequestTarget(t, target, 7, "acme", "widgets")
}

func TestResolvePullRequestTargetRejectsIssueURL(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolvePullRequestTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "https://bitbucket.org/acme/widgets/issues/7", false)
	if err == nil || err.Error() != `pull request URL "https://bitbucket.org/acme/widgets/issues/7" must point to a Bitbucket pull request` {
		t.Fatalf("expected issue-url rejection, got %v", err)
	}
}

func TestResolvePullRequestTargetRejectsMismatchedFlagAndURL(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolvePullRequestTarget(context.Background(), repoSelector{
		Workspace: "Other",
		Repo:      "widgets",
		Explicit:  true,
	}, stubWorkspaceResolver{}, "https://bitbucket.org/acme/widgets/pull-requests/7", false)
	if err == nil || err.Error() != `repository workspace "Other" does not match "acme"` {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestResolvePullRequestCommentTargetUsesCommentURLRepositoryContext(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	target, err := resolvePullRequestCommentTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "", "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15", false)
	if err != nil {
		t.Fatalf("resolvePullRequestCommentTarget returned error: %v", err)
	}
	assertResolvedPullRequestCommentTarget(t, target, 7, 15, "acme", "widgets")
}

func TestResolvePullRequestCommentTargetUsesNumericCommentIDWithPRRef(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{
		Host:      "bitbucket.org",
		Workspace: "acme",
		RepoSlug:  "widgets",
	}, nil)

	target, err := resolvePullRequestCommentTarget(context.Background(), repoSelector{}, nil, "7", "15", true)
	if err != nil {
		t.Fatalf("resolvePullRequestCommentTarget returned error: %v", err)
	}
	assertResolvedPullRequestCommentTarget(t, target, 7, 15, "acme", "widgets")
}

func TestResolvePullRequestCommentTargetRejectsNumericCommentIDWithoutPRRef(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolvePullRequestCommentTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "", "15", false)
	if err == nil || err.Error() != "pull request comment ID 15 requires --pr <id-or-url>" {
		t.Fatalf("expected missing pr error, got %v", err)
	}
}

func TestResolvePullRequestCommentTargetRejectsInvalidNumericCommentID(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolvePullRequestCommentTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "", "0", false)
	if err == nil || err.Error() != `invalid pull request comment ID "0"` {
		t.Fatalf("expected invalid comment id error, got %v", err)
	}
}

func TestResolvePullRequestCommentTargetRejectsNonCommentBitbucketURL(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolvePullRequestCommentTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "", "https://bitbucket.org/acme/widgets/pull-requests/7", false)
	if err == nil || err.Error() != `pull request comment URL "https://bitbucket.org/acme/widgets/pull-requests/7" must point to a Bitbucket pull request comment` {
		t.Fatalf("expected non-comment URL rejection, got %v", err)
	}
}

func TestResolvePullRequestCommentTargetRejectsMismatchedPRRefAndCommentURL(t *testing.T) {
	t.Parallel()

	withLocalRepoContext(t, gitrepo.RepoContext{}, errors.New("not a repo"))

	_, err := resolvePullRequestCommentTarget(context.Background(), repoSelector{}, stubWorkspaceResolver{}, "https://bitbucket.org/acme/widgets/pull-requests/8", "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15", false)
	if err == nil || err.Error() != `--pr "https://bitbucket.org/acme/widgets/pull-requests/8" does not match comment target "https://bitbucket.org/acme/widgets/pull-requests/7#comment-15"` {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestRequireExplicitRepoTarget(t *testing.T) {
	t.Parallel()

	if err := requireExplicitRepoTarget(repoSelector{Workspace: "acme"}); err == nil || err.Error() != "repository is required; pass <repo>, <workspace>/<repo>, or --repo" {
		t.Fatalf("expected missing repository error, got %v", err)
	}

	if err := requireExplicitRepoTarget(repoSelector{Workspace: "acme", Repo: "widgets"}); err != nil {
		t.Fatalf("did not expect error: %v", err)
	}
}

func assertResolvedRepoTarget(t *testing.T, target resolvedRepoTarget, host, workspace, repo string) {
	t.Helper()

	if target.Workspace != workspace || target.Repo != repo || target.Host != host {
		t.Fatalf("unexpected target %+v", target)
	}
}

func assertResolvedPullRequestTarget(t *testing.T, target resolvedPullRequestTarget, id int, workspace, repo string) {
	t.Helper()

	if target.ID != id {
		t.Fatalf("unexpected pull request id %+v", target)
	}
	assertResolvedRepoTarget(t, target.RepoTarget, "bitbucket.org", workspace, repo)
}

func assertResolvedPullRequestCommentTarget(t *testing.T, target resolvedPullRequestCommentTarget, prID, commentID int, workspace, repo string) {
	t.Helper()

	if target.CommentID != commentID {
		t.Fatalf("unexpected comment id %+v", target)
	}
	assertResolvedPullRequestTarget(t, target.PRTarget, prID, workspace, repo)
}

func withLocalRepoContext(t *testing.T, repo gitrepo.RepoContext, err error) {
	t.Helper()
	lockCommandTestHooks(t)

	originalGetwd := getWorkingDirectory
	originalResolve := resolveRepoAtDir

	getWorkingDirectory = func() (string, error) {
		return "/tmp/worktree", nil
	}
	resolveRepoAtDir = func(context.Context, string) (gitrepo.RepoContext, error) {
		if err != nil {
			return gitrepo.RepoContext{}, err
		}
		return repo, nil
	}

	t.Cleanup(func() {
		getWorkingDirectory = originalGetwd
		resolveRepoAtDir = originalResolve
	})
}

func TestCoalesce(t *testing.T) {
	t.Parallel()

	if got := coalesce("", " ", "value", "other"); got != "value" {
		t.Fatalf("expected value, got %q", got)
	}
}

func TestResolveLocalRepoContextPropagatesGetwdError(t *testing.T) {
	lockCommandTestHooks(t)

	originalGetwd := getWorkingDirectory
	originalResolve := resolveRepoAtDir
	getWorkingDirectory = func() (string, error) { return "", fmt.Errorf("boom") }
	resolveRepoAtDir = originalResolve
	t.Cleanup(func() {
		getWorkingDirectory = originalGetwd
		resolveRepoAtDir = originalResolve
	})

	_, err := resolveLocalRepoContext(context.Background())
	if err == nil || err.Error() != "get working directory: boom" {
		t.Fatalf("expected getwd error, got %v", err)
	}
}
