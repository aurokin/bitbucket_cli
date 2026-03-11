# Integration Tests

These tests are manual only.

They do not run in normal `go test ./...` and should not be added to CI.

## What They Do

- Reuse a dedicated Bitbucket Cloud project and repositories if they already exist
- Create them if they do not exist
- Seed arbitrary git content into the primary and secondary repositories
- Ensure there is an open pull request in the primary repository
- Run the `bb repo clone`, `bb repo view`, `bb repo create`, `bb repo delete`, `bb pr list`, `bb pr status`, `bb pr diff`, `bb pr comment`, `bb pr view`, `bb pr create`, `bb pr checkout`, and `bb pr merge` commands against the seeded repositories

## Fixture Names

- Project key: `BBCLI`
- Project name: `bb-cli integration`
- Primary repo: `bb-cli-integration-primary`
- Secondary repo: `bb-cli-integration-secondary`

## Run

```bash
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoView -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoClone -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoCreate -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoDelete -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRCreate -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRCheckout -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRMerge -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRList -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRStatus -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRDiff -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRComment -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRView -v
```

## Notes

- The tests intentionally reuse existing fixtures to avoid unnecessary churn against the Bitbucket API.
- The `repo delete` test uses a dedicated sacrificial repository and recreates it only when needed.
- They do not delete projects, and they do not delete the primary or secondary fixture repositories.
