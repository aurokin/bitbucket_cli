# Automation

Deterministic usage patterns for agents, scripts, and CI-adjacent tooling.

Use the generated [CLI reference](./cli-reference.md) for the full command surface.

`bb` automation is built on the official Bitbucket Cloud REST API. When a wrapped command does not exist yet, prefer `bb api` against the documented Bitbucket Cloud endpoints:

- REST intro: https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- OpenAPI document: https://api.bitbucket.org/swagger.json
- Repositories: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/
- Pull requests: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- Pipelines: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/
- Issue tracker: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-issue-tracker/
- Users: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/

## Rules

- Prefer `--repo <workspace>/<repo>` over local inference.
- Use `--workspace` only to disambiguate a bare repository name.
- Use `--json` or `--json '*'` for machine parsing.
- Use `--jq` when a smaller result is enough.
- Use `--no-prompt` for mutations and all non-interactive runs.
- Do not parse the default human-readable output when structured output is available.

## Browse

```bash
bb browse --repo OhBizzle/bb-cli-integration-primary --no-browser --json url,type
bb browse README.md:12 --repo OhBizzle/bb-cli-integration-primary --branch main --no-browser --json url,type,path,line,ref
bb browse --pr 1 --repo OhBizzle/bb-cli-integration-primary --no-browser --json url,type,pr
```

## Authentication

```bash
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
bb auth status --check --json
bb config set browser 'firefox --new-window'
```

Create or rotate the token at:

- https://id.atlassian.com/manage-profile/security/api-tokens
- https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/

Validate raw current-user auth behavior with:

```bash
bb api /user --jq '{display_name, account_id}'
```

## Repository Commands

```bash
bb repo view --repo OhBizzle/bb-cli-integration-primary --json name,project_key,main_branch,html_url
bb repo create --repo OhBizzle/example-repo --project-key BBCLI --reuse-existing --json slug,name,project
bb repo clone OhBizzle/bb-cli-integration-primary /tmp/bb-cli-integration-primary --json workspace,repo,directory
bb --no-prompt repo delete OhBizzle/example-repo --yes --json workspace,repo,deleted
```

## Pipeline Commands

```bash
bb pipeline list --repo OhBizzle/bb-cli-integration-pipelines --json build_number,state,target,created_on
bb pipeline log 1 --repo OhBizzle/bb-cli-integration-pipelines --step '{step-uuid}' --json pipeline,step,log
bb --no-prompt pipeline stop 1 --repo OhBizzle/bb-cli-integration-pipelines --yes --json pipeline,stopped
bb pipeline view 1 --repo OhBizzle/bb-cli-integration-pipelines --json host,workspace,repo,pipeline,steps
```

## Pull Request Commands

```bash
bb pr list --repo OhBizzle/bb-cli-integration-primary --json id,title,state,author
bb pr view 1 --repo OhBizzle/bb-cli-integration-primary --json '*'
bb pr diff 1 --repo OhBizzle/bb-cli-integration-primary --json patch,stats
bb pr comment 1 --repo OhBizzle/bb-cli-integration-primary --body "Please add a regression test." --json id,content,links
bb --no-prompt pr create --repo OhBizzle/bb-cli-integration-primary --source feature --destination main --title "Add feature" --json id,title,state
bb pr merge 2 --repo OhBizzle/bb-cli-integration-primary --json id,title,state
bb pr close 3 --repo OhBizzle/bb-cli-integration-primary --json id,title,state
```

## Issue Commands

```bash
bb issue list --repo OhBizzle/bb-cli-integration-issues --json id,title,state
bb issue view 1 --repo OhBizzle/bb-cli-integration-issues --json '*'
bb issue create --repo OhBizzle/bb-cli-integration-issues --title "Broken flow" --body "Needs investigation." --json id,title,state
bb issue edit 1 --repo OhBizzle/bb-cli-integration-issues --priority major --json id,title,priority,state
bb issue close 1 --repo OhBizzle/bb-cli-integration-issues --json id,title,state
bb issue reopen 1 --repo OhBizzle/bb-cli-integration-issues --json id,title,state
```

## Search And Status

```bash
bb search repos bb-cli --workspace OhBizzle --json name,slug,project
bb search prs fixture --repo OhBizzle/bb-cli-integration-primary --jq '.[] | {id, title, state}'
bb search issues broken --repo OhBizzle/bb-cli-integration-issues --json id,title,state
bb status --workspace OhBizzle --limit 10 --json authored_prs,review_requested_prs,your_issues,warnings
```

## Raw REST Access

Use `bb api` when you need an official endpoint that is not wrapped yet:

```bash
bb api /user --jq '{display_name, account_id}'
bb api /2.0/repositories/OhBizzle/bb-cli-integration-primary --jq '{slug, project, mainbranch}'
bb api /2.0/repositories/OhBizzle/bb-cli-integration-primary/pullrequests --jq '.values[] | {id, title, state}'
```

## Alias And Config

```bash
bb config set output.format json
bb config get output.format --json
bb alias set pv 'pr view --repo OhBizzle/bb-cli-integration-primary'
bb alias get pv
bb extension list --json
```

## Manual Live Verification

Live Bitbucket integration tests remain manual-only. They are useful when you need to validate real API behavior or the human-readable output path:

```bash
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudHumanOutputSmoke -v
```
