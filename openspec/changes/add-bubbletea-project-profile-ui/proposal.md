## Why

`lufy-ai init --interactive` y `lufy-ai scan` ya tienen una decision de producto importante: definir la mentalidad del agente segun `project_profile.surfaces`. Hoy esa interaccion existe como prompt textual lineal, suficiente para validar la idea pero limitado para proyectos con varias superficies, monorepos, fullstack flows o ajustes repetidos.

La CLI necesita una experiencia interactiva mas clara sin romper automatizacion. Bubble Tea/Charm encaja para construir una TUI enfocada, testeable y portable, siempre que quede aislada como adapter de entrada y no se mezcle con `projectconfig`.

## What Changes

- Agregar una TUI Bubble Tea/Charm para seleccionar y ajustar `project_profile.surfaces` durante `lufy-ai init --interactive` y `lufy-ai scan`.
- Mantener `projectconfig.ProfilePrompt` como puerto de aplicacion.
- Reemplazar el prompt textual actual por un adapter TUI cuando haya TTY y `--interactive` este habilitado.
- Preservar fallback no interactivo cuando no haya TTY, stdin/stdout no sean terminales o `--interactive=false`.
- Permitir que la TUI muestre superficies detectadas, tipo, roots, stacks, frameworks y lens resultante antes de confirmar.
- Mantener compatibilidad del schema `.lufy/config/project.yaml` version 1.

## Non-Goals

- No cambiar el formato YAML ni agregar schema version 2.
- No mover reglas de dominio o merge a Bubble Tea.
- No hacer que `scan` bloquee en CI, hooks, scripts o pipes.
- No implementar un editor completo de YAML.
- No agregar una landing/marketing UI para la CLI.
- No cambiar defaults de `workflow_limits`, stacks o metodologia.

## Review Slices

### Slice 1: Adapter TUI minimo para ProfilePrompt

- Objetivo: introducir dependencias Charm, modelo Bubble Tea y adapter que produce `projectconfig.ProjectProfile`.
- Archivos esperados: `tools/lufy-cli-go/internal/cli`, `tools/lufy-cli-go/internal/tui/projectprofile` o `internal/adapters/tui/projectprofile`.
- Criterios:
  - WHEN `ProfilePrompt` recibe un `ProjectConfig`, THEN la TUI puede devolver un `ProjectProfile` sin escribir archivos.
  - WHEN stdin/stdout no son TTY, THEN se conserva el fallback no interactivo.
- Riesgo: acoplar TUI a CLI o dominio; mantener el puerto `ProfilePrompt`.

### Slice 2: Edicion de superficies

- Objetivo: permitir revisar y ajustar superficies detectadas sin editar YAML manualmente.
- Archivos esperados: adapter TUI, tests de modelo, tests CLI.
- Criterios:
  - WHEN existen varias superficies, THEN el usuario puede navegar entre ellas y cambiar el `type`.
  - WHEN cambia el `type`, THEN se recalcula `agent_lens` con `DefaultAgentLens`.
  - WHEN confirma, THEN `init/scan` escribe el perfil elegido.
- Riesgo: complejidad de interaccion; priorizar controles concretos y estados observables.

### Slice 3: Integracion CLI y fallback robusto

- Objetivo: conectar `init --interactive` y `scan` con la TUI sin romper automatizacion.
- Archivos esperados: `internal/cli/app.go`, tests CLI, docs de uso.
- Criterios:
  - WHEN `lufy-ai scan --interactive=false` corre en CI, THEN no intenta abrir TUI.
  - WHEN `lufy-ai scan` corre con TTY, THEN abre TUI por defecto segun comportamiento actual.
  - WHEN Bubble Tea devuelve cancelacion, THEN el comando sale con error accionable y no escribe cambios parciales.
- Riesgo: comportamiento diferente entre terminales; cubrir fallback y cancelacion.

## Validation

- `go test ./internal/projectconfig ./internal/cli`
- `go test ./internal/tui/...` si se crea paquete dedicado
- `scripts/validate.sh`
- `openspec validate "add-bubbletea-project-profile-ui" --strict`
- Smoke manual en terminal real para `lufy-ai init --interactive` y `lufy-ai scan`
