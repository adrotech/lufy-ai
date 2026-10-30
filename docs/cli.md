# Referencia de la CLI

La implementación está en `tools/lufy-cli-go`. `lufy-ai --help` y `lufy-ai <comando> --help` muestran la interfaz del binario instalado; en grupos se consulta también el subcomando. Varios handlers imprimen ayuda por stderr: conservá ambos streams al capturarla. Los flags y sus defaults pertenecen a esa implementación.

La superficie soportada es Codex project-local. Aunque algunos helps compartidos enumeran `global|both`, el adapter actual rechaza esos scopes. Usá `--target` explícito para otro proyecto. No todos los comandos aceptan `--json`.

## Operaciones disponibles

| Comando | Resultado y efectos |
| --- | --- |
| `init`, `scan` | Crean/actualizan configuración detectando stacks y superficies. |
| `plan` | Plan read-only de superficies, contratos y validaciones declaradas. |
| `install`, `sync` | Preview con `--dry-run`; aplican assets con `--yes`, ownership, hashes y conflictos. |
| `uninstall` | Retira assets gestionados conforme a sus policies y conserva backup. Revisar preview. |
| `status`, `info`, `doctor`, `verify` | Diagnóstico sin reparar; `verify --deep` incluye hashes. |
| `conflicts plan` | Plan read-only de conflictos y recomendaciones. |
| `merge`, `pin`, `unpin` | Resolución de `.lufy-new` o congelado de assets. Mutan estado/archivos. |
| `backup`, `restore` | Snapshot mínimo y recuperación de la generación actual. |
| `memory init/status/validate/search/capture/connect/index` | Consulta y mantenimiento de memoria Obsidian. |
| `skills ensure/refresh/status` | Mantienen o consultan el registry derivado. |
| `context scan/build/status/query/path/explain/diff/review/metrics` | Índice lexical, observaciones Git y métricas explícitas. `build` persiste derivados. |
| `evidence inspect` | Interpreta comprobaciones declaradas; no las ejecuta. |
| `candidate capture/verify` | Vincula/verifica evidencia contra un candidato Git comprometido limpio. |
| `review start/complete/rework/reopen` | Registra metadata explícita; no hace una review. |
| `run record/checkpoint/status/summary/verify/prune` | Ledger causal local. Registro, reparación y poda son mutantes. |
| `adaptive recommend/assign/yield/status` | Experimento opt-in; asignaciones/yields requieren `--record`. No lanza agentes. |
| `pr guard` | Inspecciona paths internos/ignorados en el diff. No crea PRs. |
| `lifecycle codex` | Orientación compacta de sesión para el hook del host. |
| `upgrade` | Descarga una versión fija y reemplaza el ejecutable tras verificar integridad/generación. |
| `setup` | Orquesta instalación, configuración, memoria, grafo y verify; también consulta versiones. |
| `version`, `menu` | Identidad/contrato de runtime y menú interactivo TTY. |

## Consultas habituales

```sh
lufy-ai version --runtime-contract
lufy-ai status --target /ruta/al/proyecto
lufy-ai info --target /ruta/al/proyecto --json
lufy-ai doctor --target /ruta/al/proyecto --codex codex --json
lufy-ai plan --target /ruta/al/proyecto --json
lufy-ai context review --target /ruta/al/proyecto --base origin/develop --json
lufy-ai skills status --target /ruta/al/proyecto --json
```

`plan` puede recibir `--files a,b`, `--surface <id>` o `--base <ref>`. Resuelve superficies; no genera propuestas ni autoriza acciones. Requiere `project_profile.surfaces`; si no hay superficies, ejecutá `scan` y revisá la detección. Sin `--files`, consulta el diff Git y necesita un repositorio con la referencia base disponible. Un directorio vacío todavía no ofrece ese contexto.

Los ejemplos de instalación están en [installation.md](installation.md); los de memoria/contexto en [memory-context.md](memory-context.md); los de evidencia en [evidence.md](evidence.md). No existe `result validate/transition`, selección `--methodology` ni `context coverage/trace` en el flujo actual.
