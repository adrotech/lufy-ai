# Bloque 3: observabilidad operativa de Lufy en Codex

## Why
El plan comparativo exige verificar eventos de review y métricas reales sin construir otra plataforma. El ledger existente es content-free e idempotente; Metrics agrupa inicios sólo por run, puede sobrescribir pares ambiguos y no marca timestamps faltantes de rework/reopened.

## What Changes
- Endurecer el emparejado por run/task y la degradación ante pares ambiguos o timestamps ausentes, sin cambiar el shape JSON público.
- Probar la CLI/ledger existente en fixtures: eventos completos, parciales, duplicados y privacidad.
- Documentar soporte estático frente a eventos disponibles en el runtime Codex actual. Una integración de lifecycle no demuestra producción de eventos semánticos review.*.

## Capabilities
### Modified Capabilities
- systemic-workflow: métricas operativas conservadoras.

## Alcance y aceptación
B3-AC1 WHEN hay pares independientes por task en un run THEN contar ambos sin cruzarlos; conservar compatibilidad de task vacío para pares simples.
B3-AC2 WHEN hay inicios duplicados, huérfanos o timestamps inválidos THEN partial y duración sólo de pares inequívocos; ausencia total unavailable.
B3-AC3 WHEN rework/reopened carecen de timestamp THEN partial; counts son observaciones y no tasas ni ceros operativos inferidos.
B3-AC4 WHEN se consultan métricas o reingresa una clave THEN fuentes inmutables y duplicate_noop, sin contenido sensible.
B3-AC5 WHEN Codex sólo expone lifecycle THEN informar review metrics unavailable, separando evidencia sintética de operativa.

Autorización: el usuario aprobó finalizar todos los bloques y su delivery. Programa T1; slice T2 acotado a corrección y verificación del servicio existente. Sin nuevas plataformas, productores automáticos ni cambios OpenCode/Pi.
