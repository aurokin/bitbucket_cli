package cmd

import (
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
)

func TestIssueNeedsAttention(t *testing.T) {
	t.Parallel()

	if issueNeedsAttention(bitbucket.Issue{State: "resolved"}) {
		t.Fatal("expected resolved issue to be filtered out")
	}
	if !issueNeedsAttention(bitbucket.Issue{State: "new"}) {
		t.Fatal("expected new issue to need attention")
	}
}

func TestIssueInvolvesUser(t *testing.T) {
	t.Parallel()

	user := bitbucket.CurrentUser{AccountID: "user-1"}
	issue := bitbucket.Issue{
		Reporter: bitbucket.IssueActor{AccountID: "user-1"},
	}
	if !issueInvolvesUser(user, issue) {
		t.Fatal("expected reporter match")
	}
}
