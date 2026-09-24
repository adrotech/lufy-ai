# Verificación profunda del harness — PR #230

Fecha: 2026-09-24
Rama: `codex/feat-adaptive-role-allocation`
Base: `origin/develop` (`0` commits detrás, `2` delante antes de estas correcciones locales)
Alcance: runtime adaptativo, Run Ledger, CLI, instalación/sync, assets OpenCode/Codex, Lufy SDD, bootstrap y artefactos de release.

## Resultado ejecutivo

La auditoría profunda original pasó infraestructura, portabilidad, privacidad y suites, pero detectó dos invariantes operativos faltantes: confirmación de assignments sin revalidar recommendation/capacidad/budget y leases vencidas sin transición durable de recuperación. Ambos hallazgos quedaron corregidos con regresiones permanentes, contrato/documentación alineados y spec activa sincronizada.

El gate funcional queda **validado**. El change no está cerrado: faltan commit/push, checks remotos sobre la corrección, merge, cierre de issue y archive, todos sujetos al flujo de delivery autorizado.

## Evidencia acumulada

| Área | Comando/escenario | Resultado |
| --- | --- | --- |
| suite profunda original | `go test -race ./... -count=1` | passed |
| corrección focal con race | `go test -race ./internal/adaptive/... ./internal/runledger ./internal/cli -count=1` | passed |
| stress adaptativo posterior | `go test ./internal/adaptive/... ./internal/runledger -count=30` | passed |
| calidad integral posterior | `scripts/validate.sh` | passed; coverage global 80,4%, build/coupling/assets/YAML/PR guard/whitespace verdes |
| strict SDD posterior | `lufy-ai sdd validate --change experiment-adaptive-role-allocation --strict` | valid; 48/50, digest `a22cfb12f0da8ec5041ba0c27c3716dc87f19b18fb0ea39b6086007bf12ac4bd` |
| sync SDD posterior | `lufy-ai sdd sync --change experiment-adaptive-role-allocation` | synced; 18 requirements actualizados en 3 specs activas |
| fuzzing original | Result Contract ingress, `-fuzztime=10s` | passed; 71.175 ejecuciones |
| instalaciones limpias originales | OpenCode + Lufy SDD, Codex + Lufy SDD, OpenCode + OpenSpec | passed; verify deep, reinstall y sync sin drift/conflicts/errors |
| lifecycle/supply chain original | install/verify/backup/restore, wrapper, hooks, bootstrap y artifacts multi-OS | passed |
| CI remota previa | Quality, Ubuntu, macOS, Windows, installer smoke | 5/5 success; debe repetirse sobre el commit de corrección |

Advertencias no bloqueantes conservadas: memoria Obsidian y Context Graph no inicializados en targets limpios, comportamiento esperado. Los smokes cross-build requirieron `/opt/homebrew/bin/go`. La CI remota verde corresponde al commit previo; no cuenta como evidencia remota de estas correcciones aún no publicadas.

## Resolución de hallazgos

### Resuelto — confirmación autoritativa de assignment

- La recommendation debe seguir siendo el último evento de la proyección; copiar la versión actual del ledger ya no revive una recommendation desplazada.
- El adapter revalida, dentro del mismo snapshot/CAS del append, el máximo de assignments activas y el budget acumulado del actor.
- `recommend` descuenta el budget ya consumido del perfil antes del scoring para evitar sugerencias que no podrán confirmarse.
- Regresiones: recommendation stale con versión actual, capacidad global agotada, budget del actor agotado y propagación de límites desde application.

### Resuelto — recuperación de lease vencida

- La proyección sigue siendo determinista y nunca libera por reloj implícito.
- Un yield ordinario posterior al vencimiento continúa rechazándose.
- La recuperación exige coincidencia exacta de assignment, actor, lease y expected version, más `reason: lease_expiring` y `next_status: waiting`.
- Solo el checkpoint durable `recorded`/`duplicate_noop` libera budget y reencola la demanda.
- Regresión: lease vencida recuperada de forma explícita deja cero assignments/budget activos y un waiting item.

## Coherencia y completitud

- Proposal, design, delta spec, spec activa, contrato neutral, superficies OpenCode/Codex, template, harness y documentación describen las mismas invariantes.
- Los assets root/embedded conservaron paridad y pasaron el coupling/catalog del gate integral.
- La solución mantiene Run Ledger append-only, CAS, fencing, idempotencia, metadata content-free y `gate_advanced=false`.
- No se añadieron autonomía, roles, permisos, autoridad de delivery ni liberación dependiente solo del reloj.

## Gate

- Veredicto funcional: **validado**.
- Estado del workflow: `delivery_pending`.
- PR #230: abierta; requiere publicar las correcciones y repetir checks remotos.
- Issue #222: abierta; debe cerrarse después del merge.
- Tasks SDD: 48/50; solo quedan delivery/cierre y archive.
- Próximo paso: delivery autorizado de estas correcciones; luego checks remotos, merge, cierre de issue y archive.

```yaml
schema_version: result-contract/v1
status: delivery_pending
legacy_fallback: false
executive_summary: Los dos hallazgos del harness profundo fueron corregidos, probados, documentados y sincronizados; queda delivery remoto y cierre.
artifacts:
  changed:
    - tools/lufy-cli-go/internal/adaptive/application/service.go
    - tools/lufy-cli-go/internal/adaptive/adapters/runledger.go
    - tools/lufy-cli-go/internal/cli/app_adaptive.go
    - .lufy/workflows/sdd/changes/experiment-adaptive-role-allocation/
    - .lufy/workflows/sdd/specs/adaptive-yield-protocol/spec.md
    - .lufy/contracts/adaptive-routing.md
    - docs/
  referenced:
    - PR #230
    - issue #222
ledger:
  run_id: not_applicable
  event_id: not_applicable
  status: not_applicable
evidence:
  commands:
    - command: go test -race ./internal/adaptive/... ./internal/runledger ./internal/cli -count=1
      result: passed
      notes: correcciones sin data races
    - command: go test ./internal/adaptive/... ./internal/runledger -count=30
      result: passed
      notes: stress posterior estable
    - command: scripts/validate.sh
      result: passed
      notes: coverage 80.4%, build y gates integrales verdes
    - command: lufy-ai sdd validate --change experiment-adaptive-role-allocation --strict
      result: passed
      notes: valid 48/50
    - command: lufy-ai sdd sync --change experiment-adaptive-role-allocation
      result: passed
      notes: 18 requirements sincronizados
  static:
    - Scenarios, código y regresiones contrastados; contratos root/embedded alineados.
surface_execution:
  schema_version: surface-execution-plan/v1
  source: git_diff
  primary_surface: fullstack
  mode: composed
  active_surfaces:
    - cli
    - harness-assets
    - run-ledger
  validation_rule_ids:
    - full-harness
workflow_decision:
  tier: T1
  program_tier: T1
  slice_tier: T1
  fast_path_allowed: false
  adapter_context:
    tool_id: opencode
    methodology_id: lufy-sdd
    methodology_mode: full
    methodology_required: true
    execution_mode: full-sdd
  workflow_limits_source: workflow_limits
  workflow_limits_paths:
    sizing: workflow_limits.sizing
    routing: workflow_limits.routing
    proposal_slicing: workflow_limits.proposal_slicing_strategy
    delivery_batching: workflow_limits.delivery_batch_strategy
    preflight: workflow_limits.preflight
    stop_rules: workflow_limits.stop_rules
  workload_decision_needed: true
  review_slices:
    - confirmacion autoritativa de assignment
    - recuperacion durable de leases vencidas
  preflight_status: passed
  stop_rule_status: clear
  delivery_batching_guidance: publicar las correcciones en PR #230 y exigir checks remotos antes del merge
  artifact_branching:
    status: not_needed
    stage: not_applicable
    candidate_count: 1
    reason: correcciones acotadas con criterios objetivos
    parallel_allowed: false
    requires_join: false
    candidate_isolation: not_applicable
    merge_plan_required: false
    human_escalation_triggers:
      - not_applicable
risks:
  - checks remotos aun no ejecutados sobre estas correcciones locales
next_recommended:
  owner: delivery
  action: commit, push y actualizar PR #230 con autorizacion explicita
skill_resolution:
  local_skills_used:
    - lufy-sdd-apply
    - lufy-sdd-verify
    - lufy-sdd-sync
  selected_skill_paths:
    - .agents/skills/lufy-sdd-apply/SKILL.md
    - .agents/skills/lufy-sdd-verify/SKILL.md
    - .agents/skills/lufy-sdd-sync/SKILL.md
  bootstrap_recommended: false
  notes: los actions canonicos bajo .lufy/workflows/sdd/actions no estaban presentes; se usaron los fallbacks locales bajo .lufy/sdd/actions.
```
