## Why

El instalador actual de `lufy-ai` vive principalmente en `scripts/install.sh`, lo que concentra detección de plataforma, copia de archivos, backup, merge de configuración y verificación en Bash. Esto dificulta mantener idempotencia, pruebas automatizadas, rollback seguro y soporte portable a medida que la iniciativa de roadmap `RM-013` busca endurecer y evolucionar la instalación del kit.

Este cambio propone migrar progresivamente el instalador a una CLI escrita en Go, manteniendo `scripts/install.sh` como wrapper/bootstrapper de compatibilidad para no romper el flujo actual de instalación mientras se introduce una base tipada, testeable y multiplataforma.

## What Changes

- Introducir una CLI Go `lufy-ai` como nuevo motor de instalación, verificación, backup y restore del kit.
- Mantener `scripts/install.sh` como wrapper compatible que detecta/obtiene/ejecuta el binario Go cuando esté disponible, y conserva un camino de bootstrap seguro durante la transición.
- Añadir scaffolding Go real en carpeta dedicada de tooling (`tools/lufy-cli-go/go.mod`, `tools/lufy-cli-go/cmd/lufy-ai/main.go` y paquetes `tools/lufy-cli-go/internal/...`) cuando se implemente este change; actualmente esta propuesta **no implementa** la CLI.
- Definir comandos iniciales:
  - `install`: instala o actualiza assets gestionados en un target local.
  - `verify`: valida estado de instalación, binario `engram` y archivos esperados.
  - `backup`: crea snapshots con manifest de archivos gestionados y conflictos.
  - `restore`: restaura desde un backup manifest de forma controlada.
- Establecer flags comunes y defaults seguros: `--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup`.
- Sustituir supuestos frágiles, como rutas hardcodeadas a Engram, por resolución portable con `exec.LookPath("engram")` o equivalente.
- Documentar y validar la instalación por fases, incluyendo pruebas Go (`go test ./...`, `go build ./cmd/lufy-ai`) solo después de introducir el toolchain Go.
- No hay cambios **BREAKING** previstos: el contrato de `scripts/install.sh` debe mantenerse durante la migración.

## Capabilities

### New Capabilities
- `go-cli-installer`: Contrato funcional para una CLI Go de instalación, verificación, backup y restore de `lufy-ai`, con comportamiento idempotente, backups con manifest, dry-run y wrapper Bash compatible.

### Modified Capabilities
- Ninguna. No existen specs previas bajo `openspec/specs/` que deban modificarse para este cambio.

## Impact

- Código futuro afectado: `tools/lufy-cli-go/go.mod`, `tools/lufy-cli-go/cmd/lufy-ai/main.go`, `tools/lufy-cli-go/internal/installer`, `tools/lufy-cli-go/internal/backup`, `tools/lufy-cli-go/internal/verify`, `tools/lufy-cli-go/internal/config`, `tools/lufy-cli-go/internal/platform`.
- Script afectado: `scripts/install.sh`, que pasará a actuar como wrapper/bootstrapper de compatibilidad sin dejar de soportar el flujo actual.
- Documentación afectada: `README.md` y documentación de instalación relacionada; debe preservar el banner existente `docs/assets/lufy-ai-banner.png`.
- Roadmap: el cambio implementa una pieza de `RM-013` descrita en `docs/roadmap.md`, sin requerir modificar el roadmap para proponerlo.
- Dependencias futuras: se introducirá módulo Go en la raíz con comandos reales de compilación/prueba; no debe asumirse toolchain Node/TS global.
- Seguridad y datos de usuario: la CLI no debe tocar archivos fuera de `--target` salvo autorización explícita; debe evitar sobrescribir trabajo local no gestionado y debe producir backups/restauración trazables.
