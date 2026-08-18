---
name: sdd-workflow
description: Route Lufy work through proportional T1/T2/T3 SDD using the configured OpenSpec, native Lufy SDD Full/Lite, or Express adapter.
---

# SDD Workflow

Use this skill for non-trivial work in repositories governed by Lufy.

1. Classify the request as T1, T2, or T3.
2. Use T1 for architecture, public contracts, security, cross-cutting work, or high uncertainty.
3. Use T2 for bounded behavior changes, relevant bugs, agent/skill work, or controlled refactors.
4. Use T3 for trivial local, mechanical, or documentation-only changes.
5. For T1/T2, define observable acceptance criteria and validation evidence before reporting readiness.
6. Preserve user-owned files and unrelated local changes.
7. Resolve methodology independently from tier. Use `lufy-sdd/full` for native T1 Full and `lufy-sdd/lite` for native T2 Lite when selected; never substitute OpenSpec silently.
8. Lufy SDD Full/Lite automatically maintain `change-overview.html` through new/validate/sync/archive. Report its path/status and do not invoke or offer another render action.
