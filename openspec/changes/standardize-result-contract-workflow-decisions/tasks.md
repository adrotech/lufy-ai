## 1. Contract Definition

- [x] 1.1 Define Result Contract envelope v1 in root agent guidance, including required fields, allowed statuses, compact T3 handling and legacy fallback behavior.
- [x] 1.2 Update reusable result-contract template/docs so future routed work can copy the envelope without inventing role-specific schemas.
- [x] 1.3 Document workflow decision fields for `workflow_limits` source, paths considered, workload decision, review slices, preflight, stop rules and delivery batching guidance.

## 2. Agent And Policy Updates

- [x] 2.1 Update `sdd-router` output contract to emit workflow-limit decision fields and `workload_decision_needed` from canonical `workflow_limits` inputs.
- [x] 2.2 Update `orchestrator` handoff guidance to require envelope v1, carry workflow decisions forward and normalize legacy fallback outputs.
- [x] 2.3 Update `implementer`, `validator`, `reviewer` and `delivery` guidance to return substantive routed results using envelope v1.
- [x] 2.4 Update delivery policy so delivery batching guidance remains separate from explicit Git/GH delivery authorization and remote check evidence.

## 3. Embedded Assets And Documentation

- [x] 3.1 Sync changed managed assets into `tools/lufy-cli-go/internal/assets/embedded/**`.
- [x] 3.2 Update human-facing documentation or backlog/status notes if the implemented behavior changes what is current versus future.
- [x] 3.3 Ensure installed harness assets and root assets remain aligned for agent, policy, template and spec changes.

## 4. Validation

- [x] 4.1 Run `openspec validate standardize-result-contract-workflow-decisions --strict`.
- [x] 4.2 Run `openspec validate --all --strict`.
- [x] 4.3 Run embedded catalog parity validation with `go test ./internal/assets -run TestEmbeddedCatalogMatchesRepositoryAssets -count=1` from `tools/lufy-cli-go` if managed assets or specs were synced.
- [x] 4.4 Run grouped local validation with `scripts/validate.sh` and report exact results and limitations.
