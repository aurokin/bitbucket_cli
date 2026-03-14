# Command Examples

Generated from Cobra `Example` fields and validated by the docs example tests.

## `bb alias set`

```bash
bb alias set pv 'pr view'
bb alias set rls 'pr list --state OPEN'
```

## `bb api`

```bash
bb api /user
bb api '/repositories/workspace-slug/repo-slug/pullrequests?state=OPEN'
bb api /user --jq .display_name
printf '{"name":"my-repo"}' | bb api /repositories/workspace-slug/my-repo -X POST --input -
```

## `bb auth login`

```bash
bb auth login --username you@example.com --with-token
bb auth login --username you@example.com --token $BITBUCKET_TOKEN
BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login
printf '%s\n' "$BITBUCKET_TOKEN" | bb auth login --username you@example.com --with-token
```

## `bb auth logout`

```bash
bb auth logout
bb auth logout --host bitbucket.org
```

## `bb auth status`

```bash
bb auth status
bb auth status --check --json
bb auth status --host bitbucket.org
```

## `bb branch create`

```bash
bb branch create feature/demo --repo workspace-slug/repo-slug --target main
bb branch create feature/demo --repo workspace-slug/repo-slug --target abc1234 --json '*'
```

## `bb branch delete`

```bash
bb branch delete feature/demo --repo workspace-slug/repo-slug --yes
bb --no-prompt branch delete feature/demo --repo workspace-slug/repo-slug --yes --json '*'
```

## `bb branch list`

```bash
bb branch list workspace-slug/repo-slug
bb branch list --repo workspace-slug/repo-slug --limit 50
bb branch list --repo workspace-slug/repo-slug --query 'name ~ "release"' --json branches
```

## `bb branch view`

```bash
bb branch view main --repo workspace-slug/repo-slug
bb branch view feature/demo --repo workspace-slug/repo-slug --json '*'
```

## `bb browse`

```bash
bb browse --repo workspace-slug/repo-slug
bb browse README.md:12 --repo workspace-slug/repo-slug --no-browser
bb browse --pr 1 --repo workspace-slug/repo-slug
bb browse --pipelines --repo workspace-slug/repo-slug --json '*'
bb browse a1b2c3d --repo workspace-slug/repo-slug --no-browser
```

## `bb commit approve`

```bash
bb commit approve abc1234 --repo workspace-slug/repo-slug
bb commit approve https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'
```

## `bb commit comment list`

```bash
bb commit comment list abc1234 --repo workspace-slug/repo-slug
bb commit comment list https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json comments
```

## `bb commit comment view`

```bash
bb commit comment view 15 --commit abc1234 --repo workspace-slug/repo-slug
bb commit comment view 15 --commit https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'
```

## `bb commit diff`

```bash
bb commit diff abc1234 --repo workspace-slug/repo-slug
bb commit diff abc1234 --repo workspace-slug/repo-slug --stat
bb commit diff https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json patch,stats
```

## `bb commit report list`

```bash
bb commit report list abc1234 --repo workspace-slug/repo-slug
bb commit report list https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json reports
```

## `bb commit report view`

```bash
bb commit report view my-report --commit abc1234 --repo workspace-slug/repo-slug
bb commit report view my-report --commit https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'
```

## `bb commit statuses`

```bash
bb commit statuses abc1234 --repo workspace-slug/repo-slug
bb commit checks https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json statuses
bb commit statuses abc1234 --repo workspace-slug/repo-slug --limit 50
```

## `bb commit unapprove`

```bash
bb commit unapprove abc1234 --repo workspace-slug/repo-slug
bb commit unapprove https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'
```

## `bb commit view`

```bash
bb commit view abc1234 --repo workspace-slug/repo-slug
bb commit view https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'
bb commit view abc1234
```

## `bb completion bash`

```bash
bb completion bash
```

## `bb completion fish`

```bash
bb completion fish
```

## `bb completion powershell`

```bash
bb completion powershell
```

## `bb completion zsh`

```bash
bb completion zsh
```

## `bb config get`

```bash
bb config get prompt
bb config get browser
bb config get output.format --json
```

## `bb config list`

```bash
bb config list
bb config list --json
```

## `bb config set`

```bash
bb config set prompt false
bb config set browser 'firefox --new-window'
bb config set output.format json
bb config get output.format
```

## `bb config unset`

```bash
bb config unset prompt
bb config unset browser
bb config unset output.format
```

## `bb deployment environment list`

```bash
bb deployment environment list --repo workspace-slug/pipelines-repo-slug
bb deployment environment list --repo workspace-slug/pipelines-repo-slug --json environments
```

## `bb deployment environment variable create`

```bash
bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key APP_ENV --value production
bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key SECRET_TOKEN --value-file secret.txt --secured
printf 'secret\n' | bb deployment environment variable create --repo workspace-slug/pipelines-repo-slug --environment test --key SECRET_TOKEN --value-file - --json '*'
```

## `bb deployment environment variable delete`

```bash
bb deployment environment variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --environment test --yes
bb --no-prompt deployment environment variable delete '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test --yes --json '*'
```

## `bb deployment environment variable edit`

```bash
bb deployment environment variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --environment test --value staging
bb deployment environment variable edit '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test --secured true --json '*'
```

## `bb deployment environment variable list`

```bash
bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test
bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment '{environment-uuid}' --json variables
```

## `bb deployment environment variable view`

```bash
bb deployment environment variable view '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test
bb deployment environment variable view '{variable-uuid}' --repo workspace-slug/pipelines-repo-slug --environment test --json variable
```

## `bb deployment environment view`

```bash
bb deployment environment view test --repo workspace-slug/pipelines-repo-slug
bb deployment environment view '{environment-uuid}' --repo workspace-slug/pipelines-repo-slug --json environment
```

## `bb deployment list`

```bash
bb deployment list --repo workspace-slug/pipelines-repo-slug
bb deployment list --repo workspace-slug/pipelines-repo-slug --json deployments
```

## `bb deployment view`

```bash
bb deployment view '{deployment-uuid}' --repo workspace-slug/pipelines-repo-slug
bb deployment view '{deployment-uuid}' --repo workspace-slug/pipelines-repo-slug --json deployment
```

## `bb issue attachment list`

```bash
bb issue attachment list 1 --repo workspace-slug/issues-repo-slug
bb issue attachment list https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'
```

## `bb issue attachment upload`

```bash
bb issue attachment upload 1 ./trace.txt --repo workspace-slug/issues-repo-slug
bb issue attachment upload https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 ./trace.txt ./screenshot.png --json '*'
```

## `bb issue comment create`

```bash
bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body 'Needs follow-up'
bb issue comment create https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --body-file comment.md --json '*'
printf 'Needs follow-up\n' | bb issue comment create 1 --repo workspace-slug/issues-repo-slug --body-file -
```

## `bb issue comment delete`

```bash
bb issue comment delete 3 --issue 1 --repo workspace-slug/issues-repo-slug --yes
bb --no-prompt issue comment delete 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --yes --json '*'
```

## `bb issue comment edit`

```bash
bb issue comment edit 3 --issue 1 --repo workspace-slug/issues-repo-slug --body 'Updated feedback'
bb issue comment edit 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --body-file comment.md --json '*'
```

## `bb issue comment list`

```bash
bb issue comment list 1 --repo workspace-slug/issues-repo-slug
bb issue comment list https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'
```

## `bb issue comment view`

```bash
bb issue comment view 3 --issue 1 --repo workspace-slug/issues-repo-slug
bb issue comment view 3 --issue https://bitbucket.org/workspace-slug/issues-repo-slug/issues/1 --json '*'
```

## `bb issue component list`

```bash
bb issue component list --repo workspace-slug/issues-repo-slug
bb issue component list --repo workspace-slug/issues-repo-slug --json '*'
```

## `bb issue component view`

```bash
bb issue component view 1 --repo workspace-slug/issues-repo-slug
bb issue component view 1 --repo workspace-slug/issues-repo-slug --json '*'
```

## `bb issue create`

```bash
bb issue create --repo workspace-slug/issues-repo-slug --title 'Broken flow'
bb issue create --repo workspace-slug/repo-slug --title 'Broken flow' --body 'Needs investigation'
bb issue create --title 'Request' --kind proposal --priority major --json
```

## `bb issue edit`

```bash
bb issue edit 1 --repo workspace-slug/issues-repo-slug --title 'Updated title'
bb issue edit 1 --repo workspace-slug/repo-slug --state open --priority major --json
```

## `bb issue list`

```bash
bb issue list --repo workspace-slug/issues-repo-slug
bb issue list --repo workspace-slug/repo-slug
bb issue list --state open --json id,title,state
```

## `bb issue milestone list`

```bash
bb issue milestone list --repo workspace-slug/issues-repo-slug
bb issue milestone list --repo workspace-slug/issues-repo-slug --json '*'
```

## `bb issue milestone view`

```bash
bb issue milestone view 1 --repo workspace-slug/issues-repo-slug
bb issue milestone view 1 --repo workspace-slug/issues-repo-slug --json '*'
```

## `bb issue view`

```bash
bb issue view 1 --repo workspace-slug/issues-repo-slug
bb issue view 1 --repo workspace-slug/repo-slug --json
```

## `bb pipeline cache list`

```bash
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug
bb pipeline cache list --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline list`

```bash
bb pipeline list --repo workspace-slug/repo-slug
bb pipeline list --repo workspace-slug/repo-slug --state COMPLETED --json build_number,state,target
bb pipeline list --limit 5
```

## `bb pipeline log`

```bash
bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug
bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
bb pipeline log 42 --repo workspace-slug/pipelines-repo-slug --json pipeline,step,log
```

## `bb pipeline run`

```bash
bb pipeline run main --repo workspace-slug/pipelines-repo-slug
bb pipeline run --repo workspace-slug/pipelines-repo-slug --ref main --json '*'
bb pipeline run v1.2.3 --ref-type tag --repo workspace-slug/pipelines-repo-slug
```

## `bb pipeline runner list`

```bash
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug
bb pipeline runner list --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline runner view`

```bash
bb pipeline runner view '{runner-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline runner view '{runner-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline schedule create`

```bash
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --cron '0 0 12 * * ? *'
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref release --cron '0 30 9 * * ? *' --enabled=false --json '*'
bb pipeline schedule create --repo workspace-slug/pipelines-repo-slug --ref main --selector-type custom --selector-pattern nightly --cron '0 0 1 * * ? *'
```

## `bb pipeline schedule delete`

```bash
bb pipeline schedule delete '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --yes
bb --no-prompt pipeline schedule delete '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'
```

## `bb pipeline schedule disable`

```bash
bb pipeline schedule disable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule disable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline schedule enable`

```bash
bb pipeline schedule enable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule enable '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline schedule list`

```bash
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule list --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline schedule view`

```bash
bb pipeline schedule view '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug
bb pipeline schedule view '{schedule-uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline stop`

```bash
bb pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes
bb pipeline stop '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
bb --no-prompt pipeline stop 42 --repo workspace-slug/pipelines-repo-slug --yes --json pipeline,stopped
```

## `bb pipeline test-reports`

```bash
bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug
bb pipeline test-reports 42 --repo workspace-slug/pipelines-repo-slug --cases --limit 50 --json '*'
bb pipeline test-reports '{uuid}' --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}'
```

## `bb pipeline variable create`

```bash
bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file secret.txt --secured
printf 'token-value\n' | bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key CI_TOKEN --value-file - --json '*'
bb pipeline variable create --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value production
```

## `bb pipeline variable delete`

```bash
bb pipeline variable delete CI_TOKEN --repo workspace-slug/pipelines-repo-slug --yes
bb --no-prompt pipeline variable delete '{uuid}' --repo workspace-slug/pipelines-repo-slug --yes --json '*'
bb pipeline variable delete APP_ENV --repo workspace-slug/pipelines-repo-slug --yes
```

## `bb pipeline variable edit`

```bash
bb pipeline variable edit CI_TOKEN --repo workspace-slug/pipelines-repo-slug --value-file secret.txt --secured true
bb pipeline variable edit '{uuid}' --repo workspace-slug/pipelines-repo-slug --key APP_ENV --value staging --json '*'
bb pipeline variable edit APP_ENV --repo workspace-slug/pipelines-repo-slug --value production
```

## `bb pipeline variable list`

```bash
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug
bb pipeline variable list --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline variable view`

```bash
bb pipeline variable view CI_TOKEN --repo workspace-slug/pipelines-repo-slug
bb pipeline variable view '{uuid}' --repo workspace-slug/pipelines-repo-slug --json '*'
```

## `bb pipeline view`

```bash
bb pipeline view 42 --repo workspace-slug/repo-slug
bb pipeline view '{uuid}' --repo workspace-slug/repo-slug --json '*'
bb pipeline view 42
```

## `bb pr activity`

```bash
bb pr activity 7 --repo workspace-slug/repo-slug
bb pr activity https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --limit 50 --json '*'
bb pr activity https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15
```

## `bb pr checkout`

```bash
bb pr checkout 1
bb pr checkout 1 --repo workspace-slug/repo-slug
bb pr checkout https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
bb pr checkout https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
```

## `bb pr checks`

```bash
bb pr checks 7 --repo workspace-slug/repo-slug
bb pr checks https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'
bb pr statuses https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15
```

## `bb pr close`

```bash
bb pr close 1
bb pr close 1 --repo workspace-slug/repo-slug
bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json
bb pr close https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
```

## `bb pr comment delete`

```bash
bb pr comment delete https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --yes
bb --no-prompt pr comment delete 15 --pr 1 --repo workspace-slug/repo-slug --yes --json '*'
bb pr comment delete 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --yes
```

## `bb pr comment edit`

```bash
bb pr comment edit https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Updated feedback'
bb pr comment edit 15 --pr 1 --repo workspace-slug/repo-slug --body-file comment.md --json '*'
printf 'Updated feedback\n' | bb pr comment edit 15 --pr 1 --repo workspace-slug/repo-slug --body-file -
```

## `bb pr comment reopen`

```bash
bb pr comment reopen https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment reopen 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
bb pr comment reopen 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
```

## `bb pr comment resolve`

```bash
bb pr comment resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment resolve 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
bb pr comment resolve 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
```

## `bb pr comment view`

```bash
bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
bb pr comment view 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
bb pr comment view 15 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
```

## `bb pr commits`

```bash
bb pr commits 7 --repo workspace-slug/repo-slug
bb pr commits https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7 --json '*'
bb pr commits https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --limit 50
```

## `bb pr create`

```bash
bb pr create --title 'Add feature'
bb pr create --source feature --destination main --description 'Ready for review'
bb pr create --reuse-existing --json
```

## `bb pr diff`

```bash
bb pr diff 1
bb pr diff 1 --repo workspace-slug/repo-slug --stat
bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json patch,stats
bb pr diff https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --stat
```

## `bb pr list`

```bash
bb pr list
bb pr list --repo workspace-slug/repo-slug
bb pr list --repo https://bitbucket.org/workspace-slug/repo-slug
bb pr list --state ALL --json id,title,state
```

## `bb pr merge`

```bash
bb pr merge 7
bb pr merge 7 --repo workspace-slug/repo-slug
bb pr merge 7 --strategy merge_commit
bb pr merge 7 --message 'Ship feature' --close-source-branch --json
bb pr merge https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15
```

## `bb pr review approve`

```bash
bb pr review approve 7 --repo workspace-slug/repo-slug
bb pr review approve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review approve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

## `bb pr review clear-request-changes`

```bash
bb pr review clear-request-changes 7 --repo workspace-slug/repo-slug
bb pr review clear-request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review clear-request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

## `bb pr review request-changes`

```bash
bb pr review request-changes 7 --repo workspace-slug/repo-slug
bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review request-changes https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

## `bb pr review unapprove`

```bash
bb pr review unapprove 7 --repo workspace-slug/repo-slug
bb pr review unapprove https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb pr review unapprove https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
```

## `bb pr status`

```bash
bb pr status
bb pr status --repo workspace-slug/repo-slug
bb pr status --json current_branch,created,review_requested
```

## `bb pr task create`

```bash
bb pr task create 1 --repo workspace-slug/repo-slug --body 'Follow up on review feedback'
bb pr task create 1 --repo workspace-slug/repo-slug --comment 15 --body-file task.md --json '*'
bb pr task create https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --comment https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --body 'Handle this thread'
```

## `bb pr task delete`

```bash
bb pr task delete 3 --pr 1 --repo workspace-slug/repo-slug --yes
bb --no-prompt pr task delete 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --yes --json '*'
```

## `bb pr task edit`

```bash
bb pr task edit 3 --pr 1 --repo workspace-slug/repo-slug --body 'Updated follow-up'
bb pr task edit 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --body-file task.md --json '*'
```

## `bb pr task list`

```bash
bb pr task list 1 --repo workspace-slug/repo-slug
bb pr task list https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --state all --json '*'
bb pr task list 1 --repo workspace-slug/repo-slug --state resolved --limit 50
```

## `bb pr task reopen`

```bash
bb pr task reopen 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task reopen 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'
```

## `bb pr task resolve`

```bash
bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task resolve 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'
```

## `bb pr task view`

```bash
bb pr task view 3 --pr 1 --repo workspace-slug/repo-slug
bb pr task view 3 --pr https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1 --json '*'
```

## `bb pr view`

```bash
bb pr view 1
bb pr view 1 --json title,state,source,destination
bb pr view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1
bb pr view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15
```

## `bb project create`

```bash
bb project create BBCLI --workspace workspace-slug --name 'bb cli integration'
bb project create DEMO --workspace workspace-slug --name 'Demo' --visibility private --json '*'
bb project create TMP --name 'Temp project'
```

## `bb project default-reviewer list`

```bash
bb project default-reviewer list BBCLI --workspace workspace-slug
bb project default-reviewer list BBCLI --workspace workspace-slug --json default_reviewers
```

## `bb project delete`

```bash
bb project delete TMP --workspace workspace-slug --yes
bb --no-prompt project delete TMP --workspace workspace-slug --yes --json '*'
```

## `bb project edit`

```bash
bb project edit BBCLI --workspace workspace-slug --description 'Updated by automation'
bb project edit BBCLI --workspace workspace-slug --visibility public --json '*'
bb project edit BBCLI --name 'New project name'
```

## `bb project list`

```bash
bb project list workspace-slug
bb project list --workspace workspace-slug --limit 50
bb project list workspace-slug --json projects
```

## `bb project permissions group list`

```bash
bb project permissions group list BBCLI --workspace workspace-slug
bb project permissions group list BBCLI --workspace workspace-slug --json permissions
```

## `bb project permissions group view`

```bash
bb project permissions group view BBCLI developers --workspace workspace-slug
bb project permissions group view BBCLI developers --workspace workspace-slug --json permission
```

## `bb project permissions user list`

```bash
bb project permissions user list BBCLI --workspace workspace-slug
bb project permissions user list BBCLI --workspace workspace-slug --json permissions
```

## `bb project permissions user view`

```bash
bb project permissions user view BBCLI 557058:example --workspace workspace-slug
bb project permissions user view BBCLI 557058:example --workspace workspace-slug --json permission
```

## `bb project view`

```bash
bb project view BBCLI --workspace workspace-slug
bb project view BBCLI --workspace workspace-slug --json project
```

## `bb repo clone`

```bash
bb repo clone workspace-slug/repo-slug
bb repo clone --repo workspace-slug/repo-slug ./tmp/repo
bb repo clone repo-slug --workspace workspace-slug
bb repo clone https://bitbucket.org/workspace-slug/repo-slug
bb repo clone workspace-slug/repo-slug ./tmp/repo --json
```

## `bb repo create`

```bash
bb repo create workspace-slug/my-repo --project-key BBCLI
bb repo create --repo workspace-slug/my-repo --reuse-existing --json
bb repo create my-repo --workspace workspace-slug
```

## `bb repo delete`

```bash
bb repo delete workspace-slug/delete-repo-slug --yes
bb repo delete --repo workspace-slug/delete-repo-slug --yes
bb repo delete delete-repo-slug --workspace workspace-slug --yes
bb repo delete https://bitbucket.org/workspace-slug/delete-repo-slug --json
```

## `bb repo deploy-key create`

```bash
bb repo deploy-key create --repo workspace-slug/repo-slug --label ci --key-file ./id_ed25519.pub
bb repo deploy-key create --repo workspace-slug/repo-slug --label ci --key 'ssh-ed25519 AAAA...' --json key
```

## `bb repo deploy-key delete`

```bash
bb repo deploy-key delete 7 --repo workspace-slug/repo-slug --yes
bb --no-prompt repo deploy-key delete 7 --repo workspace-slug/repo-slug --yes --json '*'
```

## `bb repo deploy-key list`

```bash
bb repo deploy-key list --repo workspace-slug/repo-slug
bb repo deploy-key list --repo workspace-slug/repo-slug --json keys
```

## `bb repo deploy-key view`

```bash
bb repo deploy-key view 7 --repo workspace-slug/repo-slug
bb repo deploy-key view 7 --repo workspace-slug/repo-slug --json key
```

## `bb repo edit`

```bash
bb repo edit workspace-slug/repo-slug --description 'Updated description'
bb repo edit --repo workspace-slug/repo-slug --visibility public --json '*'
bb repo edit repo-slug --workspace workspace-slug --name 'Renamed repo'
```

## `bb repo fork`

```bash
bb repo fork workspace-slug/repo-slug --to-workspace workspace-slug --name repo-slug-fork
bb repo fork --repo workspace-slug/repo-slug --to-workspace other-workspace --reuse-existing --json '*'
bb repo fork repo-slug --workspace workspace-slug --to-workspace workspace-slug --name repo-slug-fork
```

## `bb repo hook create`

```bash
bb repo hook create --repo workspace-slug/repo-slug --url https://example.com/hook --event repo:push
bb repo hook create --repo workspace-slug/repo-slug --url https://example.com/hook --event pullrequest:created --event pullrequest:updated --json hook
```

## `bb repo hook delete`

```bash
bb repo hook delete {hook-uuid} --repo workspace-slug/repo-slug --yes
bb --no-prompt repo hook delete {hook-uuid} --repo workspace-slug/repo-slug --yes --json '*'
```

## `bb repo hook edit`

```bash
bb repo hook edit {hook-uuid} --repo workspace-slug/repo-slug --description 'Updated hook'
bb repo hook edit {hook-uuid} --repo workspace-slug/repo-slug --event repo:push --event pullrequest:created --json hook
```

## `bb repo hook list`

```bash
bb repo hook list --repo workspace-slug/repo-slug
bb repo hook list --repo workspace-slug/repo-slug --json hooks
```

## `bb repo hook view`

```bash
bb repo hook view {hook-uuid} --repo workspace-slug/repo-slug
bb repo hook view {hook-uuid} --repo workspace-slug/repo-slug --json hook
```

## `bb repo list`

```bash
bb repo list workspace-slug
bb repo list --workspace workspace-slug --limit 50
bb repo list workspace-slug --query 'name ~ "bb"' --json repos
```

## `bb repo permissions group list`

```bash
bb repo permissions group list --repo workspace-slug/repo-slug
bb repo permissions group list --repo workspace-slug/repo-slug --json permissions
```

## `bb repo permissions group view`

```bash
bb repo permissions group view developers --repo workspace-slug/repo-slug
bb repo permissions group view developers --repo workspace-slug/repo-slug --json permission
```

## `bb repo permissions user list`

```bash
bb repo permissions user list --repo workspace-slug/repo-slug
bb repo permissions user list --repo workspace-slug/repo-slug --json permissions
```

## `bb repo permissions user view`

```bash
bb repo permissions user view 557058:example --repo workspace-slug/repo-slug
bb repo permissions user view 557058:example --repo workspace-slug/repo-slug --json permission
```

## `bb repo view`

```bash
bb repo view
bb repo view --repo workspace-slug/repo-slug
bb repo view --repo https://bitbucket.org/workspace-slug/repo-slug
bb repo view --json name,project_key,main_branch
```

## `bb resolve`

```bash
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
bb resolve https://bitbucket.org/workspace-slug/repo-slug/src/main/README.md#lines-12 --json type,repo,path,line
```

## `bb search issues`

```bash
bb search issues fixture --repo workspace-slug/issues-repo-slug
bb search issues bug --repo workspace-slug/issues-repo-slug --json id,title,state
```

## `bb search prs`

```bash
bb search prs fixture --repo workspace-slug/repo-slug
bb search prs feature --repo workspace-slug/repo-slug --json id,title,state
```

## `bb search repos`

```bash
bb search repos integration --workspace workspace-slug
bb search repos bb-cli --workspace workspace-slug --json name,slug,description
```

## `bb status`

```bash
bb status
bb status --workspace workspace-slug --limit 10
bb status --json authored_prs,review_requested_prs,your_issues
```

## `bb tag create`

```bash
bb tag create v1.0.0 --repo workspace-slug/repo-slug --target main --message 'release'
bb tag create v1.0.0 --repo workspace-slug/repo-slug --target abc1234 --json '*'
```

## `bb tag delete`

```bash
bb tag delete v1.0.0 --repo workspace-slug/repo-slug --yes
bb --no-prompt tag delete v1.0.0 --repo workspace-slug/repo-slug --yes --json '*'
```

## `bb tag list`

```bash
bb tag list workspace-slug/repo-slug
bb tag list --repo workspace-slug/repo-slug --limit 50
bb tag list --repo workspace-slug/repo-slug --query 'name ~ "v1"' --json tags
```

## `bb tag view`

```bash
bb tag view v1.0.0 --repo workspace-slug/repo-slug
bb tag view release-2026 --repo workspace-slug/repo-slug --json '*'
```

## `bb version`

```bash
bb version
bb version --json
```

## `bb workspace list`

```bash
bb workspace list
bb workspace list --json workspaces
bb workspace list --jq '.workspaces[].slug'
```

## `bb workspace member list`

```bash
bb workspace member list workspace-slug
bb workspace member list --workspace workspace-slug --query 'user.account_id="123"'
bb workspace member list workspace-slug --json members
```

## `bb workspace member view`

```bash
bb workspace member view 557058:example --workspace workspace-slug
bb workspace member view '{account-uuid}' --workspace workspace-slug --json membership
```

## `bb workspace permission list`

```bash
bb workspace permission list workspace-slug
bb workspace permission list --workspace workspace-slug --query 'permission="owner"'
bb workspace permission list workspace-slug --json members
```

## `bb workspace permission view`

```bash
bb workspace permission view 557058:example --workspace workspace-slug
bb workspace permission view '{account-uuid}' --workspace workspace-slug --json membership
```

## `bb workspace repo-permission list`

```bash
bb workspace repo-permission list workspace-slug
bb workspace repo-permission list workspace-slug --repo repo-slug
bb workspace repo-permission list --workspace workspace-slug --repo workspace-slug/repo-slug --json permissions
```

## `bb workspace view`

```bash
bb workspace view workspace-slug
bb workspace view --workspace workspace-slug --json workspace
bb workspace view --json '*'
```
