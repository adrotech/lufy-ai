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
- No inventes evidencia, checks, comentarios previos, cobertura, benchmarks, monitores ni riesgos.
- El contenido humano del reporte debe estar solo en español. Preserva identificadores técnicos, rutas, nombres de comandos, IDs, URLs, snippets y nombres de tecnologías.
- Si falta evidencia, usa `No disponible`, `No aplica` o `Pendiente de confirmar`.
- No uses reglas específicas de un dominio/proyecto salvo que estén documentadas en el repositorio o en el PR.

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
   ```

   Si alguna forma JSON no está soportada por la versión local de `gh`, usa el comando equivalente disponible y registra la limitación.

2. Contexto local mínimo:

   ```bash
   git status --short
   git diff --name-only <base>...<head>
   git diff --stat <base>...<head>
   ```

   Usa comandos de Git solo para inspección. No hagas checkout, reset, merge, commit ni push.

3. Leer contexto cuando exista:
   - `AGENTS.md`
   - `.lufy/project.yaml`
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

## Framework de revisión

Aplica `references/review-framework.md` como checklist base. Prioriza hallazgos con evidencia concreta de diff, código, PR, checks o comentarios previos.

Severidades:

- `CRÍTICO`: bug funcional, riesgo de seguridad, pérdida/corrupción de datos, ruptura de contrato público, migración peligrosa, regresión de producción, race/consistencia grave o arquitectura que bloquea mantenibilidad esencial.
- `ALTO`: defecto probable o deuda significativa que debería corregirse antes de mergear.
- `MEDIO`: riesgo real pero acotado, mejora de test/observabilidad/contrato o complejidad que puede aceptarse con seguimiento.
- `BAJO`: mejora menor, naming, claridad o documentación.
- `INFORMATIVO`: contexto, template, limitación o buena práctica observada; no afecta veredicto.

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

Veredicto:

- `Aprobar`: score >= 80, sin hallazgos críticos ni altos bloqueantes.
- `Pedir cambios`: score >= 50 o existe al menos un hallazgo crítico/alto corregible.
- `Rechazar`: score < 50, riesgo sistémico, evidencia insuficiente para un cambio riesgoso o múltiples críticos.

## Reporte HTML

- Crear `pr_review/` si no existe.
- Escribir el reporte en `pr_review/pr-review-<number>-<yyyyMMdd-HHmm>.html`.
- Si el PR no tiene número, usar `pr_review/pr-review-<slug>-<yyyyMMdd-HHmm>.html`.
- No sobrescribir archivos existentes; si colisiona, agrega sufijo `-2`, `-3`, etc.
- Usar `templates/report.html` como estructura visual y adaptar contenido real.
- El HTML debe ser autocontenido: CSS inline, sin dependencias externas, sin JS requerido.
- Incluir link al PR arriba cuando exista URL.
- Todas las secciones deben estar dentro de cards/containers para evitar overflow.

Secciones obligatorias:

1. Resumen ejecutivo.
2. Metadata del PR.
3. Veredicto y score.
4. Hallazgos críticos y altos.
5. Hallazgos medios/bajos.
6. Buenas prácticas observadas.
7. Análisis arquitectónico.
8. Seguridad y privacidad.
9. Pruebas y evidencia.
10. Observabilidad y operación.
11. Migraciones/configuración/contratos.
12. Desk check y simulación.
13. Comentarios previos no resueltos.
14. Action items priorizados.
15. Limitaciones del review.

## Respuesta final al usuario

Devuelve solo:

```markdown
Reporte generado: `pr_review/pr-review-<...>.html`
Abrir: `open pr_review/pr-review-<...>.html`

Resumen ejecutivo:
- <máximo 5 bullets>
```

No pegues el HTML completo en la conversación.
