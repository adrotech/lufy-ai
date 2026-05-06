## 1. Release binary foundation

- [x] 1.1 Add `lufy-ai version` with semantic version, commit, build date, GOOS and GOARCH, including explicit development-build output when linker metadata is absent.
- [x] 1.2 Add release build configuration for supported OS/arch artifacts from `tools/lufy-cli-go/` without relying on root Node/TS tooling.
- [x] 1.3 Generate deterministic artifact names and SHA-256 checksum files for every packaged release binary.
- [x] 1.4 Add CI validation that recalculates checksums and smoke-tests `lufy-ai version` from executable release artifacts where the runner can execute them.

## 2. Bootstrap installer

- [x] 2.1 Implement a remote bootstrap installer that detects OS/arch, resolves an explicit version, downloads the matching artifact and fails safely for unsupported platforms.
- [x] 2.2 Verify artifact SHA-256 before installing or executing any downloaded binary, and block installation on checksum mismatch.
- [x] 2.3 Support version pinning via flag or documented environment variable, with `latest`/stable mode treated as an explicit convenience path.
- [x] 2.4 Support a user-selected install directory, actionable PATH guidance and safe failure when the destination requires privileges.
- [x] 2.5 Add fixture-backed or dry-run validation for bootstrap URL resolution, checksum success and checksum failure without mutating user machines.

## 3. Standalone asset strategy

- [x] 3.1 Decide and implement the first standalone asset source, preferring `go:embed` for managed assets unless bundle trade-offs justify a release bundle.
- [x] 3.2 Ensure `lufy-ai install --target <dir>` from a release binary can install managed OpenCode/OpenSpec assets without reading from the source repository checkout.
- [x] 3.3 Preserve existing install, verify, backup, restore and sync safety semantics when running from a distributed binary.
- [x] 3.4 If a release bundle is used, verify bundle integrity and manifest consistency before using assets for installation.

## 4. Release and CI integration

- [x] 4.1 Add an authorized GitHub Actions release workflow that builds, packages, checksums and publishes artifacts only from release/tag context.
- [x] 4.2 Add release artifact smokes for `install --dry-run`, temporary install and `verify --target <temp> --no-engram` using the packaged binary where executable.
- [x] 4.3 Keep `scripts/install.sh` as a strict local wrapper that delegates to `lufy-ai install` and does not grow remote-download fallback logic.
- [x] 4.4 Document Homebrew, Scoop and `go install` only as follow-up channels unless implemented and backed by release artifact checksums.

## 5. Documentation and cleanup

- [x] 5.1 Update `README.md`, `docs/getting-started.md` and `tools/lufy-cli-go/README.md` only after runtime support exists, describing no-clone install, version pinning, checksum verification and `lufy-ai verify`.
- [x] 5.2 Document `curl | bash` only alongside an inspectable alternative that downloads the script first and uses an explicit version.
- [x] 5.3 Remove or demote obsolete clone/build instructions from end-user docs once the release installer becomes the supported primary install path.
- [x] 5.4 Keep contributor clone/build instructions scoped to development workflows if they remain useful.

## 6. Validation

- [x] 6.1 Run targeted Go validation from `tools/lufy-cli-go/`, including `go test ./...` and `go build ./cmd/lufy-ai`, after runtime changes.
- [x] 6.2 Run release/bootstrap validation commands added by this change and record exact evidence.
- [x] 6.3 Run OpenSpec validation/status for this change and `git diff --check` before delivery.
