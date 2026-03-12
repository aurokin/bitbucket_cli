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
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
bb auth status --check
```

## Missing Token Scopes Or Insufficient Access

Typical failure:

```text
request denied by Bitbucket Cloud: the API token may be missing required Bitbucket scopes
```

Recovery:

```bash
bb auth status --check
bb repo view --repo OhBizzle/bb-cli-integration-primary
```

If the token is valid but still denied, create a new Bitbucket API token with the required Bitbucket scopes and store it again with `bb auth login`.

## Ambiguous Repository Resolution

Typical failure:

```text
multiple workspaces are available; pass --repo <workspace>/<repo> or add --workspace to disambiguate
```

Recovery:

```bash
bb repo view --repo OhBizzle/bb-cli-integration-primary
bb pr list --repo OhBizzle/bb-cli-integration-primary
```

Prefer `--repo <workspace>/<repo>` in automation and when you are outside a local checkout.

## No Repository In The Current Directory

Typical failure:

```text
could not determine the repository from the current directory
```

Recovery:

```bash
bb repo view --repo OhBizzle/bb-cli-integration-primary
bb browse --repo OhBizzle/bb-cli-integration-primary --no-browser
bb repo clone OhBizzle/bb-cli-integration-primary
```

## Invalid Alias Quoting

Typical failure:

```text
invalid alias "ship"
```

Recovery:

```bash
bb alias get ship
bb alias set ship 'pr create --repo OhBizzle/bb-cli-integration-primary --title "Add feature"'
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
bb repo view --repo OhBizzle/bb-cli-integration-issues
bb issue list --repo OhBizzle/bb-cli-integration-issues
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
bb status --workspace OhBizzle --repo-limit 200 --limit 50
bb pr list --repo OhBizzle/bb-cli-integration-primary
bb issue list --repo OhBizzle/bb-cli-integration-issues
```

Use narrower workspace scans or explicit repository commands when you need complete detail.
