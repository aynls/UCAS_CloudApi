# AxonHub local model context declarations

User request: advertise 196K, rather than 1M, for the local DeepSeek and Qwen
models. Target value: **196608 tokens**.

Production configuration changes on Tencent (2026-10-19):

| Active channel | Exact model ID | Metadata override |
| --- | --- | --- |
| 39 | Qwen3.8-Flash-Next-NVFP4 | limit.context = 196608 |
| 45 | DeepSeek-V4.1-Flash-1017 | limit.context = 196608 |

Both are contributor-owned local channels pointing to the existing 8443 relay.
Archived channel37 is retained unchanged. These models have no explicit global
Model records, so their previous display value was AxonHub's generic 1000000
fallback, not the backend's current public context declaration.

Only `settings.modelMetadataOverrides[exactModelID].limit.context` and each
affected channel's `updated_at` were changed. The transaction reread the rows,
checked ownership, enabled state, model membership and destination, and preserved
all other settings. Runtime code, routing, credentials, outputs and inference
limits were not changed. AxonHub's existing one-minute channel-cache refresh
applies the metadata without a service restart.

Before/after settings backup (root-only, mode0600, not committed):
`/opt/axonhub/backups/local-context-196608-20261019-035250.json`.

Verification uses the production AxonHub endpoint
`https://holocortexzero.com:10443/v1/models?include=all`, checking the two exact
IDs' `context_length`. The API uses the same resolved metadata as the model cards.
This is a display/discovery update, not a new preflight input-token rejection rule.

Verified after the existing cache refresh at approximately 11:53 Asia/Shanghai:
the authenticated production 10443 catalog returned `context_length: 196608` for
both exact model IDs. AxonHub's container start time remained
2026-10-02T14:38:20.682493078Z with restart count0. No rebuild or restart was needed.

Rollback: under a short transaction, remove only the context override added to
each target model (the backup records that neither previously had an override).
Preserve any later unrelated settings or other model fields. Let the existing
channel cache refresh, then recheck the catalog. Never restore the complete old
settings blob over subsequent edits or restart inference merely to change metadata.
