# Lufy Harness Contracts

This directory is the adapter-neutral source of truth for invariants shared by Codex, OpenCode, and future harness adapters.

## Contract Inventory

| Contract | Responsibility | Adapter overlays |
| --- | --- | --- |
| `delivery.md` | Authorization, branch safety, validation, remote checks, and closure gates | Agent permissions and Git/GH execution details |
| `result-contract.md` | Portable result envelope, evidence, risks, workflow decision, and next action | Tool-native handoff syntax |
| `pr-review/review-framework.md` | Review dimensions, severity, scoring, and desk-check guidance | Tool-specific evidence collection |
| `pr-review/report.html` | Self-contained HTML report structure and visual contract | Output-path and open/preview mechanics |

Roles and methodology flows stay in their own registries: `AGENTS.md` defines repository governance, `.lufy/sdd/` defines native Lufy SDD, and `openspec/` defines OpenSpec. Adapters render or reference these contracts without duplicating their semantics.

Codex skill resources are projected from this directory into each installed skill's `references/` and `assets/` folders. OpenCode keeps compatibility overlays under `.opencode/` for existing commands and links.
