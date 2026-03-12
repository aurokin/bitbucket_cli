# Failure And Recovery

Common failure modes, what they mean, and the next command to run.

Generated from the current recovery guidance catalog.

## Invalid, Expired, Or Revoked API Token

Typical failure:

```text
authentication failed: the stored API token may be invalid, expired, or revoked.
```

Recovery:

```bash
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
bb auth status --check
```

Create or rotate the token at https://id.atlassian.com/manage-profile/security/api-tokens

## Missing Token Scopes Or Insufficient Access

Typical failure:

```text
request denied by Bitbucket Cloud: the API token may be missing required Bitbucket scopes
```

Recovery:

```bash
bb auth status --check
bb repo view --repo workspace-slug/repo-slug
```

If the token is valid but still denied, create a new Bitbucket API token with the required Bitbucket scopes at https://id.atlassian.com/manage-profile/security/api-tokens and store it again with `bb auth login`.

## Ambiguous Repository Resolution

Typical failure:

```text
multiple workspaces are available; pass --repo <workspace>/<repo> or add --workspace to disambiguate
```

Recovery:

```bash
bb repo view --repo workspace-slug/repo-slug
bb pr list --repo workspace-slug/repo-slug
```

Prefer `--repo <workspace>/<repo>` in automation and when you are outside a local checkout.

## No Repository In The Current Directory

Typical failure:

```text
could not determine the repository from the current directory
```

Recovery:

```bash
bb repo view --repo workspace-slug/repo-slug
bb browse --repo workspace-slug/repo-slug --no-browser
bb repo clone workspace-slug/repo-slug
```

## Invalid Alias Quoting

Typical failure:

```text
invalid alias "ship"
```

Recovery:

```bash
bb alias get ship
bb alias set ship 'pr create --repo workspace-slug/repo-slug --title "Add feature"'
```

If the alias is no longer needed:

`bb alias delete ship`

## Repository Without Bitbucket Issue Tracking

Typical failure:

```text
this repository does not have Bitbucket issue tracking enabled
```

Recovery:

```bash
bb repo view --repo workspace-slug/issues-repo-slug
bb issue list --repo workspace-slug/issues-repo-slug
```

Use a repository with Bitbucket issue tracking enabled, or enable issue tracking in the repository settings.

## Bounded Cross-Repository Status Output

Typical failure:

```text
Notes
  Some workspaces hit --repo-limit.
  Some repositories do not have issue tracking enabled.
```

Recovery:

```bash
bb status --workspace workspace-slug --repo-limit 200 --limit 50
bb pr list --repo workspace-slug/repo-slug
bb issue list --repo workspace-slug/issues-repo-slug
```

Use narrower workspace scans or explicit repository commands when you need complete detail.
