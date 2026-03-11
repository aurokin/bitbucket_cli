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

## Why There Is No `bb pr reopen`

Bitbucket Cloud does not support reopening a declined pull request.

Once a pull request is declined in Bitbucket Cloud, it stays declined. The public Bitbucket Cloud pull request API exposes merge and decline operations, but it does not expose a reopen operation. Atlassian also documents this product limitation in their public issue tracker.

Because of that, `bb` does not provide a misleading `pr reopen` command that would pretend to restore the original pull request. The correct workflow is to create a new pull request from the same source and destination branches when you need to continue the work.

References:

- https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
- https://jira.atlassian.com/browse/BCLOUD-23807
