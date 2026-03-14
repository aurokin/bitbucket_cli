# bb

`bb` is a Bitbucket Cloud CLI aimed at both humans and agents.

## Install CLI

`bb` is currently installed with Go. There are no packaged release binaries yet.

Requirements:

- Go `1.25+`
- a `PATH` entry that includes `$(go env GOPATH)/bin` or your configured `GOBIN`

Fresh-machine install:

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@latest
bb version
```

Update an existing install:

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@latest
bb version
```

If you want a specific tagged release instead of the latest published version:

```bash
go install github.com/aurokin/bitbucket_cli/cmd/bb@v0.1.0
bb version
```

If you are working from a local checkout instead:

```bash
go build ./cmd/bb
./bb version
```

Update a local checkout install:

```bash
git pull --ff-only
go install ./cmd/bb
bb version
```

## Local Gate

For local development from a checkout, the regular quality gate is:

```bash
make check
```

That runs:

- `go test ./...`
- `golangci-lint run ./...`
- explicit complexity checks with `gocognit` and `gocyclo`

Additional validation targets:

- `make race`
- `make fuzz-short`

The first run installs pinned dev tools into `.tools/bin` with:

```bash
make tools
```

## Install Agent Skill

This repo also ships a reusable `bb-cli` skill for agents.

Install it from this repo with:

```bash
npx skills add https://github.com/aurokin/bitbucket_cli --skill bb-cli
```

You can also inspect the skill directly in [skills/bb-cli](./skills/bb-cli).

## Quick Start

Authenticate with an Atlassian API token:

```bash
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
bb auth status --check
```

Create or rotate the token here:

- https://id.atlassian.com/manage-profile/security/api-tokens
- https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/

Prefer explicit repository targets when you are outside a checkout or writing automation:

```bash
bb workspace list
bb workspace member list workspace-slug
bb workspace permission list workspace-slug
bb workspace repo-permission list workspace-slug --repo workspace-slug/repo-slug
bb project list workspace-slug
bb project view BBCLI --workspace workspace-slug
bb project default-reviewer list BBCLI --workspace workspace-slug
bb project permissions user list BBCLI --workspace workspace-slug
bb browse --repo workspace-slug/repo-slug --no-browser
bb repo list workspace-slug
bb repo view --repo workspace-slug/repo-slug
bb repo fork workspace-slug/repo-slug --to-workspace workspace-slug --name repo-slug-fork
bb repo hook list --repo workspace-slug/repo-slug
bb repo deploy-key list --repo workspace-slug/repo-slug
bb repo permissions user list --repo workspace-slug/repo-slug
bb branch list --repo workspace-slug/repo-slug
bb tag list --repo workspace-slug/repo-slug
bb commit view abc1234 --repo workspace-slug/repo-slug
bb commit statuses abc1234 --repo workspace-slug/repo-slug
bb commit report list abc1234 --repo workspace-slug/repo-slug
bb pipeline list --repo workspace-slug/pipelines-repo-slug
bb deployment environment list --repo workspace-slug/pipelines-repo-slug
bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test
bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key APP_ENV --value production
bb pipeline run --repo workspace-slug/pipelines-repo-slug --ref main
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug
bb pr list --repo workspace-slug/repo-slug
bb issue list --repo workspace-slug/issues-repo-slug
bb issue comment list 1 --repo workspace-slug/issues-repo-slug
bb issue attachment list 1 --repo workspace-slug/issues-repo-slug
bb issue milestone list --repo workspace-slug/issues-repo-slug
bb issue component list --repo workspace-slug/issues-repo-slug
```

If a workflow starts from a pasted Bitbucket URL, normalize it first:

```bash
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
```

## Implementation

`bb` is implemented against the official Bitbucket Cloud REST API and stays aligned with documented Bitbucket Cloud behavior instead of inventing `gh` parity where the platform does not support it.

Implementation rules:

- command behavior is built on the documented `https://api.bitbucket.org/2.0` REST surface
- the raw `bb api` command maps directly onto Bitbucket Cloud REST paths and URLs
- unsupported Bitbucket Cloud behaviors are documented explicitly instead of being approximated silently
- auth is API-token based because that is the clean, supported CLI path we could verify directly

Primary Bitbucket Cloud API references:

- Overview: https://developer.atlassian.com/cloud/bitbucket/about-bitbucket-cloud-rest-api/
- REST reference intro: https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- Canonical OpenAPI document: https://api.bitbucket.org/swagger.json

Main API groups used by `bb`:

- Repositories and project-linked repository operations: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/
- Pull requests: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- Pipelines: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/
- Issue tracker: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-issue-tracker/
- Workspaces: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-workspaces/
- Projects: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-projects/
- Users and current-account validation: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/
- Source browsing and file-oriented repository URLs: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-source/

## Typical Usage

For humans:

```bash
bb repo view
bb pr view 1
bb issue create --title "Broken flow"
bb status
```

For agents and scripts:

```bash
bb workspace list --json workspaces
bb project list workspace-slug --json projects
bb project permissions user list BBCLI --workspace workspace-slug --json permissions
bb repo list workspace-slug --json repos
bb repo edit --repo workspace-slug/repo-slug --description "Updated by automation" --json repository
bb repo fork workspace-slug/repo-slug --to-workspace workspace-slug --name repo-slug-fork --reuse-existing --json repository
bb repo hook list --repo workspace-slug/repo-slug --json hooks
bb repo deploy-key list --repo workspace-slug/repo-slug --json keys
bb repo permissions user list --repo workspace-slug/repo-slug --json permissions
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb branch list --repo workspace-slug/repo-slug --json branches
bb tag list --repo workspace-slug/repo-slug --json tags
bb commit view https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json commit
bb commit statuses abc1234 --repo workspace-slug/repo-slug --json statuses
bb commit report list abc1234 --repo workspace-slug/repo-slug --json reports
bb pr list --repo workspace-slug/repo-slug --json id,title,state,task_count,comment_count
bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug --json pipeline,steps
bb deployment environment list --repo workspace-slug/pipelines-repo-slug --json environments
bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test --json variables
bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key APP_ENV --value production --json variable
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug --json variables
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug --json schedules
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug --json caches
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug --json runners
bb search prs fixture --repo workspace-slug/repo-slug --jq '.[] | .id'
bb issue comment list 1 --repo workspace-slug/issues-repo-slug --json comments
bb issue attachment list 1 --repo workspace-slug/issues-repo-slug --json attachments
bb issue milestone list --repo workspace-slug/issues-repo-slug --json milestones
bb issue component list --repo workspace-slug/issues-repo-slug --json components
```

Key behavior:

- human-readable output is header-first, compact, and usually includes `Next:` guidance
- prefer `--repo <workspace>/<repo>` over local inference
- use `--json` or `--json '*'` for machine parsing
- use `--jq` to keep agent output token-efficient
- use `--no-prompt` for mutations and any non-interactive flow

## Additional Docs

- [Human workflows](./docs/workflows.md)
- [Automation guide](./docs/automation.md)
- [Flag matrix](./docs/flag-matrix.md)
- [Error index](./docs/error-index.md)
- [JSON field index](./docs/json-fields.md)
- [JSON shapes](./docs/json-shapes.md)
- [Failure and recovery](./docs/recovery.md)
- [CLI reference](./docs/cli-reference.md)

## Command Surface

Use the generated [CLI reference](./docs/cli-reference.md) for the full command tree and flag details. The high-level command families are:

- auth, api, branch, browse, commit, deployment, project, resolve, tag, and workspace
- repo, pipeline, pr, and issue
- search, status, config, alias, and extension

## Compared With `gh`

### What `gh` Offers That `bb` Also Offers

- Authenticated API access through `gh api` / `bb api`
- Browser navigation through `gh browse` / `bb browse`
- Repository listing, inspection, creation, editing, forking, webhook/deploy-key administration, cloning, and deletion
- Branch and tag listing, inspection, creation, and deletion
- Pipeline run triggering, listing, inspection, test reports, and repository variable management
- Pull request listing, review, status, activity, commit inspection, viewing, diffing, commenting, creation, checkout, merge, and close flows
- Repository commit viewing, diffing, approval, status inspection, comment inspection, and code-insight report inspection
- Issue listing, viewing, creation, editing, and state transitions
- Issue comment listing, creation, viewing, editing, and deletion
- Issue milestone and component inspection
- Cross-repository status summaries
- Search for repositories, pull requests, and issues
- Config defaults for prompt behavior and default output format, plus aliases and extension discovery
- Structured automation paths with `--json`
- Flexible repository targeting with local git inference, `workspace/repo`, and Bitbucket/GitHub-style URLs
- Workspace and project inspection plus project creation, editing, and deletion

### What `bb` Offers That `gh` Does Not

- Native Bitbucket Cloud repository and pull request workflows
- Bitbucket project-aware repository creation through `--project-key`
- Explicit documentation of Bitbucket Cloud platform limits when parity is impossible

### What `gh` Offers That `bb` Does Not

- Broader auth account management
- Releases and broader CI/workflow management such as dispatching, rerunning, and log-heavy workflow tooling
- Richer repository administration such as rename, archive, and sync
- Additional pull request flows such as edit, ready, update-branch, and revert
- Pull request reopen on platforms that actually support it

## Why There Is No `bb pr reopen`

Bitbucket Cloud does not support reopening a declined pull request.

Once a pull request is declined in Bitbucket Cloud, it stays declined. The public Bitbucket Cloud pull request API exposes merge and decline operations, but it does not expose a reopen operation. Atlassian also documents this product limitation in their public issue tracker.

Because of that, `bb` does not provide a misleading `pr reopen` command that would pretend to restore the original pull request. The correct workflow is to create a new pull request from the same source and destination branches when you need to continue the work.

References:

- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- https://jira.atlassian.com/browse/BCLOUD-23807

## Notes On Pipeline Behavior

- `bb` supports the Bitbucket Cloud pipeline APIs we could verify directly today: run, list, view, test reports, repository variables, schedules, runners, caches, log, and stop.
- `bb deployment` supports the official Bitbucket Cloud deployment and environment paths we could verify directly today: deployment listing, environment inspection, and deployment environment variable inspection plus create, edit, and delete.
- Bitbucket's single deployment-variable GET endpoint currently rejects the verified API-token path even though the list/create/edit/delete paths work. `bb deployment environment variable view` resolves through the list endpoint so the user-facing workflow stays reliable.
- `bb` does not provide pipeline rerun because the current Bitbucket Cloud pipeline REST docs do not expose a rerun endpoint. `bb` does not fake rerun by creating a new run behind your back.
- Raw step logs are not guaranteed for every Bitbucket pipeline step. When Bitbucket does not expose a log file for a step, `bb pipeline log` fails clearly instead of inventing synthetic output.
- Test reports are also not guaranteed for every pipeline step. When Bitbucket does not expose test reports for a step, `bb pipeline test-reports` fails clearly instead of inventing synthetic results.

## Notes On Current Behavior

- `bb status` is intentionally bounded. When a workspace scan hits `--repo-limit`, an item section hits `--limit`, or issue tracking is disabled on some repositories, the output includes notes telling you to continue with `bb pr list --repo <workspace>/<repo>` or `bb issue list --repo <workspace>/<repo>`.
- `bb browse` defaults to opening the browser. Use `--no-browser` for deterministic printing, automation, and manual smoke tests.
- Project permission mutation stays out of scope for now. `bb` only exposes explicit project permission inspection until the API-token path is verified cleanly enough to support a documented write workflow.
- Bitbucket Cloud exposes a workspace code-search endpoint in the official REST API, but on the verified full-access API-token path it currently returns `Search is not enabled for the requested account`. `bb` keeps code search out of scope until Atlassian exposes it consistently enough to support as a documented workflow.
- Bitbucket Cloud repository downloads stay out of scope for now. On the verified API-token path in the current fixture workspace, the official downloads endpoint returns `402 Payment Required` because that workspace plan does not support downloads.
- Bitbucket Cloud snippets also stay out of scope for now. On the verified API-token path in the current fixture workspace, the official snippets endpoint returns a workspace-plan limitation instead of usable snippet data.
- `bb` intentionally supports API-token login only. Browser login is out of scope unless Bitbucket Cloud exposes a cleaner CLI-safe auth path.
- `bb` does not wrap Bitbucket issue import or export jobs today. Atlassian documents those endpoints, but the current Bitbucket Cloud issue import/export endpoints reject API-token auth, so `bb` leaves them out instead of shipping a broken wrapper.
- `bb repo deploy-key` supports list, view, create, and delete. Bitbucket rejected deploy-key updates in the live API behavior we verified, so key rotation should use delete plus create instead of an `edit` command.
- `bb repo permissions` currently covers read-only inspection of explicit user and group permissions. Bitbucket’s permission write/delete docs still describe app-password-only behavior in places, so `bb` leaves repository permission mutation out of scope until the API-token path is verified live.
- `bb commit report` currently covers report inspection only. `bb` does not offer commit report creation, update, or deletion until the API-token path is verified cleanly enough to expose as a supported CLI workflow.
- `bb config` exposes the keys that affect runtime today: `prompt`, `browser`, and `output.format`. Editor and pager configuration are still not wired up.
- Alias expansion preserves shell-style quoting so aliases like `bb alias set ship 'pr create --title "Add feature"'` expand reliably for both humans and automation.
- Live Bitbucket integration tests and human-output smoke tests are manual-only. They are never part of `go test ./...` or CI.
