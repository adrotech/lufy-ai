# Bloque 5: evidencia ligada al candidato Git

## Why
Un Result Contract válido identifica su contenido y autoridad de transición, pero no demuestra que siga correspondiendo al árbol revisado. Se necesita detectar evidencia obsoleta sin reescribir contratos ni recrear CAS/receipts.

## What Changes
- Añadir `lufy-ai candidate capture|verify` read-only para candidatos comprometidos y limpios.
- Emitir sidecar `candidate-evidence/v1` con hashes de repositorio, base, candidato, diff de paths/modos, políticas canónicas/config local y evidencia canónica validada para un rol.
- Rechazar worktrees con cambios tracked/untracked de producto; permitir únicamente config local conocida, ligada por hash. Leer Git sin hooks, external diff, textconv ni optional locks.
- Detectar cambios de HEAD/base, contenido/modos/paths, política, configuración, evidencia o repositorio; no avanzar gates ni ejecutar delivery.

## Capabilities
### New Capabilities
- codex-candidate-evidence: binding explícito y verificación de vigencia.

## Aceptación y workload
WHEN la evidencia válida corresponde al mismo candidato limpio, base y políticas THEN verify devuelve current.
WHEN cambia cualquiera de esas identidades o hay cambios sin commit THEN verify rechaza sin modificar evidencia ni Git.
WHEN el sidecar es malformed, contiene claves ajenas/duplicadas o evidencia inválida THEN no se acepta como vigente.
WHEN se verifica un candidato THEN no se ejecutan hooks, modelos ni transiciones de workflow.

T1 por nuevo contrato público de evidencia. Una unidad coherente servicio/CLI/tests/docs/spec y mirror; no se agrega estado durable ni reemplaza Result Contract v1. Programa y entrega ya autorizados. workflow_limits no disponible; graph del worktree not_available, fallback dirigido a Result Contract, ledger y referencia Gentle fijada. Overview HTML not_available; artifact branching no aporta alternativas objetivas.
