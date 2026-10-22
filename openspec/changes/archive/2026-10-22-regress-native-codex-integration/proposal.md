# Bloque 3: regresión de integración nativa Codex

## Why
El piloto actual emula roles y desactiva hooks/delegación. La versión y las flags no prueban un handoff nativo. Hace falta comprobar el transporte y la recuperación con observaciones del runtime.

## What Changes
- Añadir un protocolo opt-in aislado con reviewer nativo, hooks de lifecycle y herramientas, un handoff deliberadamente vacío, recuperación sobre el mismo agente y cierre observado.
- Registrar metadatos de hooks mediante su contrato documentado; no parsear transcripts privados ni confundir texto del modelo con prueba de herramientas.
- Conservar resultados negativos/inconclusos, hashes de instrucciones, runtime elegido y fuentes del fixture. Sin evidencia completa no hay pass.

## Capabilities
### New Capabilities
- codex-native-regression: ensayo nativo reproducible y fail-closed.

## Aceptación y workload
WHEN están observados spawn/reviewer/hooks/wait vacío/recovery mismo agente/wait válido/cierre THEN el caso puede pasar si fuentes permanecen intactas y no hay errores.
WHEN falta inventario, hooks confiables o trazabilidad THEN se conserva unavailable/inconclusive sin emular silenciosamente.
WHEN el runtime observa shell, edición o mutación inesperada THEN falla aunque falte otra evidencia.

Slice T1 del programa autorizado por nuevo protocolo transversal y confianza de hooks. Una unidad revisable acopla collector, evaluador, runner, pruebas, documentación y spec/mirror: separar recorder y evaluación permitiría conclusiones sin procedencia. No hay artifacts alternativos: hooks oficiales ofrecen observación estable y transcripts no son contrato estable. Límites `not_available`; no se inventan budgets. Overview HTML `not_available`. Pruebas offline RED/GREEN primero; ensayo live explícito acotado, sin cambios globales, seguido de validación agrupada y delivery propio.
