# Verification: Slice C - Self-contained contracts and skill parity

## Scope

- Change: `complete-codex-parity-and-native-integration`
- Program tracking: GitHub issue `#218`
- Slice: C
- Gate state: `validated`
- Runtime surface: neutral contracts, effective catalog, Codex skill resources and optional skill metadata.

## Requirement Evidence

| Requirement | Evidence |
| --- | --- |
| Shared invariants have a tool-neutral source | `.lufy/contracts/README.md` inventories delivery, Result Contract, PR-review framework/template, role governance and methodology ownership. |
| Adapters do not duplicate complete contracts | `.lufy/contracts/` owns the full shared documents; `.opencode/policies/delivery.md` and `.opencode/templates/result-contract.md` are compatibility pointers. |
| Codex skills are self-contained | `git-delivery` references neutral contracts; `pr-reviewer` loads installed `references/review-framework.md` and `assets/report.html`; no effective Codex `SKILL.md` contains `.opencode/`. |
| Catalog projects reusable resources | The same neutral PR-review sources are installed both under `.lufy/contracts/` and the Codex skill resource paths, with SHA-256-managed catalog entries. |
| Agent semantics remain intact | Existing tests for native/emulated/inline execution, permission isolation and one-shot empty-result recovery remain green; the eight Codex role files are still installed. |
| Metadata remains optional | `agents/openai.yaml` is present only for `pr-reviewer` and `git-delivery`; discovery still depends on required `SKILL.md` name/description, and delivery disables implicit invocation. |
| Codex-only installation has no accidental OpenCode dependency | An isolated install produced no `.opencode` directory, no `.opencode/` references in installed Codex skills, and passed deep verification. |

## RED/GREEN Evidence

RED contract test:

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./internal/harnesscatalog -run TestCodexEffectiveCatalogIsSelfContained
```

Initial result: failed because `.agents/skills/git-delivery/SKILL.md` had a mandatory `.opencode` reference.

Focused GREEN suite:

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./internal/harnesscatalog ./internal/assets ./internal/adapters/tool/codex ./internal/skillregistry ./internal/installer ./internal/verify
```

Result: passed for every package.

Grouped suite:

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./...
```

Result: passed for every package.

Repository gate:

```text
scripts/validate.sh
```

Result: passed. Whitespace and PR guard against `origin/develop`, action pinning, release checks, workflow YAML, harness coupling, format-dispatch smoke, Go tests, coverage `80.0%`, and Go build passed. `shellcheck` was unavailable and the gate reported the omission explicitly.

Codex-only smoke:

```text
lufy-ai install --target <temp> --tool codex --methodology-tier T3:none --yes
test ! -e <temp>/.opencode
rg -n "\\.opencode/" <temp>/.agents/skills
lufy-ai verify --target <temp> --tool codex --deep
```

Result: installation and deep verification passed; the negative path/reference assertions passed; 8 role TOMLs and 18 effective skills were discovered. The expected context-graph warning remained `not_available` because the temporary target had no built graph.

SDD artifact validation:

```text
lufy-ai sdd validate --change complete-codex-parity-and-native-integration --strict --target <repo>
```

Result: `valid`, Full mode, `24/38` tasks complete after the Slice C join; the automatic overview was refreshed.

## Codex Metadata Basis

The implementation follows the official Codex skill structure: `SKILL.md` keeps required `name` and `description`, progressive resources live in `references/` and `assets/`, and optional UI/policy metadata lives in `agents/openai.yaml`.

Reference: <https://learn.chatgpt.com/es-419/docs/build-skills>

## Residual Risks

- Main OpenSpec specs still describe the historical OpenCode canonical paths; their convergence belongs to the final validated sync, not this implementation slice.
- OpenCode compatibility pointers intentionally remain installed for older commands and links.
- The context graph was unavailable locally and in the temporary smoke target; directed repository inspection and executable tests supplied the primary evidence.
- Runtime E2E across all methodology combinations remains Slice D.
