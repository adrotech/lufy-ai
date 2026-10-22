# Coherencia de decisiones del workflow proporcional

## Why

Lufy ya ofrece T1 Full, T2 Lite y T3 Express. El problema no es la ausencia de proporcionalidad, sino que riesgo, volumen de archivos, carga de review y condiciones de pausa se expresan en varias superficies con funciones y precedencias poco explícitas. El análisis comparativo con ODD identificó hipótesis de fricción, no ahorros demostrados en Lufy.

El usuario pidió comenzar las mejoras después de revisar el informe `pr_review/lufy-odd-comparativa-mejoras.html`, basado en `Gentleman-Programming/gentle-ai` (snapshot `4acaf4241beae1e0c977c92dfe406a270f8d40de`). Este change aborda únicamente su bloque 1: coherencia del contrato proporcional. No adopta ODD ni modifica permisos. La corrección de alcance del 2026-10-21 limita las mejoras a Lufy en Codex; se descartan mejoras OpenCode y la evaluación/migración a Pi.

## What Changes

- Separar riesgo/incertidumbre/impacto, que determinan tier y controles, de volumen, que informa workload y review slices.
- Explicar las funciones de los checkpoints actuales de cuatro archivos significativos y cinco archivos estimados sin sustituirlos por un nuevo número.
- Definir archivo significativo y cómo tratar fuentes gestionadas y mirrors: evitar duplicación conceptual sin ocultar el diff real al reviewer.
- Documentar precedencia entre invariantes de seguridad/autorización, configuración canónica `workflow_limits` y checkpoints del harness.
- Hacer que un plan/slice existente y pertinente cubra un checkpoint de amplitud, sin crear aprobaciones repetidas ni saltar gates independientes.
- Sincronizar contrato neutral, agentes Codex, documentación compartida pertinente y distribución embebida de esas superficies.
- Complementar pruebas de presencia/paridad con una matriz explícita de decisiones esperadas. La evaluación de agentes reales queda para el bloque siguiente.

## Capabilities

### Modified Capabilities

- `sdd-harness-routing`: distinguir clasificación por riesgo de decisiones por volumen y evitar escalado de tier por recuento aislado.
- `systemic-workflow`: precisar checkpoints de amplitud, cobertura por plan y precedencia observable de límites.

## Alcance autorizado y no objetivos

Incluye implementación aprobada y validación del bloque 1 para Codex. La solicitud de comenzar no autoriza commit, push, PR, publicación, cambios destructivos ni modificación de permisos. No incluye retirar adapters, cambiar el tool default ni desinstalar programas. Las mejoras específicas de OpenCode quedan fuera; los archivos modificados para ese adapter en esta sesión se restauraron al estado previo. Los roles YAML neutrales y su golden OpenCode también quedan sin cambios.

No se cambiarán:

- permisos de `orchestrator`, `implementer`, `validator`, `reviewer` o `delivery`;
- aprobación post-plan de T1/T2 sin fast path;
- controles de seguridad, contratos públicos, persistencia, release o delivery;
- schemas públicos, flags, defaults numéricos de configuración o contratos Result Contract;
- topología TDD/test-writer, arquitectura de memoria o productores de métricas;
- tooling/configuración del usuario mediante init, build del grafo o inicialización de memoria;
- specs activas mediante sync ni archivos de otros changes.

## Aceptación

- AC1: WHEN un cambio es mecánico, entendido y sin riesgo relevante, THEN el recuento de archivos por sí solo no aumenta el tier; los checkpoints de workload siguen evaluándose.
- AC2: WHEN una modificación de una línea afecta auth/configuración sensible, THEN sigue requiriendo router, al menos T2 y fast path deshabilitado según las reglas vigentes.
- AC3: WHEN un bloque alcanza cuatro archivos significativos, THEN se verifica cobertura por un plan/slice vigente; si falta o el alcance se amplió, se pausa en la frontera segura.
- AC4: WHEN el router estima cinco o más archivos, THEN emite una decisión de workload/slicing justificada; no aumenta tier sólo por esa cifra ni fragmenta T3 artificialmente.
- AC5: WHEN la configuración falta, THEN se reporta `not_available`, no se inventan budgets y siguen aplicando invariantes/checkpoints declarados del harness.
- AC6: WHEN un plan cubre la amplitud, THEN ello no sustituye aprobación post-plan, seguridad, validación, delivery ni una stop rule independiente.
- AC7: WHEN cambian instrucciones distribuidas de Codex o superficies compartidas pertinentes, THEN fuentes y mirrors conservan paridad y los agentes Codex expresan decisiones equivalentes al contrato neutral.

## Review slices

| Slice | Objetivo | Archivos esperados | Aceptación | Validación | Riesgo / guía |
| --- | --- | --- | --- | --- | --- |
| S1 | Contrato y matriz de decisiones | `.lufy/contracts/delivery.md`, `docs/harness-workflow.md`, pruebas dedicadas de assets, mirrors correspondientes | AC1–AC7 | Revisión de escenarios; tests de assets | Cambios normativos; revisar precedencia antes de tocar agentes |
| S2 | Propagar a Codex y superficies compartidas pertinentes | `.codex/agents/{orchestrator,sdd-router,implementer}.toml`, `AGENTS.md.template`, `AGENTS.md`, `lufy-ia.harness.md`, mirrors exactos | AC1–AC7 | Tests de contratos/paridad entre contrato y Codex | No modificar OpenCode; mantener mismo PR sólo si delivery se autoriza |
| S3 | Validación agrupada y evidencia | Pruebas pertinentes, tareas y evidencia de este change | AC1–AC7 | `scripts/validate.sh` cuando aplique; build local; validación OpenSpec estricta | No confundir tests textuales con conducta probada |

Los slices son secuenciales y forman un único contrato coherente. No hay paralelismo: S2 consume S1 y las superficies comparten reglas. Mirrors no son unidades independientes de ejecución.

## Impact

Afecta instrucciones distribuidas y su documentación, no el comportamiento del binario clasificador: la función actual de routing valida selecciones de metodología y no es un clasificador completo de solicitudes.

La matriz y las pruebas estáticas no demostrarán por sí solas cumplimiento de un LLM, mejora de latencia o ahorro de tokens. Esas mediciones requieren el piloto del bloque siguiente.

## Estado inicial y límites de evidencia

- `project.yaml` ausente: `workflow_limits` y selección efectiva por tier no disponibles. Se usa la superficie OpenSpec ya declarada por `openspec/config.yaml`, sin crear configuración ni reinterpretar defaults.
- Memoria Obsidian no inicializada y Context Graph `not_available`: continuar con evidencia del repo, sin inicializarlos.
- `openspec` ausente en PATH: validación estricta pendiente. No usar el renderer HTML ni una comprobación textual como sustituto.
- `tools/lufy-cli-go/dist/` preexistente sin rastrear: preservar.
- No hay tooling de subagentes expuesto en esta sesión: preparación documental inline, sin afirmar delegación ni revisión independiente. Si apply requiere delegación, resolver ese requisito antes de ejecutar, no simular roles.

## Estado actual y siguiente gate

El usuario aprobó continuar con S1 y autorizó instalar OpenSpec y generar el overview HTML. Se instaló `@fission-ai/openspec@1.14.3`; `openspec validate clarify-proportional-workflow-decisions --strict` reportó el change válido después de conservar los nombres de escenarios existentes en MODIFIED.

El bloqueo inicial del CLI y el de delegación están resueltos. El 2026-10-21 se reanudó S1 con herramientas de subagentes: `test-writer` preparó las regresiones y capturó RED antes de modificar fuentes. Se actualizaron contrato neutral, documentación y mirror exacto. Las pruebas de assets pasaron tras el join; S1 queda `validated` con evidencia en `tasks.md`.

S2 se adaptó al foco exclusivo en Codex el 2026-10-21. Se retiraron las ediciones de OpenCode, YAML neutrales y golden OpenCode realizadas durante la propagación inicial; se ajustaron las regresiones al alcance final. Contrato neutral, Codex, templates, harness y mirrors pertinentes conservan las mejoras. S3 confirmó nuevamente el conjunto: `scripts/validate.sh` exit 0, coverage 80.4% >= 80.0%, vet y build correctos. La comprobación de diff de las superficies OpenCode/neutrales retiradas pasó sin cambios residuales. Estado: `validated`; shell lint omitido por falta de `shellcheck`, sin cambios shell.

Siguiente bloque del plan comparativo: escenarios conductuales reales de Codex; después observabilidad y, sólo con baseline, optimización experimental. Sin sync, delivery ni archive por este change.
