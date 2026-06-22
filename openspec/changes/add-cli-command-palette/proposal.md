## Why

La CLI `lufy-ai` crecio en cantidad de comandos, subcomandos y flags. Aunque los comandos scriptables siguen siendo necesarios para CI y automatizacion, el uso diario se volvio dificil para usuarios que no recuerdan parametros como `--target`, `--dry-run`, `--yes`, `--json`, `--scope`, `--tool`, `--to`, `--backup` o subcomandos de `memory` y `context`.

Ya usamos Bubble Tea/Charm en la CLI, por lo que tiene sentido agregar un command palette interactivo cuando el usuario ejecuta `lufy-ai` sin argumentos: mostrar comandos con descripcion, permitir elegir uno, editar parametros y ejecutar el comando resultante sin memorizar flags.

## What Changes

- Agregar un command palette TUI para `lufy-ai` sin argumentos en terminal interactiva.
- Mantener el comportamiento scriptable/no-TTY: sin argumentos fuera de TTY sigue mostrando help y exit usage.
- Crear un registry declarativo de comandos, subcomandos, parametros, valores por defecto, choices y descripciones.
- Permitir seleccionar un comando con `enter`, navegar parametros, alternar booleanos, ciclar choices y escribir valores de texto.
- Mostrar la linea equivalente `lufy-ai ...` antes de ejecutar.
- Reutilizar `cli.Run` con args generados para no duplicar logica de negocio.
- Incluir los comandos principales actuales: `setup`, `init`, `scan`, `install`, `sync`, `verify`, `doctor`, `status`, `info`, `memory`, `context`, `upgrade`, `backup`, `restore`, `merge`, `pin`, `unpin`, `uninstall`, `version` y `opsx render`.

## Non-Goals

- No eliminar flags ni comandos existentes.
- No cambiar contratos de automatizacion ni CI.
- No implementar mouse/click como requisito inicial; teclado con enter/espacio es suficiente y portable.
- No ejecutar comandos destructivos sin que el usuario confirme explicitamente los flags que habilitan mutacion.

## Review Slices

### Slice 1: Registry declarativo

- Objetivo: definir comandos y parametros en una estructura reusable y testeable.
- Criterios:
  - WHEN un comando tiene flags booleanos, THEN el registry permite activarlos/desactivarlos.
  - WHEN un comando tiene choices, THEN el registry define valores validos.
  - WHEN se construyen args, THEN solo se incluyen flags habilitados o valores no vacios.

### Slice 2: TUI command palette

- Objetivo: implementar navegacion de comandos y parametros con Bubble Tea.
- Criterios:
  - WHEN el usuario ejecuta `lufy-ai` en TTY, THEN ve comandos y descripciones.
  - WHEN selecciona un comando, THEN ve sus parametros editables.
  - WHEN confirma, THEN se genera y ejecuta la linea equivalente.

### Slice 3: Integracion CLI

- Objetivo: conectar el palette solo para invocacion interactiva sin argumentos.
- Criterios:
  - WHEN `lufy-ai` se ejecuta sin args en TTY, THEN abre TUI.
  - WHEN `lufy-ai` se ejecuta sin args no-TTY, THEN conserva help/usage.
  - WHEN el usuario cancela, THEN no ejecuta ningun comando.

## Validation

- `openspec validate "add-cli-command-palette" --strict`
- `go test ./internal/cli ./internal/tui/commandpalette`
- `go test ./...` desde `tools/lufy-cli-go`
- `scripts/validate.sh`
- Smoke manual: `lufy-ai --help` sigue funcionando y `go run ./cmd/lufy-ai` abre el palette en TTY.
