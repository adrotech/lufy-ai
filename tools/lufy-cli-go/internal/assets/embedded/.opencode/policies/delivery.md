# Delivery Policy

Canonical policy for lufy-ai agents, commands, and skills.

## Roles

- `orchestrator` coordinates and routes; must not edit files or run shell commands.
- `sdd-router` classifies work into T1/T2/T3, recommends execution mode, context slice, skill status, and review workload read-only; must not edit files, run shell/Git/OpenSpec/validation commands, install external skills, or perform delivery. It routes to `explorer`, `validator`, or `delivery` when repository state, evidence, validation, or Git/GH operations are needed.
- `explorer` investigates impact and repository context read-only; must not edit files.
- `implementer` implements bounded changes and uses systemic workflow: initial context analysis, no repeated old-file rereads during normal implementation, bounded final reread of changed/affected old files, and validación agrupada at the end of a work block/proposal unless blocked, risky, or diagnosing; must not commit, push, create PRs, or update GitHub Projects.
- `validator` runs compile/test evidence and diagnoses failures read-only; must not edit files.
- `reviewer` reviews only; must not modify files.
- `delivery` owns Git/GH operations, PR creation, project sync, traceability comments, and final validation evidence.

## Branch And PR Rules

- Normal integration branch and default PR base: `develop`.
- Productive/stable branch: `main`; use it only for `develop` → `main` promotion, release, or explicitly authorized hotfix work.
- Normal work opens PRs from feature/fix/chore branches to `develop`.
- Promotion to production opens a PR from `develop` to `main` with release/promotion validation evidence.
- Stable release tags MUST match `v*` and be created from commits reachable from `origin/main`; do not tag unpromoted `develop` commits for stable releases.
- Protected PR source branches: `develop`, `main`, `master`, `development`.
- Never force push unless the user explicitly requests it.
- Report dirty or mixed worktrees before staging.
- With explicit delivery authorization, `delivery` can run `git status`, `git diff`, `git log`, `git add`, `git commit`, `git push`, and `gh` without intermediate prompts.
- Without explicit delivery authorization, return `blocked` with exact recovery instruction.

## Validation Tiers

- **Routing gate** for non-trivial or ambiguous work: classify the request as T1 Full SDD, T2 SDD Lite, or T3 Express before choosing agents, context, permissions, artifacts, and review workload. Tiers classify proposals, functionalities, and tasks by risk/uncertainty/impact; they do not authorize delivery.
- **Systemic analysis gate** for `explorer`/`implementer`: analyze existing files, dependencies, interconnections, feedback paths, and structure/behavior impact at the beginning of a coherent block/proposal.
- **Implementation gate** for `implementer`: do not reread old files repeatedly during normal implementation after the initial analysis. Reread old files only when they were modified/affected, conflicts appear, new evidence invalidates the initial analysis, scope changes, a blocker appears, or risk requires confirmation.
- **Block/proposal gate** for `implementer` and `validator`: run grouped validation at the end of all tasks in a coherent block/proposal, including tests and coverage when real commands exist for the scope. Do not run tests constantly during normal implementation.
- **Planning/OpenSpec-only fast path**: when the current micro-slice touches only 1-2 OpenSpec/docs artifacts, has no runtime/app changes, no delivery, and no security/public-contract impact, the workflow may use T2/T3 fast path even inside a broader T1 program. Expected validation is `openspec validate "<change>" --strict` when applicable plus static checkbox/file review.
- **Exception gate**: run focused rereads or validation earlier only when a blocker, risky change, feedback loop, or failure diagnosis requires it.
- **Final PR gate** for `delivery`: run the repository's real full validation suite when available (typecheck/compile, tests, coverage, linting as applicable).
- **Remote PR checks gate** for `delivery`: after `gh pr create`, consult or wait for remote PR pipelines/checks and record command evidence, for example `gh pr checks <PR>` or `gh pr view <PR> --json statusCheckRollup,mergeStateStatus,url`. If checks report `FAILURE`, `CANCELLED`, `TIMED_OUT`, `ACTION_REQUIRED`, evidence is missing, or required tooling is unavailable, do not report `delivered` or `closed`; report `blocked` with PR URL/status and recovery command. Use `delivery_pending` only when remote checks exist and are still pending without a successful conclusion.
- **PR whitespace gate** for `validator`/`delivery`: for PR-bound changes, reproduce the PR diff range against the target base. Use `git diff --check origin/develop...HEAD` for committed branch contents, or `git diff --check origin/develop` while local worktree changes are still pending. Plain `git diff --check` is insufficient because it only checks uncommitted worktree/staged changes.
- **Local grouped validation**: prefer `scripts/validate.sh` when the change scope matches this repository's Go CLI/assets workflow; it runs the PR-aware whitespace gate plus available Go validation.
- If change affects behavior, include functional evidence when practical.
- Never claim validation passed without command evidence.
- Delivery remains explicitly authorized regardless of tier. A T1/T2/T3 classification can recommend delivery readiness but cannot authorize Git/GH operations.
- Dirty or mixed worktrees must be reported before delivery/staging, but dirty state alone must not block scoped documentation/OpenSpec-only validation when no delivery is requested and there is no concrete evidence of mixed runtime changes.
- For this repo, the CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto for that CLI and must not fall back to legacy install paths.

## Workflow Limits Config

- `.lufy/project.yaml` top-level `workflow_limits` is the only canonical source for project workflow limits.
- Delivery batching decisions use `workflow_limits.delivery_batch_strategy` after a block/change is validated and delivery is explicitly authorized.
- Delivery batching guidance may be recorded before authorization, but it remains advisory and MUST NOT authorize Git/GH operations or produce `delivered`/`closed` state by itself.
- Proposal or review splitting uses `workflow_limits.proposal_slicing_strategy` before implementation/review; do not reinterpret proposal slicing as delivery batching or as Git/GH delivery authorization.
- Delivery preflight checks use `workflow_limits.preflight` when present and must be verified or reported before moving to delivery-ready, delivered, or closed states.
- Delivery stop conditions use `workflow_limits.stop_rules`; when a configured stop condition is hit, pause and report `blocked` or `delivery_pending` with the exact recovery path instead of continuing silently.
- Top-level `loc_budget` and `delivery_strategy` are legacy/non-canonical and must not drive sizing, routing, slicing, batching, preflight, stop rules, delivery authorization, or closure.

## OpenSpec Task/Block Gate

Evaluate completion at the smallest coherent delivery unit: a `tasks.md` task, an implementation block, or a review slice. Nested micro-checkboxes can track internal progress, but they do not trigger full validation, delivery, archive readiness, or closure unless explicitly declared as the coherent unit.

Use these task/block states consistently:

- `implemented`: scoped edits are applied and task checkboxes for the implementation block may be updated; proportional validation still remains.
- `validated`: real applicable validation evidence exists, or static/manual evidence is documented when no toolchain applies; delivery or sync may still remain.
- `delivery_pending`: the block is implemented/validated but needs Git/GH delivery, issue/project sync, PR, external publishing that has not been explicitly authorized/completed, or existing remote PR checks are still pending.
- `delivered`: explicitly authorized `delivery` completed the required commit, push, PR, traceability, or external sync for the requested scope, and any required remote PR checks have successful command evidence.
- `closed`: implementation, validation, required delivery, required remote PR checks, required sync, and archive/traceability preconditions are all satisfied for the declared scope.

Role boundaries for the gate:

1. `implementer` may move a coherent block to `implemented`; it must not commit, push, create PRs, update GitHub Projects, or report `closed` without validation and delivery evidence from the proper roles.
2. `validator` may move a block to `validated` with read-only command/static evidence; it must not edit files or perform Git/GH delivery.
3. `delivery` may move a validated block to `delivered` or `closed` only after explicit user authorization and only when closure evidence is complete.
4. `orchestrator` coordinates transitions and requests explicit authorization before routing Git/GH delivery.

If any gate item is missing, report the precise next state (`implemented`, `validated`, `delivery_pending`, `sync_pending`, or `blocked`) with exact recovery instruction. Delivery remains explicitly authorized regardless of tier, task completion, validation status, or user acceptance of implementation.

Substantive delivery handoffs and final delivery results use Result Contract envelope v1. Include the relevant `workflow_decision` fields from `.lufy/project.yaml` or upstream routing, especially delivery batching guidance, preflight status, stop-rule status, remote check state and authorization status.

Tasks incompletas always block archive. Do not archive a change with unchecked tasks, even with user confirmation. `migrate-installer-to-go-cli` is explicitly blocked from archive while any tasks remain incomplete.

Current active/focus spec context: `install-managed-assets-with-hash-idempotency` covers managed assets, SHA-256, manifest-driven idempotency, backup/restore, and structural verify.

## Completed Change Gate

- If a change reaches 100% tasks complete, treat it as `implemented` or `validated` according to available evidence, not automatically `closed` or archive-ready.
- If the change requires Git/GH delivery and the user has explicitly authorized it, route to `delivery` for PR creation before starting another change.
- If delivery is required but not authorized, report `delivery_pending` or `blocked` and ask for explicit authorization before Git/GH operations.
- Do not begin new change while completed-change PR is pending, unless explicitly authorized.
- Use `.opencode/templates/pr-evidence.md` or skill helpers for PR body.

## GitHub Project Sync

- Use `.opencode/skills/project-sync/project_sync.py` for board sync only when a downstream project installs that optional skill.
- Keep generated content in Spanish.
- Configure Project IDs through environment variables.
- On remote failures, return `sync_pending` with recovery command.

## Language And Format

- Human-facing delivery output, PR bodies, issue comments, and artifacts default to Spanish.
- Preserve technical identifiers, code symbols, filenames, routes, and CLI flags.
- Include ASCII diagrams in PR evidence unless change is documentation-only.
