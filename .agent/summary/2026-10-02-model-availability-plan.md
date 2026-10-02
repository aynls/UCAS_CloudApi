# Shared model panel and Codex compatibility plan

User authorized implementation on 2026-10-02 after discussing the design. Base application revision: 7a4252c5; branch: codex/model-availability-20261002. Start with a clean working tree.

1. Update Codex's default outbound client version to the upstream-reviewed 0.159.0; preserve explicit client headers and existing customizations. Verify the actual outbound request path.
2. Add project-scoped, permission-checked model success aggregates using existing completed API requests. Use a rolling 24-hour window with 24 hourly buckets as initial UI defaults. Count each request once; exclude failed/canceled requests, test requests and retries. No new database tables or retained request bodies.
3. Put model selection first: full model name, success count/ranking, green-only success history, copy and test actions. Empty history is neutral. Reuse public channel probes with an explicit selected model; results explain their channel scope. All project members can test; no owner-only gate.
4. Demote a model only on fresh dedicated negative test evidence covering its enabled candidate routes, or when it has no enabled candidate channel. Unknown and sparse histories stay discoverable. Initially use a one-hour test freshness window, clearly described in the UI. Success history never counts client errors as failures.
5. Move the existing donor/channel cards below the model section without reducing their content: contributor identity, every channel, supported models, availability, full existing remarks, expiry and quota details remain available. Preserve model capability details behind expansion; preserve onboarding and quota information with less initial visual emphasis.
6. Validate backend aggregation/privacy/alias scopes, frontend ordering and unknown states, member-accessible probes, Codex headers, donor content preservation, responsive layout and the production frontend build. Create a local preview for visual verification.
7. Commit independently reversible changes and tag the reviewed result. Production changes require a concrete verified artifact and an explicit deployment decision; this work must not restart hosts, FRP or unrelated services.

The green history represents observed successful requests, not continuous uptime. Ranking is successful request volume, not user likes. The existing authenticated project boundary remains intact.

Implementation completed locally. See `2026-10-02-model-availability-validation.md` for results and limits. Missing route mappings remain unknown rather than proving unavailability; actual fresh failed route coverage is required for demotion. Model test targets also preserve configured-model associations and per-key channel scope.
