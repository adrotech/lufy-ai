## 1. Analisis inicial sistemico

- [x] 1.1 Revisar una vez al inicio los paquetes afectados (`assets`, `platform`, `state`, `installer`, `syncer`, `backup`, `verify`, `version`) y confirmar interconexiones antes de editar.
- [x] 1.2 Confirmar el estado de paridad actual entre assets raiz y `tools/lufy-cli-go/internal/assets/embedded/`, documentando cualquier drift que deba resolverse dentro de la implementacion.

## 2. Catalogo y assets embebidos

- [x] 2.1 Implementar fingerprint estable del catalogo basado en assets de archivo ordenados y sus hashes fuente.
- [x] 2.2 Agregar test de paridad root vs embedded para target paths, kind, policy y `sourceSHA256`.
- [x] 2.3 Resolver drift detectado entre assets canonicos y embebidos sin cambiar `scripts/install.sh` ni reintroducir fallback legacy.

## 3. Path safety portable

- [x] 3.1 Endurecer `platform.EnsureRelativeSafe` para rechazar escapes con separadores Unix, Windows o mixtos.
- [x] 3.2 Agregar tests de path traversal para `../`, `..\`, separadores mixtos, paths absolutos y paths seguros esperados.

## 4. Metadata de state y backup

- [x] 4.1 Propagar metadata de `version.Current()` a `.lufy-ai/install-state.json` en install y sync.
- [x] 4.2 Reemplazar `SourceRootFingerprint` hardcodeado por el fingerprint del catalogo efectivo.
- [x] 4.3 Propagar metadata real de version a manifests de backup y actualizar tests relacionados.

## 5. Escrituras atomicas

- [x] 5.1 Crear o reutilizar helper de escritura atomica para payloads de archivos gestionados.
- [x] 5.2 Usar escritura atomica en `installer` para `copy` y `update-managed`.
- [x] 5.3 Usar escritura atomica en `syncer` para `update-managed`.
- [x] 5.4 Usar escritura atomica en `backup` para captura y restore de archivos respaldados.

## 6. Validacion final agrupada

- [x] 6.1 Revisar al final los archivos viejos modificados/afectados para coherencia con la propuesta y specs.
- [x] 6.2 Ejecutar `go test ./...` desde `tools/lufy-cli-go/`.
- [x] 6.3 Ejecutar `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/`.
- [x] 6.4 Ejecutar `git diff --check` desde la raiz.
- [x] 6.5 Reportar si coverage no se ejecuta por no existir gate/threshold definido para esta proposal.
