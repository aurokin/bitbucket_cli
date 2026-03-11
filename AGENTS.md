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
- Keep live Bitbucket integration tests manual-only. Do not add them to `go test ./...` or CI.
- Reuse the existing Bitbucket fixture project and repositories when they already exist.
- Prefer the shared `--repo` target model. Use `--workspace` only for disambiguation.
- Preserve human and agent paths: `--json`, `--jq`, and `--no-prompt`.

## Local Skills
- No repo-local skills discovered under `.claude/skills` or `plugins/*/skills`.
