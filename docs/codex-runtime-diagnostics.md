# Diagnóstico de Codex

`lufy-ai doctor --target <repo> --codex <path-or-command> --json` añade un reporte `codex_runtime` al doctor existente. Sin `--codex`, la salida mantiene su contrato anterior. Para inspeccionar el binario del PATH se usa `--codex codex`; para elegir el incluido en la aplicación se pasa su ruta literal entre comillas.

El servicio ejecuta únicamente `--version` y `features list` con un deadline total de cinco segundos, stdout limitado a 64 KiB por proceso y stderr descartado. No inicia modelos ni consulta autenticación. `path_executable` identifica el candidato del PATH para comparar con la selección explícita; una diferencia no es por sí misma un error. Los códigos de causa omiten stdout y errores crudos.

`available` exige versión parseable y al menos una flag conocida válida; `partial` conserva una de las dos evidencias y `unavailable` indica ausencia de ambas. Sólo se conservan `multi_agent`, `hooks`, `apps` y `plugins`. Una flag conocida duplicada invalida el probe de flags. `native_delegation: unknown` indica que este diagnóstico no ejecutó un handoff nativo. La configuración instalada, las flags anunciadas y una ejecución operativa son evidencias distintas.

El resultado global del doctor sigue comprobando instalación, drift, memoria y grafo: un runtime disponible no convierte un repositorio sin manifest en instalación válida. CommandContext cancela el proceso y WaitDelay limita la espera de sus pipes; no se afirma contención universal de descendientes.

## Activación local de este repositorio

El bloque 2 activó explícitamente Codex, memoria Obsidian con política Git `ignored` y grafo local. No había `project.yaml`; se creó una configuración mínima sin `workflow_limits`, por lo que no hay budgets previos que atribuirle. No se cambiaron modelos, credenciales ni permisos globales. La configuración local no se distribuye como asset del producto.

La inicialización estándar de memoria puede migrar layouts legacy. Se inicializó primero el scaffold en el worktree aislado y se copiaron exclusivamente cinco archivos nuevos al checkout principal con creación exclusiva y rechazo de symlinks/archivos preexistentes. Se preservó su layout legacy y su trabajo mixto. `memory validate` confirmó `ready` sin backlinks rotos; `context build` produjo un grafo local sin errores y excluyó archivos sensibles. Estos resultados describen el checkout inspeccionado, no prueban capacidades del runtime ni sustituyen evidencia del candidato de una PR.

Para un repositorio sin layout legacy pendiente, los comandos habituales son `lufy-ai memory init --target <repo>`, `lufy-ai memory validate --target <repo> --json` y `lufy-ai context build --target <repo> --json`. Si hay configuración existente, se revisa antes de activar opciones y se conserva su contenido. Para límites de workflow se consulta exclusivamente `workflow_limits` cuando esté configurado.
