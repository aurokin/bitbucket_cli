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

## Follow-Up UX Work

- Unified repo selector
  Standardize commands on the shared `--repo` target model so `--workspace` becomes secondary/disambiguation-only instead of a primary path.

- More compact human-readable table output for wide terminals
- Better remediation hints for common auth and repo-resolution failures

## Later Phase

- Browser login
  Add OAuth-based browser auth after the current API-token-first command set is more complete.

- Pipelines support
  Add Bitbucket Pipelines equivalents for listing runs, viewing results, and rerunning workflows.

- Issues support
  Add issue listing, viewing, creation, and state transitions.

- Browse redesign
  Add `bb browse` with `gh`-style default browser behavior while still supporting an easy URL-printing mode for agents and scripts.

- Config
  Add persistent CLI behavior/config management for prompts, editor, browser, pager, output defaults, and other operational settings.

- Aliases and extensibility
  Add user-defined command aliases plus an extension model so humans and agents can adapt workflows without changing core commands.

- Cross-repo status
  Add a `status` view for relevant pull requests, review requests, mentions, and other work across repositories.

- Search
  Add search commands for repositories, pull requests, issues, and other useful Bitbucket resources.
