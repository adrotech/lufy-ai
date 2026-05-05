## 1. Scaffolding Go y estructura base

- [x] 1.1 Crear `go.mod` en carpeta dedicada de tooling (`tools/lufy-cli-go/`) con module path coherente y versión Go soportada explícita.
- [x] 1.2 Crear `cmd/lufy-ai/main.go` (dentro de `tools/lufy-cli-go/`) como entrypoint delgado que delegue parseo/ejecución a paquetes internos.
- [x] 1.3 Crear paquetes iniciales (`internal/installer`, `internal/backup`, `internal/verify`, `internal/config`, `internal/platform`) dentro de `tools/lufy-cli-go/` con archivos mínimos compilables.
- [x] 1.4 Definir tipos compartidos mínimos para opciones de comando, resultados, errores accionables y exit codes sin acoplar lógica a `main.go`.
- [x] 1.5 Añadir tests mínimos de compilación/unidad para confirmar que `go test ./...` pasa con el scaffolding inicial.

## 2. Parser de comandos y flags

- [x] 2.1 Implementar dispatch para comandos `install`, `verify`, `backup`, `restore` y ayuda general.
- [x] 2.2 Implementar flags `--target`, `--dry-run`, `--yes`, `--no-engram` y `--backup` con defaults seguros según `design.md`.
- [x] 2.3 Mantener compatibilidad conceptual con `scripts/install.sh [target-project-dir]` mapeando argumento posicional a `--target` en el wrapper Bash posterior.
- [x] 2.4 Rechazar flags/comandos desconocidos con ayuda breve y exit code distinto de cero.
- [x] 2.5 Añadir tests unitarios del parser para comandos válidos, defaults, flags inválidos y combinaciones requeridas como `restore --backup`.

## 3. Plataforma, paths y dependencias externas

- [x] 3.1 Implementar resolución segura de `--target` a path absoluto/canonical sin permitir escrituras fuera del target planificado.
- [x] 3.2 Crear abstracción pequeña para filesystem o helpers testeables de lectura/escritura/copia/hash.
- [x] 3.3 Implementar resolución portable de Engram con `exec.LookPath("engram")` o interfaz equivalente en `internal/platform`.
- [x] 3.4 Asegurar que ninguna generación de configuración hardcodea `/opt/homebrew/bin/engram`.
- [x] 3.5 Añadir tests con resolver fake para Engram presente, ausente y omitido por `--no-engram`.

## 4. Plan de instalación y dry-run

- [x] 4.1 Definir tipos `Plan`, `Action` y `Conflict` en `internal/installer` con campos suficientes para explicar operación, origen, destino, razón y riesgo.
- [x] 4.2 Implementar construcción de plan para assets gestionados actuales: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/templates`, metadata local de `.opencode`, `AGENTS.md` desde template y `tui.json` cuando exista.
- [x] 4.3 Implementar clasificación de acciones: `mkdir`, `copy`, `skip`, `merge-json`, `backup`, `warn-conflict`, `verify`.
- [x] 4.4 Garantizar que `install --dry-run` imprime el plan y no crea, modifica, borra, clona ni respalda archivos reales.
- [x] 4.5 Añadir tests de plan en temp dirs para target vacío, target ya instalado y target con conflictos.

## 5. Assets gestionados e idempotencia

- [x] 5.1 Definir un manifest/listado inicial de assets gestionados por `lufy-ai` sin incluir archivos desconocidos del usuario.
- [x] 5.2 Implementar copia recursiva de assets gestionados preservando estructura relativa dentro del target.
- [x] 5.3 Detectar archivos idénticos por contenido/hash y reportarlos como `skip`.
- [x] 5.4 Preservar `AGENTS.md` existente y crear desde `AGENTS.md.template` solo cuando falte.
- [x] 5.5 Evitar borrado o modificación de archivos no gestionados por `lufy-ai`.
- [x] 5.6 Añadir prueba de idempotencia: instalar dos veces en temp dir y verificar que la segunda ejecución no sobrescribe ni genera conflictos falsos.

## 6. Backup manifest, rollback y restore

- [x] 6.1 Diseñar e implementar `manifest.json` versión 1 con `schemaVersion`, `createdAt`, `toolVersion`, `targetRoot` y acciones con paths relativos, hashes y estado.
- [x] 6.2 Crear backups bajo `.lufy-ai/backups/<timestamp>/` dentro del target antes de sobrescrituras o restauraciones riesgosas.
- [x] 6.3 Implementar escritura de manifest y copia de archivos previos con hashes verificables.
- [x] 6.4 Implementar `lufy-ai backup --target <dir>` para capturar estado relevante de archivos gestionados.
- [x] 6.5 Implementar `lufy-ai restore --target <dir> --backup <manifest-or-dir>` con validación de manifest, confirmación segura y soporte `--dry-run`.
- [x] 6.6 Manejar fallos parciales de `install` reportando manifest disponible e intentando rollback solo cuando sea seguro.
- [x] 6.7 Añadir tests de backup/restore en temp dirs, incluyendo restore dry-run y mismatch de target.

## 7. Configuración OpenCode y Engram

- [x] 7.1 Implementar creación de `opencode.json` cuando falte y sea requerido por la configuración instalada.
- [x] 7.2 Implementar merge conservador de `opencode.json` válido preservando claves desconocidas del usuario.
- [x] 7.3 Fallar sin sobrescribir cuando `opencode.json` existente sea inválido, con mensaje accionable.
- [x] 7.4 Integrar resolución portable de Engram en la generación/merge de configuración cuando `--no-engram` no esté activo.
- [x] 7.5 Si Engram está ausente, continuar instalación base y reportar nota accionable sin fallar.
- [x] 7.6 Añadir tests para creación, merge, JSON inválido, Engram presente, Engram ausente y `--no-engram`.

## 8. Verify command

- [x] 8.1 Implementar `lufy-ai verify --target <dir>` para validar estructura instalada y presencia de directories/files críticos.
- [x] 8.2 Validar que JSON relevante (`opencode.json`, package metadata cuando aplique) sea parseable sin asumir tooling Node global.
- [x] 8.3 Validar presencia de commands, skills, policies y plugin TUI esperados según assets gestionados.
- [x] 8.4 Reportar checks con estado claro (`ok`, `warn`, `fail`) y exit code distinto de cero ante fallos críticos.
- [x] 8.5 Respetar `--no-engram` en verificación para omitir checks obligatorios de Engram.
- [x] 8.6 Añadir tests de `verify` contra temp dirs instalado, incompleto y con JSON inválido.

## 9. Actualización de `scripts/install.sh` como wrapper

- [x] 9.1 Refactorizar `scripts/install.sh` para detectar binario Go local en `tools/lufy-cli-go/bin/lufy-ai` o CLI en PATH.
- [x] 9.2 Delegar en `lufy-ai install` cuando el binario Go esté disponible, preservando `scripts/install.sh [target-project-dir]`.
- [x] 9.3 Pasar flags compatibles (`--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup`) desde Bash hacia la CLI Go.
- [x] 9.4 Eliminar ruta legacy temporal y fallar con mensaje accionable de build cuando no exista binario Go local ni en PATH, sin descargar binarios remotos inseguros.
- [x] 9.5 Añadir validación manual o automatizada del wrapper usando un temp dir y confirmando que delega correctamente cuando existe binario Go.

## 10. Documentación

- [x] 10.1 Actualizar README para describir el estado real post-implementación: Bash wrapper + CLI Go disponible, sin presentar funcionalidades futuras como existentes.
- [x] 10.2 Preservar el banner `docs/assets/lufy-ai-banner.png` y enlaces existentes relevantes.
- [x] 10.3 Documentar comandos iniciales: `lufy-ai install`, `lufy-ai verify`, `lufy-ai backup`, `lufy-ai restore` con ejemplos seguros.
- [x] 10.4 Documentar flags y defaults: `--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup`.
- [x] 10.5 Documentar backup manifest, restore y estrategia de no sobrescritura/idempotencia.
- [x] 10.6 Documentar que Engram se resuelve desde `PATH` y que no se hardcodea `/opt/homebrew/bin/engram`.

## 11. Validación incremental y gates finales

- [x] 11.1 Ejecutar `go test ./...` después de introducir `go.mod` y paquetes Go.
- [x] 11.2 Ejecutar `go build ./cmd/lufy-ai` y conservar evidencia exacta del resultado.
- [x] 11.3 Ejecutar instalación dry-run en temp dir, por ejemplo `./lufy-ai install --target <temp> --dry-run --yes --no-engram`, y verificar que no escribe assets.
- [x] 11.4 Ejecutar instalación real en temp dir y luego `./lufy-ai verify --target <temp> --no-engram`.
- [x] 11.5 Ejecutar prueba de idempotencia instalando dos veces en temp dir y revisando salida/estado.
- [x] 11.6 Ejecutar prueba de backup/restore con conflicto controlado y manifest.
- [x] 11.7 Ejecutar `git diff --check` para validar whitespace de los cambios.
- [x] 11.8 Reportar cualquier validación no disponible sin inventar comandos ni resultados.
