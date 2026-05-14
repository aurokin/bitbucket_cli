# Jira and Confluence True-to-API CLI Specs

> Point-in-time draft: 2026-05-14 12:53 MDT. This document is a research-backed product/spec draft, not a generated reference. Re-check official Atlassian docs before implementation decisions that depend on auth support, endpoint coverage, scopes, or payload shape.

## Goal

Design two separate Go CLIs, one for Jira and one for Confluence, that follow the same philosophy as this repository's `bb` Bitbucket Cloud CLI:

- stay true to official Atlassian REST APIs instead of inventing fake `gh`-style parity
- support humans and agents with deterministic targeting, `--json`, `--jq`, and `--no-prompt`
- accept the authentication modes that Atlassian actually supports for Cloud and Data Center/Server-style organization URLs
- include a raw `api` escape hatch so uncovered official endpoints are still usable
- document unsupported behavior explicitly when Atlassian does not expose a safe API path

This is one combined planning document because the products share auth, output, configuration, and command design. The implementation should still produce two separate CLIs and two separate durable specs after this draft is reviewed:

- Jira CLI: proposed binary `jira` unless a better short name is chosen
- Confluence CLI: proposed binary `confluence`

Naming note: avoid `jj`, `cc`, and `conf`. `jj` collides mentally with the popular Jujutsu VCS CLI/library ecosystem, and `cc` commonly means Claude Code in agent workflows. Prefer names that avoid common developer/agent collisions; `confluence` is long but clear, while `conf` reads too much like config.

## Bitbucket CLI patterns to preserve

The existing `bb` CLI establishes the model we should copy.

### Product posture

`bb` is implemented against official Bitbucket Cloud REST APIs and avoids misleading approximations. The root help tells users to prefer explicit resource targets, structured output, and non-interactive flags. The README and repo instructions say:

- real platform behavior beats cross-product CLI parity
- unsupported platform behavior must be documented instead of silently approximated
- raw `bb api` maps directly onto official REST paths and URLs
- JSON is the stable agent interface
- human-readable output can include `Next:` guidance

### Command surface shape

Current `bb` root command families are:

- `auth`, `config`, `api`, `browse`, `resolve`, `status`, `search`, `alias`, `extension`, `completion`, `version`
- Bitbucket resource families: `workspace`, `project`, `repo`, `branch`, `tag`, `commit`, `pr`, `issue`, `pipeline`, `deployment`

The Jira and Confluence CLIs should keep the first group nearly identical and swap only the product resource families.

### Targeting model

`bb` prefers explicit targets, for example `--repo <workspace>/<repo>`, with local inference as a convenience rather than a requirement. The same should apply here:

- Jira: prefer `--site <site-url-or-alias>` plus explicit issue/project/filter/board keys. Local git inference is not relevant by default.
- Confluence: prefer `--site <site-url-or-alias>` plus explicit `--space`, page ID, database ID, whiteboard ID, or URL.
- If a workflow starts from a pasted Atlassian URL, run `resolve <url> --json '*'` first.

### Output and automation

Preserve these global behaviors in both CLIs:

- `--json [fields|*]` on wrapped commands
- `--jq <expr>` for JSON filtering
- `--no-prompt` for deterministic mutation flows
- table output for humans, JSON as the stable agent contract
- warnings array in JSON when inference was degraded or a product/version limit changed behavior
- stable error/recovery guidance catalog and generated docs when implemented

### Access-aware UX

Different users, service accounts, tokens, groups, and organizations will have different access levels. The CLIs must provide a good experience for the access the authenticated principal actually has, without assuming admin rights, broad product licenses, or full project/space visibility.

Design rules:

- Treat the authenticated account's permissions as the source of truth. If the API returns fewer resources than an admin would see, show the accessible subset without implying the CLI is broken.
- Distinguish **not found** from **not visible** when Atlassian exposes enough signal. When it does not, report the ambiguity honestly, for example: `not found or not visible to this account`.
- Prefer capability discovery over hard-coded assumptions. Commands that depend on optional products, scopes, permissions, or app availability should verify access through the relevant API before performing mutations.
- Make `403`, `401`, scope failures, missing product licenses, and missing project/space permissions first-class recovery cases in the error catalog.
- Keep JSON useful for agents: include `accessible`, `permission_error`, `required_scope`, `required_permission`, `site`, `token_style`, and `api_base_url` fields where known.
- Avoid admin-only defaults. Project, space, user, group, and permission-management commands should fail clearly under ordinary user tokens and suggest read-only alternatives where possible.
- For list/search/status commands, partial visibility is normal. Return the visible results plus warnings only when the API signals that access or scopes limited the result.
- For mutations, prefer preflight checks when cheap, but never rely only on preflight. The final API response remains authoritative because permissions can change between checks.

Example recovery shape:

```json
{
  "error": "permission_denied",
  "message": "The authenticated account cannot edit this Confluence page.",
  "site": "work",
  "token_style": "cloud-scoped",
  "required_scope": "write:page:confluence",
  "required_permission": "page edit permission",
  "next": "Ask a space admin for edit access, choose a token with the required scope, or retry with `confluence page view 123456 --json '*'`."
}
```

## Shared architecture for both CLIs

### Repository layout recommendation

Initial implementation should optimize for learning and product fit, not premature abstraction. It is reasonable to develop Jira and Confluence as separate CLIs first, even if they live in one workspace, then refactor shared code once the repeated seams are proven. A later monorepo can keep auth/output/config/client code DRY while producing multiple binaries.

Long-term roadmap: consider moving the existing Bitbucket CLI into the same Atlassian CLI monorepo after Jira and Confluence have stabilized. Do not force this early. The right time is after we can see which code is genuinely shared across all three products and which code should remain product-specific.

Candidate shared code:

- structured output: table rendering, `--json`, selected-field filtering, `--jq`
- config mechanics: config path, aliases, settings, site/host records, secure file permissions
- auth framework: token storage, request signing, auth status payloads, recovery guidance
- HTTP client: retries where safe, rate-limit handling, API errors, request/response tracing hooks
- pagination primitives: native product pagination adapters, not a single fake pagination model
- raw `api` command scaffolding
- URL `resolve` and `browse` framework with product-specific parsers
- generated docs/man/completions/metadata pipeline
- testing helpers: fixture HTTP servers, golden output, JSON field tests, live-test gating

Code that should stay product-specific until proven otherwise:

- resource models and payload structs
- command trees and command vocabulary
- product-specific permissions and scope mapping
- Jira workflow/transition semantics
- Confluence body-format and version semantics
- Bitbucket repository/PR/pipeline semantics

Recommended split if staying close to `bitbucket_cli`:

```text
atlassian_cli_specs_or_repo/
├── cmd/jira/               # Jira binary
├── cmd/confluence/               # Confluence binary
├── internal/atlassian/     # shared auth, site resolution, HTTP, pagination
├── internal/output/        # shared table/json/jq renderer, copied/adapted from bb
├── internal/config/        # shared config with product-specific file names
├── internal/jira/          # Jira client and models
├── internal/confluence/    # Confluence client and models
├── internal/jiracmd/       # Cobra command tree for jira
└── internal/confluencecmd/       # Cobra command tree for confluence
```

If implemented as separate repos, keep `internal/output`, auth, pagination, and config conventions synchronized through copy-with-tests or a small shared module.

### Config contract

The current `bb` config stores host credentials under the user config directory. For Jira/Confluence we need product-specific config files because the same Atlassian site can host both products and may need different API bases.

Suggested files:

- Jira: `${XDG_CONFIG_HOME:-~/.config}/jira/config.json`
- Confluence: `${XDG_CONFIG_HOME:-~/.config}/confluence/config.json`

Suggested shared schema:

```json
{
  "default_site": "work",
  "sites": {
    "work": {
      "base_url": "https://example.atlassian.net",
      "deployment": "cloud",
      "product": "jira",
      "auth_type": "api-token-basic",
      "token_style": "cloud-scoped",
      "username": "user@example.com",
      "token": "...",
      "cloud_id": "required-for-cloud-scoped-and-oauth",
      "api_base_url": "https://api.atlassian.com/ex/jira/<cloudId>",
      "updated_at": "2026-05-14T18:53:00Z"
    }
  },
  "settings": {
    "prompt": true,
    "browser": "",
    "editor": "",
    "pager": "",
    "output_format": ""
  },
  "aliases": {}
}
```

Notes:

- `base_url` is required for Jira and Confluence because Cloud classic-token REST calls are site-scoped and org URLs matter for Data Center.
- `api_base_url` is the effective REST gateway base after auth style is known. For Cloud scoped API tokens this uses `https://api.atlassian.com/ex/jira/<cloudId>` or `https://api.atlassian.com/ex/confluence/<cloudId>`. For Cloud classic tokens it uses the site URL. For Data Center PATs it uses the provided instance URL.
- `deployment` should be `cloud`, `data-center`, or `auto`. `auto` can be resolved during `auth status --check` and stored after validation.
- `token_style` should be `cloud-classic`, `cloud-scoped`, `data-center-pat`, `oauth-3lo`, or `auto`. Because Atlassian says integrations cannot distinguish scoped vs non-scoped Cloud API tokens from the token value alone, `auto` should validate by trying the configured/default URL strategy and then produce clear recovery guidance.
- Never store passwords. The legacy Atlassian Cloud username/password REST path is deprecated or disabled; support email/username + API token as Basic auth, not account password.
- For Data Center/Server-like URLs, basic auth with username/password may be present in old installations, but the CLI should not prompt for or store passwords by default. Prefer PAT bearer auth.

### Authentication contract

Support all auth modes that are true to Atlassian product APIs and practical for CLIs.

#### Cloud API tokens: classic/unscoped and scoped

Official Atlassian Cloud docs now describe two Cloud API-token styles for Jira and Confluence:

1. **Classic / general / non-scoped API tokens**
   - Created from the Atlassian account API-token page.
   - Grant the permissions available to the user, without per-token scope restriction.
   - Use HTTP Basic auth with `email@example.com:api_token`.
   - Use site-specific product URLs, for example Jira `https://<site-url>/rest/api/3/...`, Confluence v2 `https://<site>.atlassian.net/wiki/api/v2/...`, and Confluence v1 `https://<site>.atlassian.net/wiki/rest/api/...`.

2. **Scoped API tokens**
   - Created from the same Atlassian account token flow by choosing "Create API token with scopes" and selecting Jira or Confluence scopes.
   - Recommended by Atlassian because they limit token permissions.
   - Use HTTP Basic auth with `email@example.com:api_token`; this is still an API-token Basic auth path, not OAuth bearer auth.
   - Require the Atlassian API gateway URL with a Cloud ID, not the site-specific `atlassian.net` URL:
     - Jira: `https://api.atlassian.com/ex/jira/<cloudId>/<api>`
     - Confluence: `https://api.atlassian.com/ex/confluence/<cloudId>/<api>`
   - Service accounts can only create scoped API tokens.

Important implementation detail: Atlassian's Cloud 401 guidance says integrations generally cannot distinguish scoped vs non-scoped API tokens from the token value itself. If an integration asks for an API token without saying which type, Atlassian says it is usually asking for a non-scoped token. Our CLIs should therefore make token style explicit and validate the URL shape.

CLI behavior:

```bash
# Classic/general Cloud token: site-specific URLs.
jira auth login --site https://example.atlassian.net --username user@example.com --token-style cloud-classic --with-token
jira auth login --site work --url https://example.atlassian.net --username user@example.com --token-style cloud-classic --token "$ATLASSIAN_API_TOKEN"
JIRA_USERNAME=user@example.com JIRA_TOKEN="$ATLASSIAN_API_TOKEN" jira auth login --site work --url https://example.atlassian.net --token-style cloud-classic

confluence auth login --site https://example.atlassian.net/wiki --username user@example.com --token-style cloud-classic --with-token
CONFLUENCE_USERNAME=user@example.com CONFLUENCE_TOKEN="$ATLASSIAN_API_TOKEN" confluence auth login --site work --url https://example.atlassian.net/wiki --token-style cloud-classic

# Scoped Cloud token: Atlassian API gateway URLs and cloudId required.
jira auth login --site work --url https://example.atlassian.net --username user@example.com --token-style cloud-scoped --cloud-id "$ATLASSIAN_CLOUD_ID" --with-token
confluence auth login --site work --url https://example.atlassian.net/wiki --username user@example.com --token-style cloud-scoped --cloud-id "$ATLASSIAN_CLOUD_ID" --with-token
```

Implementation:

- store `auth_type: "api-token-basic"`
- store `token_style: "cloud-classic"` or `"cloud-scoped"`
- send `Authorization: Basic base64(username:token)` for both token styles
- for Cloud classic Jira, base REST URL is `https://<site>/rest/api/3`
- for Cloud classic Confluence v2, base REST URL is `https://<site>/wiki/api/v2`
- for Cloud classic Confluence v1 fallback, base REST URL is `https://<site>/wiki/rest/api`
- for Cloud scoped Jira, base gateway URL is `https://api.atlassian.com/ex/jira/<cloudId>` plus the API path expected by Atlassian docs
- for Cloud scoped Confluence, base gateway URL is `https://api.atlassian.com/ex/confluence/<cloudId>` plus the API path expected by Atlassian docs
- `auth status --check` should detect 401s that look like "scoped token used against site URL" or "classic token used against gateway URL" and print recovery guidance to rerun `auth login` with the correct `--token-style` and `--cloud-id`.

#### Data Center/Server personal access token via Bearer auth

Official Atlassian Enterprise/Data Center docs support personal access tokens for Jira and Confluence REST calls via bearer token:

```http
Authorization: Bearer <token>
```

The docs also show PAT management at:

```text
<baseUrlOfYourInstance>/rest/pat/latest/tokens
```

CLI behavior:

```bash
jira auth login --site https://jira.example.com --deployment data-center --auth-type pat --with-token
confluence auth login --site https://confluence.example.com --deployment data-center --auth-type pat --with-token
```

Implementation:

- store `auth_type: "pat-bearer"`
- do not require username for request signing
- validate with product-specific `myself` endpoint
- do not assume a Cloud `/wiki` prefix on Data Center Confluence; accept the provided base URL as authoritative

#### OAuth 2.0 3LO

Official Cloud docs recommend OAuth 2.0 3LO for distributable integrations. REST calls use `https://api.atlassian.com/ex/jira/<cloudId>/...` for Jira and product paths for Confluence through the Atlassian API gateway, after discovering accessible resources via `https://api.atlassian.com/oauth/token/accessible-resources`.

This should be a later auth mode, not MVP, because it requires an app/client distribution decision.

Spec requirement:

- do not fake OAuth by asking users to paste arbitrary browser session tokens
- if implemented, use a real 3LO app flow, store refresh/access token metadata securely where possible, and resolve `cloudId`
- expose `auth login --oauth` only after app registration and token refresh are fully implemented

#### Connect JWT and Forge app auth

These are app-framework auth modes, not user CLI auth. Do not implement in the user CLI unless we are building an app runtime. Document as out of scope.

### Global commands for both CLIs

All wrapped commands should accept the shared output flags where applicable.

```text
<bin> auth login
<bin> auth status [--check]
<bin> auth logout
<bin> config get|set|unset|list|path
<bin> api <path-or-url> [-X METHOD] [--input FILE|-] [--site ALIAS|URL] [--jq EXPR]
<bin> browse ... [--no-browser]
<bin> resolve <url> --json '*'
<bin> search ...
<bin> status ...
<bin> alias get|set|delete|list
<bin> extension list|exec
<bin> completion bash|fish|powershell|zsh
<bin> version
```

Raw API behavior:

- If the argument is an absolute URL, call it as given after validating it belongs to the configured site or Atlassian API gateway.
- If the argument starts with `/`, resolve against the product's default REST base.
- Allow product/version selectors when needed:
  - Jira: `jira api /issue/PROJ-1`, defaulting to `/rest/api/3`; maybe `--api-version 2|3` for v2 fallback.
  - Confluence: `confluence api /pages`, defaulting to `/wiki/api/v2`; `--api-version 1` for `/wiki/rest/api` fallback.

## Jira CLI spec (`jira`)

### Official API grounding

Primary Cloud docs to anchor implementation:

- Jira Cloud REST API v3 intro: `https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/`
- Jira Cloud Basic auth: `https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/`
- Atlassian account API tokens, including scoped tokens: `https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/`
- Cloud scoped-token 401 guidance: `https://support.atlassian.com/atlassian-cloud/kb/401-unauthorized-error-when-service-account-accesses-jira-or-confluence-api/`
- Jira Cloud OAuth 2.0 3LO: `https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/`
- Jira issues group: `https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/`
- Jira myself group: `https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/`

Observed official URL/auth rules:

- Basic auth direct REST URL: `https://<site-url>/rest/api/3/<resource>`
- OAuth 3LO REST URL: `https://api.atlassian.com/ex/jira/<cloudId>/rest/api/3/<resource>`
- Direct personal scripts/bots are explicitly pointed to Basic auth with email + API token.

### Jira MVP command tree

```text
jira auth
  login|logout|status
jira api
jira browse
jira resolve
jira config
jira project
  list|view|create|edit|delete
  version list|view|create|edit|delete|archive|release|unrelease
  component list|view|create|edit|delete
  role list|view
jira issue
  list|view|create|edit|transition|assign|delete
  comment list|view|create|edit|delete
  attachment list|upload|download|delete
  link list|create|delete
  remote-link list|create|delete
  watcher list|add|remove
  worklog list|create|edit|delete
jira sprint
  list|view|create|edit|start|close|move-issues
jira board
  list|view|issues|sprints|backlog
jira filter
  list|view|create|edit|delete|favorite|unfavorite
jira search
  issues
  users
  projects
jira status
```

MVP should prioritize:

1. `auth`, `api`, `resolve`, `browse`, `project list/view`, `issue list/view/create/edit/transition/comment`, `search issues`, `status`.
2. Boards/sprints only after confirming Jira Software endpoints and licenses. They are real APIs, but not every Jira site/product exposes them.
3. Admin/configuration-heavy endpoints only after permissions and payloads are verified live.

### Jira resource targeting

Recommended flags:

- `--site <alias-or-url>` on all remote commands
- issue operands accept `PROJ-123` or Jira issue URL
- project operands accept project key, project ID, or project URL where resolvable
- `--jql <query>` for issue searches
- `--project <key-or-id>` where Jira API uses project scoping
- `--account-id <id>` for user targeting, with username/email lookup helpers only where the API supports privacy-safe search

Examples:

```bash
jira issue view PROJ-123 --site work --json '*'
jira issue list --site work --jql 'assignee = currentUser() AND statusCategory != Done' --json issues,total
jira --no-prompt issue transition PROJ-123 --transition 'Done' --json issue,status
jira issue comment create PROJ-123 --body 'Investigating' --json comment
jira project list --site work --json projects
jira api /issue/PROJ-123 --jq '{key, fields: {summary: .fields.summary, status: .fields.status.name}}'
```

### Jira `resolve` behavior

Resolve should parse common Jira URLs into a stable entity object without calling APIs when URL structure is sufficient, then enrich when authenticated.

Inputs:

- `https://example.atlassian.net/browse/PROJ-123`
- `https://example.atlassian.net/jira/software/projects/PROJ/boards/12`
- `https://example.atlassian.net/issues/?jql=...`
- `https://example.atlassian.net/projects/PROJ`

Output sketch:

```json
{
  "product": "jira",
  "site": "https://example.atlassian.net",
  "type": "issue",
  "issue_key": "PROJ-123",
  "api_path": "/issue/PROJ-123",
  "browser_url": "https://example.atlassian.net/browse/PROJ-123",
  "warnings": []
}
```

### Jira unsupported/non-goals until verified

- Do not implement browser cookie/session scraping.
- Do not accept account passwords for Cloud Basic auth. Cloud docs say password auth is deprecated; use API tokens.
- Do not expose user lookup by email unless the Cloud API and privacy settings actually return it for the authenticated user/scopes.
- Do not hide Jira Software vs Jira Work Management vs Jira Service Management product boundaries. If an endpoint needs a product license or JSM REST path, put it under the right command family or document the dependency.
- Do not pretend every workflow has a single `close` or `reopen` command. Jira transitions are workflow-specific. `jira issue transition` should expose real available transitions, and convenience aliases should be explicit helpers over that API.

## Confluence CLI spec (`confluence`)

### Official API grounding

Primary Cloud docs to anchor implementation:

- Confluence Cloud REST API v2 intro: `https://developer.atlassian.com/cloud/confluence/rest/v2/intro/`
- Confluence Cloud REST API v1 intro: `https://developer.atlassian.com/cloud/confluence/rest/v1/intro/`
- Confluence Cloud Basic auth: `https://developer.atlassian.com/cloud/confluence/basic-auth-for-rest-apis/`
- Atlassian account API tokens, including scoped tokens: `https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/`
- Confluence Cloud scoped API tokens: `https://support.atlassian.com/confluence/kb/scoped-api-tokens-in-confluence-cloud/`
- Cloud scoped-token 401 guidance: `https://support.atlassian.com/atlassian-cloud/kb/401-unauthorized-error-when-service-account-accesses-jira-or-confluence-api/`
- Confluence v2 pages group: `https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/`
- Confluence v2 users group: `https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-user/`

Observed official URL/auth rules:

- Cloud Confluence v2 uses cursor-based pagination and `Link`/`_links.next` for next pages.
- Cloud Confluence v1 remains relevant for endpoints not yet present in v2.
- Cloud Basic auth direct REST examples use `https://<site>.atlassian.net/wiki/rest/api/...`.
- v2 API base is `https://<site>.atlassian.net/wiki/api/v2/...`.

### Confluence MVP command tree

```text
confluence auth
  login|logout|status
confluence api
confluence browse
confluence resolve
confluence config
confluence space
  list|view|create|edit|archive|delete
confluence page
  list|view|create|edit|delete|archive|restore
  children|ancestors|descendants
  label list|add|remove
  attachment list|upload|download|delete
  comment list|view|create|edit|delete
  version list|view|restore
  property list|view|set|delete
confluence blog
  list|view|create|edit|delete
confluence user
  current|view|search
confluence group
  list|view|members
confluence search
  cql
  pages
  spaces
confluence status
```

MVP should prioritize:

1. `auth`, `api`, `resolve`, `browse`, `space list/view`, `page list/view/create/edit`, `page children`, `search cql`, `status`.
2. Attachments/comments/labels next because they are common agent workflows.
3. Databases, whiteboards, folders, and admin operations only after v2 endpoint behavior and product availability are verified.

### Confluence resource targeting

Recommended flags:

- `--site <alias-or-url>` on all remote commands
- `--space <key-or-id>` where endpoint supports scoping
- page operands accept page ID or Confluence page URL
- `--title <title>` only with `--space` when resolving by title, because page titles are not globally unique
- `--body-format storage|atlas_doc_format|wiki|view` only where the API supports it; prefer exact API terminology rather than invented names
- `--version-message <message>` for page edits if exposed by endpoint

Examples:

```bash
confluence page view 123456 --site work --json '*'
confluence page list --site work --space ENG --json pages,_links
confluence page create --site work --space ENG --title 'Runbook' --parent 123456 --body-file runbook.html --body-format storage --json page
confluence --no-prompt page edit 123456 --title 'Runbook v2' --body-file runbook.html --body-format storage --json page
confluence search cql 'space = ENG AND title ~ "Runbook"' --json results
confluence api /pages/123456 --jq '{id, title, status}'
confluence api --api-version 1 '/content/123456?expand=body.storage,version' --jq '{id, title, version: .version.number}'
```

### Confluence `resolve` behavior

Resolve should parse common Confluence Cloud URLs and normalize them to IDs/API paths where possible.

Inputs:

- `https://example.atlassian.net/wiki/spaces/ENG/pages/123456/Page+Title`
- `https://example.atlassian.net/wiki/spaces/ENG/overview`
- `https://example.atlassian.net/wiki/search?text=...`
- legacy display URLs if encountered

Output sketch:

```json
{
  "product": "confluence",
  "site": "https://example.atlassian.net/wiki",
  "type": "page",
  "space_key": "ENG",
  "page_id": "123456",
  "api_version": "v2",
  "api_path": "/pages/123456",
  "browser_url": "https://example.atlassian.net/wiki/spaces/ENG/pages/123456/Page+Title",
  "warnings": []
}
```

### Confluence unsupported/non-goals until verified

- Do not silently convert storage HTML, ADF, wiki markup, and editor formats. Expose the exact API representation and require explicit `--body-format`.
- Do not claim perfect round-trip page editing until version handling and conflict behavior are tested. Confluence edits are versioned.
- Do not assume Cloud `/wiki` prefix for Data Center.
- Do not implement browser cookie/session scraping.
- Do not model page title as a stable ID. Prefer page ID once resolved.
- Do not overcommit v2-only coverage. Keep v1 fallback available where official v2 is incomplete.

## Shared pagination and API behavior

Implement product-native pagination helpers rather than one fake shape:

- Jira Cloud REST commonly uses `startAt`, `maxResults`, `total`, and `isLast`-style fields depending on endpoint.
- Confluence Cloud v2 uses cursor-based pagination with `Link` header and `_links.next`.
- Confluence v1 commonly uses `start`, `limit`, `size`, and `_links.next`.

CLI flags:

```text
--limit <n>          # per-page or maximum item count, documented per command
--paginate           # fetch all pages when safe
--cursor <cursor>    # Confluence v2 where applicable
--start-at <n>       # Jira where applicable
```

JSON outputs should preserve native pagination fields under `_pagination` or include raw `_links` where useful. Do not erase Atlassian's pagination model.

## Implementation phases

### Phase 1: Shared foundation

- Copy/adapt `bb` config, output rendering, `--json`, `--jq`, aliases, extension hooks, and API error handling.
- Generalize auth into Cloud `api-token-basic` for both `cloud-classic` and `cloud-scoped` token styles, plus Data Center/Server `pat-bearer` request signing.
- Add site config with `base_url`, `deployment`, and product.
- Implement `auth login/status/logout` for both CLIs.
- Implement raw `api` for both CLIs.
- Implement the shared access/error model for `401`, `403`, scope mismatch, product/license missing, and ambiguous not-found/not-visible cases before wrapping many endpoints.

### Phase 2: Jira MVP

- Implement Jira client base URL resolution for Cloud Basic auth: `<site>/rest/api/3`.
- Validate auth with `GET /myself`.
- Implement `issue view/list/create/edit/transition` and `issue comment`.
- Implement `project list/view`.
- Implement `resolve` and `browse` for issue/project URLs.
- Add generated docs and examples.

### Phase 3: Confluence MVP

- Implement Confluence v2 client base URL: `<site>/wiki/api/v2`.
- Add v1 fallback base: `<site>/wiki/rest/api`.
- Validate auth with current user endpoint.
- Implement `space list/view`, `page list/view/create/edit/children`, and `search cql`.
- Implement `resolve` and `browse` for page/space URLs.
- Add generated docs and examples.

### Phase 4: Broader endpoint coverage

- Jira: boards/sprints, filters, versions, components, attachments, worklogs, watchers.
- Confluence: attachments, comments, labels, versions, page properties, blogs, groups/users.
- Add endpoint support only after matching official docs and tests to real payloads.

### Phase 5: Optional OAuth 3LO

- Decide app registration/distribution model.
- Implement real OAuth device/browser flow only if supported by Atlassian's documented authorization code flow for our CLI needs.
- Discover accessible resources and cloud IDs via Atlassian gateway.
- Keep API-token Basic and PAT Bearer paths working independently.

## Test and verification requirements

- Unit-test URL resolution heavily with messy Atlassian URLs.
- Unit-test auth request signing without hitting the network.
- Use fixture HTTP servers for payload decoding and pagination behavior.
- Keep live tests manual-only unless disposable Atlassian sites and tokens are intentionally provisioned.
- For every mutation, test `--no-prompt` failure when required input is missing and success when supplied.
- For every human output change, test field order and `Next:` hints.
- For every wrapped command, test `--json '*'`, selected fields, and `--jq`.
- Add fixture tests for low-access users, scoped tokens missing required scopes, admin-only endpoints under non-admin tokens, and APIs that collapse unauthorized resources into 404/not-found responses.
- Add manual live-smoke scripts that can be run with read-only, contributor, admin, classic-token, scoped-token, and Data Center PAT accounts when those fixtures are available.

## Initial open questions

1. Binary names: confirm `jira`/`confluence`, or choose another pair that avoids `jj`, `cc`, and `conf` collisions?
2. Repo strategy: separate repos for release clarity first, or shared workspace/monorepo now with the understanding that abstraction should follow proven repetition?
3. OAuth 3LO: do we have or want a distributable Atlassian app registration, or should MVP be Basic API token and Data Center PAT only?
4. Data Center support depth: auth-only plus raw API first, or full wrapped commands against Data Center REST from the start?
5. Confluence body editing: should MVP accept storage HTML only first, or include ADF where official endpoints support it?
6. Bitbucket migration timing: after Jira/Confluence stabilize, should `bb` move into the same Atlassian monorepo, or should shared code be extracted without moving the existing repo?


