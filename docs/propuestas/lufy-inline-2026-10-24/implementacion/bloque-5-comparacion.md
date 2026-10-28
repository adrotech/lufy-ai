# Bloque 5: comparación exploratoria y segunda modificación

**El ensayo acotado terminó: 24 turnos CLI nuevos, 18 entregas iniciales y 6 seguimientos. La conclusión comparativa es inconclusa para una ventaja general o ahorro de costo humano.** No se repitieron muestras para obtener verde ni se cambiaron sus veredictos.

La decisión operativa es continuar con el diseño inline único que el usuario autorizó. La evidencia funcional permite seguir su evolución, pero no convierte esa elección de producto en una superioridad demostrada por benchmark. Este estudio termina con sus límites registrados; no se prolonga hasta encontrar un ganador. El retiro interno del camino anterior corresponde al bloque 6.

## Qué se comparó

- **A:** instalación estándar Codex del snapshot `1cc18d4647afc61ea1f6b1cb726ba74128fded5f`, compilado desde Git en un directorio externo: **86 assets**. No es una copia de la instalación personal antigua de 108 assets ni una variante sobrecargada con todas las referencias.
- **B:** principal inline con instrucciones esenciales y el mismo alcance, sin assets de producto.
- **C:** instalación inline actual: **16 assets**, delegación off y capacidades disponibles bajo demanda.

Las diferencias de CLI/assets son parte del tratamiento. No se atribuye un efecto a una instrucción o capacidad aislada. El baseline se guarda como evidencia del estudio; no se incorpora un motor ni un lector legacy a la instalación actual.

Se fijaron antes de los modelos los seis fixtures, fuentes, binarios, orden, aceptación y recursos. CLI de app `0.164.0-alpha.11.1`, SHA-256 `50ab38ba21d0d9f8346f32f41848382f15b556190f3c7a07e885a4fb73e379c8`; modelo solicitado **gpt-6.1-sol**, coincidente con la preferencia local. Identidad servida no atestada.

Mismo pedido, datos, permisos y criterios por familia; orden inicial ABC/BCA/CAB alternado. Un contexto nuevo por celda, hooks/apps/plugins/delegación desactivados sólo para la invocación, máximo 60 segundos por turno, sin publicación. Esto evalúa al principal bajo esas restricciones comunes; no evalúa el SDD histórico con delegación nativa ni paralelismo.

La precisión de parada se congeló antes de ejecutar: un fallo común de proceso/sandbox detendría el grupo; un timeout sería una celda incompleta conservada sin retries, sin eliminar otras variantes del denominador. Ninguna celda agotó el límite, por lo que esa distinción no cambió los resultados.

## Resultados de contenido y continuidad

| Resultado | A | B | C |
| --- | --- | --- | --- |
| Explicar precio y resultado para 100 | Correcto | Correcto | Correcto |
| Proponer exportación sin implementar | Correcto | Correcto | Correcto |
| Redactar guía y conservar referencia/comando | Correcto | Correcto | Correcto |
| Traducir documento y conservar comando | Rótulo `Example:` en inglés | Traducción completa | Rótulo `Example:` en inglés |
| Corregir filtro string/None | Oráculo passed | Oráculo passed | Oráculo passed |
| Diagnosticar multiplicación del join | Correcto | Correcto | Correcto |
| Ampliar filtro a colecciones, preservando contrato | Oráculo passed | Oráculo passed | Oráculo passed |
| Cambiar nombre de archivo en guía y conservar resto | Correcto | Correcto | Correcto |

La revisión posterior del principal registra **22 observaciones de contenido sustentadas y 2 de localización parcial**. No son 22 aprobaciones automáticas: contenido, observación de efectos y finalización se informan por separado.

En traducción, A y C conservaron el comando y el significado, pero dejaron `Example:` sin traducir. El criterio de traducir todo el texto no queda satisfecho literalmente. A la vez, «conservá el ejemplo literalmente» introduce ambigüedad sobre ese rótulo. Se registra la diferencia y la limitación del caso, sin reinterpretar la salida como éxito completo ni convertir una muestra en una regresión general de método. Una futura revisión del corpus debe precisar que sólo el comando entre backticks conserva su literalidad.

C leyó `lufy-check` en ese caso y aun así validó su propio texto esperado con el rótulo en inglés. **Invocar una skill o afirmar que se comprobó el texto no demuestra fidelidad completa.** La revisión debe contrastar contenido y criterios, no repetir la salida elegida como resultado esperado.

Los seguimientos partieron de una copia de la entrega completa de cada variante. Sólo se ajustó `targetRoot` del manifest a la copia. El parent quedó intacto. El filtro ampliado conservó los tests previos y pasó un oráculo externo para strings, None, list/tuple/set, colecciones vacías, orden y datos. Su baseline previo falló ante la nueva exigencia y luego pasó. En las tres guías, el resultado coincide exactamente con reemplazar `movimientos.csv` por `gastos.csv`, con la referencia intacta.

Son cambios sobre artefactos entregados en sesiones nuevas: observan mantenibilidad y recuperación de contexto desde archivos, no continuidad de una conversación ni ventaja de una memoria persistente.

## Controles y estados automáticos

- 24 turnos completos, 24 thread IDs diferentes, exit 0, cero timeouts y cero retries.
- Hashes de traza/registro vinculados, entradas/fuentes estables y snapshots posteriores sin cambios tardíos. Binarios A/C y runtime conservaron sus hashes.
- Cero cambios de archivos fuera de los paths declarados observados en los fixtures. No se observó delivery ni una violación crítica en las trazas revisadas. Esto no atesta todos los efectos del host ni reemplaza una auditoría de seguridad.
- El clasificador dejó **20 celdas inconclusas**, **3 needs_review** y **1 passed**. Hubo 33 comandos no clasificables por su gramática limitada: verificaciones Python, Git, hashes, heredocs y lecturas condicionales. Sus estados no se convirtieron en pass por la revisión de contenido.
- Los seis oráculos del filtro —tres iniciales y tres seguimientos— pasaron. El oráculo externo ejecuta el candidato mediante sandbox read-only después de observar denegación de escritura. No hay fallback al host.

La evaluación automática incompleta no equivale a una salida incorrecta, y la salida correcta no demuestra por sí sola todos los controles. Ambos límites permanecen visibles en los registros.

## Tiempo y uso observados

Suma descriptiva de las ocho celdas por variante —seis iniciales y dos seguimientos—:

| Variante | Tiempo local de celdas (s) | Input tokens | Cached input tokens | Output tokens |
| --- | --- | --- | --- | --- |
| A | 198.362 | 482599 | 401792 | 4858 |
| B | 150.158 | 367830 | 314240 | 3857 |
| C | 155.863 | 390345 | 336640 | 3933 |

El tiempo usa reloj monotónico e incluye snapshots y oráculos de la celda; no es latencia pura del modelo ni incluye compilación/preparación del estudio. La preparación se registra aparte. Cached input forma parte del uso reportado. Estos tokens no son precios ni una estimación monetaria.

B y C tuvieron tiempos totales cercanos en esta muestra. C usó menos input tokens en las entregas iniciales, pero más al sumar los seguimientos. **No aparece un ahorro general frente a B.** A consumió más tiempo y tokens en esta muestra, sin que ese dato establezca una ventaja causal general: hay una sola observación por celda, cachés y variación del servicio.

Costo humano, tiempo de revisión de una persona y precio monetario: **no disponibles**, conservados como null. La revisión del agente principal no se etiqueta como revisión humana ni se valora como cero. No se utilizan diferencias de tokens para compensar estas ausencias o las observaciones editoriales.

### Datos por celda

| Etapa / caso | Variante | Tiempo (s) | Input tokens | Estado automático original |
| --- | --- | --- | --- | --- |
| initial / explain-code | A | 18.584 | 56724 | inconclusive |
| initial / explain-code | B | 15.915 | 51199 | inconclusive |
| initial / explain-code | C | 12.857 | 43185 | needs_review |
| initial / propose-export | B | 24.959 | 38349 | needs_review |
| initial / propose-export | C | 25.292 | 28082 | needs_review |
| initial / propose-export | A | 40.828 | 80234 | inconclusive |
| initial / write-guide | C | 18.333 | 43805 | inconclusive |
| initial / write-guide | A | 26.457 | 90021 | inconclusive |
| initial / write-guide | B | 23.177 | 51922 | inconclusive |
| initial / translate-guide | A | 16.155 | 42046 | inconclusive |
| initial / translate-guide | B | 17.148 | 55121 | inconclusive |
| initial / translate-guide | C | 21.258 | 60028 | inconclusive |
| initial / fix-filter | B | 13.609 | 38991 | inconclusive |
| initial / fix-filter | C | 15.307 | 45052 | passed |
| initial / fix-filter | A | 27.577 | 42988 | inconclusive |
| initial / analyze-join | C | 25.807 | 61740 | inconclusive |
| initial / analyze-join | A | 25.915 | 59910 | inconclusive |
| initial / analyze-join | B | 20.368 | 51824 | inconclusive |
| followup / fix-filter | A | 27.107 | 68495 | inconclusive |
| followup / fix-filter | B | 16.727 | 40787 | inconclusive |
| followup / fix-filter | C | 18.925 | 47657 | inconclusive |
| followup / write-guide | A | 15.739 | 42181 | inconclusive |
| followup / write-guide | B | 18.255 | 39637 | inconclusive |
| followup / write-guide | C | 18.084 | 60796 | inconclusive |

## Decisiones para la evolución

1. Seguir con principal inline, permisos explícitos y comprobaciones por resultado. Los casos observados no necesitaron una cadena obligatoria de roles ni nuevos artefactos metodológicos. No se exige otra aprobación por una etapa interna.
2. La evidencia de estos fixtures no justifica obligar a cargar capacidades adicionales en toda tarea. Se mantienen bajo demanda; no se atribuye aquí beneficio a grafo, memoria o delegación, que no se compararon con/sin su uso en tareas pertinentes.
3. Conservar revisión factual/editorial y oráculos discriminantes. Una autoverificación que repite lo escrito puede omitir un requisito, como mostró el rótulo de la traducción.
4. El clasificador parcial queda como herramienta del experimento, no como nuevo gate universal del producto. No se amplió para blanquear las muestras.
5. La adopción **por comparación empírica** no queda acreditada. La dirección inline autorizada y la reducción de estructura se distinguen de esa hipótesis. No se afirma que Lufy o una variante sea mejor en general.

App desde cero, funcionalidades integradas, refactor de contratos, seguridad, operaciones/migraciones y fuentes contradictorias del corpus conceptual siguen sin una evaluación equivalente. Las seis familias pequeñas no cubren esos trabajos. Se registran como evidencia pendiente para esos contextos; no se cuentan como aprobados.

## Implementación y validación

- `comparison.py`: preparación opt-in de A/B/C, snapshots previos, orden fijo, seguimientos ligados al hash del parent y conservación de fallos/not_run. Sin `--live`, prepara fixtures externos sin invocar modelos. Nunca adopta ni entrega el producto automáticamente.
- `work.py`: soporte para fixtures ya preparados, reloj monotónico, modelo explícito y oráculo ampliado; conserva la interfaz del piloto. Esta reutilización interna no añade lectura de schemas anteriores al producto.
- Validación offline agrupada previa a la campaña: **182 tests, PASS**, incluida calibración del contrato ampliado, fuentes sintéticas incorrectas/correctas, protección de resultados existentes y preparación de 18 celdas sin modelos.
- Integración real: 24 turnos, 6 oráculos funcionales passed, 6 seguimientos observados, 24 snapshots estables y padres conservados. El código capturado no cambió durante la campaña.
- No cambió código Go actual ni el binario personal. No se repitió la suite Go sin cambios que invaliden su evidencia. Se comprueba whitespace y paridad de copias documentales al cerrar.
- Memoria durable: 17 notas, cero backlinks rotos. Sin commit, push, PR, release ni cambios de configuración global.

## Evidencia durable y fuentes

Archivo privado, fuera del repositorio, con directorio 0700 y archivos 0600:

`/Users/adrianrojas/.local/share/lufy-ai/experiments/comparison-20261028/evidence.tar.gz`

SHA-256: `e348511320fe3fa0633c41d3e9d9278839d7d1532092b5a7fa688aeb64484db3` (15692835 bytes). Contiene las trazas, fixtures, planes congelados, revisiones, binarios del producto y fuentes del evaluador/baseline. No incluye autenticación. Su repetición depende del runtime y entorno originales; no se promete portabilidad completa ni identidad del modelo servido. Los datos completos no se publican automáticamente.

En este directorio de documentación se conservan `bloque-5-comparacion-metrics.json`, `bloque-5-comparacion-review.json`, `bloque-5-comparacion-protocol.md`, `bloque-5-comparacion-validation.log` y `bloque-5-comparacion-package.json`. El protocolo y las observaciones posteriores se guardan separados.

Código: `/private/tmp/lufy-inline-20261026`, rama `codex/lufy-inline`. El checkout original conserva sus cambios de código anteriores. `bloques-1-a-5-comparacion.patch` respalda la implementación acumulada con reverse-check; el patch anterior y los informes del piloto permanecen como historia.

La captura de JSONL/usage se apoya en [OpenAI Docs de ejecución no interactiva](https://learn.chatgpt.com/docs/non-interactive-mode) y las opciones se contrastaron con el help del ejecutable observado. Las cifras y resultados de este informe provienen de las ejecuciones locales, no de esa documentación.
