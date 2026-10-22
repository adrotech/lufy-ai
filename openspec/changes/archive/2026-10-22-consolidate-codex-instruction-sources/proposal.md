# Bloque 6: instrucciones Codex coherentes

## Why
Tres roles repiten el mismo bloque proporcional sin una fuente de edición única. El mapping duplica resolución de skills y conserva una advertencia de hooks anterior al lifecycle soportado. La documentación también debe distinguir límites del rol y sandbox efectivo heredado.

## What Changes
- Identificar `.lufy/contracts/delivery.md` como autoridad y `.codex/instructions/proportional-workflow.md` como fuente compacta de wording compartido, sin cambiar sus reglas.
- Materializar ese mismo bloque inline en orchestrator/sdd-router/implementer y sus mirrors mediante renderer offline con check de drift. Mantener inline hasta que el experimento del bloque 7 aporte evidencia; no adoptar referencias por intuición.
- Unificar resolución de skills, modo de delegación y límites de lifecycle en mapping/README Codex, sin permisos nuevos ni afirmaciones de aislamiento no observado.
- Disparar CI también ante cambios sólo en .codex o el renderer, sin depender de que se haya sincronizado su mirror.
- Verificar paridad de superficies Codex y coherencia con las reglas existentes sin ejecutar modelos.

## Capabilities
### New Capabilities
- codex-instruction-consistency: edición compartida y detección de divergencia distribuida.

## Aceptación y workload
WHEN cambia el wording compartido o un mirror diverge THEN check rechaza hasta render/sync explícito.
WHEN se renderiza THEN sólo se actualizan segmentos marcados y mirrors Codex esperados, preservando instrucciones particulares y permisos.
WHEN se consulta mapping THEN diferencia hooks configurados/observados de native delegation, conserva recovery de handoff vacío y no trata una mención @rol como identidad inline.
WHEN no hay drift THEN validación offline pasa sin modelos ni cambios globales.

Slice T1 dentro de programa T1: aunque la consolidación es controlada, incorpora una nueva herramienta de distribución de instrucciones y aceptación transversal; se conserva OpenSpec completo y no se cambia al adapter T2 configurado. Sin nueva autoridad. Workload: una unidad de 19 paths fuente/mirror cubiertos por el plan con fuente, tres consumidores/mirrors, documentación, checker/tests y spec; los mirrors conservan churn visible. Aceptación y autorización cubiertas por el plan vigente. workflow_limits no disponible; graph worktree not_available, fallback dirigido a mapping/roles/catalog/tests. Overview HTML not_available; branching no necesario.
