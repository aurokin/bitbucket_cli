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
- `bb browse`
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
  - `bb issue attachment`
    - `bb issue attachment list`
    - `bb issue attachment upload`
  - `bb issue close`
  - `bb issue comment`
    - `bb issue comment create`
    - `bb issue comment delete`
    - `bb issue comment edit`
    - `bb issue comment list`
    - `bb issue comment view`
  - `bb issue component`
    - `bb issue component list`
    - `bb issue component view`
  - `bb issue create`
  - `bb issue edit`
  - `bb issue list`
  - `bb issue milestone`
    - `bb issue milestone list`
    - `bb issue milestone view`
  - `bb issue reopen`
  - `bb issue view`
- `bb pipeline`
  - `bb pipeline cache`
    - `bb pipeline cache clear`
    - `bb pipeline cache delete`
    - `bb pipeline cache list`
  - `bb pipeline list`
  - `bb pipeline log`
  - `bb pipeline run`
  - `bb pipeline runner`
    - `bb pipeline runner delete`
    - `bb pipeline runner list`
    - `bb pipeline runner view`
  - `bb pipeline schedule`
    - `bb pipeline schedule create`
    - `bb pipeline schedule delete`
    - `bb pipeline schedule disable`
    - `bb pipeline schedule enable`
    - `bb pipeline schedule list`
    - `bb pipeline schedule view`
  - `bb pipeline stop`
  - `bb pipeline test-reports`
  - `bb pipeline variable`
    - `bb pipeline variable create`
    - `bb pipeline variable delete`
    - `bb pipeline variable edit`
    - `bb pipeline variable list`
    - `bb pipeline variable view`
  - `bb pipeline view`
- `bb pr`
  - `bb pr activity`
  - `bb pr checkout`
  - `bb pr checks`
  - `bb pr close`
  - `bb pr comment`
    - `bb pr comment delete`
    - `bb pr comment edit`
    - `bb pr comment reopen`
    - `bb pr comment resolve`
    - `bb pr comment view`
  - `bb pr commits`
  - `bb pr create`
  - `bb pr diff`
  - `bb pr list`
  - `bb pr merge`
  - `bb pr review`
    - `bb pr review approve`
    - `bb pr review clear-request-changes`
    - `bb pr review request-changes`
    - `bb pr review unapprove`
  - `bb pr status`
  - `bb pr task`
    - `bb pr task create`
    - `bb pr task delete`
    - `bb pr task edit`
    - `bb pr task list`
    - `bb pr task reopen`
    - `bb pr task resolve`
    - `bb pr task view`
  - `bb pr view`
- `bb repo`
  - `bb repo clone`
  - `bb repo create`
  - `bb repo delete`
  - `bb repo view`
- `bb resolve`
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
bb api '/repositories/workspace-slug/repo-slug/pullrequests?state=OPEN'
bb api /user --jq .display_name
printf '{"name":"my-repo"}' | bb api /repositories/workspace-slug/my-repo -X POST --input -
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

Store an Atlassian API token for Bitbucket Cloud. The username should be your Atlassian account email. Humans can run `bb auth login` interactively and paste the token securely. Agents can provide `BB_EMAIL` and `BB_TOKEN`, or pass `--username` and `--token` explicitly.

Usage:

```text
bb auth login [flags]
```

Examples:

```bash
bb auth login --username you@example.com --with-token
bb auth login --username you@example.com --token $BITBUCKET_TOKEN
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
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

## `bb browse`

Open or print Bitbucket web URLs

Open Bitbucket repository, pull request, issue, commit, settings, pipelines, and source URLs. Default behavior opens the browser; use --no-browser to print the URL instead.

Usage:

```text
bb browse [target] [flags]
```

Examples:

```bash
bb browse --repo workspace-slug/repo-slug
bb browse README.md:12 --repo workspace-slug/repo-slug --no-browser
bb browse --pr 1 --repo workspace-slug/repo-slug
bb browse --pipelines --repo workspace-slug/repo-slug --json '*'
bb browse a1b2c3d --repo workspace-slug/repo-slug --no-browser
```

Flags:

- `--branch`: Branch name for source browsing
- `--commit`: Commit SHA for source browsing
- `--host`: Bitbucket host to use
- `--issue`: Open one issue by ID
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-browser`: Print the destination URL instead of opening the browser
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pipelines`: Open repository pipelines
- `--pr`: Open one pull request by ID
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--settings`: Open repository settings
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb config`

Manage bb behavior defaults

Manage persistent bb settings that affect runtime behavior today, such as prompt behavior, browser selection for `bb browse`, and the default output format.

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
bb config get browser
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
bb config set browser 'firefox --new-window'
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
bb config unset browser
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

List, view, create, edit, close, and reopen Bitbucket Cloud repository issues, and manage issue comments, attachments, milestones, and components.

Usage:

```text
bb issue
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb issue attachment`: Work with issue attachments
- `bb issue close`: Close an issue
- `bb issue comment`: Work with issue comments
- `bb issue component`: List and view issue components
- `bb issue create`: Create an issue
- `bb issue edit`: Edit an issue
- `bb issue list`: List issues for a repository
- `bb issue milestone`: List and view issue milestones
- `bb issue reopen`: Reopen an issue
- `bb issue view`: View one issue

## `bb issue attachment`

Work with issue attachments

List and upload Bitbucket issue attachments. Attachment import and export jobs remain separate platform workflows.

Usage:

```text
bb issue attachment
```

Examples:

```bash
bb issue attachment list 1 --repo workspace-slug/issues-repo-slug
bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb issue attachment list`: List attachments on an issue
- `bb issue attachment upload`: Upload attachments to an issue

## `bb issue attachment list`

List attachments on an issue

Usage:

```text
bb issue attachment list <issue-id-or-url> [flags]
```

Examples:

```bash
bb issue attachment list 1 --repo workspace-slug/issues-repo-slug
bb issue attachment list https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of issue attachments to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue attachment upload`

Upload attachments to an issue

Upload one or more files to a Bitbucket issue. Existing attachments with the same name are replaced by Bitbucket Cloud.

Usage:

```text
bb issue attachment upload <issue-id-or-url> <file>... [flags]
```

Examples:

```bash
bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug
bb issue attachment upload https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 ./trace.txt ./screenshot.png --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

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

## `bb issue comment`

Work with issue comments

List, view, create, edit, and delete Bitbucket issue comments.

Usage:

```text
bb issue comment
```

Examples:

```bash
bb issue comment list 1 --repo workspace-slug/issues-repo-slug
bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body 'Needs follow-up'
bb issue comment view 3 --issue 1 --repo workspace-slug/issues-repo-slug
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb issue comment create`: Create an issue comment
- `bb issue comment delete`: Delete an issue comment
- `bb issue comment edit`: Edit an issue comment
- `bb issue comment list`: List comments on an issue
- `bb issue comment view`: View one issue comment

## `bb issue comment create`

Create an issue comment

Usage:

```text
bb issue comment create <issue-id-or-url> [flags]
```

Examples:

```bash
bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body 'Needs follow-up'
bb issue comment create https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --body-file comment.md --json '*'
printf 'Needs follow-up\n' | bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body-file -
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

## `bb issue comment delete`

Delete an issue comment

Delete a Bitbucket issue comment. Humans must confirm the exact repository, issue, and comment unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb issue comment delete <comment-id> [flags]
```

Examples:

```bash
bb issue comment delete 3 --issue 1 --repo workspace-slug/issues-repo-slug --yes
bb --no-prompt issue comment delete 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --yes --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--issue`: Issue ID or Bitbucket issue URL
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb issue comment edit`

Edit an issue comment

Usage:

```text
bb issue comment edit <comment-id> [flags]
```

Examples:

```bash
bb issue comment edit 3 --issue 1 --repo workspace-slug/issues-repo-slug --body 'Updated feedback'
bb issue comment edit 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --body-file comment.md --json '*'
```

Flags:

- `--body-file`: Read the comment body from a file, or '-' for stdin
- `--body`: Comment body text
- `--host`: Bitbucket host to use
- `--issue`: Issue ID or Bitbucket issue URL
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue comment list`

List comments on an issue

Usage:

```text
bb issue comment list <issue-id-or-url> [flags]
```

Examples:

```bash
bb issue comment list 1 --repo workspace-slug/issues-repo-slug
bb issue comment list https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of issue comments to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue comment view`

View one issue comment

Usage:

```text
bb issue comment view <comment-id> [flags]
```

Examples:

```bash
bb issue comment view 3 --issue 1 --repo workspace-slug/issues-repo-slug
bb issue comment view 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--issue`: Issue ID or Bitbucket issue URL
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue component`

List and view issue components

List and view Bitbucket issue tracker components. Bitbucket Cloud only exposes component read APIs in the official REST surface.

Usage:

```text
bb issue component
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb issue component list`: List issue components for a repository
- `bb issue component view`: View one issue component

## `bb issue component list`

List issue components for a repository

Usage:

```text
bb issue component list [flags]
```

Examples:

```bash
bb issue component list --repo workspace-slug/issues-repo-slug
bb issue component list --repo workspace-slug/issues-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of components to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue component view`

View one issue component

Usage:

```text
bb issue component view <id> [flags]
```

Examples:

```bash
bb issue component view 1 --repo workspace-slug/issues-repo-slug
bb issue component view 1 --repo workspace-slug/issues-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue create`

Create an issue

Usage:

```text
bb issue create [flags]
```

Examples:

```bash
bb issue create --repo workspace-slug/issues-repo-slug --title 'Broken flow'
bb issue create --repo workspace-slug/repo-slug --title 'Broken flow' --body 'Needs investigation'
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
bb issue edit 1 --repo workspace-slug/issues-repo-slug --title 'Updated title'
bb issue edit 1 --repo workspace-slug/repo-slug --state open --priority major --json
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
bb issue list --repo workspace-slug/issues-repo-slug
bb issue list --repo workspace-slug/repo-slug
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

## `bb issue milestone`

List and view issue milestones

List and view Bitbucket issue tracker milestones. Bitbucket Cloud only exposes milestone read APIs in the official REST surface.

Usage:

```text
bb issue milestone
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb issue milestone list`: List issue milestones for a repository
- `bb issue milestone view`: View one issue milestone

## `bb issue milestone list`

List issue milestones for a repository

Usage:

```text
bb issue milestone list [flags]
```

Examples:

```bash
bb issue milestone list --repo workspace-slug/issues-repo-slug
bb issue milestone list --repo workspace-slug/issues-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of milestones to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb issue milestone view`

View one issue milestone

Usage:

```text
bb issue milestone view <id> [flags]
```

Examples:

```bash
bb issue milestone view 1 --repo workspace-slug/issues-repo-slug
bb issue milestone view 1 --repo workspace-slug/issues-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
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
bb issue view 1 --repo workspace-slug/issues-repo-slug
bb issue view 1 --repo workspace-slug/repo-slug --json
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

- `bb pipeline cache`: Inspect and clear pipeline caches
- `bb pipeline list`: List pipeline runs for a repository
- `bb pipeline log`: Show the log for one pipeline step
- `bb pipeline run`: Trigger a pipeline run
- `bb pipeline runner`: Inspect repository pipeline runners
- `bb pipeline schedule`: Manage pipeline schedules
- `bb pipeline stop`: Stop a running pipeline
- `bb pipeline test-reports`: View pipeline test reports
- `bb pipeline variable`: Manage repository pipeline variables
- `bb pipeline view`: View one pipeline run

## `bb pipeline cache`

Inspect and clear pipeline caches

List Bitbucket repository pipeline caches, delete one cache by UUID, or clear caches by name when the official API supports it.

Usage:

```text
bb pipeline cache
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pipeline cache clear`: Clear pipeline caches by name
- `bb pipeline cache delete`: Delete one pipeline cache by UUID
- `bb pipeline cache list`: List pipeline caches

## `bb pipeline cache clear`

Clear pipeline caches by name

Clear Bitbucket pipeline caches by cache name. Humans must confirm the exact repository and cache name unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pipeline cache clear <name> [flags]
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pipeline cache delete`

Delete one pipeline cache by UUID

Delete a Bitbucket pipeline cache by UUID. Humans must confirm the exact repository and cache UUID unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pipeline cache delete <uuid> [flags]
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pipeline cache list`

List pipeline caches

Usage:

```text
bb pipeline cache list [flags]
```

Examples:

```bash
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pipeline caches to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline list`

List pipeline runs for a repository

Usage:

```text
bb pipeline list [flags]
```

Examples:

```bash
bb pipeline list --repo workspace-slug/repo-slug
bb pipeline list --repo workspace-slug/repo-slug --state COMPLETED --json build_number,state,target
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

## `bb pipeline log`

Show the log for one pipeline step

Show the raw log for a pipeline step. If the pipeline has exactly one step, bb selects it automatically. Otherwise pass --step with a step UUID or name.

Usage:

```text
bb pipeline log <number-or-uuid> [flags]
```

Examples:

```bash
bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug
bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug --json pipeline,step,log
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--step`: Pipeline step UUID or name when a pipeline has more than one step
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline run`

Trigger a pipeline run

Trigger a Bitbucket pipeline run for a branch or tag. If no ref is provided, bb uses the current local branch when the repository target matches the current checkout.

Usage:

```text
bb pipeline run [ref] [flags]
```

Examples:

```bash
bb pipeline run main --repo workspace-slug/pipelines-repo-slug
bb pipeline run --repo workspace-slug/pipelines-repo-slug --ref main --json '*'
bb pipeline run v1.2.3 --ref-type tag --repo workspace-slug/pipelines-repo-slug
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--ref-type`: Reference type to build: branch or tag
- `--ref`: Branch or tag name to build
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline runner`

Inspect repository pipeline runners

List, inspect, and delete Bitbucket repository pipeline runners. Runner creation and update remain out of scope until the official API request shape is clearer.

Usage:

```text
bb pipeline runner
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pipeline runner delete`: Delete a pipeline runner
- `bb pipeline runner list`: List pipeline runners
- `bb pipeline runner view`: View one pipeline runner

## `bb pipeline runner delete`

Delete a pipeline runner

Delete a Bitbucket repository pipeline runner. Humans must confirm the exact repository and runner UUID unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pipeline runner delete <uuid> [flags]
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pipeline runner list`

List pipeline runners

Usage:

```text
bb pipeline runner list [flags]
```

Examples:

```bash
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pipeline runners to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline runner view`

View one pipeline runner

Usage:

```text
bb pipeline runner view <uuid> [flags]
```

Examples:

```bash
bb pipeline runner view '{runner-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline runner view '{runner-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline schedule`

Manage pipeline schedules

List, inspect, create, enable, disable, and delete Bitbucket pipeline schedules.

Usage:

```text
bb pipeline schedule
```

Examples:

```bash
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'
bb pipeline schedule disable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pipeline schedule create`: Create a pipeline schedule
- `bb pipeline schedule delete`: Delete a pipeline schedule
- `bb pipeline schedule disable`: Disable a pipeline schedule
- `bb pipeline schedule enable`: Enable a pipeline schedule
- `bb pipeline schedule list`: List pipeline schedules
- `bb pipeline schedule view`: View one pipeline schedule

## `bb pipeline schedule create`

Create a pipeline schedule

Create a Bitbucket pipeline schedule for a branch. By default bb uses the branch name as the selector pattern and the branches selector type.

Usage:

```text
bb pipeline schedule create [flags]
```

Examples:

```bash
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref release --cron '0 30 9 * * ? *' --enabled=false --json '*'
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --selector-type custom --selector-pattern nightly --cron '0 0 1 * * ? *'
```

Flags:

- `--cron`: Seven-field cron pattern in UTC
- `--enabled`: Create the schedule enabled
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--ref`: Branch name to run on the schedule
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--selector-pattern`: Pipeline selector pattern; defaults to the ref name
- `--selector-type`: Pipeline selector type, for example branches or custom
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline schedule delete`

Delete a pipeline schedule

Delete a Bitbucket pipeline schedule. Humans must confirm the exact repository and schedule UUID unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pipeline schedule delete <uuid> [flags]
```

Examples:

```bash
bb pipeline schedule delete '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --yes
bb --no-prompt pipeline schedule delete '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pipeline schedule disable`

Disable a pipeline schedule

Usage:

```text
bb pipeline schedule disable <uuid> [flags]
```

Examples:

```bash
bb pipeline schedule disable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule disable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline schedule enable`

Enable a pipeline schedule

Usage:

```text
bb pipeline schedule enable <uuid> [flags]
```

Examples:

```bash
bb pipeline schedule enable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule enable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline schedule list`

List pipeline schedules

Usage:

```text
bb pipeline schedule list [flags]
```

Examples:

```bash
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pipeline schedules to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline schedule view`

View one pipeline schedule

Usage:

```text
bb pipeline schedule view <uuid> [flags]
```

Examples:

```bash
bb pipeline schedule view '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule view '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline stop`

Stop a running pipeline

Stop a Bitbucket pipeline run. Humans must confirm the exact repository and pipeline number unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pipeline stop <number-or-uuid> [flags]
```

Examples:

```bash
bb pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes
bb pipeline stop '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
bb --no-prompt pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes --json pipeline,stopped
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pipeline test-reports`

View pipeline test reports

View Bitbucket pipeline test reports for one pipeline step. If the pipeline has exactly one step, bb selects it automatically. Otherwise pass --step with a step UUID or step name.

Usage:

```text
bb pipeline test-reports <number-or-uuid> [flags]
```

Examples:

```bash
bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug
bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug --cases --limit 50 --json '*'
bb pipeline test-reports '{uuid}' --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
```

Flags:

- `--cases`: Include individual test cases in the response
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of test cases to return when --cases is set
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--step`: Pipeline step UUID or name when a pipeline has more than one step
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline variable`

Manage repository pipeline variables

List, inspect, create, edit, and delete Bitbucket repository pipeline variables.

Usage:

```text
bb pipeline variable
```

Examples:

```bash
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug
bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured
bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pipeline variable create`: Create a repository pipeline variable
- `bb pipeline variable delete`: Delete a repository pipeline variable
- `bb pipeline variable edit`: Edit a repository pipeline variable
- `bb pipeline variable list`: List repository pipeline variables
- `bb pipeline variable view`: View one repository pipeline variable

## `bb pipeline variable create`

Create a repository pipeline variable

Usage:

```text
bb pipeline variable create [flags]
```

Examples:

```bash
bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured
printf 'token-value\n' | bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file - --json '*'
bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value production
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--key`: Pipeline variable key
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--secured`: Mark the variable as secured
- `--value-file`: Read the pipeline variable value from a file, or '-' for stdin
- `--value`: Pipeline variable value
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline variable delete`

Delete a repository pipeline variable

Delete a Bitbucket repository pipeline variable by key or UUID. Humans must confirm the exact repository and variable unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pipeline variable delete <key-or-uuid> [flags]
```

Examples:

```bash
bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes
bb --no-prompt pipeline variable delete '{uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'
bb pipeline variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --yes
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pipeline variable edit`

Edit a repository pipeline variable

Edit a Bitbucket repository pipeline variable by key or UUID. By default the existing secured flag is preserved unless --secured true or --secured false is provided.

Usage:

```text
bb pipeline variable edit <key-or-uuid> [flags]
```

Examples:

```bash
bb pipeline variable edit CI_TOKEN --repo workspace-slug/pipelines-repo-slug --value-file secret.txt --secured true
bb pipeline variable edit '{uuid}' --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value staging --json '*'
bb pipeline variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --value production
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--key`: Override the pipeline variable key
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--secured`: Set secured to true or false; defaults to the existing value
- `--value-file`: Read the pipeline variable value from a file, or '-' for stdin
- `--value`: Pipeline variable value
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline variable list`

List repository pipeline variables

Usage:

```text
bb pipeline variable list [flags]
```

Examples:

```bash
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of pipeline variables to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline variable view`

View one repository pipeline variable

Usage:

```text
bb pipeline variable view <key-or-uuid> [flags]
```

Examples:

```bash
bb pipeline variable view CI_TOKEN --repo workspace-slug/pipelines-repo-slug
bb pipeline variable view '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pipeline view`

View one pipeline run

Usage:

```text
bb pipeline view <number-or-uuid> [flags]
```

Examples:

```bash
bb pipeline view 42 --repo workspace-slug/repo-slug
bb pipeline view '{uuid}' --repo workspace-slug/repo-slug --json '*'
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

- `bb pr activity`: Show recent pull request activity
- `bb pr checkout`: Check out a pull request locally
- `bb pr checks`: Show commit statuses for a pull request
- `bb pr close`: Close a pull request without merging it
- `bb pr comment`: Add a comment to a pull request
- `bb pr commits`: List commits on a pull request
- `bb pr create`: Create a pull request
- `bb pr diff`: View a pull request diff
- `bb pr list`: List pull requests for a repository
- `bb pr merge`: Merge a pull request
- `bb pr review`: Review a pull request
- `bb pr status`: Show pull request status for a repository
- `bb pr task`: Work with pull request tasks
- `bb pr view`: View a pull request

## `bb pr activity`

Show recent pull request activity

Show recent pull request activity including comments, updates, approvals, and change requests. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr activity <id-or-url> [flags]
```

Examples:

```bash
bb pr activity 7 --repo workspace-slug/repo-slug
bb pr activity https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --limit 50 --json '*'
bb pr activity https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of activity entries to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr checkout`

Check out a pull request locally

Fetch the pull request source branch from the current repository's remote and switch to it locally. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr checkout <id-or-url> [flags]
```

Examples:

```bash
bb pr checkout 1
bb pr checkout 1 --repo workspace-slug/repo-slug
bb pr checkout https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
bb pr checkout https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
```

Flags:

- `--host`: Bitbucket host to use
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr checks`

Show commit statuses for a pull request

Show commit statuses for a pull request. This is the Bitbucket Cloud equivalent of PR checks backed by commit statuses. Accepts a numeric ID, pull request URL, or pull request comment URL.

Aliases: `statuses`

Usage:

```text
bb pr checks <id-or-url> [flags]
```

Examples:

```bash
bb pr checks 7 --repo workspace-slug/repo-slug
bb pr checks https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'
bb pr statuses https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of commit statuses to return
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr close`

Close a pull request without merging it

Close a pull request without merging it. In Bitbucket Cloud this maps to declining the pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr close <id-or-url> [flags]
```

Examples:

```bash
bb pr close 1
bb pr close 1 --repo workspace-slug/repo-slug
bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json
bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
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

Add a comment to a pull request using --body, --body-file, or --body-file - for stdin. This first pass is intentionally deterministic for agent and script usage. Use the comment subcommands to view, edit, delete, resolve, or reopen specific pull request comments.

Usage:

```text
bb pr comment <id-or-url> [flags]
```

Examples:

```bash
bb pr comment 1 --body 'Looks good'
bb pr comment 1 --repo workspace-slug/repo-slug --body-file comment.md
printf 'Ship it\n' | bb pr comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --body-file - --json
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

Subcommands:

- `bb pr comment delete`: Delete a pull request comment
- `bb pr comment edit`: Edit a pull request comment
- `bb pr comment reopen`: Reopen a pull request comment thread
- `bb pr comment resolve`: Resolve a pull request comment thread
- `bb pr comment view`: View a pull request comment

## `bb pr comment delete`

Delete a pull request comment

Delete a pull request comment. Humans must confirm the exact repository, pull request, and comment unless --yes is provided. Scripts and agents should use --yes together with --no-prompt. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.

Usage:

```text
bb pr comment delete <comment-url-or-id> [flags]
```

Examples:

```bash
bb pr comment delete https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --yes
bb --no-prompt pr comment delete 15 --pr 1 --repo workspace-slug/repo-slug --yes --json '*'
bb pr comment delete 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --yes
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL; required when the comment target is a numeric ID
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pr comment edit`

Edit a pull request comment

Edit a pull request comment. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.

Usage:

```text
bb pr comment edit <comment-url-or-id> [flags]
```

Examples:

```bash
bb pr comment edit https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Updated feedback'
bb pr comment edit 15 --pr 1 --repo workspace-slug/repo-slug --body-file comment.md --json '*'
printf 'Updated feedback\n' | bb pr comment edit 15 --pr 1 --repo workspace-slug/repo-slug --body-file -
```

Flags:

- `--body-file`: Read the comment body from a file, or '-' for stdin
- `--body`: Comment body text
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL; required when the comment target is a numeric ID
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr comment reopen`

Reopen a pull request comment thread

Reopen a previously resolved pull request comment thread. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.

Usage:

```text
bb pr comment reopen <comment-url-or-id> [flags]
```

Examples:

```bash
bb pr comment reopen https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment reopen 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
bb pr comment reopen 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL; required when the comment target is a numeric ID
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr comment resolve`

Resolve a pull request comment thread

Resolve a pull request comment thread. Bitbucket Cloud only allows resolving top-level diff comments. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.

Usage:

```text
bb pr comment resolve <comment-url-or-id> [flags]
```

Examples:

```bash
bb pr comment resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment resolve 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
bb pr comment resolve 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL; required when the comment target is a numeric ID
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr comment view`

View a pull request comment

View a pull request comment. Accepts a Bitbucket pull request comment URL directly, or a numeric comment ID together with --pr <id-or-url>.

Usage:

```text
bb pr comment view <comment-url-or-id> [flags]
```

Examples:

```bash
bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment view 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
bb pr comment view 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL; required when the comment target is a numeric ID
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr commits`

List commits on a pull request

List the commits that would be merged by the pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr commits <id-or-url> [flags]
```

Examples:

```bash
bb pr commits 7 --repo workspace-slug/repo-slug
bb pr commits https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'
bb pr commits https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --limit 50
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of commits to return
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

Show the patch for a pull request by default. Use --stat for a concise per-file summary, or --json for structured output that includes both the patch and diff stats. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr diff <id-or-url> [flags]
```

Examples:

```bash
bb pr diff 1
bb pr diff 1 --repo workspace-slug/repo-slug --stat
bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json patch,stats
bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --stat
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
bb pr list --repo workspace-slug/repo-slug
bb pr list --repo https://bitbucket.org/workspace-slug/repo-slug
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

Merge an open pull request in Bitbucket Cloud. bb uses the destination branch default merge strategy when Bitbucket exposes one, or falls back to the repository default when Bitbucket does not include strategy metadata on the pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr merge <id-or-url> [flags]
```

Examples:

```bash
bb pr merge 7
bb pr merge 7 --repo workspace-slug/repo-slug
bb pr merge 7 --strategy merge_commit
bb pr merge 7 --message 'Ship feature' --close-source-branch --json
bb pr merge https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15
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

## `bb pr review`

Review a pull request

Review a pull request using Bitbucket Cloud review actions such as approve, request-changes, and clearing your own prior review state.

Usage:

```text
bb pr review
```

Examples:

```bash
bb pr review approve 7 --repo workspace-slug/repo-slug
bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review clear-request-changes 7 --repo workspace-slug/repo-slug
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pr review approve`: Approve a pull request
- `bb pr review clear-request-changes`: Clear your prior request for changes
- `bb pr review request-changes`: Request changes on a pull request
- `bb pr review unapprove`: Withdraw your approval of a pull request

## `bb pr review approve`

Approve a pull request

Approve a pull request as the authenticated user. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr review approve <id-or-url> [flags]
```

Examples:

```bash
bb pr review approve 7 --repo workspace-slug/repo-slug
bb pr review approve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review approve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr review clear-request-changes`

Clear your prior request for changes

Clear your own prior request for changes on a pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr review clear-request-changes <id-or-url> [flags]
```

Examples:

```bash
bb pr review clear-request-changes 7 --repo workspace-slug/repo-slug
bb pr review clear-request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review clear-request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr review request-changes`

Request changes on a pull request

Request changes on a pull request as the authenticated user. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr review request-changes <id-or-url> [flags]
```

Examples:

```bash
bb pr review request-changes 7 --repo workspace-slug/repo-slug
bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr review unapprove`

Withdraw your approval of a pull request

Withdraw your own approval of a pull request. Accepts a numeric ID, pull request URL, or pull request comment URL.

Usage:

```text
bb pr review unapprove <id-or-url> [flags]
```

Examples:

```bash
bb pr review unapprove 7 --repo workspace-slug/repo-slug
bb pr review unapprove https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review unapprove https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
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
bb pr status --repo workspace-slug/repo-slug
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

## `bb pr task`

Work with pull request tasks

List, inspect, create, edit, delete, resolve, and reopen Bitbucket pull request tasks. Tasks can be attached to specific pull request comments when the Bitbucket Cloud REST API supports it.

Usage:

```text
bb pr task
```

Examples:

```bash
bb pr task list 1 --repo workspace-slug/repo-slug
bb pr task create 1 --repo workspace-slug/repo-slug --body 'Follow up on reviewer feedback'
bb pr task create 1 --repo workspace-slug/repo-slug --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Handle this thread'
bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug
```

Flags:

- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

Subcommands:

- `bb pr task create`: Create a task on a pull request
- `bb pr task delete`: Delete a pull request task
- `bb pr task edit`: Edit a pull request task
- `bb pr task list`: List tasks on a pull request
- `bb pr task reopen`: Reopen a pull request task
- `bb pr task resolve`: Resolve a pull request task
- `bb pr task view`: View a pull request task

## `bb pr task create`

Create a task on a pull request

Create a Bitbucket pull request task using --body, --body-file, or --body-file - for stdin. Use --comment to attach the task to a specific pull request comment when Bitbucket Cloud supports that linkage.

Usage:

```text
bb pr task create <pr-id-or-url> [flags]
```

Examples:

```bash
bb pr task create 1 --repo workspace-slug/repo-slug --body 'Follow up on review feedback'
bb pr task create 1 --repo workspace-slug/repo-slug --comment 15 --body-file task.md --json '*'
bb pr task create https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Handle this thread'
```

Flags:

- `--body-file`: Read the task body from a file, or '-' for stdin
- `--body`: Task body text
- `--comment`: Optional pull request comment as a numeric ID or Bitbucket pull request comment URL
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pending`: Mark the created task as pending
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr task delete`

Delete a pull request task

Delete a Bitbucket pull request task. Humans must confirm the exact repository, pull request, and task unless --yes is provided. Scripts and agents should use --yes together with --no-prompt.

Usage:

```text
bb pr task delete <task-id> [flags]
```

Examples:

```bash
bb pr task delete 3 --pr 1 --repo workspace-slug/repo-slug --yes
bb --no-prompt pr task delete 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --yes --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target
- `--yes`: Skip the confirmation prompt

## `bb pr task edit`

Edit a pull request task

Edit the body of a Bitbucket pull request task. Tasks are addressed by numeric task ID together with --pr <id-or-url>.

Usage:

```text
bb pr task edit <task-id> [flags]
```

Examples:

```bash
bb pr task edit 3 --pr 1 --repo workspace-slug/repo-slug --body 'Updated follow-up'
bb pr task edit 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --body-file task.md --json '*'
```

Flags:

- `--body-file`: Read the task body from a file, or '-' for stdin
- `--body`: Task body text
- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr task list`

List tasks on a pull request

List tasks on a Bitbucket pull request. Accepts a numeric pull request ID or Bitbucket pull request URL. Defaults to unresolved tasks; pass --state all to see everything.

Usage:

```text
bb pr task list <pr-id-or-url> [flags]
```

Examples:

```bash
bb pr task list 1 --repo workspace-slug/repo-slug
bb pr task list https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --state all --json '*'
bb pr task list 1 --repo workspace-slug/repo-slug --state resolved --limit 50
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--limit`: Maximum number of tasks to list
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--state`: Task state filter: unresolved, resolved, or all
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr task reopen`

Reopen a pull request task

Reopen a Bitbucket pull request task by updating its task state to UNRESOLVED. Tasks are addressed by numeric task ID together with --pr <id-or-url>.

Usage:

```text
bb pr task reopen <task-id> [flags]
```

Examples:

```bash
bb pr task reopen 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task reopen 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr task resolve`

Resolve a pull request task

Resolve a Bitbucket pull request task by updating its task state to RESOLVED. Tasks are addressed by numeric task ID together with --pr <id-or-url>.

Usage:

```text
bb pr task resolve <task-id> [flags]
```

Examples:

```bash
bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task resolve 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr task view`

View a pull request task

View a specific Bitbucket pull request task. Tasks are addressed by numeric task ID together with --pr <id-or-url>.

Usage:

```text
bb pr task view <task-id> [flags]
```

Examples:

```bash
bb pr task view 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task view 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--pr`: Parent pull request as an ID or Bitbucket pull request URL
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb pr view`

View a pull request

View a pull request by numeric ID, pull request URL, or pull request comment URL. Comment URLs resolve to the parent pull request.

Usage:

```text
bb pr view <id-or-url> [flags]
```

Examples:

```bash
bb pr view 1
bb pr view 1 --json title,state,source,destination
bb pr view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
bb pr view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
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
bb repo clone workspace-slug/repo-slug
bb repo clone --repo workspace-slug/repo-slug ./tmp/repo
bb repo clone repo-slug --workspace workspace-slug
bb repo clone https://bitbucket.org/workspace-slug/repo-slug
bb repo clone workspace-slug/repo-slug ./tmp/repo --json
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
bb repo create workspace-slug/my-repo --project-key BBCLI
bb repo create --repo workspace-slug/my-repo --reuse-existing --json
bb repo create my-repo --workspace workspace-slug
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
bb repo delete workspace-slug/delete-repo-slug --yes
bb repo delete --repo workspace-slug/delete-repo-slug --yes
bb repo delete delete-repo-slug --workspace workspace-slug --yes
bb repo delete https://bitbucket.org/workspace-slug/delete-repo-slug --json
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
bb repo view --repo workspace-slug/repo-slug
bb repo view --repo https://bitbucket.org/workspace-slug/repo-slug
bb repo view --json name,project_key,main_branch
```

Flags:

- `--host`: Bitbucket host to use
- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal
- `--repo`: Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL
- `--workspace`: Optional workspace slug used only to disambiguate a bare repository target

## `bb resolve`

Resolve a Bitbucket URL into a structured entity

Resolve Bitbucket repository, pull request, pull request comment, issue, commit, and source URLs into a structured entity payload without making an API request.

Usage:

```text
bb resolve <url> [flags]
```

Examples:

```bash
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb resolve https://bitbucket.org/workspace-slug/repo-slug/src/main/README.md#lines-12 --json type,repo,path,line
```

Flags:

- `--jq`: Filter JSON output using a jq expression
- `--json`: Output JSON with the specified comma-separated fields, or '*' for all fields
- `--no-prompt`: Do not prompt for missing input, even in an interactive terminal

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
bb search issues fixture --repo workspace-slug/issues-repo-slug
bb search issues bug --repo workspace-slug/issues-repo-slug --json id,title,state
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
bb search prs fixture --repo workspace-slug/repo-slug
bb search prs feature --repo workspace-slug/repo-slug --json id,title,state
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
bb search repos integration --workspace workspace-slug
bb search repos bb-cli --workspace workspace-slug --json name,slug,description
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
bb status --workspace workspace-slug --limit 10
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
