# Bloque 2: diagnóstico observable de Codex

## Why
El doctor comprueba assets estáticos, pero no distingue el ejecutable elegido, su versión y sus flags de la disponibilidad real de delegación. La memoria y el grafo del repositorio tampoco están inicializados.

## What Changes
- Añadir diagnóstico opt-in `doctor --codex <path-or-command>`: sólo `--version` y `features list`, sin shell ni modelos, con deadline total de cinco segundos y salida acotada.
- Emitir selección explícita/PATH, versión parseada, cuatro flags permitidas y códigos de causa seguros. Delegación nativa permanece `unknown` hasta evidencia operativa del bloque 3.
- Documentar y activar configuración, memoria y grafo locales preservando archivos existentes, sin instalación masiva ni cambios globales.

## Capabilities
### New Capabilities
- codex-runtime-diagnostics: diagnóstico observable y limitado.

## Aceptación y workload
WHEN no se solicita diagnóstico THEN el formato previo del doctor permanece compatible.
WHEN se solicita un ejecutable THEN se usa literalmente y se muestran evidencia parcial/ausente sin exponer stdout o errores crudos.
WHEN sólo se observan flags THEN nunca se afirma disponibilidad nativa.
WHEN se activan memoria/grafo THEN no se sobrescribe configuración existente ni se asumen budgets no configurados.

Slice T2 del programa T1 autorizado para implementar, validar y entregar. Unidad coherente de servicio, integración CLI, pruebas, documentación y spec/mirror; separar reporte y comando dejaría una capacidad incompleta. Límites iniciales `not_available`. Overview HTML `not_available`. Validación agrupada: tests offline del servicio y CLI, suite Go/coverage, validate.sh y OpenSpec estricto; probes reales sólo read-only y explícitos.
