## 1. Coverage And Go Quality

- [x] 1.1 Measure current Go coverage for `tools/lufy-cli-go` and choose an initial threshold from the observed baseline.
- [x] 1.2 Add a reproducible coverage gate script or validation step that generates `coverage.out` and fails below the configured threshold.
- [x] 1.3 Add minimal Go lint/static analysis configuration or command for the CLI scope.
- [x] 1.4 Integrate coverage and Go lint/static analysis into CI without depending on root Node/TS tooling.
- [x] 1.5 Integrate available coverage/lint checks into `scripts/validate.sh`, reporting unavailable optional tools explicitly.

## 2. Shell And Workflow Quality

- [x] 2.1 Add ShellCheck validation for `scripts/*.sh` and `tools/lufy-cli-go/scripts/*.sh`.
- [x] 2.2 Ensure ShellCheck runs in CI with pinned actions or installed tooling and does not lint Markdown/YAML snippets.
- [x] 2.3 Keep workflow YAML syntax validation documented or scripted for local use.

## 3. Multi-Platform CI

- [x] 3.1 Update `go-cli-install.yml` to run Go tests/builds on Linux, macOS and Windows or document any unsupported subset.
- [x] 3.2 Gate installer/wrapper smokes so POSIX-specific smokes run only on compatible runners.
- [x] 3.3 Preserve existing Linux smoke coverage for dry-run, install, verify, idempotence, backup/restore and wrapper delegation.

## 4. Release E2E And Regression Tests

- [x] 4.1 Add a separate post-release/manual E2E workflow or script that downloads published GitHub Release artifacts for a `v*` tag and verifies checksum/install/verify.
- [x] 4.2 Add at least one golden or structured regression test for representative installer plan output.
- [x] 4.3 Add at least one runtime/dispatch test for `cmd/lufy-ai/main.go` or equivalent CLI entrypoint behavior.

## 5. Documentation And OpenSpec

- [x] 5.1 Document the new local/CI quality gates and any optional local tool limitations.
- [x] 5.2 Update OpenSpec task checkboxes as implementation progresses.
- [x] 5.3 Sync embedded managed assets if root managed artifacts change.

## 6. Validation And Delivery

- [x] 6.1 Run `scripts/validate.sh` from repository root.
- [x] 6.2 Run any new coverage/lint/shellcheck scripts directly if they are not fully covered by `scripts/validate.sh`.
- [x] 6.3 Validate workflow YAML syntax after workflow edits.
- [x] 6.4 Verify `openspec status --change improve-ci-quality-gates` before delivery.
