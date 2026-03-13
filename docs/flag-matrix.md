# Flag Matrix

Generated from the `bb` command tree.

This is a compact view of common automation and targeting flags across executable commands.

| Command | `--json` | `--jq` | `--no-prompt` | `--host` | `--repo` | `--workspace` | Other notable flags |
|---|---|---|---|---|---|---|---|
| `bb alias delete` |  |  | yes |  |  |  |  |
| `bb alias get` |  |  | yes |  |  |  |  |
| `bb alias list` | yes | yes | yes |  |  |  |  |
| `bb alias set` |  |  | yes |  |  |  |  |
| `bb api` |  | yes | yes | yes |  |  | `--input`, `--method` |
| `bb auth login` |  |  | yes | yes |  |  | `--default`, `--token`, `--username`, `--with-token` |
| `bb auth logout` |  |  | yes | yes |  |  |  |
| `bb auth status` | yes | yes | yes | yes |  |  | `--check` |
| `bb browse` | yes | yes | yes | yes | yes | yes | `--branch`, `--commit`, `--issue`, `--no-browser`, `--pipelines`, `--pr`, `--settings` |
| `bb config get` | yes | yes | yes |  |  |  |  |
| `bb config list` | yes | yes | yes |  |  |  |  |
| `bb config path` |  |  | yes |  |  |  |  |
| `bb config set` |  |  | yes |  |  |  |  |
| `bb config unset` |  |  | yes |  |  |  |  |
| `bb extension exec` |  |  | yes |  |  |  |  |
| `bb extension list` | yes | yes | yes |  |  |  |  |
| `bb issue close` | yes | yes | yes | yes | yes | yes | `--message`, `--state` |
| `bb issue create` | yes | yes | yes | yes | yes | yes | `--body`, `--kind`, `--priority`, `--title` |
| `bb issue edit` | yes | yes | yes | yes | yes | yes | `--body`, `--kind`, `--priority`, `--state`, `--title` |
| `bb issue list` | yes | yes | yes | yes | yes | yes | `--limit`, `--state` |
| `bb issue reopen` | yes | yes | yes | yes | yes | yes | `--message`, `--state` |
| `bb issue view` | yes | yes | yes | yes | yes | yes |  |
| `bb pipeline list` | yes | yes | yes | yes | yes | yes | `--limit`, `--state` |
| `bb pipeline log` | yes | yes | yes | yes | yes | yes | `--step` |
| `bb pipeline stop` | yes | yes | yes | yes | yes | yes | `--yes` |
| `bb pipeline view` | yes | yes | yes | yes | yes | yes |  |
| `bb pr checkout` |  |  | yes | yes | yes | yes |  |
| `bb pr close` | yes | yes | yes | yes | yes | yes |  |
| `bb pr comment` | yes | yes | yes | yes | yes | yes | `--body-file`, `--body` |
| `bb pr create` | yes | yes | yes | yes | yes | yes | `--close-source-branch`, `--description`, `--destination`, `--draft`, `--reuse-existing`, `--source`, `--title` |
| `bb pr diff` | yes | yes | yes | yes | yes | yes | `--stat` |
| `bb pr list` | yes | yes | yes | yes | yes | yes | `--limit`, `--state` |
| `bb pr merge` | yes | yes | yes | yes | yes | yes | `--close-source-branch`, `--message`, `--strategy` |
| `bb pr status` | yes | yes | yes | yes | yes | yes | `--limit` |
| `bb pr view` | yes | yes | yes | yes | yes | yes |  |
| `bb repo clone` | yes | yes | yes | yes | yes | yes |  |
| `bb repo create` | yes | yes | yes | yes | yes | yes | `--description`, `--name`, `--private`, `--project-key`, `--reuse-existing` |
| `bb repo delete` | yes | yes | yes | yes | yes | yes | `--yes` |
| `bb repo view` | yes | yes | yes | yes | yes | yes |  |
| `bb resolve` | yes | yes | yes |  |  |  |  |
| `bb search issues` | yes | yes | yes | yes | yes | yes | `--limit` |
| `bb search prs` | yes | yes | yes | yes | yes | yes | `--limit` |
| `bb search repos` | yes | yes | yes | yes |  | yes | `--limit` |
| `bb status` | yes | yes | yes | yes |  | yes | `--limit`, `--repo-limit` |
| `bb version` | yes | yes | yes |  |  |  |  |
