## 1. Estructura del skill

- [x] 1.1 Crear `.opencode/skills/pr.creator/` con estructura estándar tipo Anthropic y `SKILL.md` como punto de entrada.
- [x] 1.2 Añadir recursos/templates autocontenidos dentro del skill para generar título sugerido y cuerpo Markdown de Pull Request.
- [x] 1.3 Documentar en `SKILL.md` inputs esperados, outputs, límites y prohibición de ejecutar `git`, `gh`, sync remoto o mutaciones de delivery.

## 2. Template y contenido del PR

- [x] 2.1 Implementar el template de PR en español con resumen, `Why`, tarea asociada, evidencia de pruebas, `Monitors`, `Migraciones`, riesgos/follow-ups y checklist/notas de validación.
- [x] 2.2 Implementar la preferencia por `openspec/changes/<change>/proposal.md` para derivar resumen funcional y `why` cuando exista un change-id o contexto OpenSpec.
- [x] 2.3 Definir fallback explícito para contexto incompleto usando `Pendiente de confirmar`, `No configurado` o `No aplica` sin inventar evidencia.

## 3. Modo manual

- [x] 3.1 Documentar el flujo manual para que una persona o agente invoque `pr.creator` con change-id, diff, evidencia, tickets, monitores o notas del cambio.
- [x] 3.2 Asegurar que el modo manual produzca solo contenido de PR —título sugerido y cuerpo Markdown— sin ejecutar delivery.

## 4. Integración con delivery

- [x] 4.1 Actualizar `.opencode/agents/delivery.md` o documentación/configuración equivalente para que `delivery` use `pr.creator` al crear PR cuando `.opencode/skills/pr.creator/` exista.
- [x] 4.2 Documentar la separación de responsabilidades: `pr.creator` estructura/genera contenido y `delivery` mantiene branch safety, validación final, commit, push, `gh pr create`, sync y reporte.
- [x] 4.3 Definir comportamiento de fallback para `delivery` si `pr.creator` no está disponible, reportando la limitación sin romper los gates de delivery.

## 5. Detección de trazabilidad, monitores y migraciones

- [x] 5.1 Documentar cómo el skill identifica tickets desde contexto/configuración disponible para Jira, GitHub Issues/Projects, Notion u otros sistemas.
- [x] 5.2 Documentar cómo el skill captura monitores o dashboards configurados/proporcionados para Grafana, New Relic, Datadog u otros sistemas.
- [x] 5.3 Implementar heurísticas documentadas para detectar migrations o cambios de tablas/schemas mediante rutas y patrones de diff como `migrations/`, `db/migrate/`, `schema.sql`, `prisma/schema.prisma`, `*.migration.*`, `CREATE TABLE` y `ALTER TABLE`.

## 6. Validación

- [x] 6.1 Verificar estáticamente que `SKILL.md` y templates cubren todos los requisitos de la spec y preservan contenido humano en español.
- [x] 6.2 Verificar que `delivery` referencia o invoca `pr.creator` para creación de PR sin transferirle permisos Git/GH.
- [x] 6.3 Ejecutar validación OpenSpec relevante para confirmar que el cambio sigue apply-ready y reportar comandos/resultados reales.
  - Nota de validación: se re-ejecutaron `openspec instructions apply --change "add-pr-creator-skill" --json` y `openspec status --change "add-pr-creator-skill"` para confirmar el estado apply-ready/all done tras los ajustes de revisión.
