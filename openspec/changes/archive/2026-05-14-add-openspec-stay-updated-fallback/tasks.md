## 1. Resolver And Manifest Design

- [x] 1.1 Define resolver data structures for OpenSpec source, version, layer and manifest metadata.
- [x] 1.2 Extend `openspec/UPSTREAM.json` with minimum compatible version/source metadata needed by the resolver.
- [x] 1.3 Keep root and embedded `UPSTREAM.json` in parity.

## 2. CLI Resolver Implementation

- [x] 2.1 Add internal Go package for OpenSpec stay-updated resolution.
- [x] 2.2 Implement PATH resolution for compatible `openspec` CLI versions.
- [x] 2.3 Implement cache resolution from `.lufy-ai/openspec-cache/<version>/manifest.json`.
- [x] 2.4 Implement embedded baseline fallback for offline standalone binaries.
- [x] 2.5 Ensure resolver and cache operations are stdlib-only and path-safe.

## 3. Cache And Atomic Writes

- [x] 3.1 Define cache manifest format and validation rules.
- [x] 3.2 Implement atomic writes for cache manifests/assets where cache writes are introduced.
- [x] 3.3 Reject corrupt cache manifests, unsafe paths and symlink escapes.

## 4. Workflow And Reporting

- [x] 4.1 Update `opsx-version` guidance or CLI reporting to include effective resolver layer.
- [x] 4.2 Add `.github/workflows/sync-openspec.yml` to propose baseline bumps by PR only.
- [x] 4.3 Ensure workflow does not automerge, tag or publish releases.

## 5. Specs, Docs And Embedded Assets

- [x] 5.1 Sync new/changed managed assets into `tools/lufy-cli-go/internal/assets/embedded/`.
- [x] 5.2 Update docs for stay-updated behavior without claiming expanded profile or release availability.
- [x] 5.3 Keep OpenSpec specs and embedded spec copies synchronized.

## 6. Validation

- [x] 6.1 Run `openspec validate add-openspec-stay-updated-fallback`.
- [x] 6.2 Run `openspec validate --all`.
- [x] 6.3 Run `go test -count=1 ./internal/assets`.
- [x] 6.4 Run resolver-focused Go tests once implemented.
- [x] 6.5 Run `scripts/validate.sh`.
- [x] 6.6 Run `git diff --check origin/develop`.
- [x] 6.7 Run sandbox smokes for offline fallback and cache resolution.
