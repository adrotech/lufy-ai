## Why

El backlog LUFY-3 pide un reporte de tiempo y ROI que convierta la actividad local de OpenCode, Git y configuración del proyecto en un HTML autocontenido para revisión offline. Hoy no existe comando, skill ni contrato para generar esos KPIs sin exponer prompts, outputs o diffs sensibles.

## What Changes

- Agregar la capability `lufy-timereport` para generar reportes HTML offline/autocontenidos desde una invocación `/lufy.timereport` respaldada por el skill `.opencode/skills/lufy.timereport/SKILL.md`.
- Definir fuentes y degradación: SQLite local de OpenCode como fuente primaria, Git como fuente secundaria, `.opencode/project.yaml` opcional para stack, y JSONL/`session_diff` fuera del alcance inicial salvo futuro opt-in explícito.
- Definir métricas sanitizadas por defecto: wall-clock, AI working time, tiempo humano activo, LOC neto, commits, tool calls, top tools, subagents, skills, fases y stack detectado.
- Establecer privacidad por defecto: no incluir prompts, outputs completos, contenidos de herramientas ni diffs/session_diff en el HTML salvo una futura opción explícita no incluida en este cambio.
- Incluir heurísticas observables para tiempos y fases, más validación con fixtures/datos sanitizados y comandos read-only.

## Capabilities

### New Capabilities
- `lufy-timereport`: Generación de reportes HTML autocontenidos de tiempo/ROI a partir de actividad local de OpenCode y Git, con métricas estructurales, privacidad por defecto y degradación explícita.

### Modified Capabilities
- None.

## Impact

- Nuevos artefactos esperados: `.opencode/skills/lufy.timereport/SKILL.md` y `.opencode/commands/lufy.timereport.md`.
- Lecturas locales read-only esperadas durante ejecución futura: `~/.local/share/opencode/opencode.db`, metadata Git del repositorio y `.opencode/project.yaml` cuando exista.
- Posible dependencia o helper de lectura SQLite/HTML durante implementación; deberá permanecer local/offline y evitar red, prompts, outputs y diffs por defecto.
- No cambia puertos, auth, schema de base de datos, contratos HTTP/API ni flujo de delivery.
