# Tasks: coherencia del workflow proporcional

## Estado y alcance

Programa T1. Este change implementa sólo el bloque 1 del informe comparativo con `Gentleman-Programming/gentle-ai`, centrado exclusivamente en Lufy para Codex. No incluye mejoras OpenCode, evaluación/migración Pi, benchmark, productores de métricas, permisos inline, cambios de aprobación ni delivery.

## 0. Propuesta y precondiciones

- [x] 0.1 Analizar specs existentes, contrato, roles, adapters, mirrors y pruebas pertinentes.
- [x] 0.2 Preparar proposal, design, deltas y plan de slices con aceptación AC1–AC7.
- [x] 0.3 Ejecutar `openspec validate clarify-proportional-workflow-decisions --strict`; OpenSpec 1.14.3 reportó el change válido tras preservar los nombres de escenarios existentes en los bloques MODIFIED.
- [x] 0.4 Confirmar el plan visible de S1 y el mecanismo de ejecución/delegación antes de apply. Se conserva la aprobación registrada; el 2026-10-21 se resolvió el bloqueo mediante un subagente `test-writer` con ownership exclusivo de las pruebas S1 y RED previo a modificar fuentes.

## 1. S1 — Contrato y matriz de decisiones

Objetivo: semántica común sin cambiar permisos, defaults, schemas o gates. Aceptación: AC1–AC7.

- [x] 1.1 Definir fuentes de autoridad, archivo significativo, checkpoint de amplitud y distinción riesgo/workload en el contrato canónico.
- [x] 1.2 Mantener matriz revisable con resultados permitidos/prohibidos y explicar el alcance textual de las pruebas.
- [x] 1.3 Preparar regresiones de contrato/paridad y evidencia RED cuando corresponda; resolver test-writer o registrar limitación aplicable antes de trabajo TDD sustantivo.
- [x] 1.4 Actualizar documentación sistémica y mirrors exactos de las fuentes modificadas.

## 2. S2 — Propagación a consumidores

Depende de S1. Objetivo: coherencia entre contrato neutral, Codex y distribución pertinente. Aceptación: AC1–AC7.

- [x] 2.1 Alinear orchestrator, sdd-router e implementer Codex preservando permisos, mapping, fallback y aprobación post-plan.
- [x] 2.2 Actualizar sólo reglas afectadas de AGENTS/template/harness y mirrors; no modificar otras convenciones.
- [x] 2.3 Ajustar regresiones AC1–AC7 a Codex y superficies compartidas pertinentes, con paridad y protección contra escalado por recuento aislado.
- [x] 2.4 Retirar las ediciones específicas OpenCode y neutrales de esta sesión; comprobar diff final y conservar el trabajo local no relacionado.

## 3. S3 — Validación agrupada y revisión

Depende del join secuencial S1/S2. No valida rendimiento de un modelo.

- [x] 3.1 Ejecutar pruebas pertinentes de assets, instructions, core y controles de seguridad/aprobación conservados tras la corrección a Codex.
- [x] 3.2 Ejecutar validación agrupada disponible con `scripts/validate.sh` cuando aplique, y build local de `tools/lufy-cli-go/bin/lufy-ai` al cambiar assets.
- [x] 3.3 Validar OpenSpec en modo estricto y revisar archivos viejos modificados/afectados y mirrors.
- [x] 3.4 Registrar evidencia de cada AC, limitaciones de conducta no evaluada y estado explícito; no declarar delivered/closed.

## 4. Gates posteriores, no autorizados por estos checkboxes

- [x] 4.1 Revisar implementación y evidencia antes de recomendar sync/delivery. Revisión manual y aceptación AC1–AC7 en `verification.md`.
- [x] 4.2 Sincronizar deltas únicamente cuando la implementación esté validada y el workflow lo autorice. Ambas specs principales y mirrors actualizados; strict y tests focalizados passed.
- [x] 4.3 Resolver autorización de delivery: el usuario respondió «Autorizo» a commit, push, creación de PR y merge a develop condicionado a checks requeridos exitosos. Ejecución delegada a delivery; la propuesta por sí sola no concedió permisos.
- [x] 4.4 Evaluar archive sólo después de satisfacer tareas y gates requeridos. PR 241 publicado, cinco checks SUCCESS para 59a1c60 y sync validado; archive autorizado para incorporarse al mismo PR. El merge y los checks del commit final permanecen como gates de delivery.

## Evidencia de preparación

- Inspección de specs `sdd-harness-routing`, `systemic-workflow` y `project-stack-config`.
- Configuración OpenSpec presente; configuración project-local y selección efectiva por tier ausentes.
- Grafo no disponible; memoria no inicializada. No se mutaron para preparar la propuesta.
- Instalación autorizada: `npm install --global @fission-ai/openspec@1.14.3`; `openspec --version` devolvió `1.14.3`. No se creó package.json, lockfile ni node_modules en el repo.
- Primera validación estricta: falló por nombres de escenarios existentes omitidos en MODIFIED. Se conservaron los nombres originales; la segunda ejecución de `openspec validate clarify-proportional-workflow-decisions --strict` reportó `is valid`.
- La comprobación estática anterior no detectó la omisión de nombres de escenarios: no debe tratarse como sustituto del CLI estricto.
- Bloqueo histórico de apply resuelto el 2026-10-21: se delegaron pruebas S1 a `test-writer` con proposal/design/deltas y AC1–AC7, sin autoridad de delivery.
- No se implementaron cambios en agentes/contratos/pruebas. Las pruebas Go de la sesión anterior son baseline histórico, no validación de esta propuesta ni de una implementación futura.

## Evidencia S1 — 2026-10-21

- Estado del bloque: `validated`; S2/S3 y gates posteriores siguen pendientes.
- `lufy-ai plan --target /Users/adrianrojas/Desktop/projects/lufy-ai --json`: exit 1 por `project.yaml` ausente. Superficies/límites configurados `not_available`; no se inicializó configuración, memoria ni grafo.
- RED del subagente: `go test ./internal/assets -run 'TestProportionalWorkflow(Contract|DecisionMatrix)S1' -count=1` desde `tools/lufy-cli-go`: exit 1 por reglas/matriz ausentes; compiló correctamente.
- GREEN tras actualizar contrato, docs y mirror: `go test ./internal/assets -count=1` desde `tools/lufy-cli-go`: passed. Incluye regresiones S1, paridad del catálogo y controles existentes de aprobación/seguridad.
- `go build -o bin/lufy-ai ./cmd/lufy-ai` desde `tools/lufy-cli-go`: passed; binario local reconstruido con el contrato distribuido.
- `openspec validate clarify-proportional-workflow-decisions --strict`: passed tras actualizar el estado del bloque.
- Handoff `test-writer`: el primer envelope omitía campos requeridos; fue corregido y `lufy-ai result validate --file /private/tmp/lufy-s1-test-writer-result.yaml --role test-writer --json` reportó `valid` antes de aceptarlo.
- `git diff --check origin/develop`: passed para cambios tracked; el test nuevo se revisó aparte, sin errores de whitespace.
- AC1/AC2: ejes y seguridad; AC3/AC4: cobertura cuatro/decisión cinco; AC5/AC6: procedencia y gates; AC7: paridad de contrato y visibilidad de mirrors. La equivalencia entre agentes se completará en S2.
- Las pruebas verifican texto/paridad, no conducta LLM, latencia ni ahorro de tokens. Validación global `scripts/validate.sh` reservada para S3 tras S2, según diseño.

## Evidencia histórica S2/S3 anterior a la corrección de alcance

- La propagación inicial incluyó OpenCode y Codex. El usuario corrigió después el alcance a 100% Codex, descartando mejoras OpenCode y la evaluación Pi. La evidencia de esta sección no valida por sí sola el resultado corregido.
- RED delegado: `go test ./internal/assets -run 'TestProportionalWorkflow.*S2' -count=1` falló por reglas ausentes y dos contradicciones de recuento. Tras el join, assets pasó en la suite global.
- Se propagaron AC1–AC7 a nueve superficies, ocho mirrors y tres roles neutrales. Se eliminaron las dos instrucciones ambiguas de escalado por recuento. Permisos, fallback, mapping y gates previos se preservan.
- Primer `scripts/validate.sh`: falló por golden obsoleto `TestRenderRouterT3NoneMatchesGolden`; se añadieron sólo las dos nuevas responsabilidades al fixture. La prueba focalizada pasó.
- Segundo `scripts/validate.sh`: exit 0; whitespace contra `origin/develop`, PR guard, action pinning, versiones, YAML, coupling, smoke, suite Go, coverage, vet y build correctos. Cobertura total 80.4% >= 80.0%. `shellcheck` ausente: shell lint omitido explícitamente por el script.
- `go build -o bin/lufy-ai ./cmd/lufy-ai` desde `tools/lufy-cli-go`: passed; binario local actualizado.
- `git diff --check --no-index /dev/null tools/lufy-cli-go/internal/assets/proportional_workflow_test.go`: sin diagnósticos de whitespace (exit 1 indica archivo añadido).
- `lufy-ai result validate --file /private/tmp/lufy-s2-test-writer-result.yaml --role test-writer --json`: valid tras actualizar evidencia del fixture.
- Auditoría estructural: archivos en roles/adapters/fixtures y mirrors existentes; sin nuevas capas ni movimientos pendientes. Revisión final dirigida de los diffs afectados y búsqueda de las dos reglas anteriores: ausentes.
- Limitación conservada: las pruebas son textuales/de paridad; no evalúan conducta real, latencia ni tokens. La migración solicitada no presupone mejoras demostradas de seguridad o rendimiento.

## Corrección de alcance — 2026-10-21

- Decisión explícita del usuario: mejoras de Lufy en este repositorio, basadas en el análisis/plan comparativo Gentleman AI, exclusivamente para Codex. Se descarta la exploración Pi y el trabajo específico OpenCode; no se creó ningún change de migración ni se instaló tooling.
- Se retiraron sólo nuestras modificaciones en seis archivos OpenCode (tres fuentes y mirrors), tres YAML neutrales y un golden OpenCode. Al inicio de esta sesión esos archivos no tenían cambios locales. No se revirtieron S1, Codex ni trabajo ajeno.
- Regresiones ajustadas por `test-writer`: tres agentes Codex y tres superficies compartidas, cinco mirrors; S1 conservado. El reporte `/private/tmp/lufy-codex-test-writer-result.yaml` pasó el validador de Result Contract; RED previo identificado como histórico.
- La regla de foco queda también en `AGENTS.md`; Obsidian no está inicializado y no se inicializa como efecto lateral.
- S3 final: `scripts/validate.sh` exit 0 para este alcance; suite Go, coverage 80.4% >= 80.0%, vet, build, paridad y controles existentes correctos. Shell lint omitido por `shellcheck` ausente, sin cambios shell en este bloque.
- `git diff --exit-code -- .opencode tools/lufy-cli-go/internal/assets/embedded/.opencode tools/lufy-cli-go/internal/instructions/roles tools/lufy-cli-go/internal/adapters/tool/opencode`: exit 0. Ninguna modificación residual específica de OpenCode ni de los roles YAML neutrales.
- Auditoría final de alcance: tres agentes Codex, contrato neutral, AGENTS/template/harness, docs y mirrors exactos; test nuevo en assets. Sin carpetas/capas requeridas pendientes ni cambios de permisos. `tools/lufy-cli-go/dist/` preservado.
- Estado final del bloque 1: `validated`, no delivered/closed. Siguientes bloques del plan original: piloto conductual de Codex, observabilidad y optimización sólo después de un baseline.
