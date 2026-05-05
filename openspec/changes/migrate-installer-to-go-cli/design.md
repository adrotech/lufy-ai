## Context

`lufy-ai` es hoy un kit instalable de OpenCode/OpenSpec, no un producto Go multiagente. El instalador principal es `scripts/install.sh`; copia `.opencode/`, plantillas, políticas, comandos, skills, plugins, `AGENTS.md` y `tui.json` hacia un repositorio destino, además de ofrecer integración con Engram si detecta el binario.

El repo no tiene `go.mod` en la raíz ni CI bajo `.github/`. Tampoco debe asumirse un toolchain Node/TS global: la validación real actual suele ser estática/documental salvo que el cambio introduzca comandos reales. La iniciativa `RM-013` del roadmap define que el installer debe migrar a una CLI Go (`lufy-ai`) con Bash reducido a wrapper/bootstrapper de compatibilidad.

El cambio debe ser progresivo: primero se crea una CLI mínima y testeable que cubra `install`, `verify`, `backup` y `restore`; después, fuera de este alcance inicial, se podrán añadir comandos como `sync`, `update`, `doctor` o una TUI.

## Goals / Non-Goals

**Goals:**

- Introducir un módulo Go real en la raíz con un binario `lufy-ai` compilable desde `cmd/lufy-ai`.
- Mover la lógica crítica de instalación fuera de Bash hacia paquetes Go testeables.
- Mantener `scripts/install.sh` como punto de entrada compatible para usuarios existentes, limitado a wrapper estricto de la CLI Go.
- Implementar comandos iniciales: `install`, `verify`, `backup`, `restore`.
- Proveer `--dry-run` con plan explícito antes de escribir, y defaults seguros que eviten sobrescribir trabajo local.
- Crear backups con manifest antes de modificar archivos existentes cuando `--backup` esté activo o el comando lo requiera por seguridad.
- Resolver Engram de forma portable con `exec.LookPath("engram")` o abstracción equivalente, sin hardcodear `/opt/homebrew/bin/engram`.
- Generar o mergear `opencode.json` sin destruir configuración existente del usuario.
- Validar por fases con comandos Go reales una vez exista el scaffolding.

**Non-Goals:**

- No implementar todavía `sync`, `update`, `doctor`, TUI Go ni distribución de releases multi-OS.
- No convertir `lufy-ai` en producto multiagente Go; el contenido instalable sigue siendo el kit OpenCode/OpenSpec existente.
- No cambiar puertos, defaults de autenticación, esquemas de base de datos ni contratos HTTP/API.
- No modificar `docs/roadmap.md` salvo que en una implementación futura se requiera trazabilidad adicional.
- No eliminar el archivo `scripts/install.sh` en esta fase, pero sí retirar su lógica legacy de instalación.
- No inventar CI en `.github/` dentro de esta propuesta; puede ser una mejora posterior o tarea separada si se autoriza.

## Decisions

### 1. Usar Go para la CLI inicial

**Decisión:** crear un binario Go `lufy-ai` en una carpeta dedicada `tools/lufy-cli-go/` con su propio `go.mod`.

**Rationale:** Go cubre bien filesystem, JSON, detección de OS, ejecución de binarios externos, pruebas unitarias y distribución como single binary. Esto coincide con `RM-013` y reduce complejidad frente a mantener lógica creciente en Bash.

**Alternativas consideradas:**

- **Mantener todo en Bash:** menor costo inicial, pero peor testabilidad, merge JSON frágil, rollback complejo y portabilidad limitada.
- **Rust:** ofrece memory-safety fuerte, pero para esta etapa el problema principal es idempotencia/backup/UX, no performance ni control fino de memoria. Go es suficiente y más directo para otro LLM implementador.

### 2. Arquitectura por paquetes internos

**Decisión:** organizar el código sugerido así:

```text
tools/lufy-cli-go/
  go.mod
  cmd/lufy-ai/main.go
  internal/cli/              # dispatch de comandos y parseo de flags si se separa de main
  internal/installer/        # planificación y aplicación de instalación
  internal/backup/           # manifest, creación de snapshots, restore
  internal/verify/           # checks de estructura, JSON y Engram
  internal/config/           # lectura/merge/escritura de opencode.json y metadata
  internal/platform/         # paths, OS detection, exec.LookPath, abstracciones filesystem
  internal/assets/           # manifest/listado de assets gestionados, si no se embebe inicialmente
```

Los paquetes deben depender de interfaces pequeñas cuando facilite pruebas, por ejemplo `FileSystem`, `CommandResolver` o `Clock`, pero sin sobrediseñar. `cmd/lufy-ai/main.go` debe ser delgado: parsea argumentos, construye dependencias y llama a servicios.

**Alternativas consideradas:**

- Un único `main.go`: rápido, pero dificulta pruebas y migración incremental.
- Framework CLI externo: útil a largo plazo, pero para primera versión se puede usar `flag`/`FlagSet` de la librería estándar y evitar dependencias hasta que haya necesidad real.

### 3. Comandos y flags iniciales

**Decisión:** soportar estos comandos mínimos:

- `lufy-ai install --target . [--dry-run] [--yes] [--no-engram] [--backup]`
- `lufy-ai verify --target . [--no-engram]`
- `lufy-ai backup --target . [--yes]`
- `lufy-ai restore --target . --backup <manifest-or-dir> [--dry-run] [--yes]`

Flags y defaults:

- `--target <dir>`: default `.`; debe resolverse a path absoluto/canonical seguro antes de operar.
- `--dry-run`: default `false`; cuando `true`, no escribe, no borra, no clona, no crea backups reales y emite un plan detallado.
- `--yes`: default `false`; cuando `false`, operaciones destructivas o conflictivas requieren confirmación interactiva. En ambientes no interactivos debe fallar con mensaje accionable si necesita confirmación.
- `--no-engram`: default `false`; cuando `true`, omite detección/generación de integración Engram.
- `--backup`: para `install`, default seguro recomendado `true` cuando existan archivos gestionados previos o conflictos; puede aceptarse como flag explícito para forzar backup incluso sin conflictos. Para `restore`, `--backup` identifica el backup a restaurar.

El parser debe rechazar flags desconocidos con ayuda breve y exit code distinto de cero.

### 4. Plan de instalación y dry-run primero

**Decisión:** `install` debe construir un `Plan` antes de escribir. El plan lista acciones como `mkdir`, `copy`, `skip`, `merge-json`, `backup`, `warn-conflict` y `verify`.

Ejemplo conceptual:

```go
type Action struct {
    Kind   string
    Source string
    Target string
    Reason string
    Risk   string
}

type Plan struct {
    TargetRoot string
    Actions    []Action
    Conflicts  []Conflict
}
```

`--dry-run` imprime el plan y termina sin mutar el filesystem. La implementación debe permitir pruebas unitarias del plan sin tocar disco real cuando sea práctico.

### 5. Assets gestionados e idempotencia

**Decisión:** la CLI debe distinguir archivos gestionados por `lufy-ai` de personalizaciones del usuario. En la primera versión puede usar una lista explícita/manifest en código o archivo del repo; idealmente evolucionará a `internal/assets` con metadata.

Reglas:

- Crear directorios requeridos dentro de `--target` si faltan.
- Copiar assets gestionados cuando no existen.
- Si existe un archivo idéntico, marcar `skip`.
- Si existe un archivo distinto y no hay estrategia de merge segura, marcar conflicto y requerir confirmación/backup antes de sobrescribir.
- No sobrescribir `AGENTS.md` existente; si falta, crear desde `AGENTS.md.template`. Si existe, reportar que se preserva o sugerir merge manual.
- No borrar archivos desconocidos del target.

### 6. Merge seguro de `opencode.json`

**Decisión:** si la CLI genera o actualiza `opencode.json`, debe hacerlo mediante parseo JSON y merge conservador.

Reglas mínimas:

- Si no existe, crear JSON válido con configuración requerida por `lufy-ai`.
- Si existe JSON válido, preservar claves desconocidas del usuario.
- Añadir/actualizar únicamente secciones gestionadas y documentadas por la CLI.
- Si existe JSON inválido, no sobrescribir automáticamente: reportar error, sugerir backup/manual fix y permitir `backup` antes de intervención futura.
- Escribir de forma atómica: archivo temporal en el mismo directorio y rename cuando sea posible.

### 7. Backup, rollback y restore con manifest

**Decisión:** antes de cambios riesgosos, crear backup bajo un directorio dentro del target, por ejemplo `.lufy-ai/backups/<timestamp>/`, con `manifest.json`.

Manifest sugerido:

```json
{
  "schemaVersion": 1,
  "createdAt": "2026-05-05T00:00:00Z",
  "toolVersion": "dev",
  "targetRoot": "/abs/path/to/target",
  "actions": [
    {
      "path": "AGENTS.md",
      "operation": "backup-before-overwrite",
      "backupPath": "files/AGENTS.md",
      "sha256Before": "...",
      "sha256After": "",
      "status": "captured"
    }
  ]
}
```

Reglas:

- El manifest usa paths relativos al target para trazabilidad y portabilidad.
- El backup no debe incluir secretos deliberadamente fuera del scope; solo archivos que la CLI va a tocar o necesita preservar.
- Si una operación falla después de crear backup, `install` debe reportar el manifest y, cuando sea seguro, intentar rollback automático de acciones ya aplicadas.
- `restore` debe validar que el manifest pertenece al target o pedir `--yes` para continuar si hay mismatch justificado.
- `restore --dry-run` muestra qué se restauraría sin escribir.

### 8. Engram portable y opt-out

**Decisión:** la integración Engram debe resolverse con `exec.LookPath("engram")` o una interfaz equivalente. Nunca debe hardcodear `/opt/homebrew/bin/engram`.

Reglas:

- Si `--no-engram` está activo, omitir toda configuración Engram.
- Si Engram existe en `PATH`, usar el path resuelto o una invocación portable según el contrato de OpenCode.
- Si no existe, dejar integración deshabilitada o comentada según soporte de configuración, y reportar instrucción de instalación; no fallar instalación base por ausencia de Engram salvo que el usuario lo solicite explícitamente en el futuro.

### 9. Relación entre Bash wrapper y binario Go

**Decisión:** `scripts/install.sh` debe quedar como wrapper compatible estricto de la CLI Go, sin fallback legacy.

Comportamiento esperado:

1. Aceptar el uso histórico `scripts/install.sh [target-project-dir]` y mapearlo a `lufy-ai install --target <dir>`.
2. Pasar flags compatibles: `--target`, `--dry-run`, `--yes`, `--no-engram` y `--backup`.
3. Si el binario Go local existe en `tools/lufy-cli-go/bin/lufy-ai`, delegar en él.
4. Si no existe binario local, buscar `lufy-ai` en `PATH` y delegar en él.
5. Si no existe binario local ni en `PATH`, fallar con instrucciones claras para compilar: `cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`.
6. No descargar/ejecutar binarios remotos sin un mecanismo explícito de integridad y autorización; si se agrega distribución remota futura, será otra propuesta.
7. No conservar lógica Bash legacy de copia, detección de stack, Engram, `copy_files`, `opencode.json` o prompts de instalación.

### 10. Validación por fases

**Decisión:** validar incrementalmente según el estado real del repo.

Fases sugeridas:

- Fase A — scaffolding: `go test ./...` y `go build ./cmd/lufy-ai` deben pasar desde `tools/lufy-cli-go/` una vez creado su `go.mod`.
- Fase B — plan/dry-run: prueba de `lufy-ai install --target <temp> --dry-run --yes --no-engram`; confirmar que no escribe en temp dir salvo artefactos permitidos por el test harness.
- Fase C — install real en temp dir: ejecutar instalación en directorio temporal y luego `lufy-ai verify --target <temp> --no-engram`.
- Fase D — idempotencia: ejecutar install dos veces y verificar que la segunda ejecución reporta `skip`/sin conflictos no esperados.
- Fase E — backup/restore: crear conflicto controlado, ejecutar backup/install y restaurar con manifest.

## Risks / Trade-offs

- **Riesgo: duplicar lógica entre Bash y Go durante la transición** → Mitigar manteniendo Bash como wrapper delgado y reduciendo gradualmente la ruta legacy.
- **Riesgo: sobrescribir personalizaciones del usuario** → Mitigar con plan previo, dry-run, conflictos explícitos, backups por defecto en cambios riesgosos y merge conservador.
- **Riesgo: rollback parcial incompleto** → Mitigar con manifest detallado, acciones atómicas cuando sea posible y restore explícito.
- **Riesgo: introducir Go sin CI** → Mitigar documentando comandos locales reales y añadiendo CI en una iniciativa separada o fase posterior autorizada.
- **Riesgo: resolver Engram distinto por plataforma** → Mitigar con `exec.LookPath("engram")`, tests con resolver fake y opt-out `--no-engram`.
- **Riesgo: crecer demasiado el alcance de la primera CLI** → Mitigar excluyendo `sync`, `update`, `doctor`, TUI y distribución remota.

## Migration Plan

1. Crear scaffolding Go mínimo sin cambiar comportamiento público del instalador Bash.
2. Implementar parser de comandos/flags y ayuda de uso.
3. Implementar plan/dry-run para `install` sin escrituras reales.
4. Añadir backup manifest y abstracciones de filesystem necesarias.
5. Implementar copia idempotente de assets gestionados dentro de `--target`.
6. Implementar generación/merge conservador de `opencode.json` y resolución portable de Engram.
7. Implementar `verify` y cubrirlo con tests unitarios/funcionales sobre temp dirs.
8. Implementar `backup`/`restore` con manifest.
9. Actualizar `scripts/install.sh` para delegar estrictamente en la CLI preservando compatibilidad de argumentos, sin fallback legacy.
10. Actualizar README/docs para describir estado real: Bash wrapper + CLI Go cuando esté implementada.

Rollback del cambio de código: ante fallos de la CLI, se debe corregir o reconstruir el binario Go; no existe fallback legacy en Bash. Rollback de instalaciones en targets de usuario debe hacerse con `lufy-ai restore --target <dir> --backup <manifest>`.

## Open Questions

- ¿El primer binario se compilará localmente desde checkout o se distribuirá como release en una fase posterior?
- ¿Dónde debe vivir el manifest definitivo de assets gestionados: archivo versionado legible (`assets-manifest.json`) o lista Go embebida?
- ¿Qué formato final necesita `opencode.json` para Engram en versiones actuales de OpenCode? La implementación debe validar contra ejemplos reales del repo antes de escribir.
- ¿Se debe mantener indefinidamente la ruta Bash legacy o removerla después de estabilizar releases del binario?
