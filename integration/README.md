# Integration Tests

These tests are manual only.

They do not run in normal `go test ./...` and should not be added to CI.

## What They Do

- Reuse a dedicated Bitbucket Cloud project and repositories if they already exist
- Create them if they do not exist
- Seed arbitrary git content into the primary and secondary repositories
- Seed a dedicated pipelines repository with `bitbucket-pipelines.yml` and ensure there is at least one reusable pipeline run
- Ensure there is an open pull request in the primary repository
- Run the `bb browse`, `bb repo clone`, `bb repo view`, `bb repo create`, `bb repo delete`, `bb pipeline list`, `bb pipeline log`, `bb pipeline stop`, `bb pipeline view`, `bb pr list`, `bb pr status`, `bb pr review`, `bb pr activity`, `bb pr commits`, `bb pr checks`, `bb pr diff`, `bb pr comment`, `bb pr task`, `bb pr close`, `bb pr view`, `bb pr create`, `bb pr checkout`, `bb pr merge`, `bb issue list`, `bb issue view`, `bb issue create`, `bb issue close`, `bb issue reopen`, and `bb status` commands against the seeded repositories
- Smoke-test the human-readable output paths for browse, repo, pipeline, pull request, issue, and search commands
- Smoke-test representative structured-output commands used by the generated docs

## Fixture Names

- Project key: `BBCLI`
- Project name: `bb-cli integration`
- Primary repo: `bb-cli-integration-primary`
- Secondary repo: `bb-cli-integration-secondary`
- Issues repo: `bb-cli-integration-issues`
- Pipelines repo: `bb-cli-integration-pipelines`

## Run

```bash
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoView -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoClone -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoCreate -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudRepoDelete -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudBrowse -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPipelineList -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPipelineLog -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPipelineStop -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPipelineView -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRCreate -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRCheckout -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRMerge -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRList -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRStatus -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRReview -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRActivity -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRCommits -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRChecks -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRDiff -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRComment -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRTaskFlow -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRClose -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudPRView -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudIssueFlow -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudStatus -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudHumanOutputSmoke -v
BB_RUN_INTEGRATION=1 go test -tags=integration ./integration -run TestBitbucketCloudGeneratedDocsSmoke -v
```

## Notes

- The tests intentionally reuse existing fixtures to avoid unnecessary churn against the Bitbucket API.
- The `repo delete` test uses a dedicated sacrificial repository and recreates it only when needed.
- They do not delete projects, and they do not delete the primary or secondary fixture repositories.
- The issue flow uses a dedicated repository with Bitbucket issue tracking enabled.
- The pipeline flow uses a dedicated repository so normal fixture pushes do not create unnecessary pipeline churn.
