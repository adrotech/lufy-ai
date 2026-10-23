# Bloque 4: optimización experimental con baseline

## Why
Reducir coordinación exige medir primero; no basta mover texto o reutilizar una aprobación. El baseline son las muestras auditadas del piloto y las garantías verificadas del ledger, con sus límites de runtime explícitos.

## What Changes
- Comparar fuente inline frente a fuente por referencia en un experimento aislado de no escritura, conservando instrucciones de rol, contrato y aceptación.
- Registrar tamaño del prompt, usage expuesto y duración observada, sin llamarlos ahorro global ni benchmark estadístico.
- Evaluar recuperación Lite/checkpoints, alcance de autorizaciones y estrategia TDD con evidencia existente y controles deterministas; distinguir desk check de experimento real.
- Mantener el workflow actual si una variante falla, pierde evidencia o no demuestra mejora reproducible. Acotar intentos; agotarlos no concede aprobación.

## Capabilities
### New Capabilities
- codex-execution-experiments: experimentación local auditada sin cambiar defaults del harness.

## Aceptación
B4-AC1 WHEN se compara un candidato THEN se conservan fuentes/aceptación y se identifica runtime y modo; no se concede permiso por next owner.
B4-AC2 WHEN hay muestras pareadas válidas THEN se muestran métricas observadas con procedencia; datos ausentes quedan unavailable.
B4-AC3 WHEN falla fidelidad, trazabilidad o no existe mejora reproducible THEN no se adopta la variante y se conserva el baseline.
B4-AC4 WHEN se recupera un checkpoint o reutiliza autorización THEN se comprueban idempotencia y alcance; una hint no avanza gates.
B4-AC5 WHEN se agota el ensayo acotado THEN se registra bloqueo/no adopción sin más ciclos de corrección ni afirmaciones de éxito.

El usuario autorizó finalizar todos los bloques. Este bloque concluye con una decisión experimental sustentada; un resultado negativo/no adopción es válido. No exige inventar una mejora ni sustituye la aceptación pendiente del bloque 2.
