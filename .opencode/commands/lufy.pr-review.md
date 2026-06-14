---
description: Genera un review HTML en español para un Pull Request existente, agnóstico de lenguaje y framework.
agent: orchestrator
---

Genera un review de PR en español, read-only y agnóstico de lenguaje/framework.

## Comportamiento del comando

- Resolver el PR desde argumento (`<número>` o URL) o pedirlo si falta.
- Usar el skill concreto `pr.reviewer`.
- No comentar, aprobar, rechazar, mergear ni modificar el PR remoto.
- Recolectar evidencia con `gh`, Git y archivos locales cuando estén disponibles.
- Ejecutar consultas locales y remotas read-only sin pedir permisos redundantes: lectura de archivos, glob/grep/list, `pwd`, `ls`, `date`, `git status/diff/log/show/branch`, `gh auth status`, `gh pr view/diff/checks` y `gh api` para inspección del PR.
- Analizar arquitectura, contratos, seguridad, pruebas, observabilidad, complejidad, migraciones/configuración, resiliencia, idempotencia y compatibilidad.
- Incluir score por dimensiones, confianza del review, riesgo de merge, test gap map y resumen orientado a autor/reviewer/tech lead/QA cuando aplique.
- Generar un HTML autocontenido solo en español dentro de `pr_review/`.
- Mantener la estética canónica del skill `pr.reviewer` alineada al overview OpenSpec `notion-dark`; no usar plantillas ad hoc.
- Profundizar el análisis con desk check por escenarios, before/after, revisión de comentarios previos, scoring explicado y resumen final/recomendación.
- Crear `pr_review/` si no existe.
- No sobrescribir reportes previos; usar nombre con PR y timestamp.
- Mantener como único efecto de escritura permitido `mkdir -p pr_review` y la creación/escritura de `pr_review/pr-review-*.html`.
- Reportar ruta, comando `open ...` y resumen ejecutivo breve.

## Ejecución recomendada

1. Usar skill `pr.reviewer`.
2. Si falta `gh`, auth o contexto de PR, reportar `blocked` con recuperación exacta.
3. Si el review queda incompleto por falta de permisos/evidencia, generar el HTML igualmente marcando limitaciones explícitas.
4. No solicitar permisos para consultas cubiertas por el allowlist del agente `reviewer`; solo solicitar autorización si aparece una acción mutante, acceso externo no relacionado o una herramienta fuera del alcance de consulta.
