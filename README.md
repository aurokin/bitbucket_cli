# bb

`bb` is a Bitbucket Cloud CLI aimed at both humans and agents.

## Current Commands

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

## Compared With `gh`

### What `gh` Offers That `bb` Also Offers

- Authenticated API access through `gh api` / `bb api`
- Repository inspection, creation, cloning, and deletion
- Pull request listing, status, viewing, diffing, commenting, creation, checkout, merge, and close flows
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
- Search and cross-repository work dashboards
- Config, aliases, and extensions
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
