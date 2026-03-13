package cmd

import (
	"bytes"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestWriteBranchListSummary(t *testing.T) {
	t.Parallel()

	payload := branchListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Branches: []bitbucket.RepositoryBranch{
			{Name: "main", Target: bitbucket.RepositoryCommit{Hash: "abc1234"}, DefaultMergeStrategy: "squash"},
		},
	}

	var buf bytes.Buffer
	if err := writeBranchListSummary(&buf, payload); err != nil {
		t.Fatalf("writeBranchListSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Next: bb branch view main --repo acme/widgets",
	)
}

func TestWriteTagListSummary(t *testing.T) {
	t.Parallel()

	payload := tagListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Tags: []bitbucket.RepositoryTag{
			{Name: "v1.0.0", Target: bitbucket.RepositoryCommit{Hash: "abc1234"}, Date: "2026-03-13T00:00:00Z"},
		},
	}

	var buf bytes.Buffer
	if err := writeTagListSummary(&buf, payload); err != nil {
		t.Fatalf("writeTagListSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Next: bb tag view v1.0.0 --repo acme/widgets",
	)
}
