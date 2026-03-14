# Changelog

## v0.2.0 - 2026-03-14

### Added

- Added Bitbucket URL resolution with `bb resolve`, plus broader support for PR URLs and PR comment URLs across PR commands.
- Added PR comment management and PR task management, including comment-thread resolve/reopen and comment-linked tasks.
- Added PR review and inspection commands: review actions, activity, commits, and checks.
- Expanded pipeline support with run, schedules, runners, caches, variables, test reports, log inspection, and stop behavior.
- Expanded issue support with issue comments, attachments, milestones, and components.
- Added repository administration commands for listing, editing, forking, webhooks, deploy keys, and permission inspection.
- Added commit, branch, tag, workspace, project, deployment, and deployment-environment command families.
- Added the `bb-cli` consumer skill and kept it aligned with the CLI’s supported Bitbucket Cloud workflows.

### Reliability And Testing

- Added a repo-local quality gate with linting, complexity checks, race runs, fuzzing, coverage, and stability targets.
- Expanded unit coverage across command helpers, selectors, summaries, shell parsing, git wrappers, and payload builders.
- Expanded manual Bitbucket Cloud integration coverage across PRs, issues, pipelines, repositories, commits, deployments, workspaces, projects, URL resolution, and human-readable output.
- Hardened pipeline stop integration behavior around Bitbucket’s terminal-state races.
- Hardened commit report integration behavior around Bitbucket’s eventual consistency.

### Refactoring

- Split large command files across PR, pipeline, deployment, issue, repo, workspace, project, status, auth, and selector families.
- Extracted shared resolver, rendering, warning, variable, and summary helpers to reduce duplication and drift.
- Added a maintained lint baseline and reduced command complexity hotspots surfaced by complexity tooling.

### Docs

- Added generated docs for CLI reference, flag matrix, JSON fields, JSON shapes, recovery guidance, and error index.
- Kept README, workflow docs, automation docs, and skill docs aligned with verified Bitbucket Cloud API-token behavior.
- Documented platform limits clearly where Bitbucket Cloud or API-token auth does not support a workflow cleanly.
