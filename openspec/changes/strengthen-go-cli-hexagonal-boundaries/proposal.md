## Why

La CLI Go ya tiene una base saludable: `internal/core/domain`, `internal/ports`, adapters de tool/metodologia, registry, tests amplios y validacion agrupada con cobertura agregada de 80.0%. Sin embargo, la arquitectura hexagonal todavia es parcial.

La revision estatica detecto que los casos de uso principales siguen mezclando reglas de aplicacion con detalles externos:

- `internal/installer`: planificacion, locks, filesystem, backup, render, manifest, verify y salida humana.
- `internal/syncer`: logica similar a installer con duplicacion operacional.
- `internal/verify`: construccion de reporte, lectura de estado, filesystem, hashes, JSON, memoria Obsidian y validacion de assets.
- `internal/projectconfig`: scanning, merge, YAML y escritura.

Esto limita SOLID, especialmente SRP y DIP, y hace que los patrones existentes funcionen bien en adapters pero no todavia en todos los use cases. El objetivo es convertir la intencion hexagonal actual en boundaries aplicables, testeables y revisables sin romper el preset `opencode` + `openspec`.

## What Changes

- Definir boundaries hexagonales estrictos para la CLI Go:
  - dominio puro;
  - casos de uso orientados a puertos;
  - adapters externos para filesystem, state, backup, runtime/config, clock y red.
- Separar servicios grandes en componentes con responsabilidad unica:
  - builders de plan;
  - executors/appliers;
  - reporters;
  - stores/providers;
  - validators/checkers.
- Reemplazar progresivamente strings magicos de acciones por tipos/constantes y estrategias de ejecucion.
- Reducir duplicacion entre `installer` y `syncer` donde haya reglas compartidas reales.
- Mantener compatibilidad observable de `install`, `sync`, `verify`, `status`, `backup`, `restore`, `uninstall` y `upgrade`.
- Establecer criterios de clean code, SOLID y patrones como gates verificables de reviewer.
- Reforzar TDD/AAA para tests nuevos o modificados durante este refactor.

## Current Assessment

### Cumplimientos actuales

- `ToolAdapter` y `MethodologyAdapter` implementan Strategy/Adapter.
- `adapters/registry` implementa Factory/Registry.
- `cli.Run` mantiene `main` delgado y despacha comandos.
- El codigo usa early returns y errores explicitos de forma consistente.
- Hay buena inversion en tests: 46 archivos `_test.go`, 7.593 lineas de tests y cobertura agregada 80.0%.

### Brechas principales

- Los use cases dependen de detalles concretos (`os`, `filepath`, `platform`, `state`, `backup`, `toolruntime`) en vez de puertos.
- `installer`, `syncer`, `verify` y `projectconfig` concentran demasiadas razones para cambiar.
- `Action.Kind` usa strings y switches grandes para comportamiento que deberia poder crecer por estrategia.
- AAA existe implicitamente en muchos tests, pero no como convencion estable.
- La evidencia TDD historica no es demostrable desde el repo; el refactor debe producir evidencia TDD cuando toque comportamiento sustantivo.

## Review Slices

### Slice 1: Puertos de aplicacion y boundaries

- Objetivo: introducir puertos para filesystem/state/backup/runtime/clock/reporting donde agreguen valor real.
- Archivos esperados: `internal/ports`, `internal/installer`, `internal/syncer`, `internal/verify`, `internal/projectconfig`, tests.
- Criterios:
  - WHEN un use case necesita leer/escribir estado, archivos, backups o runtime de tool, THEN lo hace mediante una abstraccion inyectable o un facade de aplicacion.
  - WHEN se ejecuta el preset default, THEN el comportamiento observable no cambia.
- Riesgo: sobregeneralizar; evitar puertos sin uso claro.

### Slice 2: Separacion plan/apply/report/check

- Objetivo: reducir responsabilidades en servicios grandes sin reescribir todo.
- Archivos esperados: `internal/installer`, `internal/syncer`, `internal/verify`, `internal/status`.
- Criterios:
  - WHEN se construye un plan, THEN no se mezclan escrituras reales ni salida humana salvo evidencia explicitamente modelada.
  - WHEN se aplica un plan, THEN las mutaciones quedan en executors/appliers testeables.
- Riesgo: duplicar tipos; consolidar solo donde haya comportamiento compartido real.

### Slice 3: Acciones tipadas y execution strategy

- Objetivo: reemplazar strings fragiles por tipos/constantes y preparar dispatch por estrategia cuando el switch deje de ser simple.
- Archivos esperados: `internal/installer`, `internal/syncer`, tests.
- Criterios:
  - WHEN una accion nueva se agrega al plan, THEN tiene tipo declarado, tests y semantica de confirmacion/backup.
  - WHEN el executor recibe una accion desconocida, THEN falla explicitamente.
- Riesgo: introducir abstraccion prematura; mantener el modelo chico.

### Slice 4: TDD/AAA y reviewer gates

- Objetivo: estandarizar tests nuevos o modificados con AAA observable y evidencia TDD proporcional.
- Archivos esperados: tests del paquete tocado, docs si aplica.
- Criterios:
  - WHEN se modifica comportamiento, THEN hay test RED/GREEN o razon `not_applicable`.
  - WHEN se crea un test nuevo, THEN su estructura de setup/act/assert es clara por bloques, helpers o naming.
- Riesgo: convertir tests Go idiomaticos en ceremonia; AAA puede ser explicito o estructural.

## Non-Goals

- No cambiar defaults publicos de CLI.
- No implementar soporte escribible para `codex` o `claude-code`.
- No mover rutas gestionadas ni cambiar layout instalado.
- No reescribir todos los paquetes en una unica PR.
- No imponer una libreria externa de testing.
- No exigir comentarios `Arrange/Act/Assert` cuando la estructura del test ya sea inequívoca.

## Validation

- Validacion base observada antes de crear la proposal:
  - `scripts/validate.sh` paso.
  - cobertura agregada Go: 80.0%.
  - `shellcheck` no disponible localmente; el script lo omitio.
- Cada slice implementado debera ejecutar validacion agrupada proporcional:
  - `scripts/validate.sh` para cambios de CLI/arquitectura.
  - `openspec validate "strengthen-go-cli-hexagonal-boundaries" --strict` cuando el CLI OpenSpec este disponible.
  - Revision estatica de boundaries y diffs cuando el cambio sea documental.
