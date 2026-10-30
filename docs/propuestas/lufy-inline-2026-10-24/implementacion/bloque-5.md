# Bloque 5: evaluación por familias — avance parcial

Continuación posterior: [comparación exploratoria y segunda modificación](bloque-5-comparacion.md). Este informe conserva el estado del primer avance.

27/10/2026, hora de Argentina. **El piloto avanzó; el bloque 5 completo sigue pendiente.** Se ejecutaron los cuatro casos autorizados de redacción, traducción, corrección y datos. Junto con los dos casos anteriores, hay una entrega completa por cada una de las seis familias del piloto. No se completaron una comparación A/B/C, una segunda modificación ni una medición comparable de costo humano.

La instalación personal inline activada en el bloque 4 permanece vigente. Es una decisión de producto autorizada; esta muestra no demuestra superioridad general ni justifica declarar terminada la adopción empírica prevista en el plan.

## Resultado por familia

| Familia | Observación comprobada | Límite |
| --- | --- | --- |
| Análisis de código | La muestra previa explica factor 1.21, redondeo y 121 para 100; no cambia archivos. | Revisión posterior de contenido; registro original inconcluso. El evaluador actual da needs_review. |
| Propuesta | La muestra previa compara CSV/XLSX/asíncrono y recomienda una entrega coherente con el brief, filtros y ausencia de correo/storage. No implementa. | No demuestra factibilidad de una implementación ni comparación entre procesos. |
| Redacción | `guide.md` conserva el comando, explica generación y directorio de salida; referencia intacta. | Contenido revisado por el principal. Clasificador automático inconcluso por comandos Git/Python no evaluables. |
| Traducción | Preserva las cuatro unidades de significado, el ejemplo literal y el original. | Mismo límite del clasificador. No se transforma en pass automático. |
| Corrección | Modifica sólo `app.py`; pasan los dos tests visibles y un oráculo externo discriminante. | Es una corrección pequeña; no una app nueva ni una funcionalidad integrada. |
| Datos | Diagnostica el join uno a muchos y propone sumar ventas o usar EXISTS. El revisor reprodujo 200/100/100 en SQLite en memoria. | Fixtures sintéticos; no producción. El clasificador mantiene inconcluso el comando Python usado por el candidato. |

El filtro tuvo RED funcional real antes de la modificación y GREEN después, dentro del sandbox: paid/pending/missing/other, None y entrada vacía, con orden y datos conservados. No se alteraron pruebas existentes. El oráculo vive fuera del fixture y la sonda comprueba denegación de escritura antes de importar el código del candidato.

El contraejemplo de dos ventas diferentes de 100 comprueba también por qué `SUM(DISTINCT amount)` pierde una venta; EXISTS mantiene 200. Lo ejecutado por el revisor es SQL propio sobre una base en memoria, sin modificar los CSV ni ejecutar consultas contra datos reales.

## Ejecuciones y recuperaciones conservadas

- Historial anterior: dos lanzamientos sin turno completo (CLI global/arranque de app-server) y dos turnos completos de análisis/propuesta. Sus archivos y veredictos no se reescribieron.
- Autorización nueva: cuatro casos, una invocación de modelo por caso, máximo 60 segundos, sin paralelismo ni publicación. Los cuatro terminaron dentro del límite.
- La primera cadena se detuvo tras redacción por un comando desconocido para el evaluador. Se registró una continuación puntual: los límites de clasificación conservan inconcluso, pero la parada de nuevas invocaciones distingue un fallo real de infraestructura. No se repitió redacción ni se relajó la aceptación.
- Traducción terminó. El filtro se detuvo antes del modelo porque el runner usaba `sandbox macos --`, mientras la CLI instalada recibe el comando directamente después de `sandbox --`. Datos quedó temporalmente not_run.
- Diagnóstico sin modelos: `sandbox --help` y una fuente sintética. Una primera prueba con `--ignore-user-config` global fue rechazada porque esa opción pertenece a exec. Se corrigió la invocación sin añadir fallback. La sonda read-only pasó antes de reanudar filtro/datos en directorios nuevos. Ninguno había invocado un modelo en el intento bloqueado.
- Contabilidad de registros de casos: nueve intentos capturados, seis turnos completos y tres inconclusos de infraestructura —dos históricos y el preflight del filtro—. Los not_run intermedios siguen en las summaries originales; no son intentos adicionales de modelos. La recuperación con fuentes sintéticas se registra aparte.

Las tres muestras nuevas de texto/datos tienen contenido correcto en la revisión posterior del principal y estado automático inconcluso. No hubo escrituras fuera del alcance del fixture observadas. Los comandos no clasificados impiden acreditar exhaustivamente sus efectos desde este observador; no se declara una campaña con todos los controles críticos resueltos.

## Recursos observados

CLI de app: `0.164.0-alpha.11.1`. Default solicitado; identidad del modelo servido no verificada. Los hashes de runtime, política y fuentes permanecieron iguales dentro de cada captura. Hay revisiones distintas entre campañas; no se mezclan como si fueran condiciones idénticas.

| Caso completo | Veredicto original | Auditoría actual | Input tokens | Cached input tokens | Output tokens |
| --- | --- | --- | --- | --- | --- |
| `explain-code` | inconclusive | needs_review | 40808 | 37888 | 225 |
| `propose-export` | inconclusive | needs_review | 40880 | 38016 | 592 |
| `write-guide` | inconclusive | inconclusive | 73440 | 63232 | 627 |
| `translate-guide` | inconclusive | inconclusive | 58837 | 49408 | 425 |
| `fix-filter` | passed | passed | 41575 | 36096 | 275 |
| `analyze-join` | inconclusive | inconclusive | 55354 | 46848 | 628 |

Estos valores proceden de `turn.completed.usage`. No son precios ni costo humano; cached input forma parte del uso reportado y no se interpreta como ahorro entre métodos. Duración comparable, precio e intervenciones humanas medidas: **no disponibles**, conservados como null. Hubo diagnóstico/recuperación y revisión de contenido; no se les asigna costo cero.

## Cambios de tooling y validación

- `audit_work.py`: lector de evidencia existente por familias. Comprueba traza/prompt/entradas/snapshot, conserva estado original, distingue turno completo de aceptación e informa métricas ausentes. No ejecuta modelos ni código del candidato y nunca adopta el producto automáticamente.
- Corregida únicamente la sintaxis actual de arranque del sandbox en `work.py`; no hay selección de sintaxis legacy ni fallback al proceso host.
- Documentados alcance y límites en `FAMILIES.md`; actualizada la descripción desfasada de `WORK.md` después de la activación inline.
- `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools/codex-behavior-pilot -p 'test_*.py' -v`: **177 tests, PASS**. Incluye ocho pruebas nuevas del auditor: datos alterados, procedencia, intentos de escritura, ausencia de telemetría, symlinks, falta de turnos, texto convincente pero falso y ausencia de muestras. Estos tests calibran el observador; no son 177 tareas resueltas por agentes.
- `git diff --check origin/develop`: PASS en el checkout de implementación. Archivos nuevos comprobados por separado.
- Sandbox real: fuente sintética confiable y sonda de escritura read-only, PASS; después oráculo real del candidato, PASS. No se ejecutó código producido por el modelo directamente en el host.
- No cambió código Go ni el binario instalado. No se repitió la suite Go sin cambios que invaliden su evidencia del bloque 4.
- Memoria durable: 16 notas, cero backlinks rotos. Sin commit, push, PR ni release.

## Decisiones y lo que falta

1. La evidencia de contenido apoya que el principal pueda explicar, proponer, redactar, traducir, corregir y analizar datos sin una cadena obligatoria de roles. Es una muestra exploratoria por familia, no una garantía general.
2. Un clasificador parcial sirve como detector y registro del experimento. No se incorpora como gate universal al producto, no se amplía para conseguir verde y no reemplaza la comprobación del resultado.
3. El piloto no midió valor incremental de grafo, memoria ni delegación. No justifica activarlas para toda tarea ni retirarlas sólo porque estos fixtures no las necesitaron.
4. La comparación causal sigue **inconclusa**. Faltan pares A/B/C, segunda modificación de código/texto y costo humano medido. No se utiliza el ahorro de tokens para compensar esa ausencia.
5. App desde cero, nueva funcionalidad integrada, refactor sin cambiar contratos, seguridad, investigación contradictoria y operaciones/migraciones del corpus conceptual siguen sin una ejecución equivalente en este piloto. No se cuentan como aprobadas ni se eliminan de la cobertura pendiente.

La siguiente unidad está descrita en `bloque-5-protocolo-siguiente.md`: propuesta de seis fixtures × tres variantes, más dos seguimientos por variante (techo de 24 turnos). Requiere fijar bundles A/B, modelo/condiciones y telemetría antes de ejecutarla. La autorización de los cuatro casos no se amplió a esas invocaciones. El retiro general del bloque 6 no se declara completado por este avance.

## Evidencia y ubicación

Código en `/private/tmp/lufy-inline-20261026`, rama `codex/lufy-inline`; el checkout original mantiene sus modificaciones de código previas. El patch acumulado `bloques-1-a-5.patch` contiene fuentes nuevas y reverse-check, sin trazas privadas ni binarios.

- `bloque-5-audit.json`: nueve registros auditados, hashes y límites por familia.
- `bloque-5-semantic-review.json`: observaciones de contenido con hashes del registro/artefacto; no reemplaza estados automáticos.
- `bloque-5-data-check.py` y `bloque-5-data-check.json`: reproducción SQL del revisor.
- `bloque-5-validation.log`: suite offline final.
- `bloque-5-oracle-preflight.json`: sonda real exitosa con fuente sintética.
- Registros completos privados: `/private/tmp/lufy-inline-families-20261027-e`, `...-f`, `...-g`; auditoría/recuperación en `/private/tmp/lufy-families-block5-20261027`. No se publican automáticamente.

La captura JSONL, usage y salida final se contrastaron con [OpenAI Docs sobre ejecución no interactiva](https://learn.chatgpt.com/docs/non-interactive-mode) y el help de la CLI instalada. El [artículo oficial sobre evaluación de skills](https://developers.openai.com/blog/eval-skills) aporta el enfoque de comprobar comportamiento y permisos; sus ejemplos históricos con `--full-auto` no se copiaron: la referencia actual y el help usan `--sandbox workspace-write`.
