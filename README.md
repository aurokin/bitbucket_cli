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
bb browse --repo workspace-slug/repo-slug --no-browser
bb repo view --repo workspace-slug/repo-slug
bb pipeline list --repo workspace-slug/pipelines-repo-slug
bb pr list --repo workspace-slug/repo-slug
bb issue list --repo workspace-slug/issues-repo-slug
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
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr list --repo workspace-slug/repo-slug --json id,title,state,task_count,comment_count
bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug --json pipeline,steps
bb search prs fixture --repo workspace-slug/repo-slug --jq '.[] | .id'
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

- auth, api, browse, and resolve
- repo, pipeline, pr, and issue
- search, status, config, alias, and extension

## Compared With `gh`

### What `gh` Offers That `bb` Also Offers

- Authenticated API access through `gh api` / `bb api`
- Browser navigation through `gh browse` / `bb browse`
- Repository inspection, creation, cloning, and deletion
- Pipeline run listing and inspection
- Pull request listing, review, status, activity, commit inspection, viewing, diffing, commenting, creation, checkout, merge, and close flows
- Issue listing, viewing, creation, editing, and state transitions
- Cross-repository status summaries
- Search for repositories, pull requests, and issues
- Config defaults for prompt behavior and default output format, plus aliases and extension discovery
- Structured automation paths with `--json`
- Flexible repository targeting with local git inference, `workspace/repo`, and Bitbucket/GitHub-style URLs

### What `bb` Offers That `gh` Does Not

- Native Bitbucket Cloud repository and pull request workflows
- Bitbucket project-aware repository creation through `--project-key`
- Explicit documentation of Bitbucket Cloud platform limits when parity is impossible

### What `gh` Offers That `bb` Does Not

- Broader auth account management
- Releases and broader CI/workflow management such as dispatching, rerunning, and log-heavy workflow tooling
- Richer repository administration such as list, edit, rename, fork, archive, and sync
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

- `bb` supports the Bitbucket Cloud pipeline APIs we could verify directly today: list, view, log, and stop.
- `bb` does not provide pipeline rerun because the current Bitbucket Cloud pipeline REST docs do not expose a rerun endpoint. `bb` does not fake rerun by creating a new run behind your back.
- Raw step logs are not guaranteed for every Bitbucket pipeline step. When Bitbucket does not expose a log file for a step, `bb pipeline log` fails clearly instead of inventing synthetic output.

## Notes On Current Behavior

- `bb status` is intentionally bounded. When a workspace scan hits `--repo-limit`, an item section hits `--limit`, or issue tracking is disabled on some repositories, the output includes notes telling you to continue with `bb pr list --repo <workspace>/<repo>` or `bb issue list --repo <workspace>/<repo>`.
- `bb browse` defaults to opening the browser. Use `--no-browser` for deterministic printing, automation, and manual smoke tests.
- `bb` intentionally supports API-token login only. Browser login is out of scope unless Bitbucket Cloud exposes a cleaner CLI-safe auth path.
- `bb config` exposes the keys that affect runtime today: `prompt`, `browser`, and `output.format`. Editor and pager configuration are still not wired up.
- Alias expansion preserves shell-style quoting so aliases like `bb alias set ship 'pr create --title "Add feature"'` expand reliably for both humans and automation.
- Live Bitbucket integration tests and human-output smoke tests are manual-only. They are never part of `go test ./...` or CI.
