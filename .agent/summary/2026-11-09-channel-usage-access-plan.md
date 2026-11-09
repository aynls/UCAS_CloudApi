# Approved channel usage access implementation

User approved implementation and scoped release on 2026-11-09. Base: 68e0aa43; clean branch codex/channel-usage-access-20261109. Runtime baseline: 33b35b82-security-attested on Tencent VM-0-9-ubuntu, SSH alias tencent-bj/60024. Protect hosts, FRP, Docker daemon and unrelated services.

Implement atomic all/selected-user channel access in existing settings JSON; UI adds all/selected/self with safe user search and multi-select. Existing channels remain public. Scope is evaluated against API-key owner or logged-in user, never an untrusted supplied identity. Filter candidate routes and recheck authoritative data before dispatch, including retries/sticky routes/tests. Denials return 403 without upstream calls or health degradation. Preserve ACL across settings edits/clear/duplicate, and allow donor/owner management without implicitly granting invocation access.

Available model lists and scoped test targets use permitted channels only. Keep contributor/card introductions public, but omit unauthorized channel models, model-bearing health details and tests. Do not publish allowlists. User directory exposes IDs and public nicknames/aliases only.

Model/user rankings, token statistics, historical activity, quota and donation accounting retain their existing logic and visibility. Only genuinely executed usage contributes as before; do not add visibility filters to historical statistics.

Validate backend identity/ACL persistence/filter/revocation/403, full GraphQL create/edit read/write chain, ordinary-member directory/UI, aggregation invariance and desktop/mobile behavior. Generate GraphQL bindings, build locally, commit/tag/push, deploy only AxonHub with consistent backup, checksum, protected-state guard, public/API smoke checks and artifact cleanup.
