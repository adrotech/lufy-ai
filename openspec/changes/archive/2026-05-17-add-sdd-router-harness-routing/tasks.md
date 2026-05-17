## 1. Agent Routing Model

- [x] 1.1 Add `.opencode/agents/sdd-router.md` with read-only permissions, tier decision rules, and structured output contract.
- [x] 1.2 Update `.opencode/agents/orchestrator.md` to invoke `sdd-router` before non-trivial or ambiguous implementation workflows.
- [x] 1.3 Define escalation rules from T3 to T2 and from T2 to T1.
- [x] 1.4 Define execution modes and result contract fields used by router handoffs and subagent returns.

## 2. SDD Lite Template

- [x] 2.1 Add a T2 SDD Lite template covering intent, current behavior, target behavior, scope, acceptance criteria, tasks, validation, and risks.
- [x] 2.2 Document that T2 is compact but still professional and verifiable.
- [x] 2.3 Ensure T2 acceptance criteria use observable WHEN/THEN outcomes.
- [x] 2.4 Document artifact persistence by tier: OpenSpec for T1, compact artifact or structured handoff for T2, result-only when sufficient for T3.

## 3. Skill Resolution

- [x] 3.1 Document local-first skill resolution for `.opencode/skills`.
- [x] 3.2 Add router output fields for `skill_status`, including local coverage, detected stack, bootstrap recommendation, first dry-run command, and authorization requirement.
- [x] 3.3 Document AutoSkills as optional bootstrap/fallback only, with `npx autoskills --dry-run` before any mutating command and explicit user authorization required.

## 4. Documentation And Policy

- [x] 4.1 Update `AGENTS.md` with the T1/T2/T3 model as classification of proposals, functionalities, and tasks.
- [x] 4.2 Update `.opencode/README.md` to describe the harness routing model and the role of `sdd-router`.
- [x] 4.3 Update `.opencode/policies/delivery.md` only if needed to clarify that delivery remains explicitly authorized regardless of tier.
- [x] 4.4 Document subagent isolation, context slicing, and proportional review workload.

## 5. Verification

- [x] 5.1 Verify the new spec uses core v2 delta sections and each requirement has scenarios with WHEN and THEN.
- [x] 5.2 Run grouped repository validation appropriate for documentation/agent changes.
- [x] 5.3 Review changed agent definitions for permission minimization and no accidental delivery authorization.

## 6. Documentation Refresh

- [x] 6.1 Update `README.md` to describe `sdd-router`, T1/T2/T3, SDD Lite, result contracts, and current installable assets.
- [x] 6.2 Update `docs/getting-started.md`, `docs/installation.md`, `docs/architecture.md`, `docs/status.md`, `docs/roadmap.md`, and `openspec/README.md` to remove drift and reflect the evolved harness.
- [x] 6.3 Update `tools/lufy-cli-go/README.md` to describe embedded harness assets and local installer regeneration.

## 7. Installer Embedded Assets

- [x] 7.1 Add `.opencode/templates` to the managed asset catalog so T2/result templates install with the harness.
- [x] 7.2 Sync embedded `.opencode` assets with root assets, including `sdd-router`, templates, README, policy, orchestrator, and delivery docs.
- [x] 7.3 Update embedded `AGENTS.md.template` with proportional routing, local-first skill resolution, and delivery policy guidance.
- [x] 7.4 Build a new local installer binary at `tools/lufy-cli-go/bin/lufy-ai`.

## 8. Delivery Policy Refactor

- [x] 8.1 Keep `.opencode/policies/delivery.md` as shared canonical invariants for all agents.
- [x] 8.2 Keep `.opencode/agents/delivery.md` as operational runbook and remove/avoid unnecessary duplication where practical.
- [x] 8.3 Mirror the same policy/runbook split in embedded assets.

## 9. Review Workload Harness

- [x] 9.1 Add `review_slices` guidance to `sdd-router` output and routing rules.
- [x] 9.2 Update SDD Lite and result contract templates with reviewer-friendly slices and PR split guidance.
- [x] 9.3 Document Review Workload Harness in project docs and embedded docs.
- [x] 9.4 Mirror Review Workload Harness updates into embedded installer assets and rebuild the local installer.
- [x] 9.5 Validate that review slicing remains proportional: required for T1/multi-risk T2, avoided for normal T3.
