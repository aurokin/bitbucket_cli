# Workflows

Task-oriented command recipes for humans using `bb`.

Use the generated [CLI reference](./cli-reference.md) for the full command surface.

## Authenticate

```bash
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
bb auth status --check
```

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
bb browse --repo OhBizzle/bb-cli-integration-primary --no-browser
bb repo view --repo OhBizzle/bb-cli-integration-primary
bb repo clone OhBizzle/bb-cli-integration-primary
```

## Inspect Pipelines

```bash
bb pipeline list --repo OhBizzle/bb-cli-integration-pipelines
bb pipeline view 1 --repo OhBizzle/bb-cli-integration-pipelines
bb pipeline log 1 --repo OhBizzle/bb-cli-integration-pipelines --step '{step-uuid}'
```

## Review A Pull Request

```bash
bb pr list --repo OhBizzle/bb-cli-integration-primary
bb pr view 1 --repo OhBizzle/bb-cli-integration-primary
bb pr diff 1 --repo OhBizzle/bb-cli-integration-primary --stat
bb pr comment 1 --repo OhBizzle/bb-cli-integration-primary --body "Looks good overall. Please tighten the error handling."
```

## Create And Land A Pull Request

From a local checkout:

```bash
bb pr create --source feature --destination main --title "Add feature"
bb pr view 2
bb pr merge 2 --repo OhBizzle/bb-cli-integration-primary
```

For a deterministic non-interactive flow:

```bash
bb --no-prompt pr create \
  --repo OhBizzle/bb-cli-integration-primary \
  --source feature \
  --destination main \
  --title "Add feature"
```

## Triage Issues

```bash
bb issue list --repo OhBizzle/bb-cli-integration-issues
bb issue create --repo OhBizzle/bb-cli-integration-issues --title "Broken flow" --body "Needs investigation."
bb issue view 1 --repo OhBizzle/bb-cli-integration-issues
bb issue edit 1 --repo OhBizzle/bb-cli-integration-issues --priority major
bb issue close 1 --repo OhBizzle/bb-cli-integration-issues --message "Fixed in main."
```

## Search For Work

```bash
bb search repos bb-cli --workspace OhBizzle
bb search prs fixture --repo OhBizzle/bb-cli-integration-primary
bb search issues broken --repo OhBizzle/bb-cli-integration-issues
bb status --workspace OhBizzle --limit 10
```

## Reuse Fixtures Safely

The live Bitbucket integration tests intentionally reuse the existing fixture project and repositories where possible. For destructive flows like repository deletion, use sacrificial fixtures instead of deleting the primary repositories.
