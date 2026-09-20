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

Pending deployment. No application source changes were made for this release.
