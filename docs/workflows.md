# Workflows

Task-oriented command recipes for humans using `bb`.

Use the generated [CLI reference](./cli-reference.md) for the full command surface.

`bb` is built on the official Bitbucket Cloud REST API. If you need to understand the underlying platform behavior behind a workflow, start with:

- Overview: https://developer.atlassian.com/cloud/bitbucket/about-bitbucket-cloud-rest-api/
- REST intro: https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- Pull requests: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- Repositories: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/
- Pipelines: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/
- Issue tracker: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-issue-tracker/

## Install

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@latest
bb version
```

## Update

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@latest
bb version
```

If you want a specific tagged release:

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@v0.1.0
bb version
```

## Authenticate

```bash
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
bb auth status --check
```

Create the token here before logging in:

- https://id.atlassian.com/manage-profile/security/api-tokens
- https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/

## Inspect And Clone A Repository

Inside a checkout, local git remotes are enough:

```bash
bb browse
bb browse README.md:12 --no-browser
bb repo view
bb pr list
```

Outside a checkout, prefer an explicit target:

```bash
bb browse --repo workspace-slug/repo-slug --no-browser
bb repo view --repo workspace-slug/repo-slug
bb repo clone workspace-slug/repo-slug
```

## Inspect Pipelines

```bash
bb pipeline list --repo workspace-slug/pipelines-repo-slug
bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug
bb pipeline log 1 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
```

## Review A Pull Request

```bash
bb pr list --repo workspace-slug/repo-slug
bb pr view 1 --repo workspace-slug/repo-slug
bb pr activity 1 --repo workspace-slug/repo-slug
bb pr commits 1 --repo workspace-slug/repo-slug
bb pr checks 1 --repo workspace-slug/repo-slug
bb pr diff 1 --repo workspace-slug/repo-slug --stat
bb pr review approve 1 --repo workspace-slug/repo-slug
bb pr comment 1 --repo workspace-slug/repo-slug --body "Looks good overall. Please tighten the error handling."
```

The default PR list and status views include task and comment counts so humans can spot follow-up work quickly before opening the full task or comment detail.

## Resolve Or Reopen A Comment Thread

```bash
bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment reopen 15 --pr 1 --repo workspace-slug/repo-slug
```

## Track Follow-Up With A Pull Request Task

```bash
bb pr task create 1 --repo workspace-slug/repo-slug --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body "Handle this thread"
bb pr task list 1 --repo workspace-slug/repo-slug
bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task reopen 3 --pr 1 --repo workspace-slug/repo-slug
```

## Create And Land A Pull Request

From a local checkout:

```bash
bb pr create --source feature --destination main --title "Add feature"
bb pr view 2
bb pr merge 2 --repo workspace-slug/repo-slug
```

For a deterministic non-interactive flow:

```bash
bb --no-prompt pr create \
  --repo workspace-slug/repo-slug \
  --source feature \
  --destination main \
  --title "Add feature"
```

## Triage Issues

```bash
bb issue list --repo workspace-slug/issues-repo-slug
bb issue create --repo workspace-slug/issues-repo-slug --title "Broken flow" --body "Needs investigation."
bb issue view 1 --repo workspace-slug/issues-repo-slug
bb issue edit 1 --repo workspace-slug/issues-repo-slug --priority major
bb issue close 1 --repo workspace-slug/issues-repo-slug --message "Fixed in main."
```

## Search For Work

```bash
bb search repos bb-cli --workspace workspace-slug
bb search prs fixture --repo workspace-slug/repo-slug
bb search issues broken --repo workspace-slug/issues-repo-slug
bb status --workspace workspace-slug --limit 10
```

## Reuse Fixtures Safely

The live Bitbucket integration tests intentionally reuse the existing fixture project and repositories where possible. For destructive flows like repository deletion, use sacrificial fixtures instead of deleting the primary repositories.
