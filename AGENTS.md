# Agent Instructions

## Package Manager
- Use Go tooling: `go test ./...`, `go run ./cmd/bb`, `go build ./cmd/bb`

## Commit Attribution
- AI commits MUST include:
```text
Co-Authored-By: Codex by OpenAI <codex@openai.com>
```

## Key Conventions
- Prefer real Bitbucket Cloud behavior over fake `gh` parity.
- If Bitbucket Cloud cannot support a feature, do not ship a misleading approximation by default.
- Document unsupported or impossible features in `README.md`, command help, and `ROADMAP.md` when relevant.
- Prefer the shared `--repo` target model. Use `--workspace` only for disambiguation.
- Preserve human and agent paths: `--json`, `--jq`, and `--no-prompt`.

## Deterministic Command Shape
- Prefer `--repo <workspace>/<repo>` over local inference.
- Use `--json` or `--json '*'` for machine parsing.
- Use `--jq` when a smaller payload is enough.
- Use `--no-prompt` for mutations and non-interactive runs.
- Do not parse human-readable output when JSON is available.

## Fixtures And Live Tests
- Keep live Bitbucket integration tests manual-only. Do not add them to `go test ./...` or CI.
- Reuse the existing Bitbucket fixture project and repositories when they already exist.
- Reuse or create sacrificial fixtures for destructive flows. Do not delete the primary fixtures by default.

## Documentation
- Keep `README.md` task-oriented for humans.
- Keep `AGENTS.md` concise and deterministic for automation.
- When a command has a safe next step, prefer documenting it explicitly in help text and human-readable output.
- Keep [docs/workflows.md](./docs/workflows.md) task-oriented for humans.
- Keep [docs/automation.md](./docs/automation.md) deterministic for agents.
- Keep [docs/flag-matrix.md](./docs/flag-matrix.md) generated from Cobra flags.
- Keep [docs/error-index.md](./docs/error-index.md) generated from the recovery guidance catalog.
- Keep [docs/json-fields.md](./docs/json-fields.md) generated from the current payload structs and response models.
- Keep [docs/json-shapes.md](./docs/json-shapes.md) generated from the current payload structs and JSON-centric usage.
- Keep [docs/recovery.md](./docs/recovery.md) generated from the current recovery guidance catalog.
- Regenerate generated docs with `go run ./cmd/gen-docs` when command help, flags, payload shapes, or recovery guidance change.

## Local Skills
- No repo-local skills discovered under `.claude/skills` or `plugins/*/skills`.
