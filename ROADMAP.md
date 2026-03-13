# Roadmap

## Next Commands

1. Pipeline actions and diagnostics
   - `bb pipeline schedules`
   - `bb pipeline runners`
   - `bb pipeline caches`
   - Backed by the official Bitbucket Cloud pipeline schedule, runner, and cache APIs.

2. Issue collaboration depth
   - `bb issue comment`
   - `bb issue attachments`
   - `bb issue milestones`
   - `bb issue components`
   - `bb issue import`
   - `bb issue export`
   - Backed by the official Bitbucket Cloud issue tracker APIs for comments, attachments, milestones, components, and import/export.

3. Repository administration
   - `bb repo edit`
   - `bb repo fork`
   - `bb repo list`
   - `bb repo hooks`
   - `bb repo deploy-key`
   - `bb repo permissions`
   - Backed by the official Bitbucket Cloud repository, webhook, deploy key, and permission APIs.

4. Commit and code insight surfaces
   - `bb commit view`
   - `bb commit diff`
   - `bb commit comments`
   - `bb commit approve`
   - `bb commit statuses`
   - `bb commit report`
   - Backed by the official Bitbucket Cloud commit, commit status, and report APIs.

## Later Phase

1. Refs and branch/tag workflows
   - `bb branch`
   - `bb tag`
   - Branch restrictions and branching model support where the official Cloud APIs make sense.

2. Workspace and project administration
   - `bb workspace list`
   - `bb workspace members`
   - `bb workspace permissions`
   - `bb project`
   - Backed by the official Bitbucket Cloud workspace and project APIs.

3. Search expansion
   - `bb search code`
   - Deeper workspace-scoped search workflows where the official Cloud search APIs remain usable and stable.

4. Downloads, deployments, and environments
   - `bb download`
   - `bb deployment`
   - Environment and deployment variable support where the official Cloud APIs are clear.

5. Snippets
   - `bb snippet`
   - Snippet comments, history, file views, and watch support.

6. Platform-limit follow-up
   - Continue documenting official Bitbucket Cloud limits instead of faking unsupported behavior.
   - Keep unsupported items like PR reopen, PR comment likes, and undocumented pipeline rerun out of scope unless Atlassian adds official API support.
