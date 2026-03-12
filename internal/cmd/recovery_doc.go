package cmd

import "strings"

type recoveryDocEntry struct {
	Title          string
	TypicalFailure string
	Recovery       []string
	Notes          []string
}

func recoveryDocEntries() []recoveryDocEntry {
	return []recoveryDocEntry{
		{
			Title:          "Invalid, Expired, Or Revoked API Token",
			TypicalFailure: "authentication failed: the stored API token may be invalid, expired, or revoked.",
			Recovery: []string{
				`BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login`,
				"bb auth status --check",
			},
			Notes: []string{
				"Create or rotate the token at " + atlassianAPITokenManageURL,
			},
		},
		{
			Title:          "Missing Token Scopes Or Insufficient Access",
			TypicalFailure: "request denied by Bitbucket Cloud: the API token may be missing required Bitbucket scopes",
			Recovery: []string{
				"bb auth status --check",
				"bb repo view --repo OhBizzle/bb-cli-integration-primary",
			},
			Notes: []string{
				"If the token is valid but still denied, create a new Bitbucket API token with the required Bitbucket scopes at " + atlassianAPITokenManageURL + " and store it again with `bb auth login`.",
			},
		},
		{
			Title:          "Ambiguous Repository Resolution",
			TypicalFailure: "multiple workspaces are available; pass --repo <workspace>/<repo> or add --workspace to disambiguate",
			Recovery: []string{
				"bb repo view --repo OhBizzle/bb-cli-integration-primary",
				"bb pr list --repo OhBizzle/bb-cli-integration-primary",
			},
			Notes: []string{
				"Prefer `--repo <workspace>/<repo>` in automation and when you are outside a local checkout.",
			},
		},
		{
			Title:          "No Repository In The Current Directory",
			TypicalFailure: "could not determine the repository from the current directory",
			Recovery: []string{
				"bb repo view --repo OhBizzle/bb-cli-integration-primary",
				"bb browse --repo OhBizzle/bb-cli-integration-primary --no-browser",
				"bb repo clone OhBizzle/bb-cli-integration-primary",
			},
		},
		{
			Title:          "Invalid Alias Quoting",
			TypicalFailure: `invalid alias "ship"`,
			Recovery: []string{
				"bb alias get ship",
				`bb alias set ship 'pr create --repo OhBizzle/bb-cli-integration-primary --title "Add feature"'`,
			},
			Notes: []string{
				"If the alias is no longer needed:",
				"`bb alias delete ship`",
			},
		},
		{
			Title:          "Repository Without Bitbucket Issue Tracking",
			TypicalFailure: "this repository does not have Bitbucket issue tracking enabled",
			Recovery: []string{
				"bb repo view --repo OhBizzle/bb-cli-integration-issues",
				"bb issue list --repo OhBizzle/bb-cli-integration-issues",
			},
			Notes: []string{
				"Use a repository with Bitbucket issue tracking enabled, or enable issue tracking in the repository settings.",
			},
		},
		{
			Title:          "Bounded Cross-Repository Status Output",
			TypicalFailure: "Notes\n  Some workspaces hit --repo-limit.\n  Some repositories do not have issue tracking enabled.",
			Recovery: []string{
				"bb status --workspace OhBizzle --repo-limit 200 --limit 50",
				"bb pr list --repo OhBizzle/bb-cli-integration-primary",
				"bb issue list --repo OhBizzle/bb-cli-integration-issues",
			},
			Notes: []string{
				"Use narrower workspace scans or explicit repository commands when you need complete detail.",
			},
		},
	}
}

func GenerateRecoveryDoc() (string, error) {
	var b strings.Builder
	b.WriteString("# Failure And Recovery\n\n")
	b.WriteString("Common failure modes, what they mean, and the next command to run.\n\n")
	b.WriteString("Generated from the current recovery guidance catalog.\n")

	for _, entry := range recoveryDocEntries() {
		b.WriteString("\n## ")
		b.WriteString(entry.Title)
		b.WriteString("\n\nTypical failure:\n\n```text\n")
		b.WriteString(entry.TypicalFailure)
		b.WriteString("\n```\n\nRecovery:\n\n```bash\n")
		b.WriteString(strings.Join(entry.Recovery, "\n"))
		b.WriteString("\n```\n")
		for _, note := range entry.Notes {
			b.WriteString("\n")
			b.WriteString(note)
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}
