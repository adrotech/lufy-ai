# Arquitectura inline

`develop` contiene el modelo incompatible descrito en [el flujo vigente](harness-workflow.md). La release estable anterior no instala este modelo.

La CLI Go concentra operaciones locales; el agente trabaja inline y elige exploración, propuesta, implementación y comprobaciones según el resultado solicitado. La configuración admite `execution.mode: inline` y delegación `off`, `on_request` o `auto` (default). No se instala un router ni un catálogo de roles o metodologías.

## Componentes

| Componente | Responsabilidad |
| --- | --- |
| `internal/core/domain`, `projectconfig` | Ejecución y configuración schema 2; rechazan formatos retirados. |
| `internal/adapters/tool/codex`, `harnesscatalog` | Una superficie project-local, 16 archivos seleccionados y bloque gestionado en AGENTS.md. |
| `internal/assets`, `state` | Catálogo, SHA-256, policies, ancestors y manifest schema 4. |
| `installer`, `syncer`, `uninstaller`, `verify`, `status` | Planificación, aplicación y diagnóstico conservando archivos ajenos. |
| `backup`, `upgrade` | Backups schema 2 vinculados al estado 4 y comprobación de generación antes de actualizar. |
| `contextgraph` | Índice lexical de archivos, símbolos y referencias, impacto Git y métricas opcionales de eventos. |
| `surfaceplan` | Plan read-only de superficies, contratos y comandos declarados; no los ejecuta. |
| `memory`, `skills` | Memoria Obsidian y registro derivado de cinco skills locales. |
| `evidence`, `reviewevents`, `runledger` | Evidencia explícita opcional, vinculada al candidato cuando corresponde; no autoriza acciones. |
| `adaptive` | Experimento opt-in de asignaciones acotadas; está deshabilitado por default. |

## Fuentes y datos derivados

Los archivos humanos de la raíz son la fuente de los assets; `internal/assets/embedded` es su mirror de distribución. El chequeo de acoplamiento exige paridad y rechaza archivos embebidos extra.

La configuración y la memoria son fuentes persistentes. Context Graph y el registro de skills se reconstruyen. El grafo usa extractores v3: los índices previos quedan stale. Ya no interpreta escenarios SDD ni calcula cobertura scenario-task-test; `context review` combina observaciones Git y límites de revisión declarados, con estado del índice orientativo.

## Recuperación

Las mutaciones conservan locks, hashes, rechazo de symlinks, conflictos explícitos y backups. Un restore público valida generación, hashes y payload de configuración/estado antes de escribir. El rollback interno puede recuperar exactamente un preimage dañado. La retención sólo elimina backups reconocidos de la generación actual.

La sustitución de una instalación anterior se ensaya y ejecuta como operación puntual conjunta de ejecutable, estado y proyecciones; no hay lector de estados anteriores ni migrador en el runtime. Ver [instalación](installation.md) y [estado y decisiones preservadas](status.md).
