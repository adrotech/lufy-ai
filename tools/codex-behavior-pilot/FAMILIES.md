# Evaluación por familias de resultados

Este ensayo comprueba la hipótesis de que las acciones y la evidencia deben adaptarse al pedido. No convierte la evaluación en un workflow obligatorio del producto.

`work-scenarios.json` fija seis fixtures y sus criterios antes de invocar candidatos: análisis de código, propuesta, redacción, traducción, corrección y análisis de datos. Los veinte casos de `docs/propuestas/lufy-inline-2026-10-24/evaluacion.md` siguen siendo el corpus conceptual más amplio; los seis fixtures no representan veinte ejecuciones ni cubren una app nueva, operación de datos o seguridad reales.

## Observar sin reescribir la historia

`audit_work.py` lee únicamente evidencia local ya capturada. No invoca modelos ni ejecuta código del candidato. Revisa hash de traza, prompt, entradas, snapshots y estabilidad declarada de fuentes/runtime; reevalúa alcance con el clasificador actual y deja separado el veredicto original. Las fuentes capturadas identifican versiones, sin atestar el host ni el modelo servido.

```sh
python3 tools/codex-behavior-pilot/audit_work.py /ruta/externa/campaña/caso
```

Exit 0 indica que produjo el informe; no declara aceptación. Sin argumentos muestra las seis familias como `not_run`. Los registros ausentes o alterados quedan inconclusos. `needs_review` necesita una revisión semántica explícita; un comando desconocido continúa inconcluso aunque la lectura humana del artefacto resulte favorable. El auditor no transforma un oráculo registrado en una nueva ejecución de ese oráculo.

Los tokens se toman de `turn.completed.usage`, conservando campos ausentes como `null`. No son precio ni costo humano. Sin reloj comparable, precio/modelo verificado o instrumentación de intervenciones, esas métricas se mantienen no disponibles. Los bloqueos y las ejecuciones incompletas permanecen en el informe.

## Calibración y límites

Las pruebas offline rechazan alteraciones de traza, fixture, prompt y procedencia, intentos de escritura read-only aun sin cambios finales, y aceptación por falta de muestras. Un texto correcto, uno falso y otro convincente pero infiel requieren juicio semántico; conservar los ejemplos literales no demuestra calidad. Las pruebas del filtro distinguen implementación correcta, hardcodeo, mutación de datos, symlinks y salida antes de assertions. Son calibración del evaluador con fuentes sintéticas conocidas, no resultados de un agente.

El runner histórico detiene la cadena también si el clasificador no entiende un comando. En el bloque 5 se registró por separado una continuación puntual: no repite casos completados y sólo detiene nuevas invocaciones ante timeout, fallo de proceso, lanzamiento o sandbox. El alcance y la aceptación no se relajan; los comandos desconocidos conservan el veredicto inconcluso. Este cambio de scheduling se documenta antes de continuar y no permite mezclar muestras como si tuvieran condiciones idénticas.

El oráculo usa la sintaxis de la CLI instalada `codex -c 'sandbox_mode="read-only"' sandbox -- ...`. La sonda debe observar denegación de escritura antes de importar el candidato. No hay selección automática de sintaxis antigua ni fallback al host. `exec --ignore-user-config` y las restricciones efectivas del host se comprueban por separado.

## Decisión de producto y comparación

La elección autorizada del usuario y la activación local de inline son decisiones de producto. Una comparación empírica A/B/C requiere muestras nuevas emparejadas: mismo pedido, entradas, runtime, modelo/condiciones, permisos, aceptación y recursos, con instrucciones como variable registrada. No se renombran las trazas anteriores como baseline actual ni se mezclan revisiones del evaluador para obtener un ganador.

Antes de esa comparación, fijar casos/repeticiones, orden alternado, política de parada y tolerancia de degradación. La segunda modificación de código y de texto debe partir de la entrega de cada variante y conservar sus validaciones anteriores. Si faltan pares, segunda modificación o costo humano comparable, la comparación queda inconclusa. El auditor no automatiza adopción ni eliminación de capacidades.

Grafo, memoria y delegación requieren comparaciones con/sin la capacidad en tareas que la necesiten. Este piloto no mide su valor y no justifica activarlas siempre ni retirarlas por ausencia de uso en fixtures pequeños.

La [comparación posterior](../../docs/propuestas/lufy-inline-2026-10-24/implementacion/bloque-5-comparacion.md) completó 24 turnos con `comparison.py`. Sus datos no se mezclan con el piloto. El estudio terminó con límites explícitos y sin adopción automática.

## Cobertura del clasificador, 29/10/2026

La revisión `command-effects/v2` mejora la gramática y agrega diagnósticos por evento. En una lectura offline de los 24 registros existentes (79 comandos completados), 13 de los 33 comandos antes desconocidos se reconocen como lecturas. Los otros 20 siguen inconclusos: 12 heredocs, dos expansiones, cuatro comandos no soportados y dos diffs Git sin desactivación explícita de helpers.

Esto mide cobertura de clasificación. No se ejecutaron modelos ni código de candidatos, no se reejecutaron oráculos y no se modificaron trazas, registros o veredictos históricos. El auditor informa su versión actual y mantiene `original_status`, métricas ausentes y `comparison_eligible: false`. La comparación original conserva su conclusión inconclusa sobre ventaja general y costo humano.
