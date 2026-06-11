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
- Analizar arquitectura, contratos, seguridad, pruebas, observabilidad, complejidad, migraciones/configuración, resiliencia, idempotencia y compatibilidad.
- Generar un HTML autocontenido solo en español dentro de `pr_review/`.
- Mantener la estética canónica del skill `pr.reviewer` alineada al overview OpenSpec `notion-dark`; no usar plantillas ad hoc.
- Profundizar el análisis con desk check por escenarios, before/after, revisión de comentarios previos, scoring explicado y resumen final/recomendación.
- Crear `pr_review/` si no existe.
- No sobrescribir reportes previos; usar nombre con PR y timestamp.
- Reportar ruta, comando `open ...` y resumen ejecutivo breve.

## Ejecución recomendada

1. Usar skill `pr.reviewer`.
2. Si falta `gh`, auth o contexto de PR, reportar `blocked` con recuperación exacta.
3. Si el review queda incompleto por falta de permisos/evidencia, generar el HTML igualmente marcando limitaciones explícitas.
