# Plan de preparación operativa de Lufy para Codex

Programa T1 autorizado por el usuario el 2026-10-22 para implementar, validar y entregar siete mejoras en este orden. Base inicial: `d83b8f7e710d078fa36e33dc4eebaa666f4bbf5b`. El alcance es exclusivamente Lufy en Codex. La autorización incluye commit, push, PR y merge condicionado a evidencia y checks; no concede cambios globales de modelos, credenciales ni permisos.

| Bloque | Entrega | Aceptación observable | Estado |
| --- | --- | --- | --- |
| 1 | Pruebas offline en CI | Los paths del piloto disparan CI; la validación ejecuta sus tests sin iniciar modelos ni requerir autenticación | validated |
| 2 | Diagnóstico Codex y uso de configuración local | El diagnóstico distingue configuración, ejecutable y capacidades observadas; memoria/grafo locales se activan de manera explícita sin sobrescribir archivos del usuario | planned |
| 3 | Regresión de integración nativa | Un protocolo reproducible conserva evidencia de roles, hooks, handoff y recuperación; una capacidad no expuesta queda unavailable, nunca un pass inferido | planned |
| 4 | Eventos semánticos de review | Operaciones explícitas de revisión emiten eventos content-free e idempotentes; el lifecycle genérico no se convierte en review por inferencia | planned |
| 5 | Evidencia ligada al candidato | Identidad de repositorio, base y candidato vinculan la evidencia; modificar el candidato invalida su vigencia sin reescribir el resultado original | planned |
| 6 | Instrucciones Codex coherentes | Reglas comunes, mapping y mirrors tienen una fuente identificada; comprobaciones detectan divergencia y mensajes incompatibles con el lifecycle soportado | planned |
| 7 | Experimento reproducible de eficiencia | Variantes usan fuentes, aceptación y runtime comparables; datos incompletos conservan baseline y ninguna variante se adopta automáticamente | planned |

## Ejecución y review

Cada bloque tiene su propio change OpenSpec, aceptación WHEN/THEN, evidencia y entrega. CI/diagnóstico son slices T2; decisiones transversales de autoridad o contratos conservan T1 cuando corresponda. La validación se agrupa al terminar cada bloque; sólo diagnósticos y RED/GREEN justificados se adelantan. El siguiente bloque comienza después de publicar la unidad anterior, preservando el orden solicitado. Los estados `implemented`, `validated`, `delivery_pending`, `delivered` y `closed` conservan sus gates independientes.

Se mantiene el checkout principal mixto y se trabaja en un worktree aislado. Git/GH pertenece al rol delivery; el agente principal puede ejecutar su relay revisado cuando el entorno del subagente no complete escalaciones. Las tareas de código y pruebas conservan ownership separado cuando TDD aporta evidencia. No se paralelizan mutaciones Git/GH ni se delega por volumen solamente.

`workflow_limits` inicialmente no disponible; no se infieren budgets. La configuración local del bloque 2 debe distinguir opciones explícitas de defaults. No se promete una mejora de rendimiento ni se usa un benchmark de otro runtime como prueba de Codex. Sin prueba suficiente, se conserva el comportamiento vigente y se informa el bloqueo o no adopción.
