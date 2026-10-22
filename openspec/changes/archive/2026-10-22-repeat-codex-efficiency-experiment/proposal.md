# Bloque 7: experimento reproducible de eficiencia Codex

## Why
El ensayo previo quedó incompleto antes del ajuste de aislamiento. No permite inferir ahorro. El comparador actual identifica el path del CLI pero no detecta reemplazo del binario en ese mismo path ni variaciones de argv de aislamiento entre muestras.

## What Changes
- Reutilizar el experimento inline/referencia y sus criterios de aceptación, sin ejecutar un benchmark de otro runtime ni adoptar cambios automáticamente.
- Registrar antes/después de cada caso SHA-256/versión/path del CLI y hashes de runner, escenarios e instrucciones efectivamente usadas; rechazar procedencia ausente, cambiante o distinta.
- Comparar el perfil argv conservando sandbox/flags/modelo; normalizar únicamente paths temporales de fixture/schema/salida.
- Ejecutar hasta dos pares alternados con la CLI incluida con la app, fuentes estables y modelo/esfuerzo explícitos por invocación. Parar ante el primer caso incompleto/fallido o drift; conservar evidencia y baseline.
- Publicar reporte de observaciones sintéticas revisadas. Ningún dato ausente se convierte en cero ni en mejora.

## Capabilities
### New Capabilities
- codex-efficiency-evidence: muestras comparables ligadas a procedencia observada del ejecutable/perfil/fuentes.

## Aceptación y workload
WHEN procedencia y aceptación son válidas e idénticas THEN pueden compararse prompt bytes, usage observado y duración.
WHEN cambia ejecutable/fuentes/perfil, falta finalización/usage o falla aceptación THEN se rechaza elegibilidad, se detienen nuevas llamadas y se conserva baseline.
WHEN dos pares completos sostienen una variante THEN sólo se declara soporte en esa muestra, con adopted=false.
WHEN el ensayo real queda inconcluso THEN se publica ese resultado sin reintentos para obtener verde ni cambios globales de credenciales/permisos/modelos.

T1 por contrato de evidencia experimental y nueva comparación de procedencia. Una unidad runner/tests/docs/spec/mirror con validación offline agrupada y ensayo live separado. Fuentes del bloque 6 ya publicadas; su archive no cambia instrucciones del ensayo. workflow_limits no disponible: máximo dos pares y timeout son guardas del protocolo local, no budgets de workflow inventados. Graph worktree not_available; fallback dirigido al piloto y ensayo previo. Overview HTML not_available; branching no necesario. Autorización del usuario cubre la implementación, ensayo y entrega de este bloque.
