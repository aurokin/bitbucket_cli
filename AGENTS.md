# Agent Instructions

## Package Manager
- Use Go tooling: `go test ./...`, `go run ./cmd/bb`, `go build ./cmd/bb`
- Use the committed lint config with `golangci-lint run ./...` when doing refactor, testing, or reliability passes.
- Use the repo-local gate for regular local verification: `make check`
- Use `make coverage` for coverage-guided reliability passes before adding broad new tests.
- Use `make race` for broader validation when touching shared state, hooks, or command execution paths.
- Use `make fuzz-short` when changing parser, selector, URL, or alias splitting behavior.
- Use `make stability` when refactors touch shared helpers or tests that may hide order dependence or accidental coupling.
- Bootstrap pinned local dev tools with `make tools`

## Commit Attribution
- AI commits MUST include:
```text
Co-Authored-By: Codex by OpenAI <codex@openai.com>
```
- Push after every commit: `git push origin master`

## Key Conventions
- Prefer real Bitbucket Cloud behavior over fake `gh` parity.
- If Bitbucket Cloud cannot support a feature, do not ship a misleading approximation by default.
- Document unsupported or impossible features in `README.md`, command help, and `ROADMAP.md` when relevant.
- Keep authentication API-token-first. Do not add browser login unless Bitbucket Cloud exposes a clean CLI-safe flow and the docs are updated accordingly.
- Prefer the shared `--repo` target model. Use `--workspace` only for disambiguation.
- Preserve human and agent paths: `--json`, `--jq`, and `--no-prompt`.

## Deterministic Command Shape
- Prefer `--repo <workspace>/<repo>` over local inference.
- When a task starts from a Bitbucket web URL, prefer `bb resolve <url> --json '*'` before inferring the entity manually.
- Use `--json` or `--json '*'` for machine parsing.
- Use `--jq` when a smaller payload is enough.
- Use `--no-prompt` for mutations and non-interactive runs.
- Do not parse human-readable output when JSON is available.

## Fixtures And Live Tests
- Keep live Bitbucket integration tests manual-only. Do not add them to `go test ./...` or CI.
- Reuse the existing Bitbucket fixture project and repositories when they already exist.
- Reuse or create sacrificial fixtures for destructive flows. Do not delete the primary fixtures by default.
- When changing human-readable output, add regression tests for field order and `Next:` guidance, then run the relevant manual Bitbucket smoke when the path is user-facing.
- When changing URL or entity resolution, add regression coverage for messy Bitbucket URLs and canonical URL behavior.

## Documentation
- Keep `README.md` task-oriented for humans.
- Keep `AGENTS.md` concise and deterministic for automation.
- Treat generated or repo-local skills as consumer-facing artifacts. Do not put repo-maintainer workflow instructions into a skill when they belong in `AGENTS.md`, tests, or repo docs.
- Keep the repo-local skill installable via `npx skills add https://github.com/aurokin/bitbucket_cli --skill bb-cli`.
- Keep documentation grounded in the official Bitbucket Cloud REST API. Prefer official Atlassian API links over secondary summaries.
- When documenting implementation behavior, explain which Bitbucket Cloud REST API groups back the command family and link the official docs.
- When a command has a safe next step, prefer documenting it explicitly in help text and human-readable output.
- Keep [docs/workflows.md](./docs/workflows.md) task-oriented for humans.
- Keep [docs/automation.md](./docs/automation.md) deterministic for agents.
- Keep [docs/flag-matrix.md](./docs/flag-matrix.md) generated from Cobra flags.
- Keep [docs/error-index.md](./docs/error-index.md) generated from the recovery guidance catalog.
- Keep [docs/json-fields.md](./docs/json-fields.md) generated from the current payload structs and response models.
- Keep [docs/json-shapes.md](./docs/json-shapes.md) generated from the current payload structs and JSON-centric usage.
- Keep [docs/recovery.md](./docs/recovery.md) generated from the current recovery guidance catalog.
- Regenerate generated docs with `go run ./cmd/gen-docs` when command help, flags, payload shapes, or recovery guidance change.

## Bitbucket API References
- Overview: https://developer.atlassian.com/cloud/bitbucket/about-bitbucket-cloud-rest-api/
- REST intro: https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- OpenAPI document: https://api.bitbucket.org/swagger.json
- Repositories: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/
- Pull requests: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- Pipelines: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pipelines/
- Issue tracker: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-issue-tracker/
- Users: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/

## Local Skills
- No repo-local skills discovered under `.claude/skills` or `plugins/*/skills`.
