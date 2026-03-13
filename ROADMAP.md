# Roadmap

## Next Commands

1. Commit and code insight surfaces
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
   - Keep Bitbucket Cloud issue import/export out of scope while the documented endpoints reject API-token auth.
   - Keep repository deploy-key updates out of scope while Bitbucket rejects live deploy-key update requests; use delete plus create for rotation.
   - Keep repository permission mutation out of scope while Bitbucket's permission write/delete behavior remains app-password-oriented in the published docs and unverified for API-token auth.
   - Continue documenting official Bitbucket Cloud limits instead of faking unsupported behavior.
   - Keep unsupported items like PR reopen, PR comment likes, and undocumented pipeline rerun out of scope unless Atlassian adds official API support.
