# Verification: Slice D - Runtime E2E, documentation and release readiness

## Scope

- Change: `complete-codex-parity-and-native-integration`
- Program tracking: GitHub issue `#218`
- Slice: D
- Gate state: `validated`
- Runtime surface: clean Codex targets, OpenSpec/default routing, Lufy SDD Full/Lite, T3 none, managed lifecycle and documentation.

## Runtime Matrix

| Scenario | Selection | Installed assets | Outcome |
| --- | --- | --- | --- |
| Codex + OpenSpec | default T1 full, T2 lite, T3 none | `.agents`, `.codex`, `.lufy/contracts`, `openspec` | install, second install, sync dry-run/real, skills, doctor, conflicts, backup/restore and deep verify passed |
| Codex + Lufy SDD | T1 `lufy-sdd/full`, T2 `lufy-sdd/lite`, T3 `none` | `.agents`, `.codex`, `.lufy/contracts`, `.lufy/workflows/sdd` | install and deep verify passed; no OpenSpec assets selected |
| T3 none | explicit `T3:none` in both targets | no methodology artifacts required for T3 | persisted consistently in project config/install state and passed deep verify |

## Managed Lifecycle Evidence

The OpenSpec target exercised:

```text
lufy-ai install --tool codex --yes
lufy-ai install --tool codex --yes
lufy-ai sync --tool codex --dry-run
lufy-ai sync --tool codex --yes
lufy-ai skills status --tool codex
lufy-ai doctor
lufy-ai conflicts plan
lufy-ai backup
lufy-ai restore --dry-run
lufy-ai restore --yes
lufy-ai verify --tool codex --deep
```

Result: all commands passed. The second install did not rewrite install state, real sync reported no managed changes, conflict count was zero, restore created its safety backup, and post-restore deep verify passed.

The Lufy SDD target exercised:

```text
lufy-ai sdd new --change matrix-probe --mode full --capability matrix-probe
lufy-ai sdd validate --change matrix-probe --strict
lufy-ai sdd sync --change matrix-probe
lufy-ai sdd archive --change matrix-probe
```

Result: `proposed -> valid -> synced -> archived`; `3/3` tasks complete, active spec materialized, archive created, and `change-overview.html` existed after new, validate, sync and archive.

## Runtime Discovery

- Codex CLI: `0.144.5`.
- `codex features list`: `hooks` and `multi_agent` reported stable.
- Installed role discovery: exactly 8 `.codex/agents/*.toml` files.
- Effective skill registry: `ready`, 18 skills, 3 roots, 0 warnings in the OpenSpec target; 18 skills in the Lufy SDD target.
- In environments without the Codex binary, structural catalog/verify tests remain mandatory and runtime probing degrades to `not_available`.

## Documentation

Updated `README.md`, `docs/installation.md`, `docs/architecture.md`, `docs/status.md` and `docs/roadmap.md` to describe the validated Codex core, neutral contracts, lifecycle events, progressive skill resources and the remaining delivery/release boundary. Troubleshooting remains in the installation guide and now points to adapter-aware `doctor`/`verify --deep` behavior.

## Grouped Validation

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./...
scripts/validate.sh
```

Result: passed. Coverage remained `80.0%`, Go build passed, source/embedded catalogs matched, and harness coupling passed. `shellcheck` was unavailable and reported as an explicit local omission.

## Residual Risks

- Runtime probes confirm Codex CLI capabilities and installed structure, but do not start an interactive Codex session; trust prompts and UI role presentation remain host-controlled.
- Context graph and Obsidian memory were intentionally not initialized in temporary targets; doctor reported actionable non-blocking warnings.
- The source tree is validated, but release publication and remote CI/check evidence remain delivery work.
