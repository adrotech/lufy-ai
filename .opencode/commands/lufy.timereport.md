---
description: Genera un Developer Impact Report local LUFY en HTML autocontenido.
agent: orchestrator
---

Genera `/lufy.timereport` delegando la ejecución al skill local `lufy.timereport`.

El resultado es un reporte estilo Notion para entender cómo la IA ayudó en el trabajo diario: qué pidió el usuario, qué hizo la IA, por qué lo hizo, tiempos, aprendizajes, pivots y evidencia técnica.

## Entradas

- `--output <ruta>` opcional: ruta explícita del HTML a generar. Si se omite, usar `/tmp/lufy-timereport-<timestamp>.html`.
- `--target-dir <ruta>` opcional: repositorio a analizar. Si se omite, usar el directorio actual de OpenCode.
- `--from <fecha>` / `--to <fecha>` opcionales: rango temporal en formato aceptado por Git/Python. Si se omiten, el skill usa la actividad disponible de OpenCode y Git.
- `--db <ruta>` opcional para validación local: SQLite de OpenCode alternativo. Por defecto usar `~/.local/share/opencode/opencode.db`.
- `--scope task|repo` opcional: `task` es el default y reporta la tarea original solicitada al usuario, incluyendo subagentes. `repo` conserva el reporte global del repositorio.
- `--session-id <id>` opcional: ancla explícita de sesión OpenCode. Si apunta a un subagente, el reporte sube a la sesión raíz y toma todo su árbol.
- `--tier T1|T2|T3` opcional: tier original de la tarea cuando se conoce.
- `--change <id>` opcional: spec/change OpenSpec o LUFY SDD asociado, cuando existe.

## Comportamiento

1. Carga el skill `lufy.timereport`.
2. Ejecuta el generador local/offline documentado por el skill.
3. Por defecto limita métricas a la tarea raíz actual y sus subagentes. Solo usa todo el repo con `--scope repo`.
4. Incluye un paso a paso sanitizado con qué tipo de trabajo ocurrió, por qué, duración total, tiempo IA y tiempo humano por tramo.
5. Renderiza el HTML con estética tipo Notion: propiedades de tarea, callouts, tablas tipo database y diseño claro/offline.
6. Reporta la ruta final del HTML o una degradación accionable si faltan fuentes.

## Privacidad por defecto

- No incluir prompts, respuestas completas, argumentos completos de tools, outputs de tools, contenidos de archivos, diffs ni `session_diff`.
- No leer JSONL ni `session_diff` salvo que un cambio futuro agregue opt-in explícito.
- Todas las fuentes son locales y read-only: OpenCode SQLite, Git y `.lufy/project.yaml`.

## Ejemplos

```bash
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py --output /tmp/lufy.html --target-dir "$PWD"
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py --tier T2 --change install-managed-assets-with-hash-idempotency
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py --scope repo
```
