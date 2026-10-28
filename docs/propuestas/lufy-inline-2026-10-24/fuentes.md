# Fuentes y límites de la propuesta

Fecha de síntesis: 24 de octubre de 2026. Parte de la [propuesta Lufy inline](README.md).

Se distinguen evidencia del producto, propuestas de autores, estudios con alcance limitado y experiencias de comunidad. Las decisiones de arquitectura de esta propuesta son inferencias y recomendaciones para Lufy; las fuentes no demuestran que ya funcionen aquí.

Restricción de producto corregida por su autor: Lufy sólo lo usa él y no necesita retrocompatibilidad. Esta instrucción directa elimina la necesidad de lectores antiguos, bridges, deprecación o coexistencia de motores; tiene precedencia sobre recomendaciones genéricas de migración de las fuentes.

## Evidencia local

La referencia del producto es `origin/develop` en `1cc18d4647afc61ea1f6b1cb726ba74128fded5f`, asociado a `v0.6.25`. Se leyeron archivos mediante `git show`/`git grep` sobre esa referencia, preservando el checkout local atrasado con cambios existentes.

| Evidencia | Qué fundamenta | Fuente fijada |
| --- | --- | --- |
| Modelo y defaults del harness | El acoplamiento actual a metodología por tier es de dominio, no sólo de prompt. | [harness.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/core/domain/harness.go) |
| Política de routing | T1 sin metodología se rechaza; T2 sin metodología requiere condiciones específicas. | [routing.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/core/domain/routing.go) |
| Catálogo, flags y configuración | Cambiar instrucciones solamente no retira selección metodológica de instalación/configuración. | [catalog.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/harnesscatalog/catalog.go), [harness_flags.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/cli/harness_flags.go), [model.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/projectconfig/model.go) |
| Estado instalado | El schema actual de install-state es 2. La propuesta puede reemplazarlo directamente y regenerar el estado de la única instalación, sin mantener su lector. | [state.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/state/state.go) |
| Plan de superficies | `lufy-ai plan` ya tiene una función concreta read-only que debe preservarse. | [app_plan.go](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/tools/lufy-cli-go/internal/cli/app_plan.go) |
| Normas de TDD y review | Las obligaciones de delegación, score y escenarios existen y deben retirarse expresamente del camino nuevo. | [systemic-workflow](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/openspec/specs/systemic-workflow/spec.md), [scored-stack-aware-reviewer](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/openspec/specs/scored-stack-aware-reviewer/spec.md) |
| Ensayo inline/referencia anterior | Par incompleto por timeout; no demuestra superioridad del baseline ni del candidato. Recuperación sintética e idempotencia no equivalen a conducta general de modelos. | [Informe del 22/10](https://github.com/adrotech/lufy-ai/blob/1cc18d4647afc61ea1f6b1cb726ba74128fded5f/docs/codex-execution-experiments-2026-10-22.md) |

El contexto de la investigación previa encontró grafo stale y ausencia de superficies declaradas para `plan`. Se utilizó lectura directa como fallback; esa disponibilidad local no significa que el producto carezca de esas capacidades. La orientación del usuario sobre no conservar por inercia, integrar prácticas y reconocer trabajos distintos prevalece sobre recomendaciones históricas del baseline.

Las autorizaciones narradas dentro de informes antiguos describen sus sesiones originales. No se transfieren a esta propuesta como permiso actual para implementar o entregar.

## Video y experiencias de desarrollo

### Video de HolaMundo

[«Me equivoqué, NO uses Spec Driven Development»](https://www.youtube.com/watch?v=5JjuF1zpNbI) y [repositorio hola-ai](https://github.com/nschurmann/hola-ai). En la investigación precedente se revisó la transcripción automática completa en español del video de 37:38. En la elaboración de esta propuesta, el nuevo intento de acceso web al video no permitió recuperarlo; no se presenta como una segunda visualización completa.

La experiencia sirve para cuestionar documentos anticipados y priorizar feedback sobre el software. Es testimonio del autor, no comparación controlada ni demostración de que toda especificación sea perjudicial. La propuesta conserva acuerdos y pruebas cuando resuelven incertidumbre real, y elimina el requisito universal de producir un paquete documental.

### gentle-ai / ODD / RDD

Referencia estable revisada: [gentle-ai v4.2.0](https://github.com/Gentleman-Programming/gentle-ai/releases/tag/v4.2.0), publicada el 15/10/2026; commit `4acaf4241beae1e0c977c92dfe406a270f8d40de`.

[Uso previsto](https://github.com/Gentleman-Programming/gentle-ai/blob/v4.2.0/docs/intended-usage.md) presenta ejecución inline, autorización separada del descubrimiento, registro para trabajo sustancial y delegación con motivo. [Arquitectura de review](https://github.com/Gentleman-Programming/gentle-ai/blob/v4.2.0/docs/architecture/organic-rdd.md) trata revisión de candidatos por separado.

Tomamos la simplificación y la separación de responsabilidades como insumos. No adoptamos automáticamente su archivo obligatorio para trabajo sustancial, duplicación con memoria, límites de tamaño, reglas exactas de tests ni semántica de commits. Tampoco sus benchmarks locales acreditan beneficios en Codex/Lufy.

La [historia de RDD](https://github.com/Gentleman-Programming/gentle-ai/blob/cd19f274e2abcb7c824edcaea531f13a6d849394/docs/architecture/the-organic-rdd-story.md), fijada al snapshot de main inspeccionado, relata bloqueos y pruebas de recuperación insuficientes. Fundamenta probar recuperación efectiva e identificar el binario medido; es un relato del propio proyecto. No se mezclan main y release como si tuvieran el mismo contrato.

### Otras prácticas para combinar

| Fuente primaria | Insumo para Lufy | Límite |
| --- | --- | --- |
| [Kent Beck: Augmented Coding](https://newsletter.kentbeck.com/p/augmented-coding-beyond-the-vibes) | Incrementos, responsabilidad por calidad y feedback del comportamiento. | No demuestra una única cadencia correcta de intervención humana. |
| [Matt Pocock: skills y flujo de trabajo](https://www.aihero.dev/skills-post) | Acuerdos de dominio, unidades verticales y ejemplos comprobables cuando aportan. | PRD/TDD no se convierten en rutas de toda investigación o redacción. |
| [Superpowers](https://github.com/obra/superpowers) | Clarificación, disciplina de pruebas y revisión. | No importar su pipeline completo como default. |
| [GSD Core](https://github.com/open-gsd/gsd-core) | Continuidad y contexto acotado para trabajo extenso. | Mantener estado también tiene costo; no hace falta para cada pedido. |
| [Harness Engineering, Birgitta Böckeler](https://martinfowler.com/articles/harness-engineering.html) | Separar instrucciones que orientan y comprobaciones que observan. | El tamaño del harness no demuestra su utilidad. |

## Estudios y evaluación

| Fuente | Hallazgo relevante | Qué no permite concluir |
| --- | --- | --- |
| [Are Harnesses Worth the Cost?, Naresh Jain, 13/10/2026](https://www.martinfowler.com/articles/are-harnesses-worth-the-cost.html) | En el experimento, el modelo frontier pasó con y sin harness; el modelo de menor costo mejoró con él. El costo extra depende del modelo y del caso. | Una API greenfield con brief detallado y diez repeticiones por condición no demuestra el efecto en todas las familias ni incluye todo el mantenimiento humano de Lufy. |
| [Evaluating AGENTS.md, v3](https://arxiv.org/abs/2602.11988v3) | No encuentra mejora general consistente por context files; observa costo adicional y utilidad de instrucciones específicas. | No justifica eliminar restricciones operativas ni atribuir su overhead exacto a Lufy. |
| [SlopCodeBench, v2](https://arxiv.org/abs/2603.24755v2) | Examina extensiones sucesivas: conviene medir evolución del resultado y mantenibilidad, no sólo el primer pass. | Sus resultados no son una tasa de defectos de este proyecto. Se usan 36 problemas y 196 checkpoints de v2, no las cifras de v1. |
| [METR, actualización del experimento, 24/02/2026](https://metr.org/blog/2026-02-24-uplift-update/) | Describe sesgos de selección y dificultades de medir herramientas actuales; obliga a cuidar muestra y trabajo humano. | No repetir el enlentecimiento del estudio de 2025 como conclusión universal vigente. |
| [Towards a Science of Scaling Agent Systems, v3](https://arxiv.org/abs/2512.08296v3) | Coordinación y rendimiento dependen de las tareas y capacidades; sustenta pedir una razón para delegar. | No prescribe una topología óptima para Lufy. Esta versión analiza 260 configuraciones y seis benchmarks. |
| [Anthropic: Demystifying evals for AI agents](https://www.anthropic.com/research/demystifying-evals-for-ai-agents) | Distingue evaluadores programáticos, de modelo y humanos; enfatiza observar resultados y calibrar el evaluador. | No basta aprobar un transcript convincente o confiar siempre en un juez LLM. |
| [Anthropic: Building effective agents](https://www.anthropic.com/research/building-effective-agents) | Componer soluciones simples según la necesidad. | No es evidencia de las herramientas o modelos exactos disponibles hoy. |
| [Anthropic: Multi-agent research system](https://www.anthropic.com/research/multi-agent-research-system) | Búsquedas amplias e independientes pueden beneficiarse de especialización. | Una experiencia de investigación de 2025 no prueba ventajas de delegación en código acoplado. |

## Trabajos que requieren criterios diferentes

| Fuente | Aplicación propuesta |
| --- | --- |
| [Design Council: Framework for Innovation](https://www.designcouncil.org.uk/resources/framework-for-innovation/) | Alternar exploración de posibilidades y convergencia cuando el problema lo necesita; sin imponer cuatro etapas a cada solicitud. |
| [GOV.UK: Discovery](https://www.gov.uk/service-manual/agile-delivery/how-the-discovery-phase-works) | Investigar puede terminar en una decisión de no construir. |
| [Cynefin Domains](https://cynefin.io/wiki/Cynefin_Domains) | Reconocer incertidumbre y contexto antes de elegir una intervención; no usarlo como clasificador automático de permisos. |
| [Refactoring, Martin Fowler](https://refactoring.com/) | Un refactor debe preservar comportamiento observable. |
| [Home Office: Design and content](https://design.homeoffice.gov.uk/design-and-content/content) | Evaluar escritura por propósito, audiencia, claridad y exactitud. |
| [NIST: Exploratory Data Analysis](https://www.itl.nist.gov/div898/handbook/eda/section1/eda11.htm) | Analizar datos exige hipótesis, supuestos y comprensión de la estructura. |
| [PostgreSQL 18: EXPLAIN](https://www.postgresql.org/docs/18/sql-explain.html) | `ANALYZE` ejecuta la sentencia: una herramienta de diagnóstico puede tener efectos que deben evaluarse. |

## Foros y comunidad

Se consultaron [OpenSpec Discussions #1159](https://github.com/Fission-AI/OpenSpec/discussions/1159), experiencias en [r/SpecDrivenDevelopment](https://www.reddit.com/r/SpecDrivenDevelopment/comments/1vubvmz/we_tried_specdriven_development_for_months_we/), debate en [r/ExperiencedDevs](https://www.reddit.com/r/ExperiencedDevs/comments/1xe74cu/was_spec_driven_development_a_failure/) y [Hacker News sobre AGENTS.md](https://news.ycombinator.com/item?id=47034087).

Son señales cualitativas de fricción, documentación, trazabilidad y diferencias entre equipos. No constituyen una muestra representativa ni consenso sobre reemplazar SDD. Sirven para construir casos que valga la pena comprobar, no para adjudicar tasas de éxito.

## Conclusión que la evidencia permite sostener

Hay fundamento para diseñar y evaluar un Lufy más pequeño, inline y adaptable al resultado de cada trabajo. No hay fundamento para prometer superioridad universal, ahorro cuantificado o eliminación del riesgo por cambiar de metodología. La recomendación concreta surge de combinar el acoplamiento observado en Lufy, las necesidades del usuario y estas prácticas; debe contrastarse con la aceptación y la migración propuestas.
