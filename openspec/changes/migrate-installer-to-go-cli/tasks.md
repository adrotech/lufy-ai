## 1. Scaffolding Go y estructura base

- [ ] 1.1 Crear `go.mod` en la raíz con un module path coherente para el repo y versión Go soportada explícita.
- [ ] 1.2 Crear `cmd/lufy-ai/main.go` como entrypoint delgado que delegue parseo/ejecución a paquetes internos.
- [ ] 1.3 Crear paquetes iniciales `internal/installer`, `internal/backup`, `internal/verify`, `internal/config` e `internal/platform` con archivos mínimos compilables.
- [ ] 1.4 Definir tipos compartidos mínimos para opciones de comando, resultados, errores accionables y exit codes sin acoplar lógica a `main.go`.
- [ ] 1.5 Añadir tests mínimos de compilación/unidad para confirmar que `go test ./...` pasa con el scaffolding inicial.

## 2. Parser de comandos y flags

- [ ] 2.1 Implementar dispatch para comandos `install`, `verify`, `backup`, `restore` y ayuda general.
- [ ] 2.2 Implementar flags `--target`, `--dry-run`, `--yes`, `--no-engram` y `--backup` con defaults seguros según `design.md`.
- [ ] 2.3 Mantener compatibilidad conceptual con `scripts/install.sh [target-project-dir]` mapeando argumento posicional a `--target` en el wrapper Bash posterior.
- [ ] 2.4 Rechazar flags/comandos desconocidos con ayuda breve y exit code distinto de cero.
- [ ] 2.5 Añadir tests unitarios del parser para comandos válidos, defaults, flags inválidos y combinaciones requeridas como `restore --backup`.

## 3. Plataforma, paths y dependencias externas

- [ ] 3.1 Implementar resolución segura de `--target` a path absoluto/canonical sin permitir escrituras fuera del target planificado.
- [ ] 3.2 Crear abstracción pequeña para filesystem o helpers testeables de lectura/escritura/copia/hash.
- [ ] 3.3 Implementar resolución portable de Engram con `exec.LookPath("engram")` o interfaz equivalente en `internal/platform`.
- [ ] 3.4 Asegurar que ninguna generación de configuración hardcodea `/opt/homebrew/bin/engram`.
- [ ] 3.5 Añadir tests con resolver fake para Engram presente, ausente y omitido por `--no-engram`.

## 4. Plan de instalación y dry-run

- [ ] 4.1 Definir tipos `Plan`, `Action` y `Conflict` en `internal/installer` con campos suficientes para explicar operación, origen, destino, razón y riesgo.
- [ ] 4.2 Implementar construcción de plan para assets gestionados actuales: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/templates`, metadata local de `.opencode`, `AGENTS.md` desde template y `tui.json` cuando exista.
- [ ] 4.3 Implementar clasificación de acciones: `mkdir`, `copy`, `skip`, `merge-json`, `backup`, `warn-conflict`, `verify`.
- [ ] 4.4 Garantizar que `install --dry-run` imprime el plan y no crea, modifica, borra, clona ni respalda archivos reales.
- [ ] 4.5 Añadir tests de plan en temp dirs para target vacío, target ya instalado y target con conflictos.

## 5. Assets gestionados e idempotencia

- [ ] 5.1 Definir un manifest/listado inicial de assets gestionados por `lufy-ai` sin incluir archivos desconocidos del usuario.
- [ ] 5.2 Implementar copia recursiva de assets gestionados preservando estructura relativa dentro del target.
- [ ] 5.3 Detectar archivos idénticos por contenido/hash y reportarlos como `skip`.
- [ ] 5.4 Preservar `AGENTS.md` existente y crear desde `AGENTS.md.template` solo cuando falte.
- [ ] 5.5 Evitar borrado o modificación de archivos no gestionados por `lufy-ai`.
- [ ] 5.6 Añadir prueba de idempotencia: instalar dos veces en temp dir y verificar que la segunda ejecución no sobrescribe ni genera conflictos falsos.

## 6. Backup manifest, rollback y restore

- [ ] 6.1 Diseñar e implementar `manifest.json` versión 1 con `schemaVersion`, `createdAt`, `toolVersion`, `targetRoot` y acciones con paths relativos, hashes y estado.
- [ ] 6.2 Crear backups bajo `.lufy-ai/backups/<timestamp>/` dentro del target antes de sobrescrituras o restauraciones riesgosas.
- [ ] 6.3 Implementar escritura de manifest y copia de archivos previos con hashes verificables.
- [ ] 6.4 Implementar `lufy-ai backup --target <dir>` para capturar estado relevante de archivos gestionados.
- [ ] 6.5 Implementar `lufy-ai restore --target <dir> --backup <manifest-or-dir>` con validación de manifest, confirmación segura y soporte `--dry-run`.
- [ ] 6.6 Manejar fallos parciales de `install` reportando manifest disponible e intentando rollback solo cuando sea seguro.
- [ ] 6.7 Añadir tests de backup/restore en temp dirs, incluyendo restore dry-run y mismatch de target.

## 7. Configuración OpenCode y Engram

- [ ] 7.1 Implementar creación de `opencode.json` cuando falte y sea requerido por la configuración instalada.
- [ ] 7.2 Implementar merge conservador de `opencode.json` válido preservando claves desconocidas del usuario.
- [ ] 7.3 Fallar sin sobrescribir cuando `opencode.json` existente sea inválido, con mensaje accionable.
- [ ] 7.4 Integrar resolución portable de Engram en la generación/merge de configuración cuando `--no-engram` no esté activo.
- [ ] 7.5 Si Engram está ausente, continuar instalación base y reportar nota accionable sin fallar.
- [ ] 7.6 Añadir tests para creación, merge, JSON inválido, Engram presente, Engram ausente y `--no-engram`.

## 8. Verify command

- [ ] 8.1 Implementar `lufy-ai verify --target <dir>` para validar estructura instalada y presencia de directories/files críticos.
- [ ] 8.2 Validar que JSON relevante (`opencode.json`, package metadata cuando aplique) sea parseable sin asumir tooling Node global.
- [ ] 8.3 Validar presencia de commands, skills, policies y plugin TUI esperados según assets gestionados.
- [ ] 8.4 Reportar checks con estado claro (`ok`, `warn`, `fail`) y exit code distinto de cero ante fallos críticos.
- [ ] 8.5 Respetar `--no-engram` en verificación para omitir checks obligatorios de Engram.
- [ ] 8.6 Añadir tests de `verify` contra temp dirs instalado, incompleto y con JSON inválido.

## 9. Actualización de `scripts/install.sh` como wrapper

- [ ] 9.1 Refactorizar `scripts/install.sh` para detectar binario Go local o posibilidad segura de ejecutar/compilar desde checkout con `go.mod`.
- [ ] 9.2 Delegar en `lufy-ai install` cuando el binario Go esté disponible, preservando `scripts/install.sh [target-project-dir]`.
- [ ] 9.3 Pasar flags compatibles (`--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup`) desde Bash hacia la CLI Go.
- [ ] 9.4 Mantener una ruta legacy temporal o mensaje de fallback seguro durante la transición sin descargar binarios remotos inseguros.
- [ ] 9.5 Añadir validación manual o automatizada del wrapper usando un temp dir y confirmando que delega correctamente cuando existe binario Go.

## 10. Documentación

- [ ] 10.1 Actualizar README para describir el estado real post-implementación: Bash wrapper + CLI Go disponible, sin presentar funcionalidades futuras como existentes.
- [ ] 10.2 Preservar el banner `docs/assets/lufy-ai-banner.png` y enlaces existentes relevantes.
- [ ] 10.3 Documentar comandos iniciales: `lufy-ai install`, `lufy-ai verify`, `lufy-ai backup`, `lufy-ai restore` con ejemplos seguros.
- [ ] 10.4 Documentar flags y defaults: `--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup`.
- [ ] 10.5 Documentar backup manifest, restore y estrategia de no sobrescritura/idempotencia.
- [ ] 10.6 Documentar que Engram se resuelve desde `PATH` y que no se hardcodea `/opt/homebrew/bin/engram`.

## 11. Validación incremental y gates finales

- [ ] 11.1 Ejecutar `go test ./...` después de introducir `go.mod` y paquetes Go.
- [ ] 11.2 Ejecutar `go build ./cmd/lufy-ai` y conservar evidencia exacta del resultado.
- [ ] 11.3 Ejecutar instalación dry-run en temp dir, por ejemplo `./lufy-ai install --target <temp> --dry-run --yes --no-engram`, y verificar que no escribe assets.
- [ ] 11.4 Ejecutar instalación real en temp dir y luego `./lufy-ai verify --target <temp> --no-engram`.
- [ ] 11.5 Ejecutar prueba de idempotencia instalando dos veces en temp dir y revisando salida/estado.
- [ ] 11.6 Ejecutar prueba de backup/restore con conflicto controlado y manifest.
- [ ] 11.7 Ejecutar `git diff --check` para validar whitespace de los cambios.
- [ ] 11.8 Reportar cualquier validación no disponible sin inventar comandos ni resultados.
