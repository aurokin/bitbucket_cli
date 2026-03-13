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

## Install And Update

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@latest
bb version
```

Pin automation to a specific release when you need reproducibility:

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@v0.1.0
bb version
```

## Rules

- Prefer `--repo <workspace>/<repo>` over local inference.
- Use `--workspace` only to disambiguate a bare repository name.
- Use `bb resolve <url> --json '*'` when automation needs to normalize a Bitbucket URL before choosing a command.
- Use `--json` or `--json '*'` for machine parsing.
- Use `--jq` when a smaller result is enough.
- Use `--no-prompt` for mutations and all non-interactive runs.
- Do not parse the default human-readable output when structured output is available.

## URL Resolution

```bash
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json type,repo,pr,comment,canonical_url
bb resolve https://bitbucket.org/workspace-slug/repo-slug/src/main/README.md#lines-12 --jq '{type, repo, path, line}'
```

## Authentication

```bash
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
bb auth status --check --json
```

Create or rotate the token at:

- https://id.atlassian.com/manage-profile/security/api-tokens
- https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/

Validate raw current-user auth behavior with:

```bash
bb api /user --jq '{display_name, account_id}'
```

## Representative Command Patterns

Use the generated [flag matrix](./flag-matrix.md), [CLI reference](./cli-reference.md), [JSON fields](./json-fields.md), and [JSON shapes](./json-shapes.md) for exhaustive details. Keep automation examples short and deterministic:

```bash
bb repo view --repo workspace-slug/repo-slug --json name,project_key,main_branch,html_url
bb pipeline list --repo workspace-slug/pipelines-repo-slug --json build_number,state,target,created_on
bb pr list --repo workspace-slug/repo-slug --json id,title,state,task_count,comment_count
bb pr view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --json id,title,state
bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --json '*'
bb pr comment resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --json '*'
bb pr task create 1 --repo workspace-slug/repo-slug --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body "Handle this thread" --json '*'
bb --no-prompt pr create --repo workspace-slug/repo-slug --source feature --destination main --title "Add feature" --json id,title,state
bb issue list --repo workspace-slug/issues-repo-slug --json id,title,state
bb search repos bb-cli --workspace workspace-slug --json name,slug,project
bb search prs fixture --repo workspace-slug/repo-slug --jq '.[] | {id, title, state, task_count, comment_count}'
bb status --workspace workspace-slug --limit 10 --json authored_prs,review_requested_prs,your_issues,warnings
bb api /user --jq '{display_name, account_id}'
bb api /2.0/repositories/workspace-slug/repo-slug --jq '{slug, project, mainbranch}'
bb config set output.format json
bb alias set pv 'pr view --repo workspace-slug/repo-slug'
bb extension list --json
bb browse --pr 1 --repo workspace-slug/repo-slug --no-browser --json url,type,pr
```

## Manual Live Verification

Live Bitbucket integration tests remain manual-only. They are useful when you need to validate real API behavior or the human-readable output path:

```bash
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudHumanOutputSmoke -v
```
