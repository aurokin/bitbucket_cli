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
- Workspaces: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-workspaces/
- Projects: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-projects/

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
bb repo list workspace-slug
bb browse --repo workspace-slug/repo-slug --no-browser
bb repo view --repo workspace-slug/repo-slug
bb repo edit --repo workspace-slug/repo-slug --description "Updated description"
bb repo fork workspace-slug/repo-slug --to-workspace workspace-slug --name repo-slug-fork --reuse-existing
bb repo hook list --repo workspace-slug/repo-slug
bb repo deploy-key list --repo workspace-slug/repo-slug
bb repo permissions user list --repo workspace-slug/repo-slug
bb repo clone workspace-slug/repo-slug
```

Bitbucket rejected deploy-key updates in the live API behavior we verified, so rotate deploy keys by deleting and re-creating them instead of expecting an in-place edit flow.

Repository permission mutation also stays out of scope for now. Bitbucket's permission write/delete docs still describe app-password-only behavior in places, so `bb` only exposes explicit permission inspection until the API-token path is verified live.

## Inspect Workspaces And Projects

```bash
bb workspace list
bb workspace view workspace-slug
bb workspace member list workspace-slug
bb workspace permission list workspace-slug
bb workspace repo-permission list workspace-slug --repo workspace-slug/repo-slug
bb project list workspace-slug
bb project view BBCLI --workspace workspace-slug
bb project default-reviewer list BBCLI --workspace workspace-slug
bb project permissions user list BBCLI --workspace workspace-slug
bb project create TMP --workspace workspace-slug --name "Temp project"
bb project edit TMP --workspace workspace-slug --description "Updated by automation"
bb project delete TMP --workspace workspace-slug --yes
```

Project permission mutation also stays out of scope for now. `bb` only exposes explicit project permission inspection until the API-token path is verified live for documented write behavior.

## Manage Branches And Tags

```bash
bb branch list --repo workspace-slug/repo-slug
bb branch view main --repo workspace-slug/repo-slug
bb branch create feature/demo --repo workspace-slug/repo-slug --target main
bb tag list --repo workspace-slug/repo-slug
bb tag create v1.0.0 --repo workspace-slug/repo-slug --target main --message "release"
```

These commands are backed by the official Bitbucket Cloud refs APIs:

- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-refs/

## Inspect A Commit

```bash
bb commit view abc1234 --repo workspace-slug/repo-slug
bb commit diff abc1234 --repo workspace-slug/repo-slug --stat
bb commit statuses abc1234 --repo workspace-slug/repo-slug
bb commit report list abc1234 --repo workspace-slug/repo-slug
```

These commands are backed by the official Bitbucket Cloud commits, commit statuses, and reports APIs:

- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/
- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commit-statuses/
- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-reports/

## Inspect Pipelines

```bash
bb pipeline list --repo workspace-slug/pipelines-repo-slug
bb deployment environment list --repo workspace-slug/pipelines-repo-slug
bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test
bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key APP_ENV --value production
bb deployment environment variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --environment test --value staging
bb deployment environment variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --environment test --yes
bb pipeline run --repo workspace-slug/pipelines-repo-slug --ref main
bb pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug
bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug
bb pipeline test-reports 1 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
bb pipeline log 1 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug
```

When you stop a pipeline, Bitbucket can still finish the run in another terminal state before it becomes `STOPPED`. `bb pipeline stop` reports the final observed pipeline state so you can tell the difference between a true stop and a late stop request.

Bitbucket Cloud repository downloads remain out of scope for this CLI on the verified API-token path because the official downloads endpoint currently returns a workspace-plan `402 Payment Required` response on the fixture workspace.

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
bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body "Needs follow-up."
bb issue comment list 1 --repo workspace-slug/issues-repo-slug
bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug
bb issue attachment list 1 --repo workspace-slug/issues-repo-slug
bb issue milestone list --repo workspace-slug/issues-repo-slug
bb issue component list --repo workspace-slug/issues-repo-slug
bb issue edit 1 --repo workspace-slug/issues-repo-slug --priority major
bb issue close 1 --repo workspace-slug/issues-repo-slug --message "Fixed in main."
```

Bitbucket Cloud documents issue import and export jobs, but the current endpoints reject API-token auth. `bb` leaves those workflows out instead of wrapping a path that cannot succeed with the CLI's supported auth model.

## Search For Work

```bash
bb search repos bb-cli --workspace workspace-slug
bb search prs fixture --repo workspace-slug/repo-slug
bb search issues broken --repo workspace-slug/issues-repo-slug
bb status --workspace workspace-slug --limit 10
```

## Reuse Fixtures Safely

The live Bitbucket integration tests intentionally reuse the existing fixture project and repositories where possible. For destructive flows like repository deletion, use sacrificial fixtures instead of deleting the primary repositories.
