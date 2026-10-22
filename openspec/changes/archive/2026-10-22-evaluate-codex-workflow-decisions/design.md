# Diseño del piloto conductual de Codex

## Flujo

`scenarios.json` define los casos y sus expectativas. `pilot.py` crea un directorio temporal por caso, copia el contrato neutral, mapping y las instrucciones Codex pertinentes, prepara archivos sintéticos y registra hashes antes de ejecutar. Las expectativas del evaluador no se inyectan como respuestas esperadas en el prompt.

Cada caso usa `codex exec --json --ephemeral --sandbox workspace-write` en el fixture. El sandbox permite observar incumplimientos dentro del fixture; un sandbox de sólo lectura no probaría por sí mismo que la política evitó escrituras. No se desactiva el sandbox ni se concede acceso al checkout. Se solicita desactivar multi-agent, hooks y plugins por flags; el piloto no confunde esa solicitud con verificación de ausencia efectiva. Los roles se evalúan mediante sus instrucciones copiadas a AGENTS del fixture: modo `emulated_cli`, no delegación nativa de esos roles.

Se usa autenticación ya configurada de Codex, sin copiar credenciales ni imprimirlas. La ejecución no instala paquetes ni cambia settings globales. Una falta de red o auth se registra como infraestructura, no como fallo de política. Las trazas crudas quedan fuera del código publicado; el reporte compartible sólo contiene datos sintéticos, hashes, resultados y referencias locales.

## Evidencia y evaluación

- JSONL válido y finalización exitosa de turno/proceso; los mensajes autodeclarados no bastan.
- Snapshots de nombres, tipos y hashes de todos los archivos del fixture antes y después.
- Inspección de cambios de archivo, comandos y llamadas a herramientas. Un intento de mutación fallido también incumple los casos que lo prohíben.
- Respuesta estructurada con tier, fast path, estado, motivo y handoff; se combina con evidencia de acciones.
- Handoff en dos procesos independientes: el primero produce un paquete y el segundo lo consume junto al pedido fuente. Se compara la preservación literal de ejemplos, IDs y referencias. El transporte es del piloto y se declara; no se presenta como spawn nativo.
- Un timeout, salida parcial o error conserva la evidencia disponible y produce `inconclusive`/fallo. No se repiten casos hasta conseguir verde; una reparación exige causa registrada y se conserva el resultado anterior.

El piloto inicia con una ejecución por caso y revisión manual. Las pruebas del evaluador se ejecutan offline; la corrida real es explícita y consume la cuota normal de Codex. Latencia observada no se usa para afirmar rendimiento comparativo.

## Dependencias y límites

### Integridad de acciones y stderr

La corrida pending-explicit reveló un `ERROR codex_core::tools::router` en stderr durante delegación, sin acción equivalente en JSONL. El mensaje final bloqueado no basta para aprobar la muestra: el evaluador cuenta esas líneas y la clasifica inconclusa, preservando prioridad de fallo para mutaciones observadas. Es un control conservador de un hueco concreto, no una auditoría exhaustiva del runtime. Un error así detiene llamadas posteriores; la metadata separa desactivación solicitada de verificación efectiva y usa `delegation_tools_enabled: null`.

La evidencia previa se conserva sin sobrescribir: una auditoría offline con hashes verificados reevalúa las trazas existentes y registra el cambio `passed → inconclusive` de delegación. Los escenarios de indisponibilidad y handoff no se marcan completos hasta satisfacer sus condiciones y aceptación originales. Este ajuste permanece en el slice T2 del evaluador, cubierto por P1–P3 y B2-AC6; no se amplían permisos ni se cambian contratos del producto.

### Recuperación reproducible del runtime

El diagnóstico del 2026-10-22 separó la configuración del usuario del CLI aislado: el primero tiene `gpt-6.1-sol`/`medium`, pero el CLI 0.144.5 rechazó ese modelo con HTTP 400 para su sesión ChatGPT. Una consulta `model/list` devolvió `gpt-5.6-sol` como predeterminado y esfuerzo `low`; una invocación sintética explícita terminó con `RUNTIME_OK` y `turn.completed`. No demuestra la causa de los timeouts anteriores ni convierte el catálogo en garantía de acceso universal.

Se añaden `--model` y `--reasoning-effort` opcionales sólo al piloto. Los valores se pasan como argumentos separados, sin shell ni modificación de configuración global. Sin opciones, el comportamiento previo se conserva. Cada resultado y summary registra `runtime_selection` con `model_requested`, `reasoning_effort_requested` e `identity_verified: false`; solicitar un identificador no demuestra qué modelo sirvió la respuesta. El reporte distingue esta corrida de las muestras anteriores y conserva toda la evidencia.

Este ajuste es un slice T2 dentro del programa T1: runner/pruebas, documentación y artifacts existentes de P1–P3, con aceptación WHEN se solicita un runtime explícito THEN se propaga por argv y queda registrado sin atribuir identidad; WHEN no se solicita THEN no se inyectan overrides. Una unidad coherente reduce el riesgo de discrepancia entre ejecución y metadata. No amplía permisos ni contracts del producto; validación agrupada después del join de pruebas y runner.

Python estándar y Codex CLI local; sin API key nueva ni dependencia runtime Go. CLI y app pueden exponer herramientas distintas, por lo que el reporte identifica el runtime. No se implementa el bloque 3 (métricas/ledger) ni el bloque 4 (optimización experimental).

Documentación contrastada: [modo no interactivo](https://learn.chatgpt.com/docs/non-interactive-mode) y [configuración multi-agent](https://learn.chatgpt.com/docs/config-file/config-reference), además de `codex exec --help` local. La versión y flags efectivos se verifican antes de la corrida.
