# Native reasoning compatibility release — 2026-10-20

## Scope

AxonHub's OpenAI Responses adapter now consumes native `response.reasoning_text.delta` and `.done` events, plus non-stream `reasoning` content items of type `reasoning_text`. These are the events emitted by the local Qwen and DeepSeek services. The adapter continues to accept `reasoning_summary_text` as a fallback and avoids duplicating reasoning when both forms are present. It does not change the backend model or reasoning policy.

The related 8443 gateway compatibility fix is revision `e11a33d` in `relay8443-deepseek-20261018`: successful Responses SSE/JSON envelopes convert integral floating-point `created_at` and `completed_at` values to JSON integers for AxonHub's Go decoder. Other event payloads remain natively forwarded.

## Source and verification

- AxonHub code revision: `9c5ecf51a447895506537d54bbe313cdb8201ceb` (`fix(axonhub): read native reasoning text from local models`).
- `go test ./transformer/openai/responses -count=1` passed, including native streaming, aggregation, non-stream content, and summary fallback cases.
- Clean source archive built on Mac for Linux amd64 with `CGO_ENABLED=0`; exact previous production frontend assets were embedded unchanged.
- Runtime version: `v1.0.0-beta5-ucas.17+reasoning.9c5ecf51`. Docker image: `ucas-cloudapi:9c5ecf51-reasoning`, ID `sha256:d1da834efa6134449fabbe01fb6ff00361ae76c00d1aa361035b8aeba7aa0d4c`.
- Image archive SHA-256: `c40d34a15bfa5df3aa92012206228327e80ba99c7869fc7020be605bb85ae1cb`.

## Deployment and limits of evidence

Only the Tencent AxonHub container was replaced. Release directory: `/opt/axonhub/releases/20261020-9c5ecf51-reasoning`. The pre-cutover SQLite snapshot was checked (`quick_check=ok`, zero foreign-key errors); its SHA-256 is `8cc1d1f3894cafb142706a57c02eb0923622826115e09db1276d21fad1e9b74c`. Public `https://holocortexzero.com:10443/health` reported the exact new version/revision and HTTP 200, the root page returned 200, and the container restart count was zero. Protected services and unrelated containers matched the preflight fingerprints. The previous image `ucas-cloudapi:c235adf4-models` and Compose override were retained for rollback.

After deployment, the user retested through AxonHub and reported “ok了”. The reply did not separately identify Qwen and DeepSeek outcomes, so it establishes successful user-side retest, not model-by-model production proof. Focused adapter tests cover the native event form shared by both models. No HCZ/HPC host or FRP service was restarted.
