# Model panel / Codex release — 2026-10-02

## Scope and source

The user approved local compilation followed by upload to the existing Tencent server and requested GitHub synchronization. Target is only the AxonHub container on Tencent VM-0-9-ubuntu (SSH 60024), loopback 8090 and public HTTPS 10443. No host, Docker daemon, FRP, HAProxy, sing-box, HCZ, HPC or unrelated container restart is permitted.

- Source/application revision: `c235adf48eecda1e592dbdf73c92cfc99120865c`.
- GitHub branch: `codex/model-availability-20261002`; source tag: `save-2026-10-02-model-availability`.
- Previous image: `ucas-cloudapi:7a4252c5-stepfun`, revision `7a4252c5c6341f55a546880653831b53a0ab4504`.
- New image: `ucas-cloudapi:c235adf4-models`.
- New image ID: `sha256:fb9302e8d6eec5f40f60b81afeca2fdcdd7a8e7608cf6a595f22e101a294efb4`.
- Runtime version: `v1.0.0-beta5-ucas.16+models.c235adf4`.
- Remote release directory: `/opt/axonhub/releases/20261002-c235adf4-models`.
- Local release directory: `/Users/apple/codex/.ucas-release-20261002`.

## Build and safeguards

A clean `git archive` of the exact revision was exported on the Mac. Frontend production build passed using the existing lockfile-compatible local dependencies. Go was cross-compiled with `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`, with explicit revision/version/build time and embedded frontend. The verified previous scratch runtime's CA/timezone/config files were preserved; only the application binary was replaced.

The finished archive is 40,275,727 bytes; SHA-256 `8c2b7a6eb45346bd3f076e5dd056498eb26d33656068b70d5ee0f6aba8ab8c63`. Server preflight verified size, checksum, image architecture, source revision, entrypoint and isolated `build-info` execution with no network or volumes. There is no server-side build or package install.

Initial disk: 59 GiB total, 42 GiB used, 16 GiB available (73%); inode use 18%. Live data approximately 821 MiB. Database quick_check passed, foreign-key errors zero, all four privacy storage flags remained false.

Before cutover, the helper fingerprinted protected services, configuration, unrelated containers, boot ID, and AxonHub port/mount/environment wiring. It preserves the previous Compose override and image, then stops only AxonHub and creates one verified consistent SQLite snapshot. There are no schema/migration changes relative to the prior deployment. Automatic fallback changes only the application image; it does not restore an old database and discard writes accepted after cutover.

## Result

Deployment completed at 2026-10-02 14:38:20 UTC (22:38:20 China time). New container restart count is zero; startup logs contain no detected panic, fatal or migration-failure lines.

- Internal and public `/health` report exact application revision `c235adf48eecda1e592dbdf73c92cfc99120865c` and the new runtime version.
- Public root HTML, three directly referenced assets and forgot-password SPA fallback pass. Seven StepFun/model-panel/localization bundles match the locally built SHA-256 values.
- Anonymous `/v1/models`, `/admin/provider-quotas` and `/admin/campus/resources` return 401.
- Post-cutover database quick_check is `ok`, foreign-key errors zero, core row counts preserved, four privacy flags unchanged. AxonHub ports, mounts, environment hashes and protected service/container/config fingerprints match the baseline; the protected fingerprint still matches after cleanup.
- A consistent 428,613,632-byte snapshot was verified and compressed to 114,056,425 bytes. Decompressed SHA-256: `b150270c43b0033af8c475bbb112bcfbc52e7a25ae54da1f6bdf7918e2041cba`; compressed SHA-256: `7d0873a751093b423aae9ea3ba93cc065983f108c93aa5f102ab8fdb67359915`.
- The uncompressed backup duplicate and uploaded image archive were removed only after verification; the finished archive remains on the Mac. The remote release folder is approximately 109 MiB. Disk free space is 16,102,060,032 bytes (approximately 15 GiB), usage 74%.
- The previous image and Compose override remain available for rollback. No Docker prune, live-data deletion, cloud build cache, dependency installation or unrelated-service restart occurred.

The deployed application tag is `save-2026-10-02-model-availability-deployed`, pointing to the actual application revision rather than the later documentation commit. The GitHub branch also receives this release record.

Authenticated live upstream inference is a separate verification boundary: the source and Codex header tests confirm default Version 0.159.0, but this release record must not claim the specific gpt-6-astra upstream incident is fixed without a successful actual channel test.
