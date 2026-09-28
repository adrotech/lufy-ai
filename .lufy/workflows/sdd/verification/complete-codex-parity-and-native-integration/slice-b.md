# Verification: Slice B - Native Codex lifecycle and policy

## Scope

- Change: `complete-codex-parity-and-native-integration`
- Program tracking: GitHub issue `#218`
- Slice: B
- Gate state: `validated`
- Runtime surface: Codex project hooks, execpolicy rules, CLI lifecycle command, deep verify and doctor.

## Requirement Evidence

| Requirement | Evidence |
| --- | --- |
| Codex installation provides a useful native lifecycle | `.codex/hooks.json` activa `SessionStart`, `SubagentStop`, `Stop` y `SessionEnd` mediante `lufy-ai lifecycle codex`, con alternativa Windows, timeouts acotados y degradación cuando la CLI no está disponible. |
| Lifecycle protects private content | `internal/codexlifecycle` sólo consume metadata explícita, no lee `transcript_path`, limita `additionalContext` y prueba que no filtra un marcador privado. |
| Subagent and stop hooks preserve workflow gates | Las pruebas cubren payload vacío y estado de un change activo; el hook informa recuperación y conserva `tasks.md` sin mutaciones. |
| Codex rules enforce conservative command policy | `codex execpolicy check` devuelve `prompt` para `git commit` y `gh pr create`, `forbidden` para `git reset --hard`, y ninguna coincidencia para `git status --short`. |
| Deep diagnostics are adapter-aware | Un target Codex-only recién instalado pasa `verify --deep` y `doctor`; ambos validan `.codex` y no solicitan hooks/plugins OpenCode. |

## GREEN Evidence

Focused tests:

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./internal/codexlifecycle ./internal/codexsurface ./internal/cli ./internal/verify ./internal/governance
```

Result: passed for every package.

Grouped suite:

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./...
```

Result: passed outside the filesystem/network sandbox required by the existing loopback `httptest` fixtures in `upgrade` and `versioncheck`.

Repository gate:

```text
GOCACHE=/private/tmp/lufy-go-cache LUFY_AI_VALIDATE_BASE=develop scripts/validate.sh
```

Result: passed outside the loopback-restricted sandbox. Whitespace and PR guard against `origin/develop`, action pinning, release checks, workflow YAML, harness coupling, format-dispatch smoke, Go tests, `go vet`, coverage `80.0%` and Go build passed. `shellcheck` was not available and the repository gate reported that omission explicitly.

Runtime probes:

```text
codex --version
codex features list
codex execpolicy check --pretty --rules .codex/rules/lufy.rules -- git commit -m probe
codex execpolicy check --pretty --rules .codex/rules/lufy.rules -- git status --short
codex execpolicy check --pretty --rules .codex/rules/lufy.rules -- git reset --hard
codex execpolicy check --pretty --rules .codex/rules/lufy.rules -- gh pr create --base develop
```

Result: Codex CLI `0.144.5`; `hooks` and `multi_agent` are stable; policy decisions were `prompt`, no match, `forbidden`, and `prompt`, respectively.

Codex-only smoke:

```text
lufy-ai install --target <temp> --scope project --tool codex --yes
lufy-ai verify --target <temp> --tool codex --deep
lufy-ai doctor --target <temp>
```

Result: install, deep verify and doctor passed; the diagnostics reported the four Codex lifecycle events and conservative rules without OpenCode recovery.

SDD artifact validation:

```text
lufy-ai sdd validate --change complete-codex-parity-and-native-integration --strict --target <repo>
```

Result: `valid`, Full mode, `16/38` tasks complete after the Slice B join.

## Compatibility Decision

The official configuration reference documents `agents.max_concurrent_threads_per_session`, but Codex CLI `0.144.5` parsed that scalar as an `AgentRoleToml` and rejected the project config. Because the limit is optional, Slice B removes both legacy `max_threads`/`max_depth` and omits the new limit until the supported runtime accepts it. `codex features list` is the runtime gate for the shipped config.

## Residual Risks

- Project hooks remain subject to Codex project trust and can be declined by the user; Lufy does not bypass that boundary.
- `SessionStart` ensures the derived skill registry and is intentionally the only lifecycle event that writes derived state.
- Hooks rely on the installed `lufy-ai` binary containing the lifecycle command; managed asset rollout and release packaging remain delivery work.
- Contracts and skills that still depend on `.opencode` are deferred to Slice C.
