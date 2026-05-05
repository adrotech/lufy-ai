## 1. Asset catalog y manifest schema

- [x] 1.1 Inventariar assets reales del repo fuente dentro del alcance: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `AGENTS.md`, `tui.json`, `openspec/` y metadata requerida.
- [x] 1.2 Definir el set raíz permitido para evitar copiar archivos fuera de alcance o temporales.
- [x] 1.3 Diseñar tipos Go para `Asset`, `AssetCatalog`, hashes SHA-256 y políticas por asset.
- [x] 1.4 Diseñar schema v1 de `.lufy-ai/install-state.json` con `schemaVersion`, `toolVersion`, `sourceChangeID`, timestamps, target informativo y lista de assets.
- [x] 1.5 Implementar lectura/escritura atómica del install state con validación de schema soportado.

## 2. Source/target resolver

- [x] 2.1 Implementar resolver de source root del checkout usando marcadores esperados (`AGENTS.md`, `.opencode/`, `openspec/config.yaml`) sin hardcodear rutas locales.
- [x] 2.2 Implementar resolución de `--target` a path absoluto/canonical seguro con default `.`.
- [x] 2.3 Normalizar todos los paths del catálogo como relativos y rechazar paths absolutos o con escape `..`.
- [x] 2.4 Detectar y rechazar symlinks peligrosos que escapen del source root o target root.
- [x] 2.5 Añadir tests de resolución para source válido, target relativo, target absoluto, path escape y symlink peligroso.

## 3. Plan builder/dry-run

- [x] 3.1 Definir tipos `Plan`, `Action`, `Conflict` y campos de hash/razón/severidad.
- [x] 3.2 Implementar expansión del catálogo de directorios a archivos hashables.
- [x] 3.3 Clasificar archivos ausentes como `copy` y directorios ausentes como `create-dir`.
- [x] 3.4 Clasificar archivos idénticos por hash como `skip`.
- [x] 3.5 Clasificar cambios upstream gestionados como `backup` + `update-managed` cuando el target no tenga drift local.
- [x] 3.6 Clasificar archivos no gestionados o con drift local como `conflict`.
- [x] 3.7 Garantizar que `install --dry-run` imprime el plan fiel y no escribe directorios, archivos, backups ni manifests.
- [x] 3.8 Añadir tests de plan para target vacío, target ya instalado, upstream cambiado, drift local y archivo no gestionado.

## 4. Install apply idempotente

- [x] 4.1 Implementar aplicación ordenada de acciones: crear directorios, copiar archivos ausentes, actualizar gestionados, escribir estado.
- [x] 4.2 Calcular y registrar hashes source/target después de cada copy/update.
- [x] 4.3 Hacer que una segunda ejecución sin cambios produzca solo `skip`/sin mutaciones.
- [x] 4.4 Escribir `.lufy-ai/install-state.json` de forma atómica después de una aplicación exitosa.
- [x] 4.5 Manejar errores parciales reportando acciones aplicadas, backup disponible y recovery recomendado.
- [x] 4.6 Añadir tests de integración con temp dirs para instalación completa y reinstalación idempotente.

## 5. Conflict handling

- [x] 5.1 Definir mensajes accionables para conflictos no gestionados y drift local.
- [x] 5.2 Asegurar que conflictos bloquean sobrescritura automática aunque `--yes` esté presente, salvo estrategia explícitamente soportada.
- [x] 5.3 Reportar exit code distinto de cero cuando el plan contenga conflictos bloqueantes en apply real.
- [x] 5.4 Permitir que `--dry-run` muestre conflictos sin mutar ni fallar de forma destructiva.
- [x] 5.5 Añadir tests para archivo existente no gestionado, archivo gestionado modificado localmente y estado corrupto.

## 6. Backup/restore multiasset

- [x] 6.1 Extender backup para capturar todos los archivos que serán tocados por `update-managed` o restore.
- [x] 6.2 Guardar backups bajo `.lufy-ai/backups/<timestamp>/` con paths relativos y manifest `manifest.json`.
- [x] 6.3 Registrar hashes before/after, acción causante, status de captura y errores por asset.
- [x] 6.4 Implementar `restore --backup <manifest-or-dir>` para restaurar solo paths registrados dentro del target.
- [x] 6.5 Implementar `restore --dry-run` con plan de restauración sin escrituras.
- [x] 6.6 Crear backup del estado actual antes de restaurar si restore sobrescribirá archivos existentes.
- [x] 6.7 Añadir tests multiasset de backup, restore real, restore dry-run, target mismatch y path escape en manifest.

## 7. Verify estructural

- [x] 7.1 Ampliar `verify --target` para validar directorios y archivos críticos del catálogo.
- [x] 7.2 Validar `.lufy-ai/install-state.json`: JSON válido, schema soportado, assets esperados y timestamps presentes.
- [x] 7.3 Recalcular hashes destino y comparar contra manifest/catálogo.
- [x] 7.4 Reportar estados `ok`, `warn` y `fail` con exit code distinto de cero ante fallos críticos.
- [x] 7.5 Verificar que no haya rutas Engram hardcodeadas generadas por este slice y respetar `--no-engram`.
- [x] 7.6 Añadir tests de verify para instalación completa, asset faltante, drift local, manifest corrupto y target movido.

## 8. Tests unit/integration con temp dirs

- [x] 8.1 Añadir tests unitarios para hashing, catálogo, path safety y estado.
- [x] 8.2 Añadir tests del plan builder sin tocar disco real cuando sea práctico.
- [x] 8.3 Añadir tests de integración con temp dirs para install completo y verify.
- [x] 8.4 Añadir test de idempotencia instalando dos veces y comparando plan/estado.
- [x] 8.5 Añadir test de upstream changed simulando cambio de source hash y `update-managed` con backup.
- [x] 8.6 Añadir test de conflicto no gestionado y drift local.
- [x] 8.7 Añadir test de backup/restore multiasset.

## 9. Docs

- [x] 9.1 Actualizar documentación de instalación para describir assets gestionados reales y estado `.lufy-ai/install-state.json`.
- [x] 9.2 Documentar acciones del plan (`create-dir`, `copy`, `skip`, `update-managed`, `conflict`, `backup`).
- [x] 9.3 Documentar reglas de idempotencia y cuándo se requiere intervención manual.
- [x] 9.4 Documentar backup/restore multiasset y ejemplos seguros con `--dry-run`.
- [x] 9.5 Documentar que `scripts/install.sh` sigue siendo wrapper estricto y que la lógica vive en la CLI Go.

## 10. Validación final

- [x] 10.1 Ejecutar `go test ./...` desde `tools/lufy-cli-go/`.
- [x] 10.2 Ejecutar `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/`.
- [x] 10.3 Ejecutar install `--dry-run` contra temp dir y confirmar que no escribe assets.
- [x] 10.4 Ejecutar install real contra temp dir y luego `verify --target <temp> --no-engram`.
- [x] 10.5 Ejecutar reinstalación contra el mismo temp dir y confirmar idempotencia por `skip`/sin mutaciones.
- [x] 10.6 Ejecutar escenario de conflicto controlado y confirmar que no sobrescribe.
- [x] 10.7 Ejecutar backup/restore multiasset con temp dirs.
- [x] 10.8 Ejecutar `git diff --check` y reportar resultados reales.
- [x] 10.9 Actualizar evidencia en handoff final: validación final PASS (`git diff --check`, `cd tools/lufy-cli-go && go test ./...`, `cd tools/lufy-cli-go && go build ./cmd/lufy-ai`); delivery PR B autorizado en rama `feat/managed-assets-hash-idempotency` contra `main`.
