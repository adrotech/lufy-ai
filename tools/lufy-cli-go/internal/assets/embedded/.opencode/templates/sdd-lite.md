# T2 SDD Lite Template

Use this template for bounded work that needs specification discipline but not Full SDD/OpenSpec.

## Intent

- User goal:
- Why now:

## Current Behavior

- What happens today:
- Relevant files or components:

## Target Behavior

- Desired outcome:
- Non-goals:

## Scope

- In scope:
- Out of scope:
- Escalate to T1 if:

## Review Workload

- Review workload: none | focused | full
- Keep as one deliverable if:
- Split into slices if:

## Review Slices

Use only when slicing reduces reviewer cognitive load or risk.

### Slice 1: <name>

- Objective:
- Expected files or areas:
- Acceptance criteria:
  - WHEN <observable trigger> THEN <observable outcome>.
- Validation:
- Primary review risk:
- PR guidance: same_pr | separate_pr_recommended | separate_pr_required_if_authorized

## Acceptance Criteria

- WHEN <observable trigger> THEN <observable outcome>.
- WHEN <edge case or failure mode> THEN <expected handling>.

## Structural Acceptance

- Field: `structural_acceptance`
- Source: user_prompt | project_profile | spec | mixed | not_available
- Expected directories/layers:
  - <components/pages/hooks/utils/constants/services/types/index/controller/service/repository/domain/usecase/ports/adapters or not_applicable>
- Forbidden root files:
  - <pattern that would violate requested structure or not_applicable>
- Normalization:
  - <page vs pages or other singular/plural decision; ask user if ambiguous>
- Gate:
  - Validation/review must report `blocked` or `needs_revision` if mandatory structure is missing, files remain in forbidden root locations, or follow-up is assumed without explicit user confirmation.

## Tasks

- [ ] Explore focused impact only if needed.
- [ ] Convert explicit folder/layer instructions into structural acceptance checks.
- [ ] Implement the bounded change.
- [ ] Update tests or docs directly tied to the change when applicable.
- [ ] Audit structural acceptance per affected feature/surface.
- [ ] Run grouped validation available for this scope.

## Validation

- Expected commands:
- Static/manual checks:
- Unavailable validation and reason:

## Risks

- Risk:
- Mitigation:

## Result Contract

- Objective:
- Actions performed:
- Evidence:
- Structural acceptance audit:
  - <feature/surface>: <satisfied/missing/blocked details>
- Optional overview/render: offered_pending | generated | skipped_by_user | not_available; include command/path only when the selected methodology and tool adapter provide one. Use skipped_by_user only after an explicit user decline.
- Review slices completed:
- Risks/follow-ups:
- State: ready | blocked | escalated | pending_validation
- Recommended next action:
