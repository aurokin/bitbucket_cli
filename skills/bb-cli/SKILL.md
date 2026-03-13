---
name: bb-cli
description: Bitbucket Cloud CLI reference and workflow guide for this repository's `bb` command. Use when a task mentions Bitbucket, `bb`, Bitbucket Cloud workflows, or Bitbucket web URLs such as `https://bitbucket.org/...`, including pull request URLs and pull request comment links. Also use when operating, extending, documenting, testing, or troubleshooting `bb`, including installation, API-token authentication, repository targeting, structured output, pull requests, issues, pipelines, browse flows, generated docs, and Bitbucket Cloud REST API alignment.
---

# bb CLI

Use this skill when work centers on this repository's `bb` Bitbucket Cloud CLI, when a request explicitly mentions Bitbucket or Bitbucket Cloud, or when the task includes a Bitbucket web URL such as a repository, pull request, pull request comment, issue, commit, or source link.

## Quick Start

- Install or update with `go install github.com/aurokin/bitbucket_cli/cmd/bb@latest`.
- Install the reusable skill with `npx skills add https://github.com/aurokin/bitbucket_cli --skill bb-cli`.
- Authenticate with `bb auth login`.
- Validate auth with `bb auth status --check`.
- Run local builds with `go build ./cmd/bb`.
- Run local commands with `go run ./cmd/bb ...`.

## Operating Rules

- Prefer real Bitbucket Cloud behavior over `gh` parity.
- If Bitbucket Cloud does not support a feature, do not ship a misleading approximation by default.
- Keep authentication API-token-first. Do not add browser login unless Bitbucket Cloud exposes a clean CLI-safe flow and the docs are updated.
- Prefer `--repo <workspace>/<repo>` over local inference.
- Use `--workspace` only for disambiguation.
- When a task starts from a Bitbucket web URL, prefer `bb resolve <url>` before choosing the follow-up command.
- Preserve both human and agent paths: `--json`, `--jq`, and `--no-prompt`.
- Keep live Bitbucket integration tests manual-only. Do not add them to `go test ./...` or CI.
- Reuse existing Bitbucket fixture repos when they already exist.
- Reuse or create sacrificial fixtures for destructive flows.
- Push after every commit when working in this repository.

## Pick The Right Reference

Read the smallest relevant doc first.

- Product overview and installation: [README.md](../../README.md)
- Human task flows: [docs/workflows.md](../../docs/workflows.md)
- Agent and automation patterns: [docs/automation.md](../../docs/automation.md)
- Full command surface: [docs/cli-reference.md](../../docs/cli-reference.md)
- Flag summary: [docs/flag-matrix.md](../../docs/flag-matrix.md)
- JSON field selection: [docs/json-fields.md](../../docs/json-fields.md)
- JSON payload shapes: [docs/json-shapes.md](../../docs/json-shapes.md)
- Recovery guidance: [docs/recovery.md](../../docs/recovery.md)
- Error catalog: [docs/error-index.md](../../docs/error-index.md)
- Repo conventions: [AGENTS.md](../../AGENTS.md)

## Human Workflow

- Prefer header-first output and `Next:` guidance in command responses.
- Prefer `bb auth login` for interactive token entry.
- Prefer local repo inference only when already inside the intended checkout.
- If a human asks "how do I install or update this", point them to the documented `go install` flow.
- If a command lacks Bitbucket support, explain the limitation clearly and link the official Bitbucket docs when relevant.

## Agent Workflow

- Prefer explicit commands:

```bash
bb --no-prompt <command> --repo workspace-slug/repo-slug --json ...
bb ... --jq '...'
```

- Do not parse human-readable output when structured output is available.
- Use `bb api` when the wrapped command surface does not cover an official endpoint yet.
- If output can be large, prefer `--jq` to return only the needed fields.
- If a command may prompt, force `--no-prompt`.
- If the task begins from a pasted Bitbucket URL, resolve it first:

```bash
bb resolve <url> --json '*'
```

## Implementation Workflow

When changing the CLI:

1. Read the relevant command file and the smallest matching doc.
2. Keep behavior grounded in the official Bitbucket Cloud REST API.
3. Update human and agent docs when behavior, flags, examples, payloads, or recovery guidance change.
4. Regenerate generated docs with `go run ./cmd/gen-docs`.
5. Run `go test ./...`.
6. Keep live integration tests manual-only.
7. Push after each commit.

When creating or changing commands:

- Prefer shared target resolution and shared output helpers over one-off parsing.
- Prefer repo context in human output.
- Prefer deterministic JSON for agent-facing flows.
- Add warnings when local inference degrades instead of silently falling back.
- When changing human-readable output, add regression tests for field order and `Next:` guidance.
- When changing URL or entity resolution, add regression coverage for messy Bitbucket URLs and canonical URL behavior.
- When changing user-facing output or URL handling, run the relevant manual Bitbucket smoke before closing the task.

## Bitbucket API Grounding

When behavior is unclear, use the official Bitbucket Cloud REST API docs first:

- Overview: https://developer.atlassian.com/cloud/bitbucket/about-bitbucket-cloud-rest-api/
- REST intro: https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- OpenAPI document: https://api.bitbucket.org/swagger.json
- Repositories: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/
- Pull requests: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- Pipelines: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/
- Issue tracker: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-issue-tracker/
- Users: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/
- Source: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-source/

## Validation

- Validate the skill itself with `python3 /home/auro/.agents/skills/skill-creator/scripts/quick_validate.py skills/bb-cli`.
- Validate CLI docs with `go run ./cmd/gen-docs`.
- Validate code with `go test ./...`.
- Validate real user-facing output with the manual smoke when relevant:

```bash
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudHumanOutputSmoke -v
```

## Common Patterns

Deterministic repo-scoped operation:

```bash
bb repo view --repo workspace-slug/repo-slug --json name,main_branch,html_url
bb pr list --repo workspace-slug/repo-slug --json id,title,state
bb pipeline list --repo workspace-slug/pipelines-repo-slug --json build_number,state
```

Raw API fallback:

```bash
bb api /user --jq '{display_name, account_id}'
```
