## 1. Preparacion

- [x] Confirmar baseline con `go test ./internal/projectconfig ./internal/cli`.
- [x] Revisar comportamiento actual de `init --interactive`, `scan` y `scan --interactive=false`.
- [x] Confirmar ubicacion del adapter TUI sin introducir Bubble Tea en `internal/projectconfig`.

## 2. Dependencias y boundaries

- [x] Agregar dependencias Charm necesarias en `tools/lufy-cli-go/go.mod`.
- [x] Crear paquete TUI dedicado para `project_profile`.
- [x] Mantener `projectconfig.ProfilePrompt` como puerto de aplicacion.

## 3. Modelo Bubble Tea

- [x] Implementar modelo con lista de superficies detectadas.
- [x] Permitir navegar superficies.
- [x] Permitir cambiar `type` de superficie.
- [x] Recalcular `AgentLens` al cambiar `type`.
- [x] Implementar confirmacion y cancelacion.

## 4. Integracion CLI

- [x] Conectar `lufy-ai init --interactive` al adapter TUI.
- [x] Conectar `lufy-ai scan` al adapter TUI cuando haya TTY.
- [x] Mantener `scan --interactive=false` y entornos no TTY sin bloqueo.
- [x] Reportar cancelacion con error accionable sin escritura parcial.

## 5. Tests y documentacion

- [x] Agregar tests del modelo TUI.
- [x] Agregar tests del adapter/fallback no interactivo.
- [x] Ajustar tests CLI existentes.
- [x] Documentar el flujo en README o docs/getting-started cuando aplique.

## 6. Validacion final

- [x] Ejecutar `go test ./internal/projectconfig ./internal/cli`.
- [x] Ejecutar tests del paquete TUI nuevo.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `openspec validate "add-bubbletea-project-profile-ui" --strict`.
- [x] Realizar smoke manual en terminal real para `init --interactive` y `scan`.
