# Detección para `pr.creator`

Estas reglas son heurísticas documentales para organizar contenido del PR. No sustituyen revisión humana ni validación de delivery.

## OpenSpec

- Si se recibe `change-id`, buscar `openspec/changes/<change-id>/proposal.md`.
- Usar `## Why` para la sección `Why` del PR.
- Usar `## What Changes`, capacidades e impacto para `Resumen`.
- Si el archivo no existe o no es legible, usar contexto proporcionado y marcar supuestos como `Pendiente de confirmar`.

## Tracking

Usar solo referencias presentes en contexto, proposal, tareas, diff, notas o configuración explícita:

- Jira: URLs de Jira o claves tipo `[A-Z][A-Z0-9]+-[0-9]+`.
- GitHub: URLs de issues/projects, referencias `owner/repo#123` o `#123` cuando el repo sea inequívoco.
- Notion: URLs `notion.so` o IDs/páginas proporcionadas.
- Otros: links o IDs con sistema identificado por el usuario/agente.

Fallbacks:

- `No configurado`: no hay herramienta de tracking conocida.
- `No aplica`: el cambio no requiere tracking externo según contexto.
- `Pendiente de confirmar`: hay una señal incompleta o ambigua.

## Monitors

Capturar monitores solo cuando estén explícitamente configurados/proporcionados:

- Grafana: URLs o nombres de dashboards/panels.
- New Relic: links, alertas, entidades, dashboards o NRQL proporcionados.
- Datadog: monitor IDs, dashboards, SLOs o links proporcionados.
- Otros: sistema, nombre y link proporcionados.

Fallbacks:

- `No configurado`: no hay sistema de observabilidad conocido.
- `No aplica`: cambio documental/config local sin impacto operativo esperado.
- `Pendiente de confirmar`: se menciona observabilidad sin enlace o sistema claro.

## Migraciones y schema

Marca `Detectadas` si archivos modificados o diff contienen cualquiera de estos patrones:

- Rutas: `migrations/`, `migration/`, `db/migrate/`, `database/migrations/`, `schema.sql`, `structure.sql`, `prisma/schema.prisma`, `db/schema.rb`.
- Nombres: `*.migration.*`, `*migration*.sql`, `*migrations*.sql`, `*.ddl.sql`.
- Diff/contenido SQL: `CREATE TABLE`, `ALTER TABLE`, `DROP TABLE`, `CREATE INDEX`, `ALTER INDEX`, `DROP INDEX`, `CREATE SCHEMA`, `ALTER SCHEMA`.
- ORM/schema: cambios a `model` en Prisma, entidades persistentes o archivos de schema explícitos cuando el contexto lo confirme.

Marca `Pendiente de confirmar` si hay señales ambiguas de persistencia sin patrón de migración claro, por ejemplo:

- Cambios en rutas `repository`, `store`, `persistence`, `dao`, `entity`, `entities`, `models` sin migration visible.
- Nuevos campos persistentes o queries DDL/DML relevantes sin plan de migración.

Marca `No detectadas` si no aparece ningún patrón y aclara: `No se detectaron migraciones; revisión heurística sobre rutas/diff disponibles`.
