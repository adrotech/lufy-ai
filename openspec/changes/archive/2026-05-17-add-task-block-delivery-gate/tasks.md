## 1. Policy and shared terminology

- [x] 1.1 Update `.opencode/policies/delivery.md` to define the task/block gate, state semantics, and explicit delivery authorization boundary.
- [x] 1.2 Update `AGENTS.md` and `AGENTS.md.template` where needed so project-wide guidance distinguishes `implemented`, `validated`, `delivery_pending`, `delivered`, and `closed`.

## 2. Agent role alignment

- [x] 2.1 Update `orchestrator` guidance to route task/block state transitions and request explicit delivery authorization before `delivery`.
- [x] 2.2 Update `implementer` guidance so implementation completion reports `implemented`/validation pending rather than closed, and never performs Git/GH delivery.
- [x] 2.3 Update `validator` guidance so validation remains read-only and reports `validated`, `delivery_pending`, `blocked`, or equivalent next state.
- [x] 2.4 Update `delivery` guidance so authorized Git/GH work moves validated blocks to `delivered`/`closed` only when required closure evidence exists.

## 3. OpenSpec skills and command workflow

- [x] 3.1 Update `openspec-apply-change` skill and `/opsx-apply` command language to mark block outcomes as `implemented` or validation pending, not closed.
- [x] 3.2 Update `openspec-verify-change` skill and `/opsx-verify` command language to evaluate block-scoped proportional validation and report delivery-pending states without performing delivery.
- [x] 3.3 Update `openspec-archive-change` skill and `/opsx-archive` command language to reject archive when closure gates, delivery, sync, or validation remain unresolved.

## 4. Consistency checks and managed assets

- [x] 4.1 Review related `.opencode` commands, templates, and skill references for stale “tasks complete means closed/archive-ready” wording and update only affected references.
- [x] 4.2 If changed harness assets are managed by the installer catalog, update embedded assets/catalog outputs required by the repository workflow.

## 5. Validation and handoff

- [x] 5.1 Run proportional validation for documentation/workflow changes, including OpenSpec status/verification commands and repository validation commands that are available for this scope.
- [x] 5.2 Record final evidence and next state: `validated`, `delivery_pending`, `delivered`, `closed`, `blocked`, or an explicitly documented equivalent.
