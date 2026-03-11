# Roadmap

## Next Commands

- `bb repo delete`
  Delete repositories explicitly, but never as part of the default integration flow.

- `bb pr status`
  Show relevant pull request status for the current repository or explicit target.

- `bb pr diff`
  Show pull request diffs in a form that works well for humans and agents.

- `bb pr comment`
  Add pull request comments from flags, stdin, or editor-driven flows.

- `bb pr close`
  Close pull requests explicitly without merging them.

- `bb pr reopen`
  Reopen closed pull requests.

- `bb browse`
  Open or print repository and pull request URLs for browser handoff.

## Follow-Up UX Work

- Unified repo selector
  Add a shared `--repo <workspace>/<repo>` style selector across commands.

- More compact human-readable table output for wide terminals
- Better remediation hints for common auth and repo-resolution failures

## Later Phase

- Browser login
  Add OAuth-based browser auth after the current API-token-first command set is more complete.

- Pipelines support
  Add Bitbucket Pipelines equivalents for listing runs, viewing results, and rerunning workflows.

- Issues support
  Add issue listing, viewing, creation, and state transitions.

- Config, aliases, and extensibility
  Add persistent CLI config, user-defined shortcuts, and an extension model for agent-heavy workflows.
