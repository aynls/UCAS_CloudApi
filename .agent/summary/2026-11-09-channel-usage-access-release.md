# Channel usage access release — 2026-11-09

## Deployed result

- Branch: `codex/channel-usage-access-20261109`, pushed to `aynls/UCAS_CloudApi`.
- Production source: `747108205a70fd9a28c695db949ac0441c2a7a82`.
- Version: `v1.0.0-beta5-ucas.18+access.74710820`.
- Image: `ucas-cloudapi:74710820-access`.
- Image ID: `sha256:d90df1cdea081ae186752ae885819c42d1ddf2c596942349adea50adb1d4ae1e`.
- Archive SHA256: `94af8c9a694a18c36828d34d7a34b8845b673fe89ab55b145a4ea69390799639`.
- Binary SHA256: `b5bec28db8398bdeeeea4a5495248a604de02dd36dd57b54c5ab662f3bd1b181`.
- Local clean-source build for Linux amd64, CGO disabled, Go 1.27.3; frontend embedded from the verified production build. GraphQL generation used Go 1.26.10 because the existing generator cannot decode Go 1.27 export data.

## Behavior

Channel settings JSON now stores `usageAccess` with `all` or `selected` and stable user IDs. Missing settings remain public. The UI's “Only me” shortcut stores the current user in the selected list. Empty selected lists, invalid modes and nonexistent/inactive users are rejected. Settings clears, unrelated edits, template updates and duplication preserve restrictions.

The authenticated API-key owner or logged-in user determines invocation access. System owners retain management rights without automatically gaining invocation access. Selection filters current persisted ACLs; transformation and the final executor recheck access, including after queueing. Denials return 403, do not reach upstream and do not trigger failure probes or availability verdicts. Existing streams are not interrupted.

Available model lists retain a model if another authorized channel serves it. Public contributor cards retain attribution and descriptions; unauthorized channel model lists, test actions and model-bearing health details are omitted by the server. Allowlists are restricted to management reads. The paginated member directory returns only ID, public display name and self marker. Existing ranking, historical usage, token, quota and donation accounting queries were not changed.

## Verification

- Full tests passed for `internal/server/biz`, `orchestrator`, `api` and `gql`.
- Added regressions cover API-key identity precedence, legacy public behavior, owner invocation restrictions, atomic validation, settings preservation, duplication, stale-cache revocation, model alternatives, sticky/alias candidate filtering, queued/stream dispatch rejection, public projection, directory privacy and historical token-total invariance.
- Seven targeted frontend tests passed, including actual settings serialization for creation/duplication and unrelated settings edits.
- Production frontend build passed. The full TypeScript check has 147 pre-existing diagnostics; baseline/current diagnostic counts by file and code match. It is not reported as a clean TypeScript check.
- An isolated local server and simulated upstream verified actual API models, authorized completion, unauthorized 403, immediate revocation, denied probes and no upstream calls for rejected requests.
- Headless Chrome verified the ordinary-donor search/multiselect/self controls, desktop/mobile layout and restricted catalog. A separate real form save verified the outgoing GraphQL ACL, accepted persistence and restored values after page reload.
- Final production verification: healthy public endpoint, expected revision/compiler/architecture, all 575 packaged JavaScript assets match, public SPA assets load, anonymous protected endpoints remain 401, ordinary-member directory/project resources return the expected safe projection. Production still had three existing public channels; no production donation was modified for testing.

The initial UI deployment omitted the ACL from the settings merge helper. Supplemental save-chain inspection caught this, and `74710820` corrected it before completion. A production read confirmed zero channels created in the interval and zero restricted channels; no newly created donation was affected. The final image includes the correction and serialization regression tests.

## Operations and retained recovery evidence

Only the AxonHub container was recreated. Host boot ID, protected service PIDs/restart counters/config hashes and unrelated container identities/states matched their baselines. No host, FRP, SSH, HAProxy or Docker-daemon restart was performed.

SQLite quick checks and foreign-key checks passed; core table counts were preserved. Existing privacy settings remain disabled for live previews, chunk storage and request/response body storage. Production has approximately 20 GiB free after cleanup.

Release evidence is retained on Tencent in `/home/ubuntu/axonhub-release-20261109-74710820`; the pre-feature snapshot is retained in `/home/ubuntu/axonhub-release-20261109-773ad976`. Both SQLite snapshots were gzip-compressed and their decompressed SHA256/size verified before removing the uncompressed copies. Uploaded image archives were removed after verification. The prior runtime images remain available.

Snapshot SHA256 before the final correction: `5a738868213731260a725e3ef7d5d556d3cf2b83821e32dacd5eec386c8ec2ef`. Pre-feature snapshot SHA256: `8369d92e0a0b6d781fb5a4d2175a54ac205bd4b3078e9de6d4a3a37f296355c0`.

Do not automatically restore an old database on application rollback: that would discard later writes. A pre-ACL binary does not enforce restrictions, so it must not be restored while restricted channels are enabled. The retained deployment script fails closed in that case; use an ACL-aware repair build instead.

## Commits

- `d2b9e7d1 chore(plan): record approved channel usage permissions`
- `c63cff37 feat(channels): enforce per-user usage access and private model visibility`
- `773ad976 feat(ui): add searchable channel access selection and scoped resource cards`
- `74710820 fix(channels): preserve usage permissions in form submissions`

The final documentation commit and `save-2026-11-09-channel-usage-access` tag identify this release record; the live binary identifies `74710820`.
