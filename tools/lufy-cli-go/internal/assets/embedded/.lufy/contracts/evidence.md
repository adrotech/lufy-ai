# Evidencia pertinente al resultado

El principal elige comprobaciones que puedan demostrar lo pedido. No hay una cantidad universal de escenarios, un score de aprobación ni un rol obligatorio. Una respuesta ordinaria puede describir resultado, observaciones y límites en prosa.

| Resultado | Qué comprobar |
| --- | --- |
| Código o nueva funcionalidad | Comportamiento solicitado, regresiones relevantes, límites y contratos afectados. TDD puede hacerlo el mismo ejecutor cuando ayude a descubrir o preservar el comportamiento. |
| Texto o traducción | Fidelidad al material, exactitud factual, coherencia editorial, restricciones y formato solicitado. Un build no demuestra estas propiedades. |
| Investigación o análisis | Fuentes pertinentes, contraste de afirmaciones, diferencias entre hechos e inferencias y límites de la evidencia. |
| Datos | Consultas reproducibles, procedencia, unidades, nulos, duplicados, agregaciones y efectos autorizados. |
| Operación o recuperación | Estado posterior observado y errores pendientes. Haber lanzado un comando o recomendado una recuperación no demuestra que funcionó. |
| Propuesta o diseño | Coherencia con la necesidad, alternativas y tradeoffs relevantes, factibilidad y preguntas abiertas. Terminar una propuesta no exige implementarla. |

Comprueba temprano lo que resuelve incertidumbre y agrupa suites completas cuando corresponda. Cambiar el candidato, la fuente o una regla relevante invalida las comprobaciones afectadas. Un fallo crítico no se compensa con otras comprobaciones aprobadas.

La revisión independiente aporta otra perspectiva cuando el riesgo o el usuario la requieran. No convierte TDD en un handoff ni autoriza delegación por sí sola. Respeta `execution.delegation`, las instrucciones del usuario y las capacidades reales del host.

## Integraciones opcionales

Los comandos `evidence inspect`, `candidate` y `review` aceptan un archivo JSON de comprobaciones declaradas. Este archivo sirve a esas integraciones; no es un formato obligatorio para responder ni una nueva máquina de estados del trabajo.

```json
{
  "checks": [
    {"name": "Regresión de la CLI", "result": "passed", "evidence": "go test ./internal/cli: exit 0"},
    {"name": "Revisión visual", "result": "not_applicable", "reason": "No cambió una interfaz visual"}
  ],
  "findings": []
}
```

Cada comprobación identifica una observación (`passed`, `failed` o `inconclusive`) o explica por qué no aplica (`not_applicable`). Los hallazgos tienen severidad `critical`, `high`, `medium` o `low`. No se aceptan scores, estados de workflow ni claims de un rol.

La inspección deriva una conclusión de lo declarado: `supported`, `failed`, `inconclusive` o `not_applicable`. Su exit 0 indica que pudo interpretar el archivo; no ejecuta pruebas ni confirma la veracidad de las afirmaciones. El agente conserva la responsabilidad de observar las fuentes y resultados reales. Una comprobación no aplicable no se presenta como aprobada y un resultado inconcluso no se transforma en éxito.

`candidate capture/verify` vincula el fingerprint de observaciones y hallazgos a identidad del repositorio, base, commit/tree del candidato y políticas vigentes. Mantiene las restricciones de lectura Git sin efectos laterales. Una vinculación vigente puede contener una comprobación fallida: vigencia y calidad son propiedades distintas.

`review` registra sólo metadata y fingerprints, con idempotencia de operaciones explícitas. Rechaza `--result passed` si hay fallos, incertidumbre, ausencia de comprobaciones aplicables o hallazgos críticos/altos. No guarda el texto de la evidencia, no avanza gates ni concede permiso para entregar.
