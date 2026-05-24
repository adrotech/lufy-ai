## 1. Reviewer Contract

- [x] 1.1 Update `.opencode/agents/reviewer.md` with L1-L5 severity definitions and approval gate rules.
- [x] 1.2 Add weighted scoring categories and required category breakdown in Result Contract envelope v1.
- [x] 1.3 Add stack-aware guidance for `.opencode/project.yaml` anti-patterns, coverage thresholds and observability libraries.

## 2. Workflow Behavior

- [x] 2.1 Add desk-check guidance requiring at least eight scenarios for substantive T1/T2 reviews.
- [x] 2.2 Clarify reviewer escalation to `validator` for missing command evidence and to `delivery` only after explicit authorization.
- [x] 2.3 Preserve T3 proportional handling with explicit `not_applicable` for heavy desk-check/scoring detail when appropriate.

## 3. Managed Asset Sync

- [x] 3.1 Sync the embedded reviewer asset under `tools/lufy-cli-go/internal/assets/embedded/.opencode/agents/reviewer.md`.
- [x] 3.2 Update role/docs references if reviewer output contract changes.
- [x] 3.3 Verify root and embedded reviewer assets remain aligned.

## 4. Validation

- [x] 4.1 Run `openspec validate add-scored-stack-aware-reviewer --strict`.
- [x] 4.2 Run `openspec validate --all --strict`.
- [x] 4.3 Run `go test ./internal/assets -run TestEmbeddedCatalogMatchesRepositoryAssets -count=1` from `tools/lufy-cli-go` after managed asset changes.
- [x] 4.4 Run `scripts/validate.sh` and report exact results and limitations.
