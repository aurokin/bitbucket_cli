---
name: bb-cli
description: Bitbucket Cloud CLI skill for this repository's `bb` command. Use when a task mentions Bitbucket, Bitbucket Cloud, `bb`, or Bitbucket web URLs like `https://bitbucket.org/...`, including workspace, project, repository, pull request, pull request comment, issue, commit, and source links. Covers installation, API-token authentication, URL resolution, browse flows, workspaces, projects, repositories, pull requests, comments, tasks, pipelines, issues, search, status, structured output, and official Bitbucket Cloud REST API fallback with `bb api`.
---

# bb CLI

Use this skill for Bitbucket Cloud work through this repository's `bb` CLI.

## Use This Skill When

- The task mentions Bitbucket, Bitbucket Cloud, or `bb`
- The task includes a `bitbucket.org/...` URL
- The task is about workspaces, projects, repositories, PRs, PR comments, PR tasks, issues, pipelines, search, status, or browsing in Bitbucket Cloud
- The task needs structured Bitbucket CLI output for an agent
- The task needs official Bitbucket Cloud REST API fallback through `bb api`

## Do Not Use This Skill When

- The task is plain local git work like commit, rebase, push, or branch management
- The task targets GitHub, GitLab, or Bitbucket Server/Data Center instead of Bitbucket Cloud
- The task needs a browser-only Bitbucket behavior that the official Cloud API does not expose
- The task can only be done by inventing fake `gh` parity instead of using real Bitbucket Cloud behavior

## Quick Setup

Install or update:

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@latest
bb version
```

Authenticate with an Atlassian API token:

```bash
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
bb auth status --check
```

Alternative token entry:

```bash
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
```

Create or rotate the token at:

- https://id.atlassian.com/manage-profile/security/api-tokens
- https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/

## Core Rules

- Prefer `--repo <workspace>/<repo>` over local inference.
- Use `--workspace` only to disambiguate a bare repository name.
- Use `--json` or `--json '*'` for machine parsing.
- Use `--jq` when a smaller result is enough.
- Use `--no-prompt` for mutations and non-interactive runs.
- Do not parse human-readable output when structured output is available.
- When a task starts from a Bitbucket URL, run `bb resolve <url>` first.
- If `bb` does not wrap an official endpoint yet, use `bb api`.

## First Commands To Reach For

Normalize a Bitbucket URL:

```bash
bb resolve <url> --json '*'
```

Print a deterministic browser URL instead of opening it:

```bash
bb browse --repo workspace-slug/repo-slug --no-browser --json url,type
bb browse --pr 7 --repo workspace-slug/repo-slug --no-browser --json url,type,pr
```

Inspect repository or pipeline state:

```bash
bb repo view --repo workspace-slug/repo-slug --json name,main_branch,html_url
bb pipeline list --repo workspace-slug/pipelines-repo-slug --json build_number,state,target,created_on
bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug --json pipeline,steps
```

Inspect workspace or project state:

```bash
bb workspace list --json workspaces
bb workspace member list workspace-slug --json members
bb project list workspace-slug --json projects
bb project view BBCLI --workspace workspace-slug --json project
bb project permissions user list BBCLI --workspace workspace-slug --json permissions
```

Inspect pull requests:

```bash
bb pr list --repo workspace-slug/repo-slug --json id,title,state,task_count,comment_count
bb pr view 7 --repo workspace-slug/repo-slug --json '*'
bb pr diff 7 --repo workspace-slug/repo-slug --json patch,stats
bb pr status --json authored_prs,review_requested_prs,current_branch_pr,warnings
```

Work from PR comment URLs directly:

```bash
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr comment resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr task create 7 --repo workspace-slug/repo-slug --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --body "Handle this thread" --json '*'
```

Issues, search, and status:

```bash
bb issue list --repo workspace-slug/issues-repo-slug --json id,title,state
bb search prs bugfix --repo workspace-slug/repo-slug --jq '.[] | {id, title, state}'
bb status --workspace workspace-slug --limit 10 --json authored_prs,review_requested_prs,your_issues,warnings
```

Raw API fallback:

```bash
bb api /user --jq '{display_name, account_id}'
bb api /2.0/repositories/workspace-slug/repo-slug/pullrequests --jq '.values[] | {id, title, state}'
```

## Recommended Agent Workflow

1. If the task starts from a pasted Bitbucket URL, resolve it first with `bb resolve <url> --json '*'`.
2. Prefer explicit `--repo workspace/repo` targeting.
3. Prefer `--json` plus `--jq` over human-readable output.
4. Use command-specific flows first, then fall back to `bb api` for uncovered official endpoints.
5. If a command might prompt, force `--no-prompt`.

## Output Expectations

- Human output is header-first and often includes a `Next:` hint.
- JSON output is the stable agent interface.
- Some commands include `warnings` in JSON when local inference degraded or context could not be resolved cleanly.
- `bb resolve` and `bb browse --no-browser` are the preferred starting points for pasted Bitbucket URLs.

## Bitbucket Cloud Limits To Respect

- Do not invent unsupported behavior.
- Bitbucket Cloud does not support reopening a declined PR, so `bb` does not provide a fake `pr reopen`.
- Bitbucket Cloud does not expose PR comment likes/reactions through the official Cloud REST API.
- Browser login is intentionally out of scope; auth is API-token-first.
- Project permission mutation is intentionally out of scope until the API-token path is verified cleanly enough to support as a documented workflow.
- If a pipeline step has no exposed raw log file, `bb pipeline log` should fail clearly instead of fabricating output.

## API Grounding

Use the official Bitbucket Cloud REST API docs first when behavior is unclear:

- Overview: https://developer.atlassian.com/cloud/bitbucket/about-bitbucket-cloud-rest-api/
- REST intro: https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- OpenAPI document: https://api.bitbucket.org/swagger.json
- Repositories: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/
- Pull requests: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- Pipelines: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/
- Issue tracker: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-issue-tracker/
- Workspaces: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-workspaces/
- Projects: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-projects/
- Users: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/
- Source: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-source/
