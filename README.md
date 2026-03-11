# bb

`bb` is a Bitbucket Cloud CLI aimed at both humans and agents.

## Current Commands

- `bb version`
- `bb config list`
- `bb config get`
- `bb config set`
- `bb config unset`
- `bb config path`
- `bb alias list`
- `bb alias get`
- `bb alias set`
- `bb alias delete`
- `bb extension list`
- `bb extension exec`
- `bb search repos`
- `bb search prs`
- `bb search issues`
- `bb status`
- `bb auth login`
- `bb auth logout`
- `bb auth status`
- `bb api`
- `bb repo view`
- `bb repo create`
- `bb repo clone`
- `bb repo delete`
- `bb pr list`
- `bb pr status`
- `bb pr view`
- `bb pr diff`
- `bb pr comment`
- `bb pr create`
- `bb pr checkout`
- `bb pr merge`
- `bb pr close`
- `bb issue list`
- `bb issue view`
- `bb issue create`
- `bb issue edit`
- `bb issue close`
- `bb issue reopen`

## Compared With `gh`

### What `gh` Offers That `bb` Also Offers

- Authenticated API access through `gh api` / `bb api`
- Repository inspection, creation, cloning, and deletion
- Pull request listing, status, viewing, diffing, commenting, creation, checkout, merge, and close flows
- Issue listing, viewing, creation, editing, and state transitions
- Cross-repository status summaries
- Search for repositories, pull requests, and issues
- Config defaults for prompt behavior and default output format, plus aliases and extension discovery
- Structured automation paths with `--json`
- Flexible repository targeting with local git inference, `workspace/repo`, and Bitbucket/GitHub-style URLs

### What `bb` Offers That `gh` Does Not

- Native Bitbucket Cloud repository and pull request workflows
- Bitbucket project-aware repository creation through `--project-key`
- Explicit documentation of Bitbucket Cloud platform limits when parity is impossible

### What `gh` Offers That `bb` Does Not

- Browser login and broader auth account management
- `browse`
- Issues, releases, and CI/workflow commands
- Richer repository administration such as list, edit, rename, fork, archive, and sync
- Additional pull request flows such as review, checks, edit, ready, update-branch, and revert
- Pull request reopen on platforms that actually support it

## Why There Is No `bb pr reopen`

Bitbucket Cloud does not support reopening a declined pull request.

Once a pull request is declined in Bitbucket Cloud, it stays declined. The public Bitbucket Cloud pull request API exposes merge and decline operations, but it does not expose a reopen operation. Atlassian also documents this product limitation in their public issue tracker.

Because of that, `bb` does not provide a misleading `pr reopen` command that would pretend to restore the original pull request. The correct workflow is to create a new pull request from the same source and destination branches when you need to continue the work.

References:

- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- https://jira.atlassian.com/browse/BCLOUD-23807

## Notes On Current Behavior

- `bb status` is intentionally bounded. When a workspace scan hits `--repo-limit`, an item section hits `--limit`, or issue tracking is disabled on some repositories, the output includes notes telling you to continue with `bb pr list --repo <workspace>/<repo>` or `bb issue list --repo <workspace>/<repo>`.
- `bb config` only exposes keys that affect runtime today: `prompt` and `output.format`. Browser, editor, and pager configuration are not wired up yet and are not exposed as working settings.
- Alias expansion preserves shell-style quoting so aliases like `bb alias set ship 'pr create --title "Add feature"'` expand reliably for both humans and automation.
