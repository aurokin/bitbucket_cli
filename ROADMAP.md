# Roadmap

## Next Commands

1. Workspace and project administration
   - `bb workspace list`
   - `bb workspace members`
   - `bb workspace permissions`
   - `bb project`
   - Backed by the official Bitbucket Cloud workspace and project APIs.

## Later Phase

1. Search expansion
   - `bb search code`
   - Deeper workspace-scoped search workflows where the official Cloud search APIs remain usable and stable.

2. Downloads, deployments, and environments
   - `bb download`
   - `bb deployment`
   - Environment and deployment variable support where the official Cloud APIs are clear.

3. Snippets
   - `bb snippet`
   - Snippet comments, history, file views, and watch support.

4. Platform-limit follow-up
   - Keep Bitbucket Cloud issue import/export out of scope while the documented endpoints reject API-token auth.
   - Keep repository deploy-key updates out of scope while Bitbucket rejects live deploy-key update requests; use delete plus create for rotation.
   - Keep repository permission mutation out of scope while Bitbucket's permission write/delete behavior remains app-password-oriented in the published docs and unverified for API-token auth.
   - Keep commit report mutation out of scope while the API-token path remains unverified as a supported `bb` workflow.
   - Keep branch restrictions and branching model work for a later pass after the core `branch` and `tag` workflows.
   - Continue documenting official Bitbucket Cloud limits instead of faking unsupported behavior.
   - Keep unsupported items like PR reopen, PR comment likes, and undocumented pipeline rerun out of scope unless Atlassian adds official API support.
