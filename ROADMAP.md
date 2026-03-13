# Roadmap

## Next Commands

1. Snippets
   - `bb snippet`
   - Snippet comments, history, file views, and watch support.

## Later Phase

1. Platform-limit follow-up
   - Keep Bitbucket Cloud issue import/export out of scope while the documented endpoints reject API-token auth.
   - Keep `bb search code` out of scope while the official workspace code-search endpoint is not reliably enabled on the verified API-token path.
   - Keep repository downloads out of scope while the verified API-token path on the current fixture workspace returns a workspace-plan `402 Payment Required` response.
   - Keep repository deploy-key updates out of scope while Bitbucket rejects live deploy-key update requests; use delete plus create for rotation.
   - Keep repository permission mutation out of scope while Bitbucket's permission write/delete behavior remains app-password-oriented in the published docs and unverified for API-token auth.
   - Keep project permission mutation out of scope while the API-token path remains unverified as a documented write workflow.
   - Keep commit report mutation out of scope while the API-token path remains unverified as a supported `bb` workflow.
   - Keep branch restrictions and branching model work for a later pass after the core `branch` and `tag` workflows.
   - Continue documenting official Bitbucket Cloud limits instead of faking unsupported behavior.
   - Keep unsupported items like PR reopen, PR comment likes, and undocumented pipeline rerun out of scope unless Atlassian adds official API support.
