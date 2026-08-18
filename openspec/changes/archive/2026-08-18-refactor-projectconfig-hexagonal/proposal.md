## Why

`internal/projectconfig/config.go` concentra demasiadas responsabilidades: modelos YAML, defaults, scanner de stacks, scanner de superficies, merge/rescan, store filesystem, validacion, reporting y helpers de deteccion. El archivo ya funciona y tiene cobertura, pero el crecimiento de `project_profile.surfaces`, `scan` y prompts interactivos confirma que el paquete necesita boundaries mas claros antes de sumar mas detectores o una UI TUI real.

La intencion hexagonal del CLI ya existe en otras zonas (`domain`, `ports`, adapters y registry), pero `projectconfig` todavia mezcla caso de uso con detalles externos. Separarlo ahora reduce riesgo de regresiones cuando se agreguen Bubble Tea, mas stacks, mas superficies o reglas de agentes.

## What Changes

- Separar `internal/projectconfig` en componentes con responsabilidades claras sin cambiar el YAML ni la CLI publica.
- Extraer modelos y defaults puros de configuracion.
- Extraer merge/rescan y drift reporting como logica testeable.
- Extraer detectores de stacks y superficies como strategies registrables.
- Mantener `Service` como orquestador de caso de uso, dependiente de interfaces chicas.
- Preservar comportamiento observable de:
  - `lufy-ai init`;
  - `lufy-ai init --rescan`;
  - `lufy-ai init --interactive`;
  - `lufy-ai scan`.
- Mantener compatibilidad del schema `.lufy/config/project.yaml` version 1.

## Non-Goals

- No cambiar el formato YAML generado.
- No agregar Bubble Tea ni dependencia externa de UI en este cambio.
- No modificar managed assets fuera de lo necesario para docs/specs.
- No cambiar defaults de stacks, surfaces, `workflow_limits`, metodologia o tool.
- No migrar datos de usuario existentes salvo preservar campos desconocidos como hoy.

## Review Slices

### Slice 1: Modelos, defaults y validacion

- Objetivo: mover tipos `ProjectConfig`, `Stack`, `ProjectProfile`, `WorkflowLimits`, defaults y validacion a archivos separados del paquete.
- Archivos esperados: `internal/projectconfig/*.go`, tests existentes.
- Criterios:
  - WHEN se marshaliza una config detectada, THEN el YAML generado es equivalente al actual.
  - WHEN se carga una config con campos extra, THEN los campos preservables siguen intactos.
- Riesgo: churn mecanico; validar con snapshots/fixtures o comparaciones estructurales.

### Slice 2: Rescan/merge como policy aislada

- Objetivo: extraer `RescanMerger`, drift items y merge de overrides a un componente dedicado.
- Archivos esperados: `internal/projectconfig/rescan*.go`, tests de rescan.
- Criterios:
  - WHEN `--rescan` detecta nuevos stacks/surfaces, THEN los agrega sin borrar overrides manuales.
  - WHEN `workflow_limits` o `project_profile` tienen overrides, THEN se preservan segun reglas actuales.
- Riesgo: cambiar semantica de drift report; mantener mensajes accionables.

### Slice 3: Detectores como strategies

- Objetivo: reemplazar el scanner monolitico por detectores independientes para Go, JS/TS, Python, JVM, unsupported stacks, infra y surfaces.
- Archivos esperados: `internal/projectconfig/detectors/*.go` o archivos equivalentes en el paquete.
- Criterios:
  - WHEN se agrega un detector nuevo, THEN no requiere tocar un switch central grande salvo registro.
  - WHEN un detector falla o no aplica, THEN no bloquea otros detectores.
- Riesgo: abstraccion prematura; mantener interfaces pequenas.

### Slice 4: Service y adapters de IO

- Objetivo: dejar `Service.Run` como orquestador con dependencias inyectables y filesystem store aislado.
- Archivos esperados: `internal/projectconfig/service.go`, `store.go`, tests de CLI/service.
- Criterios:
  - WHEN CLI llama `init` o `scan`, THEN el service coordina scan/merge/prompt/write sin conocer detalles internos de cada detector.
  - WHEN tests usan fakes, THEN pueden cubrir merge y store sin IO real salvo pruebas de adapter.
- Riesgo: sobrefragmentacion; mantener nombres claros y funciones puras cuando alcance.

## Validation

- `go test ./internal/projectconfig ./internal/cli`
- `scripts/validate.sh`
- `openspec validate "refactor-projectconfig-hexagonal" --strict`
- Revision estatica de que el diff no cambia defaults publicos ni schema YAML.
