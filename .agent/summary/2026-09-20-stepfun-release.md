# StepFun release, 2026-09-20

## Scope and baseline

- Authorized action: deploy reviewed PR #23 to Tencent Cloud; minimize server disk use.
- Application revision: `7a4252c5c6341f55a546880653831b53a0ab4504`.
- Previous revision: `f23677410ba2a077d62bde002ad71f26461da53c`.
- Previous image: `ucas-cloudapi:f2367741-password-reset-ucas15`.
- Target: Tencent `VM-0-9-ubuntu`, AxonHub container only, loopback port 8090 and existing public port 10443.
- Production Compose: `/opt/axonhub/compose.yml` plus `/opt/axonhub/compose.ucas.yml`.
- Existing volume: `axonhub_axonhub_data`; preserve configuration and credentials.
- Initial disk: root 59 GiB, 91% used, approximately 5.6 GiB available.
- Never restart the host, Docker daemon, FRP, HAProxy, sing-box, HCZ, HPC or other containers.

## Validation and build

- Clean application source exported with `git archive` at the exact application revision.
- Frontend: Node 20.19.5, pnpm 10.15.0, frozen dependency lockfile; build passed.
- Five frontend channel/model-selection tests passed after installing local dependencies.
- Targeted channel, donor and password-reset tests passed in `internal/server/biz` and `internal/server/db`.
- Previous review also passed the OpenAI transformer package tests.
- Live database preflight: `quick_check=ok`, no foreign-key errors; all four privacy flags remain false.
- Build and packaging take place on the Mac; only a finished Linux/amd64 image is uploaded.
- The runtime preserves the prior image's CA bundle, timezone archive, entrypoint and environment, replacing the application binary with the newly built binary and embedded frontend.

## Cutover plan

- Record protected service PIDs, restart counts, configuration hashes and other container identities.
- Verify archive size and SHA-256, image ID, architecture and source revision before stopping AxonHub.
- Stop only AxonHub, preserve raw SQLite/WAL/SHM files and create a verified SQLite backup.
- Change only the image reference in the production override; recreate only AxonHub with no build/pull.
- Verify health revision, served frontend hashes, unauthenticated access boundaries, database integrity and unchanged protected services.
- Preserve the old image and rollback configuration; remove only the newly uploaded archive after success and compress this release's raw backup to conserve space.

## Result

Deployment completed. No application source changes were made for this release.

- New container started at `2026-09-20T12:34:56.38715742Z` (20:34:56 China time), restart count 0.
- Image: `ucas-cloudapi:7a4252c5-stepfun`.
- Image ID: `sha256:14b124fe5e9dd2e2d2d3974c881ee107c80ee10ea8894125eaad5c949da8a261`.
- Runtime version: `v1.0.0-beta5-ucas.15+stepfun.7a4252c5`.
- Uploaded archive: 40,263,917 bytes; SHA-256 `8b614de77f0f4d840cee67b80fa3ef55b0e338fe0a0f27257c80426befb74115`.
- The new image ran an isolated `build-info` check with no network, no volumes and a read-only root before cutover.
- Internal and public health report the exact target revision.
- Public root HTML and three directly referenced assets return 200; six StepFun-containing JS bundles match the local build hashes.
- Forgot-password SPA fallback works. Anonymous models, provider quotas and campus resources return 401.
- Post-cutover and post-cleanup database checks passed: `quick_check=ok`, zero foreign-key errors, core row counts preserved, privacy flags unchanged.
- All protected service/configuration/container fingerprints match the pre-cutover baseline. Fresh HPC SSH using its configured key succeeds. The HCZ endpoint answered SSH, but the default local key did not authenticate; no HCZ host changes were made.
- New container logs contain no detected panic, fatal or migration failure.
- No StepFun upstream credentials were entered and no real StepFun inference was sent. Live authenticated member UI/probe checks were not performed; relevant local regression tests passed.

## Storage and rollback material

- Remote release directory: `/opt/axonhub/releases/20260920T123123Z-7a4252c5-stepfun`.
- The uploaded image archive was removed only after loading and successful verification; the finished archive remains on the Mac.
- Raw database and consistent SQLite snapshot were compressed at low CPU priority; each decompressed SHA-256 was checked before removing its uncompressed duplicate.
- Consistent snapshot: `axonhub.predeploy.sqlite.gz`, 109,026,642 bytes. Decompressed SHA-256: `02908ecd2241ff3ea3c401259371a913c55d24cb60c8940183d80d2084f3384b`.
- Original raw database: `raw-db/axonhub.db.gz`, 109,026,635 bytes. Decompressed SHA-256: `1a3262fe934d98258a5501e1cc2994bd968746491753426f7b78363a21f9175c`.
- AxonHub stopped gracefully and no WAL/SHM files remained to copy at snapshot time.
- `compressed-backups.json` records both compressed and original checksums. The release helper can expand and verify a compressed snapshot when its rollback path needs it; the original cutover helper is retained separately.
- Previous image and Compose override remain available. Restoring the predeployment database later would discard subsequent writes and must not be done casually.
- Release footprint after cleanup: approximately 209 MiB, down from 819 MiB. Root disk now has approximately 5.2 GiB free (91% used).
- No Docker prune, unrelated image/cache deletion, live database deletion or server-side build was performed.

Application tag: `save-2026-09-20-stepfun-deployed` points to the deployed application revision. Release record commits are on local branch `codex/deploy-stepfun-20260920`.
