# Diagnóstico y recuperación

Empezá por identificar ejecutable, target y formato, sin aplicar cambios:

```sh
lufy-ai version --runtime-contract
lufy-ai status --target /ruta/al/proyecto
lufy-ai doctor --target /ruta/al/proyecto --json
lufy-ai verify --target /ruta/al/proyecto --deep
```

| Observación | Siguiente acción |
| --- | --- |
| Schema no soportada o generación mezclada | Conservar estado y ejecutable. Revisar [recuperación entre generaciones](installation.md); no editar números de schema para forzar aceptación. |
| Drift o conflicto de un asset | `conflicts plan`, inspeccionar local/ancestor/`.lufy-new` y decidir merge o pin explícito. |
| Archivo propio extra | Conservarlo; la ubicación no lo vuelve gestionado. Ver [ownership](adr/0002-managed-directories-extra-files.md). |
| `plan` sin superficies o base Git | Ejecutar `scan`, revisar `project_profile.surfaces` y proporcionar una base Git válida o archivos explícitos con `--files`. |
| Skills stale | `skills ensure --target <repo>` reconstruye el registry. |
| Contexto stale | `context build --target <repo> --json` reconstruye derivados si se necesitan. |
| Backlinks rotos | Inspeccionar notas/enlaces, corregirlos y ejecutar `memory index` y `memory validate`. |
| Hook sin CLI disponible | Agregar el ejecutable actual al PATH del host que inicia Codex y comprobar su identidad. |
| Candidate rechazado | Revisar limpieza, políticas tracked y restricciones Git. No ocultar archivos ni desactivar controles para aceptar. |
| Proyección de run stale | `run verify`; usar `--repair` sólo si la fuente es íntegra. |

## Runtime Codex

```sh
lufy-ai doctor --target /ruta/al/proyecto --codex codex --json
```

`--codex` acepta comando o ruta literal. El probe sólo ejecuta `--version` y `features list`: plazo total de cinco segundos, stdout acotado y stderr descartado. No inicia modelos ni consulta autenticación. Conserva las flags conocidas `multi_agent`, `hooks`, `apps` y `plugins` y reporta `available`, `partial` o `unavailable` según la evidencia obtenida.

`native_delegation: unknown` significa que no hubo un handoff nativo en este diagnóstico. Configuración, flags anunciadas y delegación operativa son evidencias diferentes. Un runtime disponible tampoco convierte un target sin manifest en instalación válida. Si PATH y ruta explícita difieren, comprobá cuál ejecuta el host antes de atribuir capacidades al proyecto.

## Fallos durante una mutación

Conservá el error y la ruta del backup. Un rollback automático sólo recupera la preimagen cubierta por ese backup; el restore público exige payload válido de la generación actual. Revisá su preview y verificá el estado posterior. Las carpetas desconocidas, backups de otra generación y trabajo ajeno no deben limpiarse para obtener un diagnóstico verde.

Los [informes históricos](history/README.md) documentan incidentes concretos; sus rutas privadas y procedimientos puntuales no son comandos generales de reparación.
