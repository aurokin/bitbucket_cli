# Product Plan: `bb` Bitbucket CLI

## Goal

Build a Bitbucket CLI that feels closer to `gh` than a thin API wrapper.

The product has two primary user types from day one:

- Humans working interactively in a terminal
- Agents and scripts that need stable, machine-readable output

That means structured output is not an enhancement. It is core product surface area.

## Product Principles

1. Match the best parts of `gh`
   - Clear subcommands
   - Strong repo inference from local git remotes
   - Good defaults for interactive use
   - Consistent help text and exit codes

2. Treat automation as a first-class use case
   - `--json` available in the first pass for key commands
   - `--jq` available in the first pass for client-side filtering
   - Stable field names and predictable error output
   - Minimal noise on stdout when structured output is requested

3. Start narrow, but design for expansion
   - Target Bitbucket Cloud first
   - Keep provider boundaries clean enough to add Bitbucket Server/Data Center later

4. Prefer one static binary
   - Use Go for distribution, cross-platform support, and CLI ergonomics

## Scope Decision

### Phase 1 Target

Bitbucket Cloud only.

Reason:

- Faster path to a usable product
- Simpler auth and API model
- Avoid early abstraction cost across Cloud and Server

### Future Compatibility

Design the service layer so a second provider can be added later without rewriting command logic.

## MVP Command Set

### Authentication

- `bb auth login`
- `bb auth status`
- `bb auth logout`

### Repository

- `bb repo view`

### Pull Requests

- `bb pr list`
- `bb pr view`
- `bb pr create`
- `bb pr checkout`

### Raw API Access

- `bb api`

This is enough to validate the product in real workflows without overreaching into issues, pipelines, or workspace admin on the first pass.

## First-Pass Output Requirements

Structured output must ship with the first usable release.

### Required Behavior

- Read-oriented commands support `--json`
- Read-oriented commands support `--jq`
- JSON output goes to stdout with no extra decoration
- Human-oriented errors go to stderr
- Non-zero exit codes are used consistently

### Initial `--json` Coverage

- `bb auth status --json`
- `bb repo view --json`
- `bb pr list --json`
- `bb pr view --json`
- `bb api` returns raw response data by default, with flags for response shaping later

### `--jq` Notes

- Implement client-side filtering against JSON output
- Support simple field projection and array filtering first
- Avoid inventing a second query language
- If behavior differs from `gh`, document it explicitly

## User Experience

### Human-Focused

- Infer workspace/repo from git remotes when inside a repository
- Render concise tables by default for list commands
- Provide interactive prompts for missing values in `pr create`
- Open browser when that is the simplest auth path

### Agent-Focused

- Every automation-friendly command documents its JSON schema
- Output field names remain stable across patch releases
- Commands can run non-interactively without hidden prompts when flags are supplied
- Errors include machine-usable codes where practical

## Technical Architecture

### Language and Libraries

- Go
- `cobra` for command structure
- `viper` only if config complexity justifies it; otherwise keep config handling minimal
- `gojq` or equivalent for `--jq`

### Package Layout

- `cmd/bb`
  - CLI entrypoint and command registration
- `internal/config`
  - Host config, auth state, output defaults
- `internal/auth`
  - Login flow, token management, secure storage integration
- `internal/bitbucket`
  - Cloud API client and typed services
- `internal/git`
  - Remote parsing, branch detection, repo inference
- `internal/output`
  - Tables, JSON rendering, stderr/stdout discipline, tty handling
- `internal/pr`
  - Pull request workflows and data shaping

### Core Design Rules

- Commands should be thin
- API client logic should not know about terminal rendering
- JSON field selection should happen from typed response models, not ad hoc maps where avoidable
- Command behavior should be testable without live API access

## Authentication Plan

### First Pass

Support app password or token-based auth for Bitbucket Cloud if that is the fastest reliable path.

If OAuth or browser-based login is feasible without major friction, prefer it for human UX, but do not block MVP on a more complex auth flow.

### Requirements

- Multi-host aware config model
- Secure token storage where practical
- Clear `auth status` introspection

## Pull Request Workflow Plan

### `bb pr list`

- Defaults to current repo
- Filters by state and author in later iterations
- Table output by default
- `--json` and `--jq` in first pass

### `bb pr view`

- Accept PR ID or infer from current branch later
- Show concise summary for humans
- Full structured representation for automation

### `bb pr create`

- Source branch defaults to current branch
- Destination defaults to repository main branch when discoverable
- Title/body can come from flags, prompt, or editor in later iterations

### `bb pr checkout`

- Fetch PR source branch and switch local branch
- Keep behavior explicit and easy to reason about

## Milestones

### Milestone 0: Product Skeleton

- Initialize Go module
- Set up CLI root command
- Establish config paths and error model
- Define output interfaces with `--json` and `--jq` in the base command plumbing

### Milestone 1: Auth and Repo Resolution

- Implement `auth login`, `auth status`, `auth logout`
- Parse git remotes and infer Bitbucket Cloud repo context

### Milestone 2: Read Commands

- Implement `repo view`
- Implement `pr list`
- Implement `pr view`
- Implement `api`
- Ship first-pass JSON schemas and jq support

### Milestone 3: Write Commands

- Implement `pr create`
- Implement `pr checkout`

### Milestone 4: Polish

- Shell completion
- Pager support where useful
- Better help output
- More filtering and field selection

## Testing Strategy

### Unit Tests

- Remote URL parsing
- Repo inference
- JSON rendering
- `--jq` behavior
- Config loading and host selection

### Integration Tests

- Mock Bitbucket API responses
- Verify stdout/stderr split
- Verify exit codes
- Verify non-interactive command behavior

### Manual Validation

- Test against a real Bitbucket Cloud workspace before widening scope

## Risks

1. Bitbucket auth may force an early decision between ideal UX and fastest delivery.
2. Bitbucket Cloud API shapes may not map cleanly to a stable, compact CLI JSON schema.
3. Repo inference will need to handle SSH and HTTPS remotes consistently.
4. `--jq` introduces real product expectations early, so output contracts must be disciplined.

## Non-Goals For MVP

- Full parity with `gh`
- Bitbucket Server/Data Center support
- Pipelines, issues, snippets, and workspace administration
- Plugin architecture

## Immediate Next Steps

1. Initialize the repository and commit this plan.
2. Create the Go module and root command skeleton.
3. Define the shared output contract for table, JSON, and jq-filtered responses.
4. Implement repo inference before network-heavy commands.
