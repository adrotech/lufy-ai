> Informe histórico del modelo anterior. Sus resultados se conservan; sus comandos y roles no describen la rama inline actual. Ver [estado vigente](status.md).

# Experimentos de ejecución de Lufy en Codex — 2026-10-22

## Decisión

Se conserva el workflow vigente (`keep_baseline`, `adopted=false`). El ensayo real no obtuvo un par completo: la primera muestra baseline agotó 180 segundos. No se ejecutó el candidato ni el segundo par. No hay evidencia para adoptar entrega por referencia, cambiar routing TDD o afirmar ahorro de tokens, coste o latencia.

## Protocolo y resultados

El protocolo permite dos pares del caso auto-chain, alternando orden inline/referencia. Las fuentes, instrucciones, runtime, sandbox y aceptación deben coincidir. Se reevalúan JSONL, stderr y snapshots, sin confiar en el pass guardado; falta de procedencia, usage o finalización impide elegibilidad. Dos pares válidos con menos bytes y tokens de entrada y sin mayor duración sólo sostendrían al candidato en esa muestra; nunca lo adoptan automáticamente.

| Observación | Evidencia |
| --- | --- |
| Runtime solicitado | CLI incluida con ChatGPT, `0.163.0-alpha.10.2`; `gpt-6.1-sol` / `medium`, identidad servida no verificada |
| Baseline real | 700 bytes de prompt, 180032 ms hasta timeout, tokens de entrada unavailable |
| Candidato | No ejecutado por parada de infraestructura; métricas unavailable |
| Resultado | Par incompleto, no elegible; mantener baseline |
| Evidencia local | `/private/tmp/lufy-codex-context-experiment-permitted-20261022/summary.json` |

Una primera invocación exterior falló antes del agente por `Operation not permitted` al iniciar el app-server dentro del sandbox de herramientas. Se preservó en `/private/tmp/lufy-codex-context-experiment-20261022`. La segunda contó con permisos para iniciar el CLI; el agente conservó su sandbox workspace-write. No se repitió el experimento tras el timeout. El ensayo precedió al ajuste `agents.enabled=false` del piloto; su argv queda conservado y no se mezcla con muestras posteriores como si fueran pareadas.

## Recuperación, autorización y TDD

Se reutilizó el checkpoint existente del ledger con un fixture sintético: la primera llamada produjo `recorded`; la repetición con la misma clave, `duplicate_noop`; hashes de fuentes sin cambios y `run verify` saludable. Evidencia: `/private/tmp/lufy-block4-recovery-20261022.json`. Esto verifica idempotencia del mecanismo; no promueve gates reales ni convierte una hint en aprobación.

La autorización humana de finalizar todos los bloques cubre implementación y delivery del plan de Lufy para Codex. No autoriza otros proyectos, OpenCode/Pi, cambios de seguridad o permisos por un next owner. El caso auto-chain del piloto conserva una aprobación post-plan explícitamente ausente y comprueba el bloqueo; el contenido sintético no hereda la autorización de esta sesión.

La comparación de TDD especializado frente a unidades completas es un desk check, sin ensayo causal pareado de modelos: el test-writer independiente detectó cruces y ambigüedad en métricas, y falsos positivos de trazabilidad/procedencia en los evaluadores; root implementó y ejecutó GREEN. Es evidencia de defectos detectados y separación de responsabilidades, no de menor coste/latencia de esa estrategia. Se conserva el routing existente.

El receptor de handoff se recuperó de un productor completo previamente revalidado, preservando fuentes y ejemplos; un productor incompleto o alterado se rechaza. La recuperación conserva evidencia histórica y aprobación requerida; su resultado se registra en el reporte B2, separado de este experimento.

## Validación y límites

`python3 -m unittest discover -s tools/codex-behavior-pilot -p 'test_*.py' -v`: 68/68 passed, incluyendo 19 pruebas de comparación y 10 nuevas de recuperación/desactivación. RED delegado precedió a implementación; adversariales expusieron procedencia falsa y tipos inválidos antes del fix. No sustituyen conducta real de modelos.

Configuración `workflow_limits`, memoria y grafo no disponibles. El tamaño del ensayo es un límite de este protocolo, no un budget configurado del proyecto. No se inicializaron servicios, instalaron herramientas ni modificaron credenciales/settings globales. Las trazas completas permanecen locales; sólo se publica evidencia sintética revisada.
