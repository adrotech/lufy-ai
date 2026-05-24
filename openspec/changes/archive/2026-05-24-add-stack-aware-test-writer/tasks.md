## 1. Agent Definition

- [x] 1.1 Create `.opencode/agents/test-writer.md` as a subagent with bounded permissions, no delivery authority and Result Contract envelope v1 output.
- [x] 1.2 Define the RED -> GREEN -> TRIANGULATE -> REFACTOR workflow, including explicit `not_applicable`, `not_available` and `blocked` evidence states.
- [x] 1.3 Add stack-aware guidance for reading test commands, coverage thresholds and anti-patterns from `.opencode/project.yaml` without inventing fallback toolchains.

## 2. Workflow Integration

- [x] 2.1 Update `.opencode/agents/implementer.md` to delegate substantive T1/T2 test work to `test-writer` or record why delegation is not applicable.
- [x] 2.2 Update `.opencode/agents/validator.md` to block or escalate T1/T2 validation when required TDD evidence is missing or incomplete.
- [x] 2.3 Ensure T3 trivial, mechanical or documentation-only changes can mark TDD as `not_applicable` without forced delegation.

## 3. Managed Asset Sync

- [x] 3.1 Add the embedded `test-writer` agent asset under `tools/lufy-cli-go/internal/assets/embedded/.opencode/agents/`.
- [x] 3.2 Sync embedded copies of changed `implementer` and `validator` agent assets.
- [x] 3.3 Verify root and embedded agent assets remain aligned for installed harness behavior.

## 4. Validation

- [x] 4.1 Run `openspec validate add-stack-aware-test-writer --strict`.
- [x] 4.2 Run `openspec validate --all --strict`.
- [x] 4.3 Run `go test ./internal/assets -run TestEmbeddedCatalogMatchesRepositoryAssets -count=1` from `tools/lufy-cli-go` after managed asset changes.
- [x] 4.4 Run the grouped local validation command applicable to this repo scope, or document why it is not available.
