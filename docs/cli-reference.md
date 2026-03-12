# CLI Reference

Generated from the `bb` command tree.

Use this file for the full command surface. Keep [README.md](../README.md) focused on workflows.

## Command Tree

- `bb alias`
  - `bb alias delete`
  - `bb alias get`
  - `bb alias list`
  - `bb alias set`
- `bb api`
- `bb auth`
  - `bb auth login`
  - `bb auth logout`
  - `bb auth status`
- `bb config`
  - `bb config get`
  - `bb config list`
  - `bb config path`
  - `bb config set`
  - `bb config unset`
- `bb extension`
  - `bb extension exec`
  - `bb extension list`
- `bb issue`
  - `bb issue close`
  - `bb issue create`
  - `bb issue edit`
  - `bb issue list`
  - `bb issue reopen`
  - `bb issue view`
- `bb pipeline`
  - `bb pipeline list`
  - `bb pipeline view`
- `bb pr`
  - `bb pr checkout`
  - `bb pr close`
  - `bb pr comment`
  - `bb pr create`
  - `bb pr diff`
  - `bb pr list`
  - `bb pr merge`
  - `bb pr status`
  - `bb pr view`
- `bb repo`
  - `bb repo clone`
  - `bb repo create`
  - `bb repo delete`
  - `bb repo view`
- `bb search`
  - `bb search issues`
  - `bb search prs`
  - `bb search repos`
- `bb status`
- `bb version`

## `bb alias`

Manage command aliases

Manage persistent command aliases stored in the bb config file.

Usage:

```text
bb alias
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb alias delete`: Delete an alias
- `bb alias get`: Show one configured alias
- `bb alias list`: List configured aliases
- `bb alias set`: Create or replace an alias

## `bb alias delete`

Delete an alias

Aliases: `remove`, `rm`

Usage:

```text
bb alias delete <name>
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb alias get`

Show one configured alias

Usage:

```text
bb alias get <name>
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb alias list`

List configured aliases

Usage:

```text
bb alias list [flags]
```

Flags:

- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb alias set`

Create or replace an alias

Usage:

```text
bb alias set <name> <expansion...>
```

Examples:

```bash
bb alias set pv 'pr view'
bb alias set rls 'pr list --state OPEN'
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb api`

Make an authenticated Bitbucket API request

Make an authenticated Bitbucket Cloud API request. Use this for workflows that are not yet covered by a dedicated bb command.

Usage:

```text
bb api <path-or-url> [flags]
```

Examples:

```bash
bb api /user
bb api '/repositories/OhBizzle/bb-cli-integration-primary/pullrequests?state=OPEN'
bb api /user --jq .display_name
printf '{"name":"my-repo"}' | bb api /repositories/OhBizzle/my-repo -X POST --input -
```

Flags:

- `--host`: Bitbucket host to use
- `--input`: Read request body from a file, or '-' for stdin
- `--jq`: Filter JSON output using a jq expression
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `-X`, `--method`: HTTP method

## `bb auth`

Manage authentication

Manage Bitbucket Cloud authentication using Atlassian API tokens.

Aliases: `login-manager`

Usage:

```text
bb auth
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb auth login`: Store credentials for a Bitbucket host
- `bb auth logout`: Remove stored credentials for a Bitbucket host
- `bb auth status`: Show stored authentication status

## `bb auth login`

Store credentials for a Bitbucket host

Store an Atlassian API token for Bitbucket Cloud. The username should be your Atlassian account email.

Usage:

```text
bb auth login [flags]
```

Examples:

```bash
bb auth login --username you@example.com --with-token
bb auth login --username you@example.com --token $BITBUCKET_TOKEN
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
```

Flags:

- `--default`: Set this host as the default
- `--host`: Bitbucket host to configure
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--token`: Atlassian API token to store
- `--username`: Atlassian account email associated with the API token
- `--with-token`: Read the API token from stdin

## `bb auth logout`

Remove stored credentials for a Bitbucket host

Usage:

```text
bb auth logout [flags]
```

Examples:

```bash
bb auth logout
bb auth logout --host bitbucket.org
```

Flags:

- `--host`: Bitbucket host to log out from
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb auth status`

Show stored authentication status

Usage:

```text
bb auth status [flags]
```

Examples:

```bash
bb auth status
bb auth status --check --json
bb auth status --host bitbucket.org
```

Flags:

- `--check`: Validate stored credentials with the Bitbucket API
- `--host`: Only show status for a specific host
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb config`

Manage bb behavior defaults

Manage persistent bb settings that affect runtime behavior today, such as prompt behavior and the default output format.

Usage:

```text
bb config
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb config get`: Get one effective bb setting
- `bb config list`: List effective bb settings
- `bb config path`: Show the bb config file path
- `bb config set`: Set a persistent bb setting
- `bb config unset`: Unset a persistent bb setting

## `bb config get`

Get one effective bb setting

Usage:

```text
bb config get <key> [flags]
```

Examples:

```bash
bb config get prompt
bb config get output.format --json
```

Flags:

- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb config list`

List effective bb settings

Usage:

```text
bb config list [flags]
```

Examples:

```bash
bb config list
bb config list --json
```

Flags:

- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb config path`

Show the bb config file path

Usage:

```text
bb config path
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb config set`

Set a persistent bb setting

Usage:

```text
bb config set <key> <value>
```

Examples:

```bash
bb config set prompt false
bb config set output.format json
bb config get output.format
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb config unset`

Unset a persistent bb setting

Usage:

```text
bb config unset <key>
```

Examples:

```bash
bb config unset prompt
bb config unset output.format
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb extension`

Discover and run external bb commands

Discover and run external commands named bb-<name> from PATH.

Usage:

```text
bb extension
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb extension exec`: Run an external bb command
- `bb extension list`: List discovered external bb commands

## `bb extension exec`

Run an external bb command

Usage:

```text
bb extension exec <name> [args...]
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb extension list`

List discovered external bb commands

Usage:

```text
bb extension list [flags]
```

Flags:

- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

## `bb issue`

Work with repository issues

List, view, create, edit, close, and reopen Bitbucket Cloud repository issues.

Usage:

```text
bb issue
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb issue close`: Close an issue
- `bb issue create`: Create an issue
- `bb issue edit`: Edit an issue
- `bb issue list`: List issues for a repository
- `bb issue reopen`: Reopen an issue
- `bb issue view`: View one issue

## `bb issue close`

Close an issue

Resolve an issue by moving it to the resolved state.

Usage:

```text
bb issue close <id> [flags]
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--message`: Optional issue change message
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Target issue state
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue create`

Create an issue

Usage:

```text
bb issue create [flags]
```

Examples:

```bash
bb issue create --repo OhBizzle/bb-cli-integration-issues --title 'Broken flow'
bb issue create --repo OhBizzle/bb-cli-integration-primary --title 'Broken flow' --body 'Needs investigation'
bb issue create --title 'Request' --kind proposal --priority major --json
```

Flags:

- `--body`: Issue body text
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--kind`: Issue kind
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--priority`: Issue priority
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--title`: Issue title
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue edit`

Edit an issue

Usage:

```text
bb issue edit <id> [flags]
```

Examples:

```bash
bb issue edit 1 --repo OhBizzle/bb-cli-integration-issues --title 'Updated title'
bb issue edit 1 --repo OhBizzle/bb-cli-integration-primary --state open --priority major --json
```

Flags:

- `--body`: Issue body text
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--kind`: Issue kind
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--priority`: Issue priority
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Issue state
- `--title`: Issue title
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue list`

List issues for a repository

Usage:

```text
bb issue list [flags]
```

Examples:

```bash
bb issue list --repo OhBizzle/bb-cli-integration-issues
bb issue list --repo OhBizzle/bb-cli-integration-primary
bb issue list --state open --json id,title,state
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of issues to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Filter issues by state
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue reopen`

Reopen an issue

Reopen an issue by moving it back to the new state.

Usage:

```text
bb issue reopen <id> [flags]
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--message`: Optional issue change message
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Target issue state
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue view`

View one issue

Usage:

```text
bb issue view <id> [flags]
```

Examples:

```bash
bb issue view 1 --repo OhBizzle/bb-cli-integration-issues
bb issue view 1 --repo OhBizzle/bb-cli-integration-primary --json
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline`

Work with Bitbucket Pipelines runs

List and inspect Bitbucket Pipelines runs for one repository.

Aliases: `pipelines`

Usage:

```text
bb pipeline
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pipeline list`: List pipeline runs for a repository
- `bb pipeline view`: View one pipeline run

## `bb pipeline list`

List pipeline runs for a repository

Usage:

```text
bb pipeline list [flags]
```

Examples:

```bash
bb pipeline list --repo OhBizzle/bb-cli-integration-primary
bb pipeline list --repo OhBizzle/bb-cli-integration-primary --state COMPLETED --json build_number,state,target
bb pipeline list --limit 5
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pipelines to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Filter pipelines by pipeline state name
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline view`

View one pipeline run

Usage:

```text
bb pipeline view <number-or-uuid> [flags]
```

Examples:

```bash
bb pipeline view 42 --repo OhBizzle/bb-cli-integration-primary
bb pipeline view '{uuid}' --repo OhBizzle/bb-cli-integration-primary --json '*'
bb pipeline view 42
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr`

Work with pull requests

List, inspect, create, check out, merge, and summarize Bitbucket pull requests.

Aliases: `pull-request`, `pullrequest`

Usage:

```text
bb pr
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pr checkout`: Check out a pull request locally
- `bb pr close`: Close a pull request without merging it
- `bb pr comment`: Add a comment to a pull request
- `bb pr create`: Create a pull request
- `bb pr diff`: View a pull request diff
- `bb pr list`: List pull requests for a repository
- `bb pr merge`: Merge a pull request
- `bb pr status`: Show pull request status for a repository
- `bb pr view`: View a pull request

## `bb pr checkout`

Check out a pull request locally

Fetch the pull request source branch from the current repository's remote and switch to it locally.

Usage:

```text
bb pr checkout <id-or-url> [flags]
```

Examples:

```bash
bb pr checkout 1
bb pr checkout 1 --repo OhBizzle/bb-cli-integration-primary
bb pr checkout https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1
```

Flags:

- `--host`: Bitbucket host to use
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr close`

Close a pull request without merging it

Close a pull request without merging it. In Bitbucket Cloud this maps to declining the pull request.

Usage:

```text
bb pr close <id-or-url> [flags]
```

Examples:

```bash
bb pr close 1
bb pr close 1 --repo OhBizzle/bb-cli-integration-primary
bb pr close https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1 --json
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr comment`

Add a comment to a pull request

Add a comment to a pull request using --body, --body-file, or --body-file - for stdin. This first pass is intentionally deterministic for agent and script usage.

Usage:

```text
bb pr comment <id-or-url> [flags]
```

Examples:

```bash
bb pr comment 1 --body 'Looks good'
bb pr comment 1 --repo OhBizzle/bb-cli-integration-primary --body-file comment.md
printf 'Ship it\n' | bb pr comment https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1 --body-file - --json
```

Flags:

- `--body-file`: Read the comment body from a file, or '-' for stdin
- `--body`: Comment body text
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr create`

Create a pull request

Create a pull request in Bitbucket Cloud. When run interactively, bb prompts for missing fields. The source branch defaults to the current branch and the destination defaults to the repository main branch.

Usage:

```text
bb pr create [flags]
```

Examples:

```bash
bb pr create --title 'Add feature'
bb pr create --source feature --destination main --description 'Ready for review'
bb pr create --reuse-existing --json
```

Flags:

- `--close-source-branch`: Close the source branch when the pull request is merged
- `--description`: Pull request description
- `--destination`: Destination branch; defaults to the repository main branch
- `--draft`: Create the pull request as a draft
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--reuse-existing`: Return an existing matching open pull request instead of creating a new one
- `--source`: Source branch; defaults to the current git branch
- `--title`: Pull request title; defaults to the source branch name
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr diff`

View a pull request diff

Show the patch for a pull request by default. Use --stat for a concise per-file summary, or --json for structured output that includes both the patch and diff stats.

Usage:

```text
bb pr diff <id-or-url> [flags]
```

Examples:

```bash
bb pr diff 1
bb pr diff 1 --repo OhBizzle/bb-cli-integration-primary --stat
bb pr diff https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1 --json patch,stats
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--stat`: Show a concise per-file diff summary instead of the full patch
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr list`

List pull requests for a repository

Usage:

```text
bb pr list [flags]
```

Examples:

```bash
bb pr list
bb pr list --repo OhBizzle/bb-cli-integration-primary
bb pr list --repo https://bitbucket.org/OhBizzle/bb-cli-integration-primary
bb pr list --state ALL --json id,title,state
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pull requests to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Filter pull requests by state: OPEN, MERGED, DECLINED, SUPERSEDED, or ALL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr merge`

Merge a pull request

Merge an open pull request in Bitbucket Cloud. bb uses the destination branch default merge strategy when Bitbucket exposes one, or falls back to the repository default when Bitbucket does not include strategy metadata on the pull request.

Usage:

```text
bb pr merge <id-or-url> [flags]
```

Examples:

```bash
bb pr merge 7
bb pr merge 7 --repo OhBizzle/bb-cli-integration-primary
bb pr merge 7 --strategy merge_commit
bb pr merge 7 --message 'Ship feature' --close-source-branch --json
```

Flags:

- `--close-source-branch`: Close the source branch when the pull request is merged
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--message`: Merge commit message
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--strategy`: Merge strategy to use; required when Bitbucket does not expose a default
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr status`

Show pull request status for a repository

Show pull request status for one repository, including the current branch pull request when available, open pull requests created by you, and open pull requests requesting your review.

Usage:

```text
bb pr status [flags]
```

Examples:

```bash
bb pr status
bb pr status --repo OhBizzle/bb-cli-integration-primary
bb pr status --json current_branch,created,review_requested
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of open pull requests to inspect for status
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr view`

View a pull request

Usage:

```text
bb pr view <id-or-url> [flags]
```

Examples:

```bash
bb pr view 1
bb pr view 1 --json title,state,source,destination
bb pr view https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb repo`

Work with Bitbucket repositories

Inspect, create, clone, and delete Bitbucket repositories.

Aliases: `repos`, `repository`

Usage:

```text
bb repo
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb repo clone`: Clone a Bitbucket repository locally
- `bb repo create`: Create a repository in Bitbucket Cloud
- `bb repo delete`: Delete a Bitbucket repository
- `bb repo view`: Show repository information

## `bb repo clone`

Clone a Bitbucket repository locally

Clone a Bitbucket repository over HTTPS using the configured API token. Prefer --repo <workspace>/<repo> for explicit targeting; use --workspace only to disambiguate a bare repository name. The origin remote is rewritten after cloning so the token is not stored in git config.

Usage:

```text
bb repo clone [repository] [directory] [flags]
```

Examples:

```bash
bb repo clone OhBizzle/bb-cli-integration-primary
bb repo clone --repo OhBizzle/bb-cli-integration-primary ./tmp/repo
bb repo clone bb-cli-integration-primary --workspace OhBizzle
bb repo clone https://bitbucket.org/OhBizzle/bb-cli-integration-primary
bb repo clone OhBizzle/bb-cli-integration-primary ./tmp/repo --json
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb repo create`

Create a repository in Bitbucket Cloud

Create a repository in Bitbucket Cloud. Prefer --repo <workspace>/<repo> for explicit targeting; use --workspace only to disambiguate a bare repository name. Use --reuse-existing when the command may be run repeatedly.

Usage:

```text
bb repo create [repository] [flags]
```

Examples:

```bash
bb repo create OhBizzle/my-repo --project-key BBCLI
bb repo create --repo OhBizzle/my-repo --reuse-existing --json
bb repo create my-repo --workspace OhBizzle
```

Flags:

- `--description`: Repository description
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--name`: Display name for the repository
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--private`: Create the repository as private
- `--project-key`: Bitbucket project key for the repository
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--reuse-existing`: Return the existing repository instead of failing when it already exists
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb repo delete`

Delete a Bitbucket repository

Delete a Bitbucket repository in Bitbucket Cloud. Prefer --repo <workspace>/<repo> for explicit targeting; use --workspace only to disambiguate a bare repository name. Humans must confirm the exact workspace/repository unless --yes is provided. Scripts and agents should use --yes together with --no-prompt when they need deterministic behavior.

Usage:

```text
bb repo delete [repository] [flags]
```

Examples:

```bash
bb repo delete OhBizzle/bb-cli-delete-command-target --yes
bb repo delete --repo OhBizzle/bb-cli-delete-command-target --yes
bb repo delete bb-cli-delete-command-target --workspace OhBizzle --yes
bb repo delete https://bitbucket.org/OhBizzle/bb-cli-delete-command-target --json
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb repo view`

Show repository information

Show repository information from Bitbucket Cloud. When run inside a git checkout, local remote details are included in the output.

Usage:

```text
bb repo view [flags]
```

Examples:

```bash
bb repo view
bb repo view --repo OhBizzle/bb-cli-integration-primary
bb repo view --repo https://bitbucket.org/OhBizzle/bb-cli-integration-primary
bb repo view --json name,project_key,main_branch
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb search`

Search repositories, pull requests, and issues

Search Bitbucket Cloud repositories, pull requests, and issues using Bitbucket query filters behind the scenes.

Usage:

```text
bb search
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb search issues`: Search issues in one repository
- `bb search prs`: Search pull requests in one repository
- `bb search repos`: Search repositories in a workspace

## `bb search issues`

Search issues in one repository

Usage:

```text
bb search issues <query> [flags]
```

Examples:

```bash
bb search issues fixture --repo OhBizzle/bb-cli-integration-issues
bb search issues bug --repo OhBizzle/bb-cli-integration-issues --json id,title,state
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of issues to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb search prs`

Search pull requests in one repository

Usage:

```text
bb search prs <query> [flags]
```

Examples:

```bash
bb search prs fixture --repo OhBizzle/bb-cli-integration-primary
bb search prs feature --repo OhBizzle/bb-cli-integration-primary --json id,title,state
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pull requests to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb search repos`

Search repositories in a workspace

Usage:

```text
bb search repos <query> [flags]
```

Examples:

```bash
bb search repos integration --workspace OhBizzle
bb search repos bb-cli --workspace OhBizzle --json name,slug,description
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of repositories to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--workspace`: Workspace slug to search; inferred when only one workspace is available

## `bb status`

Show cross-repository pull request and issue status

Show authored pull requests, pull requests requesting your review, and open issues that involve you across accessible repositories.

Usage:

```text
bb status [flags]
```

Examples:

```bash
bb status
bb status --workspace OhBizzle --limit 10
bb status --json authored_prs,review_requested_prs,your_issues
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum items to return per status section
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo-limit`: Maximum repositories to scan per workspace
- `--workspace`: Limit status aggregation to one workspace

## `bb version`

Show bb version information

Usage:

```text
bb version [flags]
```

Examples:

```bash
bb version
bb version --json
```

Flags:

- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
