---
name: pr.reviewer
description: Revisa Pull Requests existentes y genera un reporte HTML en español, agnóstico de lenguaje/framework. Usar cuando el usuario pide revisar o auditar un PR, generar un reporte de review, o evaluar arquitectura, pruebas, seguridad, observabilidad y riesgos de un PR.
license: MIT
compatibility: OpenCode skill autocontenido; requiere `gh` para PRs remotos y puede degradar con evidencia local.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: pr.reviewer

Revisa un Pull Request existente y genera un reporte HTML autocontenido en español. El review es agnóstico de lenguaje y framework: aplica principios universales de ingeniería, arquitectura, seguridad, pruebas, observabilidad, resiliencia y mantenibilidad; usa señales stack-aware solo cuando el repositorio las haga evidentes.

## Límites

- Modo read-only: no edites archivos de código, no comentes en GitHub, no apruebes, no rechaces, no mergees y no ejecutes delivery.
- Puedes crear `pr_review/` y escribir el HTML del reporte dentro de esa carpeta.
- El consentimiento de revisión cubre consultas locales y remotas read-only del repositorio y del PR: lectura de archivos, glob/grep/list, `pwd`, `ls`, `date`, `git status`, `git diff`, `git log`, `git show`, `git branch`, `gh auth status`, `gh pr view`, `gh pr diff`, `gh pr checks` y `gh api` para comentarios/threads cuando `gh pr view` no alcance.
- No pidas permiso por cada consulta read-only normal del review; pide permiso solo si necesitas una acción fuera del allowlist, acceso externo no relacionado con el PR/repositorio o una operación con efectos secundarios.
- No inventes evidencia, checks, comentarios previos, cobertura, benchmarks, monitores ni riesgos.
- El contenido humano del reporte debe estar solo en español. Preserva identificadores técnicos, rutas, nombres de comandos, IDs, URLs, snippets y nombres de tecnologías.
- Si falta evidencia, usa `No disponible`, `No aplica` o `Pendiente de confirmar`.
- No uses reglas específicas de un dominio/proyecto salvo que estén documentadas en el repositorio o en el PR.

## Permisos esperados

- El agente `reviewer` debe tener permisos suficientes para ejecutar herramientas de consulta sin prompts repetidos: `read`, `glob`, `grep`, `list`, `webfetch`, utilidades locales inocuas, comandos `git` read-only y comandos `gh pr`/`gh api` de inspección.
- El único write permitido por este skill es crear `pr_review/` y escribir `pr_review/pr-review-*.html` con el reporte autocontenido.
- Están fuera de alcance: checkout, reset, merge, rebase, commit, push, publicar comentarios, aprobar/rechazar PRs, mergear PRs, ejecutar scripts, package managers, builds, tests o descargas externas.
- Si OpenCode solicita permiso para una consulta cubierta por este allowlist, considera que la configuración del agente está incompleta y reporta la limitación en el HTML.

## Inputs esperados

- PR como número, URL o referencia compatible con `gh`.
- Opcional: base branch, repo, foco del review, archivos críticos, criterios de negocio o riesgos conocidos.

## Recolección de evidencia

1. Resolver PR y metadata:

   ```bash
   gh pr view <PR> --json number,title,url,state,author,baseRefName,headRefName,mergeStateStatus,changedFiles,additions,deletions,commits,labels,reviews,reviewDecision,statusCheckRollup,body
   gh pr diff <PR>
   gh pr view <PR> --comments --json comments,reviews,reviewThreads
   gh pr checks <PR>
   # Fallback read-only si la version local de gh no expone un campo requerido:
   gh api <endpoint-del-PR-o-review-thread>
   ```

   Si alguna forma JSON no está soportada por la versión local de `gh`, usa el comando equivalente disponible y registra la limitación.

2. Contexto local mínimo:

   ```bash
    git status --short
    git diff --name-only <base>...<head>
    git diff --stat <base>...<head>
    lufy-ai pr guard --base <base>
    ```

    Usa comandos de Git solo para inspección. No hagas checkout, reset, merge, commit ni push. Si `lufy-ai pr guard` no está disponible, usa el fallback read-only `git diff --name-only <base>...<head> -- | git check-ignore -v --no-index --stdin` y revisa manualmente los prefijos internos `openspec/`, `.lufy/`, `.lufy-ai/`, `pr_review/`.

3. Leer contexto cuando exista:
   - `AGENTS.md`
   - `.lufy/config/project.yaml`
   - README o docs relevantes cercanos a los archivos modificados
   - `.github/PULL_REQUEST_TEMPLATE*`
   - políticas locales de delivery/review
   - specs OpenSpec/LUFY relacionadas si el PR las referencia

## Detección de stack

Detecta tecnologías por archivos y manifests, sin acoplar el skill a un lenguaje:

- Go: `go.mod`, `*.go`
- TypeScript/JavaScript: `package.json`, `tsconfig*.json`, `*.ts`, `*.tsx`, `*.js`, `*.jsx`
- Python: `pyproject.toml`, `requirements*.txt`, `*.py`
- Java/Kotlin: `pom.xml`, `build.gradle*`, `*.java`, `*.kt`
- Rust: `Cargo.toml`, `*.rs`
- Infra: `Dockerfile`, `docker-compose*.yml`, Terraform, Helm, Kubernetes, CI YAML

Usa esa detección para elegir ejemplos y checks, pero nunca bloquees por convenciones que el repo no declare.

## Matriz stack-aware y audiencia

Adapta el review al público principal del PR. Si el PR cruza varias superficies, marca `fullstack` o `multi-surface` y cubre contratos entre capas.

| Superficie | Señales | Foco de review | Audiencia principal |
|------------|---------|----------------|---------------------|
| Frontend | rutas UI, componentes, hooks, CSS, assets, tests browser | estados loading/empty/error, accesibilidad, responsive, contratos API, boundaries por feature, performance percibida | autor frontend, reviewer UI, QA |
| Backend | handlers/controllers, servicios, repositorios, jobs, APIs, DB | contratos, validación, dominio, auth/authz, transacciones, idempotencia, persistencia, observabilidad | backend reviewer, tech lead, SRE |
| Fullstack | cambios coordinados UI + API + datos | compatibilidad frontend/backend, serialización, estados de error, rollout, flags, versionado, contract tests | tech lead, QA, release owner |
| Infra/CI | Docker, Terraform, Helm, workflows, env vars, secrets | seguridad, rollback, ambientes, permisos, reproducibilidad, costo operativo, impacto en pipeline | infra/SRE, release owner |
| Mobile | app nativa/híbrida, permisos, stores, offline/cache | lifecycle, permisos, offline, performance, crash reporting, compatibilidad de versiones, release channels | mobile reviewer, QA mobile |
| CLI | comandos, flags, instaladores, scripts de usuario | UX de comandos, compatibilidad de flags, errores accionables, idempotencia, filesystem, cross-platform | CLI maintainer, soporte |
| Library/SDK | API pública, paquetes, exports, ejemplos | semver, backwards compatibility, typings/docs, deprecations, ejemplos, consumer ergonomics | maintainers, consumers |

Incluye un resumen por audiencia:

- Autor del PR: correcciones concretas o follow-ups aceptables.
- Reviewer humano: focos que debe mirar primero.
- Tech lead: riesgo de merge/release y tradeoffs.
- QA/release: escenarios manuales o automáticos a validar.

## Profundidad por tamaño y riesgo

- PR pequeño o mecánico: review completo, desk check reducido si no hay comportamiento observable.
- PR mediano: review completo por áreas modificadas y test gap map.
- PR grande o multi-objetivo: divide en slices por superficie/riesgo; si no puedes cubrir todo, declara `Cobertura parcial` y prioriza archivos críticos.
- PR crítico: eleva profundidad cuando toca auth, permisos, datos personales, dinero, migraciones, contratos públicos, infra de deploy, concurrencia o procesamiento masivo.

## Test gap map

Para cada cambio funcional relevante, registra:

| Comportamiento cambiado | Evidencia existente | Evidencia faltante | Riesgo cubierto |
|-------------------------|--------------------|--------------------|-----------------|
| ... | tests/checks/manual/No disponible | test o validación sugerida | contrato, edge, seguridad, rollback, etc. |

## Framework de revisión

Aplica `references/review-framework.md` como checklist base. Prioriza hallazgos con evidencia concreta de diff, código, PR, checks o comentarios previos.

Severidades unificadas:

- `CRÍTICO` (`L1`): bug funcional, riesgo de seguridad, pérdida/corrupción de datos, ruptura de contrato público, migración peligrosa, regresión de producción, race/consistencia grave o arquitectura que bloquea mantenibilidad esencial.
- `ALTO` (`L2`): defecto probable, falta de evidencia esencial, deuda significativa o riesgo de release que debería corregirse antes de mergear.
- `MEDIO` (`L3`): riesgo real pero acotado, mejora de test/observabilidad/contrato o complejidad que puede aceptarse con seguimiento explícito.
- `BAJO` (`L4`): mejora menor, naming, claridad, documentación o simplificación local.
- `INFORMATIVO` (`L5`): contexto, limitación, buena práctica observada o follow-up opcional; no afecta por sí solo el veredicto.

## Desk check obligatorio

El reporte debe incluir simulación de datos/flujo. Si no hay suficiente contexto, marca el desk check como `INCOMPLETO` y explica qué falta.

Escenarios mínimos adaptables:

- Camino feliz principal.
- Entrada inválida o incompleta.
- Dependencia externa o persistencia fallando.
- Edge case relevante: null/empty/zero/boundary/concurrencia/tamaño grande.
- Retry/idempotencia cuando el cambio pueda reprocesarse.
- Migración/configuración cuando el PR cambie schema, flags, env vars o infraestructura.

Para cada escenario, traza capas genéricas:

| Capa | Operación | Entrada | Salida esperada | Estado |
|------|-----------|---------|-----------------|--------|
| Entrada/adaptador | Parseo/validación | ... | ... | OK/FAIL |
| Aplicación/caso de uso | Orquestación | ... | ... | OK/FAIL |
| Dominio/reglas | Regla aplicada | ... | ... | OK/FAIL |
| Infra/dependencia | Repo/cliente/evento | ... | ... | OK/FAIL |
| Salida | Respuesta/estado/evento/métrica | ... | ... | OK/FAIL |

## Scoring

Calcula score de 0 a 100 con dimensiones ponderadas:

| Dimensión | Peso |
|-----------|------|
| Arquitectura y diseño | 20% |
| Correctitud funcional y contratos | 20% |
| Pruebas y evidencia | 15% |
| Seguridad y privacidad | 15% |
| Observabilidad y operación | 10% |
| Mantenibilidad y complejidad | 10% |
| Desk check | 10% |

Además del score de calidad, calcula:

- `Confianza del review`: `Alta`, `Media` o `Baja`, según completitud de diff, acceso a comentarios/checks, contexto local, evidencia de pruebas y tamaño del PR.
- `Riesgo de merge`: `Bajo`, `Medio` o `Alto`, según severidades, checks, tamaño, áreas críticas, migraciones/configuración, contratos públicos y rollout.

Veredicto:

- `Aprobar`: score >= 80, sin hallazgos críticos ni altos bloqueantes.
- `Pedir cambios`: score >= 50 o existe al menos un hallazgo crítico/alto corregible.
- `Rechazar`: score < 50, riesgo sistémico, evidencia insuficiente para un cambio riesgoso o múltiples críticos.

## Profundidad mínima del análisis

El reporte no debe ser un resumen superficial del diff. Debe leer el PR como lo haría un reviewer humano senior y dejar evidencia accionable suficiente para decidir merge/no-merge.

- Analiza el cambio por capas: entrada/adaptador, aplicación/caso de uso, dominio/reglas, persistencia/dependencias y salida/contrato.
- Explica el flujo antes/después cuando el PR modifica comportamiento observable, contratos, permisos, persistencia, jobs, eventos o integraciones.
- Evalúa el template/body del PR cuando exista: WHY, alcance, issue/ticket, test plan, evidencias, migraciones/configuración y stacked PRs/follow-ups. Si el repo usa otro template, registra `No aplica` en vez de inventar incumplimientos.
- Revisa comentarios/reviews previos y clasifícalos como `resuelto`, `pendiente`, `no verificable` o `no aplica`, con una acción concreta.
- Para cada hallazgo medio/alto/crítico, incluye evidencia, impacto, escenario de reproducción o razonamiento de fallo, recomendación y criterio de aceptación.
- Si el PR incluye paths ignorados por `.gitignore` o metadata interna sin override explícito, repórtalo como hallazgo mínimo `MEDIO` (`L3`) con evidencia de `lufy-ai pr guard` o `git check-ignore -v --no-index --stdin`. Eleva a `ALTO` (`L2`) si expone secretos, contenido privado, ruido significativo de release o contradice una política de delivery del repo.
- Incluye al menos una sección de `Buenas prácticas observadas` cuando el PR tenga decisiones correctas; no todo el reporte debe ser punitivo.
- El desk check debe cubrir escenarios reales del dominio del PR. Usa 5 escenarios como mínimo cuando el cambio sea funcional; si el alcance es documental o mecánico, explica por qué aplica una simulación reducida.
- El score debe estar justificado por dimensión. No basta un número global.
- Incluye `Test gap map`, `Confianza del review`, `Riesgo de merge` y resumen por audiencia cuando el PR tenga cambios funcionales o multi-superficie.

## Reporte HTML

- Crear `pr_review/` si no existe.
- Escribir el reporte en `pr_review/pr-review-<number>-<yyyyMMdd-HHmm>.html`.
- Si el PR no tiene número, usar `pr_review/pr-review-<slug>-<yyyyMMdd-HHmm>.html`.
- No sobrescribir archivos existentes; si colisiona, agrega sufijo `-2`, `-3`, etc.
- Usar `templates/report.html` como estructura visual canónica y adaptar contenido real.
- El HTML debe ser autocontenido: CSS inline, sin dependencias externas, sin JS requerido.
- Incluir link al PR arriba cuando exista URL.
- Todas las secciones deben estar dentro de cards/containers para evitar overflow.

### Contrato visual obligatorio

El reporte debe mantener la estética unificada del overview OpenSpec `notion-dark`:

- Usar el hero navy/deep navy con título grande, contenedor central `1180px`, fondo `surface`, cards blancas con borde `hairline`, radio `12px`, sombras suaves y variables CSS compatibles con `templates/report.html`.
- No generar una plantilla ad hoc gris/azul ni cards con radio mayor a `12px`.
- No cambiar la escala visual principal salvo para responsive. En desktop, el título principal debe conservar la jerarquía de hero y el gauge debe aparecer dentro de una card destacada.
- Mantener badges, tablas, `details`, código y findings con estilos de la plantilla base.
- Si el reporte necesita secciones adicionales, agrégalas dentro de `.card`, `.issue`, `details` o contenedores equivalentes ya definidos por la plantilla.
- Antes de escribir el HTML, verificar mentalmente que aparecen estos marcadores visuales: `--navy`, `--navy-deep`, `--surface`, `.gauge`, `.scoregrid`, `.issue`, `.final-summary`.

Secciones obligatorias:

1. Resumen ejecutivo.
2. Metadata del PR.
3. Veredicto y score.
4. Confianza del review y riesgo de merge.
5. Hallazgos críticos y altos.
6. Hallazgos medios/bajos.
7. Buenas prácticas observadas.
8. Análisis arquitectónico.
9. Seguridad y privacidad.
10. Pruebas y evidencia.
11. Test gap map.
12. Observabilidad y operación.
13. Migraciones/configuración/contratos.
14. Desk check y simulación.
15. Comentarios previos no resueltos.
16. Resumen por audiencia.
17. Action items priorizados.
18. Limitaciones del review.
19. Resumen final y recomendación.

### Secciones recomendadas para PRs funcionales

Cuando haya cambios funcionales, contratos públicos, datos sensibles, seguridad, persistencia o integraciones, incluye además:

- Validación del template/body del PR.
- Puntos de revisiones anteriores.
- Before/After del comportamiento.
- Tabla de scoring por dimensión con peso, score y justificación.
- Test gap map por comportamiento cambiado.
- Riesgo de merge, confianza del review y resumen para autor/reviewer/tech lead/QA.
- Cierre ejecutivo final que diga explícitamente si conviene aprobar, pedir cambios o rechazar, y cuál es el próximo paso exacto.

### Control de calidad antes de entregar

Antes de responder al usuario, inspecciona el HTML generado y confirma:

- Contiene `Resumen ejecutivo`, `Desk check`, `Test gap map`, `Action items`, `Limitaciones` y `Resumen final`.
- Contiene la estética canónica (`--navy`, `--navy-deep`, `.gauge`, `.final-summary`).
- Cada hallazgo alto/crítico tiene evidencia y recomendación concreta.
- El cierre no contradice el veredicto ni el score.
- No hay URLs remotas de assets, CDNs, scripts externos ni contenido en idioma distinto del español, salvo identificadores técnicos o citas del PR.

## Respuesta final al usuario

Devuelve solo. Reporta la ruta generada como hipervínculo Markdown clicable y conserva `open <ruta>` como fallback:

```markdown
Reporte generado: [pr_review/pr-review-<...>.html](pr_review/pr-review-<...>.html)
Abrir: `open pr_review/pr-review-<...>.html`

Resumen ejecutivo:
- <máximo 5 bullets>
```

No pegues el HTML completo en la conversación.
