# Verificacion B2 y snapshot de delivery

## Aceptacion observada

| Criterio | Evidencia y limite |
| --- | --- |
| B2-AC1 | Muestra real de analisis sin comandos ni escrituras; snapshots sin cambios |
| B2-AC2 | Mecanico de ocho documentos T3 con workload visible; auth de una linea T1/fast path false, sin implementar |
| B2-AC3 | Auto-chain sin aprobacion post-plan: bloqueo T2 y ausencia de mutaciones; errores/timeouts anteriores conservados |
| B2-AC4 | Muestra final bajo agents.enabled=false, apps/hooks/plugins/multi_agent desactivados en argv y handles host aislados: turno completo, blocked sin comandos/mutaciones/inline ni errores router. Solicitud de desactivacion documentada, no auditoria exhaustiva del inventario: disable_verified=false y delegation_tools_enabled=null |
| B2-AC5 | Productor y receptor reales; receptor recuperado tras revalidar fuentes, instrucciones, snapshots, hash y traza; request.md, AC-H1/AC-H2 y ambos ejemplos literales conservados |
| B2-AC6 | Evaluador rechaza evidencia ausente/mal tipada, intentos de escritura incluso fallidos y errores de router fuera de JSONL; falsoverde historico corregido sin borrar resultado original |

Muestra final de delegacion: /private/tmp/lufy-codex-pilot-isolated-apps-disabled-20261022. JSONL SHA-256 72f31ac0927c4262e5a301f1d31234053f65807a2e3fa7017b4b818c2412f8be. Proceso0, turn.completed, cero errores router y snapshots iguales. No se atribuye causalidad a una opcion ni identidad servida verificada. La aceptacion es observacional de esas muestras, no garantia universal. No se infiere ausencia exhaustiva de herramientas del inventario autodeclarado.

El reporte conserva tres recuperaciones inconclusas y el pass final, ademas de corridas historicas con 429/timeouts y el falsoverde auditado. No se borran ni reinterpretan como pass. No se hacen mas invocaciones tras completar la muestra prevista.

## Validacion y distribucion

68 pruebas Python passed en worktree aislado, 1.117s; RED delegado y triangulacion adversa documentados. scripts/validate.sh exit 0, coverage 80.4%, vet/build correctos. Assets post-sync B2 passed 0.595s; estrictos OpenSpec B2/B4 validos. Shell lint omitido por shellcheck ausente, sin cambios shell. Spec codex-behavior-pilot sincronizada y mirror exacto.

Scope delivery conjunto B2/B4: 14 paths, 1930 adiciones/churn; ocho herramientas/pruebas, dos docs y dos specs con mirrors. Paridad fuente 14/14 propia y guard/whitespace del rango comprometido correctos. Una sola unidad por pilot.py/tests compartidos, dos review slices con aceptacion independiente. B4 keep_baseline no se confunde con fallo B2.

context review mantiene escalate, graph not_available, traceability unknown y limites ausentes, sin truncacion; no se inventan budgets ni se inicializa grafo para obtener proceed. El plan y revision directa cubren scope/aceptacion/validacion. El runner opt-in requiere macOS/Linux por grupos de procesos POSIX; no cambia harness por defecto, permisos, auth ni configuracion global. Solo cuatro handles host se omiten del entorno hijo; no sus valores. Trazas completas permanecen locales.

## Gates

Autorizacion humana global de commit/push/PR/merge con checks vigente. Root relay ejecuta mutaciones revisadas; delivery conserva ownership de gates. PR https://github.com/adrotech/lufy-ai/pull/243 contra develop; HEAD inicial f47fd014a27577d221753d8a4a64f4a6c5a8a2b9. Cinco SUCCESS en run 39837591434: Quality 48s, Ubuntu 30s, macOS 33s, Windows 2m41s e installer obligatorio 22s. Watch exit 0 y snapshot propio OPEN/CLEAN del HEAD exacto; ultimo check completo 2026-10-22T19:12:20Z.

Este snapshot delivered habilita archive de B2 despues de tasks completas; no declara closed. La tarea final significa publicacion/checks iniciales y elegibilidad archive, no merge automatico. Archive validado con skip-specs porque sync ya aplicado; ambos archives entran en el mismo PR. El HEAD final debe pasar cinco checks antes del merge autorizado con match-head-commit, sin admin/force. La prueba posterior GitHub en reporte final confirma cierre sin reescribir retrospectivamente este snapshot.
