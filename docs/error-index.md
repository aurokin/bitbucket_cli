# Error Index

Generated from the current recovery guidance catalog.

Use this file when you have an error fragment and want the fastest likely next command.

| Error fragment | Recovery focus | First next command |
|---|---|---|
| `authentication failed: the stored API token may be invalid, expired, or revoked.` | Invalid, Expired, Or Revoked API Token | `BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login` |
| `request denied by Bitbucket Cloud: the API token may be missing required Bitbucket scopes` | Missing Token Scopes Or Insufficient Access | `bb auth status --check` |
| `multiple workspaces are available; pass --repo <workspace>/<repo> or add --workspace to disambiguate` | Ambiguous Repository Resolution | `bb repo view --repo workspace-slug/repo-slug` |
| `could not determine the repository from the current directory` | No Repository In The Current Directory | `bb repo view --repo workspace-slug/repo-slug` |
| `invalid alias "ship"` | Invalid Alias Quoting | `bb alias get ship` |
| `this repository does not have Bitbucket issue tracking enabled` | Repository Without Bitbucket Issue Tracking | `bb repo view --repo workspace-slug/issues-repo-slug` |
| `Notes Some workspaces hit --repo-limit. Some repositories do not have issue tracking enabled.` | Bounded Cross-Repository Status Output | `bb status --workspace workspace-slug --repo-limit 200 --limit 50` |
