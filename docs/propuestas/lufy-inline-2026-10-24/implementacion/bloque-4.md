# Bloque 4: instalación personal inline

Estado: implementado, validado y activado localmente el 27/10/2026. No se realizó commit, push, PR, release ni promoción de ramas.

La instalación personal de `/Users/adrianrojas/Desktop/projects/lufy-ai` usa configuración schema 2, manifest schema 3, ejecución `inline` y delegación `off`, preservando la preferencia previa de no paralelizar. Se conserva la ausencia explícita de `workflow_limits`, junto con memoria, contexto y restricciones útiles del proyecto.

## Sustitución aplicada

- Inventario previo: **108 assets gestionados, cero drift y cero faltantes**. Se comprobaron los hashes antes de preparar y antes de aplicar.
- Catálogo actual: **16 assets**, **5 skills**. Retiradas 51 instrucciones/tooling gestionados; conservados 44 documentos gestionados históricos y el resto de propuestas, memoria y datos propios.
- Sustituidos conjuntamente `~/.local/bin/lufy-ai`, `tools/lufy-cli-go/bin/lufy-ai` y `tools/lufy-cli-go/lufy-ai`, configuración, manifest, ancestors, guía `AGENTS.md`, hooks y mirrors seleccionados. Se regeneraron registry de skills y grafo de contexto.
- Los tres ejecutables tienen SHA-256 `b39e5c6e4bf319860d184124d1ed1ea11b4d07c573584ab71089d22b897479d9`. Son una compilación de desarrollo (`dev`) de la implementación sobre la base `1cc18d4647afc61ea1f6b1cb726ba74128fded5f`, no un release publicado; ese commit identifica la base, no cambios aún sin commit.
- La previsualización se generó fuera del proyecto y no escribió en la instalación. El procedimiento es puntual, vinculado al snapshot real mediante hashes, journal y preimágenes; no existe un migrador permanente ni lectura de formatos anteriores en el runtime.

## Correcciones necesarias para el resultado

`upgrade` comprueba que los checksums declaren el contrato actual de ejecución/configuración/state antes de descargar el artefacto y reemplazar el binario. El proceso de release genera esa declaración desde los schemas canónicos. Una generación anterior se rechaza conservando el ejecutable actual. Esto comprueba formato declarado y checksum, no añade una firma ni demuestra el comportamiento de un binario externo.

El ensayo integrado descubrió que `sync` rechazaba una configuración actual sólo por existir `.lufy/sdd`, una fuente histórica de este repositorio. El preflight comprueba las rutas activas de configuración/state y sus lectores tipados validan schemas; los documentos no activan metodologías. Retirados el comando `migrate-layout` y la migración de `setup`. Pruebas de regresión comprueban que sync/setup conservan la fuente sin crear `.lufy/workflows/sdd`.

Actualizadas las guías gestionadas y el wrapper estricto `scripts/install.sh`. Una instalación nueva no recibe OpenSpec ni una cadena de roles desde el catálogo inline.

## Evidencia

- `GOCACHE=/private/tmp/lufy-inline-go-cache scripts/validate.sh`: **PASS, exit 0**. Suite Go, cobertura **80,4%** con mínimo 80,0% intacto, 169 pruebas Python offline, `go vet`, build y controles estáticos disponibles, incluyendo whitespace contra `origin/develop`. Los tests `httptest` se ejecutaron con permiso para loopback.
- RED observado antes de corregir `sync`: `TestSyncIgnoresHistoricalSddSource` falló con «layout anterior no soportado». GREEN en la suite final. El test de upgrade reprodujo previamente el reemplazo por una generación anterior y pasa con el guard actual.
- Ensayo del snapshot: drift detiene antes de escribir; interrupción deja estado parcial distinguible; reanudación es idempotente, sin bloques duplicados; edición posterior detiene restauración sin sobrescribirla. Verify profundo y memoria válidos.
- Recuperación conjunta adicional desde una copia de los recursos del respaldo durable, sin depender del candidato temporal: después de cambiar los tres binarios, config, manifest y proyecciones, se restauraron **754 archivos y 269 directorios**, con bytes/modos comprobados. La CLI anterior `v0.6.25` volvió a leer el manifest schema 2. Fue un ensayo aislado; la instalación personal conserva la versión nueva.
- Release local `file://` de la generación actual: upgrade de una copia del binario y sync pasaron sin reponer retirados. Un checksum sin declaración actual se rechazó sin cambiar el ejecutable. `shasum -a 256 -c` acepta la cabecera generada.
- Instalación real: `verify --deep --json` devuelve `ok: true`, **0 failures y 0 warnings**; `status` registra **0 drift y 0 faltantes**. Sync desde catálogo embebido y desde fuente del repositorio pasa; el wrapper resuelve el binario local actual.
- **1.350 archivos preexistentes protegidos** conservaron sus hashes antes/después de sync, incluyendo código Go ajeno, OpenSpec, documentos, memoria, runtime y `.opencode`. HEAD, rama e índice coincidieron en esa prueba. En el control posterior de cierre cambió el hash binario del índice, con `git diff --cached --raw` vacío; no se reescribió para forzar el hash anterior. HEAD y rama permanecen iguales y no hay cambios preparados. Después se registró deliberadamente esta documentación y una nueva nota durable de memoria con su índice: **15 notas, cero backlinks rotos**.
- El hook `SessionStart` invocado directamente reporta 5 skills ready y la memoria actual. No creó eventos automáticos de runtime. Esto verifica CLI/configuración; no demuestra recarga de caches del host ni calidad de nuevas ejecuciones con modelos.

## Respaldo y restauración

Respaldo privado fuera del repositorio:

`/Users/adrianrojas/.local/share/lufy-ai/backups/inline-20261027`

Incluye snapshot completo, hashes/modos, payload actual para reanudación, plan, previsualización, journal, preimágenes físicas y script puntual. El archivo `snapshot.tar.gz` tiene SHA-256 `53142f48ce32f02fd680bb7a6adf6e400e11ee92941cee38bc1ca2f961b77abe`. No se incorpora contenido privado del respaldo a Git.

Para volver deliberadamente al conjunto anterior, con permisos sobre la instalación personal:

```sh
python3 /Users/adrianrojas/.local/share/lufy-ai/backups/inline-20261027/replace_personal.py restore-live
```

El procedimiento rechaza modificaciones posteriores de assets afectados. No es el `restore` mínimo de la CLI ni promete interoperabilidad entre schemas. Los documentos propios y la memoria no se sustituyen con archivos antiguos durante este rollback. El snapshot completo queda disponible como recuperación adicional.

## Ubicación del código y pendientes

La implementación fuente permanece en `/private/tmp/lufy-inline-20261026`, rama `codex/lufy-inline`, con el patch acumulado `bloques-1-a-4.patch`. El checkout original conserva su rama y sus modificaciones de código previas; **no debe recompilarse desde ese código anterior para reemplazar el nuevo ejecutable** antes de integrar la implementación. La activación local no equivale a integrar/publicar el código ni terminar toda la refactorización.

El bloque 5 continúa con la evaluación por familias de tareas y evidencia comparable. No se ejecutaron modelos nuevos ni se reintentaron los ensayos nativos previamente rechazados. La limpieza general de módulos/archivos históricos fuera del catálogo activo sigue en el bloque 6.

## Archivos de evidencia

- `bloque-4-preview.md`: inventario legible de operaciones.
- `bloque-4-validation.log`: validación agrupada final.
- `bloque-4-rehearsal-result.json`, `bloque-4-full-restore-result.json` y `bloque-4-upgrade-result.json`: resultados de ensayos.
- `bloque-4-live-result.json` y `bloque-4-live-commands.json`: resultado posterior y comandos de la instalación real.
- `bloque-4-replace-personal.py`, `bloque-4-full-restore-rehearsal.py`, `bloque-4-check-live.py`, `bloque-4-upgrade-rehearsal.py`: procedimiento puntual y comprobaciones, dependientes de los recursos privados del respaldo; no son comandos del producto.
- `bloques-1-a-4.patch`: respaldo de código e instrucciones de los cuatro bloques sobre la base citada, con archivos nuevos y reverse-check contra el checkout de implementación.
