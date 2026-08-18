# Proposal: Completar Lufy SDD Full

## Why

`lufy-sdd/full` hoy es seleccionable e instala una estructura de directorios, pero no puede reemplazar a OpenSpec: no crea changes, no valida artifacts, no sincroniza deltas con specs activas y no archiva con gates. El adapter incluso delega su verificación al installer estructural. Esta diferencia entre el nombre `full` y el comportamiento real deja a los repos dependientes de OpenSpec para el ciclo T1.

La siguiente evolución debe convertir Lufy SDD en una metodología local, portable y sin runtime externo obligatorio, conservando los contratos de calidad que ya usa el harness: requisitos delta, escenarios `WHEN`/`THEN`, separación entre implementación y cierre, y mutaciones seguras.

## What Changes

- Definir un workflow Lufy SDD v1 bajo `.lufy/workflows/sdd/` con configuración gestionada, specs activas, changes, decisiones, verificación y archive.
- Agregar un motor Go nativo para descubrir, crear, inspeccionar y validar changes sin depender del binario `openspec`.
- Agregar `lufy-ai sdd new|status|validate|sync|archive` con salida humana y JSON determinística, y creación Full/Lite explícita.
- Generar automáticamente un `change-overview.html` autocontenido junto a los Markdown, sin requerir otro comando o skill, y refrescarlo durante validate, sync y archive.
- Integrar Full/Lite nativo en `sdd-router`, orchestrator, roles especialistas y Result Contract para que el harness respete la metodología seleccionada por tier en vez de asumir OpenSpec.
- Reutilizar deltas `ADDED`, `MODIFIED` y `REMOVED`, y exigir escenarios testables con `WHEN` y `THEN`, para facilitar interoperabilidad y migración.
- Sincronizar deltas hacia specs activas con preflight completo, rechazo de ambigüedad y escrituras recuperables.
- Archivar únicamente changes válidos, con tareas completas y deltas sincronizados, preservando evidencia de verificación.
- Convertir `VerifyWorkflow` del adapter `lufy-sdd` en una verificación real de la superficie instalada.
- Instalar acciones/skills Lufy SDD para los tool adapters escribibles sin introducir dependencias OpenSpec en una selección exclusivamente Lufy SDD.
- Documentar convivencia y migración manual desde OpenSpec; no migrar silenciosamente artifacts existentes.

## Capabilities

### New Capabilities

- `lufy-sdd-workflow`: lifecycle nativo de changes y specs para Lufy SDD Full.

### Modified Capabilities

- `tier-methodology-routing`: `lufy-sdd/full` deja de ser solo una foundation estructural y declara assets ejecutables y verificables.

## Impact

- CLI: nuevo namespace `lufy-ai sdd`, selector `--mode full|lite` en creación y paquete interno dedicado.
- Managed assets: nuevos archivos bajo `.lufy/workflows/sdd/`, más wrappers de comandos/skills filtrados por metodología.
- Adapter: validación semántica y reporte accionable en vez de un check informativo fijo.
- Harness: routing methodology-aware, handoffs con overview automático y delivery de artifacts Lufy SDD user-owned sin falsos positivos de metadata interna.
- Compatibilidad: `openspec` continúa siendo el default; seleccionar `lufy-sdd` no elimina ni migra directorios OpenSpec preexistentes.
- Seguridad: nombres de change/capability validados, rechazo de symlinks y paths inseguros, preflight antes de sync/archive y rollback acotado ante fallas de escritura.

## Non-Goals

- No implementar ni ampliar el adapter Claude Code.
- No descargar plugins, schemas o workflows remotos.
- No borrar `openspec/` ni convertir changes automáticamente.
- No prometer compatibilidad byte a byte con el CLI OpenSpec.
- No sumar templates específicos por stack ni nuevos subagentes de dominio.
- No convertir el HTML en fuente de verdad ni introducir un comando público exclusivo para renderizarlo.
