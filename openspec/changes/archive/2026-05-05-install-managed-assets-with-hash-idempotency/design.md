## Context

`migrate-installer-to-go-cli` ya introdujo la base Go en `tools/lufy-cli-go` y dejó `scripts/install.sh` como wrapper estricto que delega en `lufy-ai install`. Este cambio no reemplaza esa decisión: la profundiza con el primer instalador completo de assets gestionados, idempotente por contenido/hash y trazable mediante estado persistente en el proyecto destino.

El repositorio fuente es el kit `lufy-ai`; el proyecto destino es cualquier directorio pasado por `--target`. La CLI debe copiar assets reales desde el checkout fuente hacia el destino sin asumir toolchains Node/TS globales ni rutas absolutas locales. Engram sigue siendo opcional/resuelto por `PATH` cuando aplique; este slice no debe hardcodearlo ni convertirlo en prerequisito.

## Goals / Non-Goals

**Goals:**

- Instalar el conjunto completo de assets gestionados: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `AGENTS.md`, `tui.json`, `openspec/` y metadatos `.lufy-ai/` necesarios.
- Definir y usar un catálogo de assets con paths relativos, tipo de entrada, hashes de source y política de manejo.
- Resolver de forma segura el source root del repo y el target project root.
- Construir un plan/dry-run fiel con acciones `create-dir`, `copy`, `skip`, `update-managed`, `conflict` y `backup`.
- Aplicar instalaciones idempotentes por hash, registrando `.lufy-ai/install-state.json`.
- Ampliar backup/restore a todos los assets tocados.
- Ampliar `verify` para validar estructura, manifest y hashes del target.
- Mantener el wrapper Bash estricto como delegador, sin fallback legacy.

**Non-Goals:**

- No construir distribución binaria, release artifacts, auto-update ni descarga remota.
- No implementar TUI o UX interactiva avanzada.
- No implementar multiagente externo ni orquestación fuera del kit OpenCode/OpenSpec.
- No implementar sync cloud ni sincronización remota de estado.
- No cambiar defaults de autenticación, puertos, esquemas de base de datos ni contratos públicos no relacionados.

## Architecture in `tools/lufy-cli-go`

La estructura sugerida mantiene `cmd/lufy-ai/main.go` delgado y concentra reglas en paquetes internos:

```text
tools/lufy-cli-go/
  cmd/lufy-ai/main.go
  internal/assets/       # catálogo, escaneo source, hashes, políticas por asset
  internal/installer/    # plan builder, apply idempotente, conflictos
  internal/state/        # install-state.json y schema/versionado
  internal/backup/       # snapshots multiasset, restore y rollback manual
  internal/verify/       # checks estructurales y de hashes
  internal/platform/     # path safety, symlinks, clock, filesystem, repo root resolver
```

`main.go` debe parsear flags, construir dependencias (`FileSystem`, `Clock`, `Hasher`, resolvers) y delegar a servicios. Las reglas de negocio quedan en `installer`, `assets`, `state`, `backup` y `verify`, cubiertas por tests con temp dirs.

## Asset catalog and hash model

El catálogo debe representar cada asset gestionado con metadata suficiente para planificar y verificar:

```go
type Asset struct {
    ID          string // estable, por ejemplo ".opencode/agents/orchestrator.md"
    SourceRel   string // relativo al source root
    TargetRel   string // relativo al target root
    Kind        string // file o dir
    Policy      string // managed, metadata, preserve-on-conflict
    SourceSHA256 string // para archivos; dirs se expanden a archivos
}
```

Los directorios del alcance se expanden a archivos individuales durante el scan para que cada copia/verificación sea hashable. El catálogo puede generarse en runtime desde el checkout fuente, pero el set raíz permitido debe estar explícito para evitar copiar archivos no deseados. Los hashes deben calcularse sobre bytes exactos del archivo fuente y destino, usando SHA-256.

## Source root and target root resolution

**Source root:** resolver desde la ubicación del ejecutable/desarrollo de forma portable. En modo checkout, la CLI puede detectar el repo root ascendiendo hasta encontrar marcadores esperados como `AGENTS.md`, `.opencode/` y `openspec/config.yaml`. Si se ejecuta desde un binario distribuido en el futuro, esa estrategia podrá cambiar en otra propuesta.

**Target root:** `--target` default `.` debe resolverse a ruta absoluta/canonical segura. Todas las acciones usan paths relativos normalizados contra ese root. La CLI MUST rechazar cualquier acción cuyo destino normalizado escape de `--target`.

Reglas de seguridad:

- No seguir symlinks que apunten fuera del source root o target root.
- No escribir mediante symlink de destino si puede escapar de `--target`.
- No operar con paths absolutos provenientes del catálogo.
- No tocar archivos fuera de `--target`, salvo lecturas del source root.

## Plan and dry-run

`install` debe construir un plan antes de escribir. `--dry-run` imprime ese plan y no crea directorios, archivos, backups ni manifests.

Acciones soportadas:

- `create-dir`: directorio requerido ausente dentro del target.
- `copy`: archivo ausente que puede copiarse desde source.
- `skip`: destino existente idéntico por hash.
- `update-managed`: destino gestionado previamente cuyo source upstream cambió; requiere backup previo.
- `conflict`: destino existente no gestionado, o destino gestionado cuyo contenido fue modificado localmente y no coincide con el hash esperado; no se sobrescribe sin decisión explícita.
- `backup`: snapshot planificado antes de `update-managed`, restore o cualquier mutación riesgosa.

El plan debe incluir `SourceRel`, `TargetRel`, `Reason`, `CurrentHash`, `SourceHash`, `PreviousInstalledHash` cuando aplique, y severidad/exit code esperado si hay conflictos.

## Idempotency rules

La decisión por archivo se basa en el catálogo actual, el contenido destino y `.lufy-ai/install-state.json`:

1. **Archivo ausente:** `copy`; después de aplicar se registra source hash y target hash.
2. **Archivo igual por hash al source actual:** `skip`; se puede refrescar el estado si falta o está incompleto.
3. **Archivo gestionado previo y cambió upstream:** si `install-state.json` registra que el target no fue modificado localmente y el source hash anterior difiere del source actual, planificar `backup` + `update-managed`.
4. **Archivo existe pero no es gestionado:** `conflict`; no sobrescribir automáticamente.
5. **Archivo gestionado previo con modificación local:** si el target hash actual no coincide con el último hash instalado/esperado, `conflict` aunque el source haya cambiado.
6. **Estado corrupto o incompatible:** tratar como no confiable, reportar `conflict` o `verify` failure según operación.

Estas reglas deben hacer que dos ejecuciones consecutivas sin cambios produzcan solo `skip`/sin mutaciones.

## Install state manifest

La CLI debe escribir `.lufy-ai/install-state.json` dentro del target con schema versionado. Forma sugerida:

```json
{
  "schemaVersion": 1,
  "toolVersion": "dev",
  "sourceChangeID": "install-managed-assets-with-hash-idempotency",
  "sourceRootFingerprint": "sha256-or-dev-fingerprint",
  "installedAt": "2026-05-05T00:00:00Z",
  "updatedAt": "2026-05-05T00:00:00Z",
  "targetRoot": "/absolute/target/path",
  "assets": [
    {
      "id": ".opencode/agents/orchestrator.md",
      "sourceRel": ".opencode/agents/orchestrator.md",
      "targetRel": ".opencode/agents/orchestrator.md",
      "sourceSHA256": "...",
      "targetSHA256": "...",
      "installedAt": "2026-05-05T00:00:00Z",
      "lastAction": "copy"
    }
  ]
}
```

`targetRoot` es informativo; las decisiones deben usar paths relativos para portabilidad. Si el target se mueve, `verify` puede advertir mismatch sin destruir estado.

## Backup and restore

Antes de cualquier `update-managed`, restore o mutación riesgosa, la CLI debe crear backup bajo `.lufy-ai/backups/<timestamp>/` con `manifest.json` y copias de los archivos previos tocados.

El backup manifest debe registrar paths relativos, hashes before/after cuando estén disponibles, acción que motivó el backup, estado de captura y errores. `restore` debe aceptar `--backup <manifest-or-dir>`, validar schema, listar acciones en `--dry-run`, respaldar el estado actual antes de restaurar si va a sobrescribir, y restaurar solo paths registrados dentro del target.

## Structural verify

`verify --target <dir>` debe validar:

- Presencia de directorios raíz gestionados y archivos críticos.
- Parseo válido de `.lufy-ai/install-state.json` y schema soportado.
- Correspondencia entre catálogo actual, manifest y archivos destino.
- Hashes de archivos gestionados: `ok` si coinciden, `warn` si falta estado recuperable, `fail` si falta un asset crítico o hay drift no registrado.
- Que los backups referenciados por estado/última ejecución existan cuando sean requeridos para recovery.
- Que no se requiera Engram para pasar instalación base y que no haya rutas hardcodeadas a `/opt/homebrew/bin/engram` generadas por este slice.

## Relationship with `scripts/install.sh`

`scripts/install.sh` permanece como wrapper estricto: detecta el binario local o en `PATH`, mapea argumentos compatibles y delega en `lufy-ai install`. Este cambio no debe añadir lógica Bash de copia, backup, verify ni detección de assets. Cualquier comportamiento nuevo vive en la CLI Go.

## Security considerations

- No tocar fuera de `--target`.
- No seguir symlinks peligrosos ni escribir a través de escapes de path.
- Hacer backup antes de mutar archivos existentes gestionados.
- No sobrescribir conflictos sin decisión explícita del usuario (`--yes` solo puede aprobar estrategias permitidas; no debe convertir conflictos no gestionados en overwrite silencioso).
- No hardcodear Engram ni fallar instalación base por ausencia de Engram.
- Escribir manifests de forma atómica cuando sea posible.

## Migration plan

1. Añadir catálogo/manifest schema sin aplicar cambios destructivos.
2. Implementar resolución segura de source/target.
3. Implementar plan/dry-run y validar clasificación de acciones.
4. Implementar apply idempotente con estado `.lufy-ai/install-state.json`.
5. Añadir manejo de conflictos y backups multiasset.
6. Ampliar restore y verify.
7. Cubrir con tests unitarios/integración en temp dirs.
8. Actualizar documentación de instalación real y evidencia de validación.

## Open Questions

- ¿El catálogo inicial será un archivo JSON/YAML versionado dentro de `tools/lufy-cli-go` o una lista Go generada en build/test? La implementación debe elegir la opción más simple y testeable para este slice.
- ¿Qué política exacta debe aplicar `--yes` sobre conflictos no gestionados? Recomendación: no sobrescribirlos en este slice; requerir futura opción explícita si se desea adopción/force.
- ¿Se debe incluir `.opencode/package.json` como metadata necesaria si `agent-observatory` depende de él? La implementación debe confirmar assets reales antes del catálogo final.
