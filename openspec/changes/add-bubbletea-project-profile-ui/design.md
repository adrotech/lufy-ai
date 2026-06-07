## Contexto

La deteccion surface-aware ya separa stack tecnico de mentalidad del agente. El punto debil actual es la interaccion: el prompt textual solo permite elegir una superficie primaria y no escala bien a multiples roots o proyectos mixtos.

El refactor de `projectconfig` dejo un puerto claro:

```go
type ProfilePrompt func(ProjectConfig) (ProjectProfile, error)
```

La TUI debe implementar ese puerto desde la capa CLI/adapters. `projectconfig` no debe importar Bubble Tea, Bubbles ni Lip Gloss.

## Arquitectura propuesta

Mantener estas responsabilidades:

- `internal/cli`: parsea flags, decide si debe usar adapter interactivo y delega al service.
- `internal/projectconfig`: escanea, mergea, valida, persiste y expone modelos/puertos.
- `internal/tui/projectprofile` o `internal/adapters/tui/projectprofile`: contiene Bubble Tea model/update/view y conversion hacia `ProjectProfile`.

Dependencias esperadas:

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`

El adapter debe recibir `io.Reader`, `io.Writer` y una funcion de TTY detection inyectable para tests. La TUI no debe usar globals salvo el wiring final del CLI.

## UX propuesta

Pantalla principal:

- Lista de superficies detectadas con `id`, `type`, roots y stacks.
- Panel compacto con `agent_lens.primary_concerns` y `validation_expectations`.
- Controles de teclado documentados por la UI de forma minima: navegar, cambiar tipo, toggle activa, confirmar, cancelar.

Flujo:

1. Cargar superficies detectadas desde `ProjectConfig`.
2. Mostrar resumen editable.
3. Permitir cambiar `type` entre `frontend`, `backend`, `mobile`, `cli`, `infra`, `library`, `fullstack`.
4. Recalcular `AgentLens` al cambiar tipo.
5. Confirmar y devolver `ProjectProfile`.
6. Cancelar con error accionable sin escribir cambios.

## Fallback no interactivo

La TUI solo corre si:

- `--interactive` esta habilitado;
- stdin y stdout son TTY;
- el entorno no indica modo CI/headless.

Si no se cumplen las condiciones, se conserva la deteccion automatica y se emite un mensaje corto como hoy.

## Testing

- Tests puros del Bubble Tea model: estados, navegacion, cambio de tipo, confirmacion y cancelacion.
- Tests del adapter: TTY falso conserva perfil detectado.
- Tests CLI: `init --interactive`, `scan`, `scan --interactive=false`.
- Validacion integrada con `scripts/validate.sh`.

## Riesgos

- Dependencias TUI aumentan superficie de build cross-platform.
- Tests pueden volverse fragiles si dependen de render exacto.
- Terminales no TTY o CI podrian bloquear si el gating falla.

Mitigacion: modelo testeable sin snapshots extensos, TTY detection inyectable, fallback headless por defecto seguro y smoke manual en terminal real.
